package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

// SeasonsHandler represents the API endpoint to receive all
// available Jeopardy Game Seasons
func (s *server) SeasonsHandler(w http.ResponseWriter, r *http.Request) {
	// Sort keys in reverse order to deliver a stable order
	var keys []string
	for k := range s.Cache.SeasonCache {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	// Load season data from cache
	seasons := []json.RawMessage{}
	for _, k := range keys {
		seasons = append(seasons, json.RawMessage(s.Cache.SeasonCache[k]))
	}

	// Convert it to JSON
	// or not in case of failure ;)
	b, err := json.Marshal(seasons)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: %+v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(b))
}

// SeasonHandler represents the API endpoint to all games from
// a single Jeopardy Game Season
func (s *server) SeasonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Check if we have a seasonID in the request
	if len(vars["seasonID"]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: seasonID missing - usage '/season/{seasonID}'")
		return
	}

	// Set prefix and suffix to get only the right cache entries.
	// We are searching the data in the GameCache, but we are not interested
	// in the actual game data. We are interested in the overview data.
	// prefix will be set to something like "games/Season_1/"
	// suffix to something like "/overview.json"
	prefix := s.GamesPath + string(os.PathSeparator) + vars["seasonID"] + string(os.PathSeparator)
	suffix := string(os.PathSeparator) + "overview.json"

	// Sort keys in reverse order to deliver a stable order
	var keys []string
	for k := range s.Cache.GameCache {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	// Load game overview data from the cache
	// Yep we know, we could build the cache in a more smarter way
	// But hey, who cares? "Make it work, make it fast, make it beautiful".
	// Right now it works, and it is fast enough. And it is late.
	// So the new rule is "Make it work, make it fast enough and be pragmatic".
	// Feel free to quote me. Andy Grunwald wrote this at 2019-02-20, 11:53 pm.
	games := []json.RawMessage{}
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) && strings.HasSuffix(k, suffix) {
			games = append(games, json.RawMessage(s.Cache.GameCache[k]))
		}
	}

	// Convert it to JSON
	// or not in case of failure ;)
	b, err := json.Marshal(games)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: %+v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(b))
}

// SeasonGameHandler represents the API endpoint to retrieve one single game
func (s *server) SeasonGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Check if we have a gameID in the request
	if len(vars["gameID"]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: gameID missing - usage '/game/{gameID}'")
		return
	}

	// A normal game ID looks like 'Season_1---2019-02-18'
	// We applied another hack here. We introduced a "special character".
	// '---' will be replaced by a '/'
	// The reason is to avoid issues with the HTTP Router when there is a '/'
	// in the gameID. There is 100% a way to fix this.
	// But again, "Make it work, and so on".
	gamesID := strings.Replace(vars["gameID"], "---", string(os.PathSeparator), 1)
	p := s.GamesPath + string(os.PathSeparator) + gamesID + string(os.PathSeparator) + "game.json"

	// Check if the game in the cache.
	// If not, throw an error.
	if v, ok := s.Cache.GameCache[p]; ok {
		b := json.RawMessage(v)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(b))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: key %+v not found in GameCache\n", p)
}
