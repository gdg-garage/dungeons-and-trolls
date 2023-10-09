package main

import (
	"context"
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"math"
	"net"
	"net/http"
	"sort"
	"strings"
)

const apiKeyFieldName = "X-API-key"

type MapCellExt struct {
	mapObjects  *api.MapObjects
	distance    int
	lineOfSight bool
}

type PlainPos struct {
	PositionX int32
	PositionY int32
}

func PlainPosFromApiPos(position *api.Position) PlainPos {
	return PlainPos{
		PositionX: position.PositionX,
		PositionY: position.PositionY,
	}
}

func calculateDistanceAndLineOfSight(currentMap *api.Level, currentPosition *api.Position) map[PlainPos]MapCellExt {
	// distance to obstacles used for line of sight
	distanceToFirstObstacle := make(map[float32]float32)
	currentPlainPos := PlainPosFromApiPos(currentPosition)

	// map for resulting map positions with distance and line of sight
	resultMap := make(map[PlainPos]MapCellExt)
	// fill map with map objects
	for _, objects := range currentMap.Objects {
		resultMap[PlainPosFromApiPos(objects.Position)] = MapCellExt{
			mapObjects:  objects,
			distance:    -1,
			lineOfSight: false,
		}
	}

	//log.Info().Msgf("Original map -> (player: A, no data / free: ' ', wall: w, spawn: *, stairs: s, unknown: ?)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if currentPosition.PositionX == x && currentPosition.PositionY == y {
	//			row += "A"
	//		} else if !found {
	//			row += " "
	//		} else if cell.mapObjects.IsSpawn != nil && *cell.mapObjects.IsSpawn {
	//			row += "*"
	//		} else if cell.mapObjects.IsStairs {
	//			row += "s"
	//		} else if cell.mapObjects.IsFree {
	//			row += " "
	//		} else if cell.mapObjects.IsWall {
	//			row += "w"
	//		} else {
	//			row += "?"
	//		}
	//	}
	//	log.Info().Msgf("Map row: %s (y = %d)", row, y)
	//}

	// standard BFS stuff
	visited := make(map[PlainPos]bool)
	var queue []PlainPos

	// start from player
	// add current node to queue and add its distance to final map
	queue = append(queue, currentPlainPos)
	cell, found := resultMap[currentPlainPos]
	mapObjects := &api.MapObjects{}
	if found {
		mapObjects = cell.mapObjects
	}
	resultMap[currentPlainPos] = MapCellExt{
		mapObjects:  mapObjects,
		distance:    0,
		lineOfSight: true,
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		nodeVisited, found := visited[node]
		if !found || !nodeVisited {
			visited[node] = true

			// Enqueue all unvisited neighbors
			for _, neighbor := range getNeighbors(node) {
				// neighbors can be out of the map,
				cell, found := resultMap[neighbor]
				// must be in bounds
				// must not be visited
				// must be free
				if isInBounds(currentMap, &neighbor) && !visited[neighbor] && (!found || cell.mapObjects.IsFree) {
					mapObjects := &api.MapObjects{
						IsFree: true,
					}
					if found {
						mapObjects = cell.mapObjects
					}
					distance := resultMap[node].distance + 1
					lineOfSight := getLoS(currentMap, resultMap, distanceToFirstObstacle, currentPosition, neighbor)
					resultMap[neighbor] = MapCellExt{
						mapObjects:  mapObjects,
						distance:    distance,
						lineOfSight: lineOfSight,
					}
					queue = append(queue, neighbor)
				}
			}
		}
	}

	//log.Info().Msgf("Map with distances -> (player: A, no data: !, not reachable: ~, distance < 10: 0-9, distance >= 10: +)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if makePosition(x, y) == currentPosition {
	//			row += "A"
	//		} else if !found {
	//			row += "!"
	//		} else if cell.distance < 10 {
	//			row += fmt.Sprintf("%d", cell.distance)
	//		} else if cell.distance == math.MaxInt32 {
	//			row += "~"
	//		} else {
	//			row += "+"
	//		}
	//	}
	//	//b.Logger.Debugf("Map row: %s (y = %d)", row, y)
	//}

	//log.Info().Msgf("Map with line of sight -> (player: A, no data: !, line of sight: ' ', wall: w, no line of sight: ~)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if currentPosition.PositionX == x && currentPosition.PositionY == y {
	//			row += "A"
	//		} else if !found {
	//			row += "!"
	//		} else if cell.lineOfSight {
	//			row += " "
	//		} else if cell.mapObjects.IsWall {
	//			row += "w"
	//		} else {
	//			row += "~"
	//		}
	//	}
	//	log.Info().Msgf("Map row: %s (y = %d)", row, y)
	//}

	return resultMap
}

func makePosition(x int32, y int32) PlainPos {
	return PlainPos{
		PositionX: x,
		PositionY: y,
	}
}

func getNeighbors(pos PlainPos) []PlainPos {
	return []PlainPos{
		makePosition(pos.PositionX-1, pos.PositionY),
		makePosition(pos.PositionX+1, pos.PositionY),
		makePosition(pos.PositionX, pos.PositionY-1),
		makePosition(pos.PositionX, pos.PositionY+1),
	}
}

func isInBounds(currentMap *api.Level, pos *PlainPos) bool {
	return pos.PositionX >= 0 && pos.PositionX < currentMap.Width && pos.PositionY >= 0 && pos.PositionY < currentMap.Height
}

func getLoS(currentLevel *api.Level, resultMap map[PlainPos]MapCellExt, distanceToFirstObstacle map[float32]float32, pos1 *api.Position, pos2 PlainPos) bool {
	// get the center of the cell
	x1 := float32(pos1.PositionX) + 0.5
	y1 := float32(pos1.PositionY) + 0.5
	x2 := float32(pos2.PositionX) + 0.5
	y2 := float32(pos2.PositionY) + 0.5

	// slope := float32(y2-y1) / float32(x2-x1)
	distance := math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
	// angle in radians
	slope := float32(math.Atan2(float64(y2-y1), float64(x2-x1)))
	// angleDegrees := angleRadians * 180 / math.Pi

	// TODO: somehow round the value to prevent cache misses
	losDist, found := distanceToFirstObstacle[slope]
	if found {
		//	log.Info().Msgf("LoS: found in cache",
		//		"playerPosition", pos1,
		//		"position", pos2,
		//		"slope", slope,
		//		"distance", distance,
		//		"lineOfSightDistance", losDist,
		//		"lineOfSight", distance < float64(losDist),
		//	)
		//	return distance < float64(losDist)
	}
	losDist = rayTrace(currentLevel, resultMap, slope, x1, y1, x2, y2)
	distanceToFirstObstacle[slope] = losDist
	//log.Info().Msgf("LoS: calculated",
	//	"playerPosition", pos1,
	//	"position", pos2,
	//	"slope", slope,
	//	"distance", distance,
	//	"lineOfSightDistance", losDist,
	//	"lineOfSight", distance < float64(losDist),
	//)
	return distance < float64(losDist)
}

func rayTrace(currentLevel *api.Level, resultMap map[PlainPos]MapCellExt, slope float32, x1 float32, y1 float32, x2 float32, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1

	// Calculate absolute values of dx and dy
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	// Determine the sign of movement along x and y
	sx := float32(1)
	sy := float32(1)
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}

	// Initialize error variables
	e := dx - dy
	x := x1
	y := y1

	for {
		// TODO: any mapping needed here?
		pos := getPositionsForFloatCoords(x, y)
		// Check the current cell for obstacles or objects
		cell, found := resultMap[pos]

		// obstacle hit if end of map OR not free
		if !isInBounds(currentLevel, &pos) || (found && !cell.mapObjects.IsFree) {
			dist := math.Sqrt(float64((x-x1)*(x-x1) + (y-y1)*(y-y1)))
			return float32(dist)
		}

		// Calculate the next step
		e2 := 2 * e
		if e2 > -dy {
			e -= dy
			x += sx
		}
		if e2 < dx {
			e += dx
			y += sy
		}
	}
}

func getPositionsForFloatCoords(x float32, y float32) PlainPos {
	// I was worried about what position to return if the float values are exactly on the border between two positions.
	// But it looks like this works fine.
	// NOTE: This might be something to adjust if we see weird line of sight.
	//			 E.g. if we see different LoS on right and left side of player or obstacle.
	return PlainPos{
		PositionX: int32(x),
		PositionY: int32(y),
	}
}

func getToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("cannot read request metadata (api key is missing)")
	}
	tokens := md.Get(apiKeyFieldName)
	if len(tokens) != 1 {
		return "", fmt.Errorf("incorrect number of auth tokens: %d", len(tokens))
	}
	return tokens[0], nil
}

type server struct {
	api.UnsafeDungeonsAndTrollsServer
	G *dungeonsandtrolls.Game
}

func filterGameState(game *dungeonsandtrolls.Game, g *api.GameState, level *int32, position *api.Position) {
	// filter monsters for non-monster players
	var keptLevels []*api.Level
	for _, l := range g.Map.Levels {
		if level != nil && l.Level != *level {
			continue
		}
		for _, o := range l.Objects {
			for _, m := range o.Monsters {
				dungeonsandtrolls.HideNonPublicMonsterFields(game, m)
			}
			for _, p := range o.Players {
				for _, e := range p.Effects {
					gameobject.FilterEffect(e)
				}
			}
			for _, e := range o.Effects {
				gameobject.FilterEffect(e)
			}
		}

		if position != nil {
			distInfo := calculateDistanceAndLineOfSight(l, position)
			for p, i := range distInfo {
				l.PlayerMap = append(l.PlayerMap, &api.PlayerSpecificMap{
					Position: &api.Position{
						PositionX: p.PositionX,
						PositionY: p.PositionY,
					},
					LineOfSight: i.lineOfSight,
					Distance:    int32(i.distance),
				})
			}
		}

		keptLevels = append(keptLevels, l)
	}
	g.Map.Levels = keptLevels
	if level != nil && *level != 0 {
		// Show shop only on 0th floor
		g.ShopItems = []*api.Item{}
	} else {
		hideUnidentifiedItems(game, g)
	}
}

func filterMonsterGameState(game *dungeonsandtrolls.Game, g *api.GameState) {
	g.ShopItems = []*api.Item{}
}

func hideUnidentifiedItems(game *dungeonsandtrolls.Game, g *api.GameState) {
	for _, i := range g.ShopItems {
		if i.Unidentified == nil {
			continue
		}
		if !*i.Unidentified {
			continue
		}
		gameobject.HideUnidentifiedFields(i)
	}
}

func isBlocking(blocking *bool) bool {
	if blocking == nil {
		return true
	}
	return *blocking
}

func (s *server) gameState(ctx context.Context, params *api.GameStateParams, level *int32) (*api.GameState, error) {
	token, err := getToken(ctx)

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	s.G.GameLock.RUnlock()

	// GameContext is special and is not blocking by default also it is blocking before the actual work
	if params.Blocking != nil && *params.Blocking {
		s.G.WaitForNextTick(tick)
	}

	s.G.GameLock.RLock()
	g, ok := proto.Clone(&s.G.Game).(*api.GameState)
	g.MaxLevel = s.G.MaxLevelReached
	if !ok {
		return nil, fmt.Errorf("cloning GameState failed")
	}
	s.G.GameLock.RUnlock()

	if params.Items != nil && !*params.Items {
		g.ShopItems = []*api.Item{}
	}

	if params.FogOfWar != nil && *params.FogOfWar {
		for _, l := range g.Map.Levels {
			lc, err := s.G.GetCachedLevel(l.Level)
			if err != nil {
				return g, fmt.Errorf("level cache retrieval failed for level %d: %s", l.Level, err.Error())
			}
			for x, vy := range lc.Fow {
				for y, fow := range vy {
					l.FogOfWar = append(l.FogOfWar, &api.FogOfWarMap{
						Position: &api.Position{
							PositionX: x,
							PositionY: y,
						},
						FogOfWar: fow,
					})
				}
			}
		}
	}

	// token not found
	if err != nil || len(token) == 0 {
		filterGameState(s.G, g, level, nil)
		return g, nil
	}
	// token is present
	p, err := s.G.GetPlayerByKey(token)
	if err != nil {
		return nil, err
	}
	if !p.IsAdmin {
		if strings.HasPrefix(p.GetName(), "leonidas") {
			filterGameState(s.G, g, level, gameobject.CoordinatesToPosition(p.GetPosition()))
		} else if level != nil {
			filterGameState(s.G, g, level, gameobject.CoordinatesToPosition(p.GetPosition()))
		} else if level == nil {
			filterGameState(s.G, g, &p.GetPosition().Level, gameobject.CoordinatesToPosition(p.GetPosition()))
		}
		g.Character = p.Character
		g.CurrentPosition = gameobject.CoordinatesToPosition(p.GetPosition())
		g.CurrentLevel = &p.GetPosition().Level
	} else {
		// Monster admin
		filterMonsterGameState(s.G, g)
	}

	return g, nil
}

func (s *server) Game(ctx context.Context, params *api.GameStateParams) (*api.GameState, error) {
	return s.gameState(ctx, params, nil)
}

func (s *server) Players(ctx context.Context, params *api.PlayersParams) (*api.PlayersInfo, error) {
	var players []*api.Character

	// Maybe block
	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	s.G.GameLock.RUnlock()
	// Players are special and is not blocking by default also it is blocking before the actual work
	if params.Blocking != nil && *params.Blocking {
		s.G.WaitForNextTick(tick)
	}

	// read players
	s.G.GameLock.RLock()
	for _, p := range s.G.Players {
		players = append(players, p.Character)
	}
	s.G.GameLock.RUnlock()

	return &api.PlayersInfo{Players: players}, nil
}

func (s *server) Levels(ctx context.Context, params *api.PlayersParams) (*api.AvailableLevels, error) {
	var levels []int32

	// Maybe block
	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	s.G.GameLock.RUnlock()
	// Levels are special and is not blocking by default also it is blocking before the actual work
	if params.Blocking != nil && *params.Blocking {
		s.G.WaitForNextTick(tick)
	}

	// read levels
	s.G.GameLock.RLock()
	for _, l := range s.G.Game.Map.Levels {
		levels = append(levels, l.Level)
	}
	s.G.GameLock.RUnlock()

	sort.Slice(levels, func(i, j int) bool { return levels[i] < levels[j] })

	return &api.AvailableLevels{Levels: levels}, nil
}

func (s *server) GameLevel(ctx context.Context, params *api.GameStateParamsLevel) (*api.GameState, error) {
	return s.gameState(ctx, &api.GameStateParams{
		Blocking: params.Blocking,
		Items:    params.Items,
		FogOfWar: params.FogOfWar,
	}, &params.Level)
}

func (s *server) Register(ctx context.Context, user *api.User) (*api.Registration, error) {
	r, err := handlers.RegisterUser(s.G, user)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *server) Buy(ctx context.Context, identifiers *api.IdentifiersWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Buy(s.G, identifiers.Identifiers, token)
	s.G.GameLock.RUnlock()
	if isBlocking(identifiers.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) PickUp(ctx context.Context, identifier *api.IdentifierWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.PickUp(s.G, identifier.Identifier, token)
	s.G.GameLock.RUnlock()
	if isBlocking(identifier.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) Move(ctx context.Context, coordinates *api.PositionWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Move(s.G, coordinates.Position, token)
	s.G.GameLock.RUnlock()
	if isBlocking(coordinates.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) Respawn(ctx context.Context, res *api.RespawnWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Respawn(s.G, token)
	s.G.GameLock.RUnlock()
	if isBlocking(res.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) Skill(ctx context.Context, skill *api.SkillUseWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Skill(s.G, skill.SkillUse, token)
	s.G.GameLock.RUnlock()
	if isBlocking(skill.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) Commands(ctx context.Context, commands *api.CommandsBatchWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Commands(s.G, commands.CommandsBatch, token)
	s.G.GameLock.RUnlock()
	if isBlocking(commands.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) MonstersCommands(ctx context.Context, commands *api.CommandsForMonstersWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.MonsterCommands(s.G, commands.CommandsForMonsters, token)
	s.G.GameLock.RUnlock()
	if isBlocking(commands.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) Yell(ctx context.Context, message *api.MessageWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.Yell(s.G, message.Message, token)
	s.G.GameLock.RUnlock()
	if isBlocking(message.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func (s *server) AssignSkillPoints(ctx context.Context, attributes *api.AttributesWithParams) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	s.G.GameLock.RLock()
	tick := s.G.Game.Tick
	// TODO add player write lock
	err = handlers.AssignAttributes(s.G, attributes.Attributes, token)
	s.G.GameLock.RUnlock()
	if isBlocking(attributes.Blocking) {
		s.G.WaitForNextTick(tick)
	}

	return &emptypb.Empty{}, err
}

func main() {
	// err := discord.SendAPIKeyToUser("API KEY", "tivvit")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("")
	// }

	g, err := dungeonsandtrolls.CreateGame()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8081))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	api.RegisterDungeonsAndTrollsServer(s, &server{G: g})
	log.Printf("server listening at %v", lis.Addr())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal().Msgf("failed to serve: %v", err)
		}
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8081",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Msgf("Failed to dial server: %s", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			header := request.Header.Get(apiKeyFieldName)
			md := metadata.Pairs(apiKeyFieldName, header)
			return md
		}))
	err = api.RegisterDungeonsAndTrollsHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal().Msgf("Failed to register gateway: %s", err)
	}

	gwServer := &http.Server{
		Addr:    ":8080",
		Handler: gwmux,
	}

	log.Info().Msg("Serving gRPC-Gateway on http://0.0.0.0:8080")
	log.Fatal().Err(gwServer.ListenAndServe())
}
