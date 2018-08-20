package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/barbarossa92/tanks_app/tanks_api/maps"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var broadcast = make(chan Message) // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var hashmapWalls = [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 2}, {3, 2}, {4, 2}, {5, 2}, {6, 5}, {7, 5}, {8, 5}, {9, 5},
	{0, 13}, {1, 13}, {2, 13}, {3, 13}, {4, 13}, {4, 14}, {4, 15}, {4, 16}, {7, 19}, {7, 18}, {7, 17}, {7, 16}, {7, 15},
	{4, 9}, {5, 9}, {6, 9}, {7, 9}, {7, 10}, {8, 10}}
var hashmap maps.Map = *maps.CreateMap(20, 10, hashmapWalls)
var mutex sync.Mutex

func main() {

	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/get-users", GetMapUsers).Methods("GET")
	apiRouter.HandleFunc("/create-tank", UsernameRequiredDecorator(CreateTank)).Methods("POST")
	apiRouter.HandleFunc("/delete-tank", UsernameRequiredDecorator(DeleteTank)).Methods("POST")
	apiRouter.HandleFunc("/move-tank", UsernameRequiredDecorator(MoveTank)).Methods("POST")
	apiRouter.HandleFunc("/fire", UsernameRequiredDecorator(RocketFire)).Methods("POST")
	// Configure websocket route
	r.HandleFunc("/ws", handleConnections)
	// Start listening for incoming chat messages
	go handleMessages(&mutex)
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Get query username from url
	username, ok := r.URL.Query()["username"]
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// Register our new client
	if ok {
		hashmap.Clients[username[0]] = ws
	}
	data := hashmap.GetData()
	ws.WriteJSON(data)
	ws.SetReadDeadline(time.Now().Add(24 * time.Hour))
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(hashmap.Clients, username[0])
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages(mutex *sync.Mutex) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// log.Printf("%v", msg)
		username := msg.Username
		command := msg.Message
		if command == "create" {
			hashmap.CreateTank(username, "user", mutex)
		} else if command == "up" || command == "down" || command == "right" || command == "left" {
			hashmap.StepUser(username, command, mutex)
		} else if command == "delete" {
			hashmap.DeleteTank(username, mutex)
		} else if command == "fire" {
			go hashmap.RocketFire(username, mutex)
		}
	}
}
