package main

import (
	"dungeons-and-trolls/dungeonsandtrolls"
	"encoding/json"
	"github.com/rs/zerolog/log"
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

	log.Info().Msg("Starting server")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal().Err(err)
	}
}
