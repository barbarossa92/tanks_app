package main

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
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

type User struct {
	name    string
	coords  [2]int
	murders int8
	deaths  int8
	conn    *websocket.Conn
	mu      sync.Mutex
}

type Map struct {
	mapWidth  int
	mapHeight int
	schema    [][]interface{}
	users     map[string]User
}

var hashmap Map = Map{mapWidth: 20, mapHeight: 10, schema: [][]interface{}{}, users: make(map[string]User)}

type Rocket struct {
	tank string
}

type Tank struct {
	route    string
	name     string
	tankType string
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	hashmap.schema = make([][]interface{}, hashmap.mapHeight)
	for i := 0; i < hashmap.mapHeight; i++ {
		hashmap.schema[i] = make([]interface{}, hashmap.mapWidth)
		for j := 0; j < hashmap.mapWidth; j++ {
			hashmap.schema[i][j] = "null"
		}
	}
	hashmap.schema[1][1] = "wall"
	hashmap.schema[1][2] = "wall"
	hashmap.schema[1][3] = "wall"
	hashmap.schema[2][1] = "wall"
	log.Println(hashmap.schema)
	// SetMaps()
	// Start listening for incoming chat messages
	go handleMessages()
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// Register our new client
	clients[ws] = true
	ws.WriteJSON(hashmap.schema)
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func findNullRect(m *Map) [2]int {
	randWidth := random(0, m.mapWidth)
	randHeigth := random(0, m.mapHeight)
	if m.schema[randHeigth][randWidth] == "null" {
		coords := [2]int{randHeigth, randWidth}
		return coords
	}
	return findNullRect(m)
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		log.Printf("%v", msg)
		username := msg.Username
		command := msg.Message
		if command == "create" {
			CreateTank(username)
			sendToClients()
		} else if command == "up" || command == "down" || command == "right" || command == "left" {
			StepUser(username, command)
			sendToClients()
		} else if command == "delete" {
			DeleteTank(username)
			sendToClients()
		} else if command == "fire" {
			go rocketFire(username)
		}
	}
}

func StepUser(username, route string) {
	user := hashmap.users[username]
	coords := user.coords
	tank := hashmap.schema[coords[0]][coords[1]]
	if tank.(map[string]interface{})["route"] == route {
		if route == "up" && coords[0] > 0 && hashmap.schema[coords[0]-1][coords[1]] == "null" {
			tank.(map[string]interface{})["coords"] = [2]int{coords[0] - 1, coords[1]}
			hashmap.schema[coords[0]-1][coords[1]] = tank
			hashmap.schema[coords[0]][coords[1]] = "null"
			user.coords[0]--
			hashmap.users[username] = user
		} else if route == "down" && coords[0] < hashmap.mapHeight-1 && hashmap.schema[coords[0]+1][coords[1]] == "null" {
			tank.(map[string]interface{})["coords"] = [2]int{coords[0] + 1, coords[1]}
			hashmap.schema[coords[0]+1][coords[1]] = tank
			hashmap.schema[coords[0]][coords[1]] = "null"
			user.coords[0]++
			hashmap.users[username] = user
		} else if route == "right" && coords[1] < hashmap.mapWidth-1 && hashmap.schema[coords[0]][coords[1]+1] == "null" {
			tank.(map[string]interface{})["coords"] = [2]int{coords[0], coords[1] + 1}
			hashmap.schema[coords[0]][coords[1]+1] = tank
			hashmap.schema[coords[0]][coords[1]] = "null"
			user.coords[1]++
			hashmap.users[username] = user
		} else if route == "left" && coords[1] > 0 && hashmap.schema[coords[0]][coords[1]-1] == "null" {
			tank.(map[string]interface{})["coords"] = [2]int{coords[0], coords[1] - 1}
			hashmap.schema[coords[0]][coords[1]-1] = tank
			hashmap.schema[coords[0]][coords[1]] = "null"
			user.coords[1]--
			hashmap.users[username] = user
		} else {
			return
		}
	} else {
		hashmap.schema[coords[0]][coords[1]].(map[string]interface{})["route"] = route
	}
	return
}

func CreateTank(username string) {
	coords := findNullRect(&hashmap)
	tank := Tank{route: "right", name: username, tankType: "user"}
	tankMap := make(map[string]interface{})
	tankMap["route"] = tank.route
	tankMap["name"] = tank.name
	tankMap["tankType"] = tank.tankType
	tankMap["coords"] = coords
	hashmap.schema[coords[0]][coords[1]] = tankMap
	hashmap.users[tank.name] = User{coords: coords, murders: 0, deaths: 0}
}

func DeleteTank(username string) {
	var user User = hashmap.users[username]
	coords := user.coords
	hashmap.schema[coords[0]][coords[1]] = "null"
	delete(hashmap.users, username)
}

func checkRect(rect interface{}) bool {
	switch rect {
	case "wall", "tank", "dummy":
		return false
	}
	return true
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func sendToClients() {
	// Send it out to every client that is currently connected
	for client := range clients {
		err := client.WriteJSON(hashmap.schema)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func rocketFire(username string) {
	user := hashmap.users[username]
	coords := user.coords
	tank := hashmap.schema[coords[0]][coords[1]]
	route := tank.(map[string]interface{})["route"]
	rocket := Rocket{tank: username}
	rocketMap := map[string]interface{}{"tank": rocket.tank}
	var arrLevel, indexHeight, indexWidth, tmpVal int
	var equalVal bool
	hasRoute := false
	for {
		switch {
		case route == "up":
			arrLevel = 0
			indexHeight = coords[0] - 1
			indexWidth = coords[1]
			equalVal = coords[0]-1 >= 0
			hasRoute = true
			tmpVal = -1
		case route == "down":
			arrLevel = 0
			indexHeight = coords[0] + 1
			indexWidth = coords[1]
			equalVal = coords[0]+1 <= hashmap.mapHeight-1
			hasRoute = true
			tmpVal = 1
		case route == "left":
			arrLevel = 1
			indexHeight = coords[0]
			indexWidth = coords[1] - 1
			equalVal = coords[1]-1 >= 0
			hasRoute = true
			tmpVal = -1
		case route == "right":
			arrLevel = 1
			indexHeight = coords[0]
			indexWidth = coords[1] + 1
			equalVal = coords[1]+1 <= hashmap.mapWidth-1
			hasRoute = true
			tmpVal = 1
		}
		if hasRoute {
			if equalVal {
				nextRect := hashmap.schema[indexHeight][indexWidth]
				if nextRect != "null" {
					if nextRect == "wall" {
						hashmap.schema[coords[0]][coords[1]] = "null"
						sendToClients()
						break
					} else if _, ok := nextRect.(map[string]interface{})["route"]; ok {
						victimName := hashmap.schema[indexHeight][indexWidth].(map[string]interface{})["name"]
						hashmap.schema[indexHeight][indexWidth] = "null"
						if _, ok := hashmap.schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
							hashmap.schema[coords[0]][coords[1]] = "null"
						}
						victim := hashmap.users[victimName.(string)]
						victim.deaths++
						user.murders++
						sendToClients()
						break
					}
				} else {
					if _, ok := hashmap.schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
						hashmap.schema[coords[0]][coords[1]] = "null"
					}
					hashmap.schema[indexHeight][indexWidth] = rocketMap
					coords[arrLevel] = coords[arrLevel] + tmpVal
				}
			} else {
				if _, ok := hashmap.schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
					hashmap.schema[coords[0]][coords[1]] = "null"
					sendToClients()
				}
				break
			}
		}
		sendToClients()
		time.Sleep(30 * time.Millisecond)
	}
}
