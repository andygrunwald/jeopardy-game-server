package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
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

// webServer represents the data structure that
// keeps everything together for the webServer
// incl. the web sockets
type webServer struct {
	Listen         string
	ButtonHits     chan buttonHit
	SocketUpgrader websocket.Upgrader
	SocketClients  map[*websocket.Conn]bool
}

// buttonHit represents the message that will be sent
// once a button/buzzer was hit
type buttonHit struct {
	// Color is the color of the buzzer that was hit
	// see constants button* above
	Color string
}

const (
	// buttonRed represents a physical buzzer in color red
	buttonRed string = "red"
	// buttonRed represents a physical buzzer in color green
	buttonGreen string = "green"
	// buttonRed represents a physical buzzer in color blue
	buttonBlue string = "blue"
	// buttonRed represents a physical buzzer in color yellow
	buttonYellow string = "yellow"
)

func main() {
	addr := ":8000"
	if v := os.Getenv("JGAME_SRV_ADDR"); len(v) > 0 {
		addr = v
	}

	log.Printf("Jeopardy Game Server ... %v, commit %v, built at %v\n", version, commit, date)
	log.Println("Jeopardy Game Server ... starting ...")

	//
	// Start web socket and webserver
	//
	buttonHits := make(chan buttonHit, 4)
	httpServer := &webServer{
		Listen:     addr,
		ButtonHits: buttonHits,
		SocketUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// This is not good idea. Why?
				// See https://github.com/gorilla/websocket/issues/367
				// But we (assume to) run locally on a
				// RaspberryPi. And we want to make it work for now ;)
				return true
			},
		},
		SocketClients: make(map[*websocket.Conn]bool),
	}

	go httpServer.startWebserver(addr)
	go httpServer.socketBroadcast()

	// Usage of https://gobot.io/for dealing with
	// physical buzzers.
	//
	// Whatever you do with the GPIO pins
	// the raw BCM2835 pinout mapping to Raspberry Pi at
	// https://godoc.org/github.com/stianeikeland/go-rpio
	// is super helpful.
	r := raspi.NewAdaptor()
	red := gpio.NewButtonDriver(r, "40")
	green := gpio.NewButtonDriver(r, "38")
	blue := gpio.NewButtonDriver(r, "36")
	//yellow := gpio.NewButtonDriver(r, "32")

	work := func() {
		red.On(gpio.ButtonPush, func(data interface{}) {
			log.Println("Button red pressed")
			msg := buttonHit{
				Color: buttonRed,
			}
			buttonHits <- msg
		})

		green.On(gpio.ButtonPush, func(data interface{}) {
			log.Println("Button green pressed")
			msg := buttonHit{
				Color: buttonGreen,
			}
			buttonHits <- msg
		})

		blue.On(gpio.ButtonPush, func(data interface{}) {
			log.Println("Button blue pressed")
			msg := buttonHit{
				Color: buttonBlue,
			}
			buttonHits <- msg
		})

		/*
			yellow.On(gpio.ButtonPush, func(data interface{}) {
				log.Println("Button yellow pressed")
				msg := buttonHit{
					Color: buttonYellow,
				}
				buttonHits <- msg
			})
		*/
	}

	robot := gobot.NewRobot("buttonBot",
		[]gobot.Connection{r},
		//[]gobot.Device{red, green, blue, yellow},
		[]gobot.Device{red, green, blue},
		work,
	)

	robot.Start()
}

// startWebserver will start the webserver incl.
// websocket - Nothing more.
func (httpServer *webServer) startWebserver(addr string) {
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

	log.Printf("Starting websocket on %s ...\n", httpServer.Listen)
	r.HandleFunc("/socket", httpServer.websocketHandler)

	// Static file server
	fs := http.FileServer(http.Dir("games"))
	r.PathPrefix("/").Handler(fs)

	http.Handle("/", r)

	log.Printf("Jeopardy Game Server ... starting on %s... Done\n", addr)
	if err := http.ListenAndServe(addr, handlers.LoggingHandler(os.Stdout, r)); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// websocketHandler registers new websocket clients
func (httpServer *webServer) websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := httpServer.SocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Register new client
	httpServer.SocketClients[c] = true
}

// socketBroadcast will send a single message
// to all connected websocket clients (broadcasting).
func (httpServer *webServer) socketBroadcast() {
	log.Println("Starting socket broadcast ...")
	for {
		msg := <-httpServer.ButtonHits

		jsonMessage, _ := json.Marshal(msg)
		log.Printf("Broadcasting message: %v\n", string(jsonMessage))

		// Send to every client that is currently connected
		for client := range httpServer.SocketClients {
			err := client.WriteMessage(websocket.TextMessage, jsonMessage)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(httpServer.SocketClients, client)
			}
		}
	}
}
