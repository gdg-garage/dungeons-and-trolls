package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func gameHandler(game *dungeonsandtrolls.Game, w http.ResponseWriter, r *http.Request) {
	gameJson, err := json.Marshal(game)
	if err != nil {
		http.Error(w, `{"message": "response marshal failed"}`, http.StatusInternalServerError)
		log.Err(err)
		return
	}
	_, err = w.Write(gameJson)
	if err != nil {
		http.Error(w, `{"message": "response write failed"}`, http.StatusInternalServerError)
		log.Err(err)
		return
	}
}

func actionHandler(game *dungeonsandtrolls.Game, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// TODO log and so on
		return
	}
	var mc dungeonsandtrolls.MoveCommand
	err = json.Unmarshal(body, &mc)
	if err != nil {
		return
	}
	// game.Inputs["player 1"] = []dungeonsandtrolls.CommandI{mc}
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, User-Agent")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		fn(w, r)
	}
}

type server struct {
	api.UnsafeDungeonsAndTrollsServer
	G *dungeonsandtrolls.Game
}

func (s *server) Game(ctx context.Context, params *api.GameStateParams) (*api.GameState, error) {
	return &s.G.Game, nil
}

func (s *server) Register(ctx context.Context, user *api.User) (*api.Registration, error) {
	r, err := handlers.RegisterUser(s.G, user)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *server) Buy(ctx context.Context, identifiers *api.Identifiers) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handlers.Buy(s.G, identifiers)
}

func (s *server) Equip(ctx context.Context, identifier *api.Identifier) (*emptypb.Empty, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *server) AssignSkillPoints(ctx context.Context, attributes *api.Attributes) (*emptypb.Empty, error) {
	return nil, nil
}
func (s *server) Move(ctx context.Context, coordinates *api.Coordinates) (*emptypb.Empty, error) {
	return nil, nil
}
func (s *server) Respawn(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}
func (s *server) Skill(ctx context.Context, spell *api.SkillUse) (*emptypb.Empty, error) {
	return nil, nil
}
func (s *server) Commands(ctx context.Context, commands *api.CommandsBatch) (*emptypb.Empty, error) {
	return nil, nil
}
func (s *server) MonstersCommands(ctx context.Context, commands *api.CommandsForMonsters) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *server) Yell(ctx context.Context, message *api.Message) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, handlers.Yell(s.G, message)
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

	http.HandleFunc("/", addDefaultHeaders(func(w http.ResponseWriter, r *http.Request) {
		gameHandler(g, w, r)
	}))
	http.HandleFunc("/actions", addDefaultHeaders(func(w http.ResponseWriter, r *http.Request) {
		actionHandler(g, w, r)
	}))

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

	gwmux := runtime.NewServeMux()
	// Register Greeter
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
