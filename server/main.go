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
	"net"
	"net/http"
	"sort"
	"strings"
)

const apiKeyFieldName = "X-API-key"

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

func filterGameState(game *dungeonsandtrolls.Game, g *api.GameState, level *int32) {
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

	// token not found
	if err != nil || len(token) == 0 {
		filterGameState(s.G, g, level)
		return g, nil
	}
	// token is present
	p, err := s.G.GetPlayerByKey(token)
	if err != nil {
		return nil, err
	}
	if !p.IsAdmin {
		if strings.HasPrefix(p.GetName(), "leonidas") {
			filterGameState(s.G, g, level)
		} else if level != nil {
			filterGameState(s.G, g, level)
		} else if level == nil {
			filterGameState(s.G, g, &p.GetPosition().Level)
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

	// read players
	s.G.GameLock.RLock()
	for _, l := range s.G.Game.Map.Levels {
		levels = append(levels, l.Level)
	}
	s.G.GameLock.RUnlock()

	sort.Slice(levels, func(i, j int) bool {
		return i < j
	})

	return &api.AvailableLevels{Levels: levels}, nil
}

func (s *server) GameLevel(ctx context.Context, params *api.GameStateParamsLevel) (*api.GameState, error) {
	return s.gameState(ctx, &api.GameStateParams{
		Blocking: params.Blocking,
		Items:    params.Items,
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
