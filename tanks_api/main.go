package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tanks_app/tanks_api/maps"
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

var hashmap = maps.Map{MapWidth: 20, MapHeight: 10, Schema: [][]interface{}{}, Users: make(map[string]maps.User), Clients: make(map[string]*websocket.Conn)}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	hashmap.Schema = make([][]interface{}, hashmap.MapHeight)
	for i := 0; i < hashmap.MapHeight; i++ {
		hashmap.Schema[i] = make([]interface{}, hashmap.MapWidth)
		for j := 0; j < hashmap.MapWidth; j++ {
			hashmap.Schema[i][j] = "null"
		}
	}
	hashmap.Schema[1][1] = "wall"
	hashmap.Schema[1][2] = "wall"
	hashmap.Schema[1][3] = "wall"
	hashmap.Schema[2][1] = "wall"
	hashmap.Schema[7][2] = "wall"
	hashmap.Schema[7][3] = "wall"
	hashmap.Schema[7][4] = "wall"
	var mutex sync.Mutex
	// SetMaps()
	// Start listening for incoming chat messages
	go handleMessages(&mutex)
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
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
	data := make(map[string]interface{})
	data["map"] = hashmap.Schema
	data["log"] = hashmap.Log
	ws.WriteJSON(data)
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
		log.Printf("%v", msg)
		username := msg.Username
		command := msg.Message
		data := make(map[string]interface{})
		data["map"] = hashmap.Schema
		if command == "create" {
			hashmap.CreateTank(username)
			log := hashmap.WriteToLog(strings.Split(username, "-")[0] + " вошел в игру.")
			data["log"] = log
			hashmap.SendToClients(mutex, data)
		} else if command == "up" || command == "down" || command == "right" || command == "left" {
			hashmap.StepUser(username, command)
			hashmap.SendToClients(mutex, data)
		} else if command == "delete" {
			hashmap.DeleteTank(username)
			log := hashmap.WriteToLog(strings.Split(username, "-")[0] + " вышел из игры.")
			data["log"] = log
			hashmap.SendToClients(mutex, data)
		} else if command == "fire" {
			go hashmap.RocketFire(username, mutex)
		}
	}
}
