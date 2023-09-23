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

func filterGameState(game *dungeonsandtrolls.Game, g *api.GameState) {
	// filter monsters for non-monster players
	for _, l := range g.Map.Levels {
		for _, o := range l.Objects {
			for _, m := range o.Monsters {
				dungeonsandtrolls.HideNonPublicMonsterFields(game, m)
			}
		}
	}
	hideUnidentifiedItems(game, g)
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

func (s *server) Game(ctx context.Context, params *api.GameStateParams) (*api.GameState, error) {
	token, err := getToken(ctx)
	g, ok := proto.Clone(&s.G.Game).(*api.GameState)
	if !ok {
		return nil, fmt.Errorf("cloning GameState failed")
	}
	// token not found
	if err != nil || len(token) == 0 {
		filterGameState(s.G, g)
		return g, nil
	}
	// token is present
	p, err := s.G.GetPlayerByKey(token)
	if err != nil {
		return nil, err
	}
	if !p.IsAdmin {
		filterGameState(s.G, g)
	}
	g.Character = p.Character
	g.CurrentPosition = gameobject.CoordinatesToPosition(p.Position)
	g.CurrentLevel = p.Position.Level
	return g, nil
}

func (s *server) Register(ctx context.Context, user *api.User) (*api.Registration, error) {
	r, err := handlers.RegisterUser(s.G, user)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *server) Buy(ctx context.Context, identifiers *api.Identifiers) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Buy(s.G, identifiers, token)
}

func (s *server) PickUp(ctx context.Context, identifier *api.Identifier) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.PickUp(s.G, identifier, token)
}

func (s *server) Move(ctx context.Context, coordinates *api.Position) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Move(s.G, coordinates, token)
}

func (s *server) Respawn(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Respawn(s.G, token)
}

func (s *server) Skill(ctx context.Context, skill *api.SkillUse) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Skill(s.G, skill, token)
}

func (s *server) Commands(ctx context.Context, commands *api.CommandsBatch) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Commands(s.G, commands, token)
}

func (s *server) MonstersCommands(ctx context.Context, commands *api.CommandsForMonsters) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Yell(ctx context.Context, message *api.Message) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.Yell(s.G, message, token)
}

func (s *server) AssignSkillPoints(ctx context.Context, attributes *api.Attributes) (*emptypb.Empty, error) {
	token, err := getToken(ctx)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, handlers.AssignAttributes(s.G, attributes, token)
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
