package main

import (
	"log"
	"math/rand"
	"net/http"
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

// Define our message object
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type User struct {
	coords  [2]int
	murders int8
	deaths  int8
}

type Map struct {
	mapWidth  int
	mapHeight int
	schema    [][]interface{}
	users     map[string]User
}

var hashmap Map = Map{mapWidth: 15, mapHeight: 15, schema: [][]interface{}{}, users: make(map[string]User)}

type Rocket struct {
	tank string
}

type Tank struct {
	route    string
	name     string
	tankType string
}

// func rocketFire(username string) {
// 	coords := hashmap.users[username]
// 	tank := hashmap.schema[coords[0]][coords[1]]
// 	route := tank.(map[string]interface{})["route"]
// 	rocket := Rocket{tank: username}
// 	for {
// 		if route == "up" {
// 			if hashmap.schema[coords[0]][coords[1]] == "tank" {
// 				hashmap.schema[coords[0]-1][coords[1]] = "null"
// 			} else if hashmap.schema[coords[0]-1][coords[1]] == "wall" {
// 				continue
// 			} else {
// 				hashmap.schema[coords[0]-1][coords[1]] = "rocket"
// 			}
// 			if hashmap.schema[coords[0]][coords[1]] == "rocket" {
// 				hashmap.schema[coords[0]][coords[1]] = "null"
// 			}
// 			coords = [2]int{coords[0] - 1, coords[1]}
// 		} else if route == "down" {
// 			hashmap.schema[coords[0]+1][coords[1]] = "rocket"
// 			if hashmap.schema[coords[0]][coords[1]] == "rocket" {
// 				hashmap.schema[coords[0]][coords[1]] = "null"
// 			}
// 			coords = [2]int{coords[0] + 1, coords[1]}
// 		} else if route == "right" {
// 			hashmap.schema[coords[0]][coords[1]+1] = "rocket"
// 			if hashmap.schema[coords[0]][coords[1]] == "rocket" {
// 				hashmap.schema[coords[0]][coords[1]] = "null"
// 			}
// 			coords = [2]int{coords[0], coords[1] + 1}
// 		} else if route == "left" {
// 			hashmap.schema[coords[0]][coords[1]-1] = "rocket"
// 			if hashmap.schema[coords[0]][coords[1]] == "null" {
// 				hashmap.schema[coords[0]][coords[1]] = "null"
// 			}
// 			coords = [2]int{coords[0], coords[1] - 1}
// 		} else {
// 			return
// 		}
// 		for client := range clients {
// 			err := client.WriteJSON(hashmap.schema)
// 			if err != nil {
// 				log.Printf("error: %v", err)
// 				client.Close()
// 				delete(clients, client)
// 			}
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

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
		if command == "barbarossa" {
			go BarbarossaBot()
			return
		} else if command == "create" {
			CreateTank(username)
		} else if command == "up" || command == "down" || command == "right" || command == "left" {
			StepUser(username, command)
		}
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
}

func BarbarossaBot() {
	var coords [2]int
	coords = findNullRect(&hashmap)
	tank := Tank{route: "up", name: "barbarossa", tankType: "bot"}
	hashmap.schema[coords[0]][coords[1]] = tank

	for {
		step := random(1, 5)
		log.Print(step)
		if step == 1 {
			if coords[0]-1 >= 0 {
				if checkRect(hashmap.schema[coords[0]-1][coords[1]]) {
					hashmap.schema[coords[0]][coords[1]] = "null"
					hashmap.schema[coords[0]-1][coords[1]] = tank
					coords = [2]int{coords[0] - 1, coords[1]}
				}
			}
		} else if step == 2 {
			if coords[0]+1 <= hashmap.mapHeight {
				if checkRect(hashmap.schema[coords[0]+1][coords[1]]) {
					hashmap.schema[coords[0]][coords[1]] = "null"
					hashmap.schema[coords[0]+1][coords[1]] = tank
					coords = [2]int{coords[0] + 1, coords[1]}
				}
			}
		} else if step == 3 {
			if coords[1]-1 >= 0 {
				if checkRect(hashmap.schema[coords[0]][coords[1]-1]) {
					hashmap.schema[coords[0]][coords[1]] = "null"
					hashmap.schema[coords[0]][coords[1]-1] = tank
					coords = [2]int{coords[0], coords[1] - 1}
				}
			}
		} else {
			if coords[1]+1 <= hashmap.mapWidth {
				if checkRect(hashmap.schema[coords[0]][coords[1]+1]) {
					hashmap.schema[coords[0]][coords[1]] = "null"
					hashmap.schema[coords[0]][coords[1]+1] = tank
					coords = [2]int{coords[0], coords[1] + 1}
				}
			}
		}
		for client := range clients {
			err := client.WriteJSON(hashmap.schema)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		time.Sleep(1 * time.Second)
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
