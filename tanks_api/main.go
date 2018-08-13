package main

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[string]*websocket.Conn) // connected clients
var broadcast = make(chan Message)             // broadcast channel
var data = make(map[string]interface{})

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
}

type Map struct {
	mapWidth  int
	mapHeight int
	schema    [][]interface{}
	users     map[string]User
	log       []string
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

func (m *Map) writeToLog(message string) []string {
	m.log = append(m.log, message)
	return m.log
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
	hashmap.schema[7][2] = "wall"
	hashmap.schema[7][3] = "wall"
	hashmap.schema[7][4] = "wall"
	data["map"] = hashmap.schema
	data["log"] = hashmap.log
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
		clients[username[0]] = ws
	}
	data["map"] = hashmap.schema
	data["log"] = hashmap.log
	ws.WriteJSON(data)
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, username[0])
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

func handleMessages(mutex *sync.Mutex) {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		log.Printf("%v", msg)
		username := msg.Username
		command := msg.Message
		if command == "create" {
			CreateTank(username)
			log := hashmap.writeToLog(strings.Split(username, "-")[0] + " вошел в игру.")
			data["log"] = log
			sendToClients(mutex, data)
		} else if command == "up" || command == "down" || command == "right" || command == "left" {
			StepUser(username, command)
			sendToClients(mutex, data)
		} else if command == "delete" {
			DeleteTank(username)
			log := hashmap.writeToLog(strings.Split(username, "-")[0] + " вышел из игры.")
			data["log"] = log
			sendToClients(mutex, data)
		} else if command == "fire" {
			go rocketFire(username, mutex)
		}
	}
}

func StepUser(username, route string) {
	user := hashmap.users[username]
	coords := user.coords
	tank := hashmap.schema[coords[0]][coords[1]]
	if _, ok := tank.(map[string]interface{}); ok {
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
	hashmap.users[tank.name] = User{name: tank.name, coords: coords, murders: 0, deaths: 0}
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

func sendToClients(mutex *sync.Mutex, data interface{}) {
	// Send it out to every client that is currently connected
	mutex.Lock()
	for client := range clients {
		err := clients[client].WriteJSON(data)
		if err != nil {
			log.Printf("error: %v", err)
			clients[client].Close()
			delete(clients, client)
		}
	}
	mutex.Unlock()
}

func rocketFire(username string, mutex *sync.Mutex) {
	user, ok := hashmap.users[username]
	if ok {
		coords := user.coords
		tank := hashmap.schema[coords[0]][coords[1]]
		rocket := Rocket{tank: username}
		rocketMap := map[string]interface{}{"tank": rocket.tank}
		var arrLevel, indexHeight, indexWidth, tmpVal int
		var equalVal bool
		hasRoute := false
		route := tank.(map[string]interface{})["route"]
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
							if _, ok := hashmap.schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
								hashmap.schema[coords[0]][coords[1]] = "null"
								sendToClients(mutex, data)
							}
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
							hashmap.users[victimName.(string)] = victim
							hashmap.users[username] = user
							log := hashmap.writeToLog(strings.Split(username, "-")[0] + " убил " + strings.Split(victimName.(string), "-")[0])
							data["log"] = log
							sendToClients(mutex, data)
							clients[victimName.(string)].WriteJSON(map[string]bool{"dead": true})
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
						sendToClients(mutex, data)
					}
					break
				}
			}
			sendToClients(mutex, data)
			time.Sleep(30 * time.Millisecond)
		}
		return
	} else {
		log.Printf("User %v is not found!", username)
		return
	}
}

// func BarbarossaBot() {
// 	coords := findNullRect(&hashmap)
// 	hashmap.users["BarbarossaBot"] = User{name: "BarbarossaBot", coords: coords, murders: 0, deaths: 0}
// 	for {

// 	}
// }
