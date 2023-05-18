package main

import (
	"encoding/json"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
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
	game.Inputs["player 1"] = []dungeonsandtrolls.CommandI{mc}
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

func main() {
	g, err := dungeonsandtrolls.CreateGame()
	if err != nil {
		log.Fatal().Err(err)
	}

	http.HandleFunc("/", addDefaultHeaders(func(w http.ResponseWriter, r *http.Request) {
		gameHandler(g, w, r)
	}))
	http.HandleFunc("/actions", addDefaultHeaders(func(w http.ResponseWriter, r *http.Request) {
		actionHandler(g, w, r)
	}))

	log.Info().Msg("Starting server")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal().Err(err)
	}
}
