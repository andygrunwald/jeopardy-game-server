package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// cache represents the caching instance to avoid I/O calls
type cache struct {
	// SeasonCache keeps a list of all available Jeopardy seasons
	// This means in detail the 'games/{SEASON-NAME}/overview.json' files.
	// The games/{SEASON-NAME} is the key. The content of the respective
	// overview.json file the value.
	SeasonCache map[string][]byte
	// GameCache keeps all the game data
	// This means in detail the 'games/{SEASON-NAME}/{GAME-NAME}/(overview|game).json' files.
	// The full path to the game file is the key. The content of the respective
	// (overview|game).json file the value.
	GameCache map[string][]byte
}

// server represents the HTTP server context
// All HTTP endpoints are handled by this server.
type server struct {
	// GamesPath is where the Jeopardy games are stored
	GamesPath string
	// Cache is the caching instance that holds all games
	Cache *cache
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	addr := ":8000"
	if v := os.Getenv("JGAME_SRV_ADDR"); len(v) > 0 {
		addr = v
	}

	log.Printf("Jeopardy Game Server ... %v, commit %v, built at %v\n", version, commit, date)
	log.Println("Jeopardy Game Server ... starting ...")

	// Determine the path to load the games from
	gamesPath := "games"
	if v := os.Getenv("JEOPARDY_GAMES"); len(v) > 0 {
		gamesPath = v
	}
	log.Printf("Jeopardy Game Server ... Using games from path '%s'\n", gamesPath)

	// Load Jeopardy seasons and games into local in-memory cache
	log.Println("Jeopardy Game Server ... Creating games cache ...")
	seasonCache := make(map[string][]byte)
	gameCache := make(map[string][]byte)
	cache := &cache{
		SeasonCache: seasonCache,
		GameCache:   gameCache,
	}

	err := buildCache(gamesPath, cache)
	if err != nil {
		panic(err)
	}

	log.Println("Jeopardy Game Server ... Creating games cache ... Done")
	log.Printf("Jeopardy Game Server ... Loaded %d seasons and %d games into cache\n", len(cache.SeasonCache), len(cache.GameCache))

	// Define server and start HTTP interface
	s := &server{
		GamesPath: gamesPath,
		Cache:     cache,
	}

	r := mux.NewRouter()
	r.HandleFunc("/seasons", s.SeasonsHandler).Methods("GET")
	r.HandleFunc("/season/{seasonID}", s.SeasonHandler).Methods("GET")
	r.HandleFunc("/game/{gameID}", s.SeasonGameHandler).Methods("GET")
	http.Handle("/", r)

	log.Printf("Jeopardy Game Server ... starting on %s... Done\n", addr)
	log.Fatal(http.ListenAndServe(addr, handlers.LoggingHandler(os.Stdout, r)))
}
