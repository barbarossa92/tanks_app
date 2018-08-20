package maps

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	Name    string
	Coords  [2]int
	Murders int
	Deaths  int
}

type Map struct {
	MapWidth  int
	MapHeight int
	Schema    [][]interface{}
	Users     map[string]User
	Log       []string
	Clients   map[string]*websocket.Conn
}

type Rocket struct {
	Tank string
}

type Tank struct {
	Route    string
	Name     string
	TankType string
}

func (m *Map) WriteToLog(message string) []string {
	datetime := time.Now().Format("2006/01/02/ 15:04:05")
	m.Log = append(m.Log, datetime+" "+message)
	return m.Log
}

func (m *Map) FindNullRect() [2]int {
	randWidth := random(0, m.MapWidth)
	randHeigth := random(0, m.MapHeight)
	if m.Schema[randHeigth][randWidth] == "null" {
		coords := [2]int{randHeigth, randWidth}
		return coords
	}
	return m.FindNullRect()
}

func (m *Map) RatingRefresh() map[string]map[string]int {
	rating := make(map[string]map[string]int)
	for u, k := range m.Users {
		rating[u] = make(map[string]int, 2)
		rating[u]["murders"] = k.Murders
		rating[u]["deaths"] = k.Deaths
	}
	return rating
}

func CreateMap(width, height int, walls [][2]int) *Map {
	hashmap := Map{MapWidth: width, MapHeight: height, Schema: [][]interface{}{}, Users: make(map[string]User), Clients: make(map[string]*websocket.Conn)}
	hashmap.Schema = make([][]interface{}, hashmap.MapHeight)
	for i := 0; i < hashmap.MapHeight; i++ {
		hashmap.Schema[i] = make([]interface{}, hashmap.MapWidth)
		for j := 0; j < hashmap.MapWidth; j++ {
			hashmap.Schema[i][j] = "null"
		}
	}
	for _, wall := range walls {
		hashmap.Schema[wall[0]][wall[1]] = "wall"
	}
	return &hashmap
}

func (m *Map) StepUser(username, route string, mutex *sync.Mutex) {
	user := m.Users[username]
	coords := user.Coords
	tank := m.Schema[coords[0]][coords[1]]
	if _, ok := tank.(map[string]interface{}); ok {
		if tank.(map[string]interface{})["route"] == route {
			if route == "up" && coords[0] > 0 && m.Schema[coords[0]-1][coords[1]] == "null" {
				tank.(map[string]interface{})["coords"] = [2]int{coords[0] - 1, coords[1]}
				m.Schema[coords[0]-1][coords[1]] = tank
				m.Schema[coords[0]][coords[1]] = "null"
				user.Coords[0]--
				m.Users[username] = user
			} else if route == "down" && coords[0] < m.MapHeight-1 && m.Schema[coords[0]+1][coords[1]] == "null" {
				tank.(map[string]interface{})["coords"] = [2]int{coords[0] + 1, coords[1]}
				m.Schema[coords[0]+1][coords[1]] = tank
				m.Schema[coords[0]][coords[1]] = "null"
				user.Coords[0]++
				m.Users[username] = user
			} else if route == "right" && coords[1] < m.MapWidth-1 && m.Schema[coords[0]][coords[1]+1] == "null" {
				tank.(map[string]interface{})["coords"] = [2]int{coords[0], coords[1] + 1}
				m.Schema[coords[0]][coords[1]+1] = tank
				m.Schema[coords[0]][coords[1]] = "null"
				user.Coords[1]++
				m.Users[username] = user
			} else if route == "left" && coords[1] > 0 && m.Schema[coords[0]][coords[1]-1] == "null" {
				tank.(map[string]interface{})["coords"] = [2]int{coords[0], coords[1] - 1}
				m.Schema[coords[0]][coords[1]-1] = tank
				m.Schema[coords[0]][coords[1]] = "null"
				user.Coords[1]--
				m.Users[username] = user
			} else {
				return
			}
		} else {
			m.Schema[coords[0]][coords[1]].(map[string]interface{})["route"] = route
		}
		m.SendToClients(mutex)
	}
	return
}

func (m *Map) CreateTank(username, tankType string, mutex *sync.Mutex) map[string]interface{} {
	coords := m.FindNullRect()
	tank := Tank{Route: "right", Name: username, TankType: tankType}
	tankMap := make(map[string]interface{})
	tankMap["route"] = tank.Route
	tankMap["name"] = tank.Name
	tankMap["tankType"] = tank.TankType
	tankMap["coords"] = coords
	m.Schema[coords[0]][coords[1]] = tankMap
	m.Users[tank.Name] = User{Name: tank.Name, Coords: coords, Murders: 0, Deaths: 0}
	m.WriteToLog(strings.Split(username, "-")[0] + " вошел в игру.")
	m.SendToClients(mutex)
	return tankMap
}

func (m *Map) DeleteTank(username string, mutex *sync.Mutex) (bool, string) {
	user, ok := m.Users[username]
	if !ok {
		return false, username + " is not found."
	}
	coords := user.Coords
	m.Schema[coords[0]][coords[1]] = "null"
	delete(m.Users, username)
	m.WriteToLog(strings.Split(username, "-")[0] + " вышел из игры.")
	m.SendToClients(mutex)
	return true, ""
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func (m *Map) SendToClients(mutex *sync.Mutex) {
	// Send it out to every client that is currently connected
	mutex.Lock()
	data := m.GetData()
	for client := range m.Clients {
		err := m.Clients[client].WriteJSON(data)
		if err != nil {
			log.Printf("error: %v", err)
			m.Clients[client].Close()
			delete(m.Clients, client)
		}
	}
	mutex.Unlock()
}

func (m *Map) RocketFire(username string, mutex *sync.Mutex) {
	user, ok := m.Users[username]
	if ok {
		coords := user.Coords
		tank := m.Schema[coords[0]][coords[1]]
		rocket := Rocket{Tank: username}
		rocketMap := map[string]interface{}{"tank": rocket.Tank}
		var arrLevel, indexHeight, indexWidth, tmpVal int
		var equalVal bool
		hasRoute := false
		tankMap, ok := tank.(map[string]interface{})
		if !ok {
			return
		}
		route := tankMap["route"]
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
				equalVal = coords[0]+1 <= m.MapHeight-1
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
				equalVal = coords[1]+1 <= m.MapWidth-1
				hasRoute = true
				tmpVal = 1
			}
			if hasRoute {
				if m.Schema[coords[0]][coords[1]] == "null" {
					break
				}
				if equalVal {
					nextRect := m.Schema[indexHeight][indexWidth]
					if nextRect != "null" {
						if nextRect == "wall" {
							if _, ok := m.Schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
								m.Schema[coords[0]][coords[1]] = "null"
								m.SendToClients(mutex)
							}
							break
						} else if _, ok := nextRect.(map[string]interface{})["route"]; ok {
							victimName := m.Schema[indexHeight][indexWidth].(map[string]interface{})["name"]
							m.Schema[indexHeight][indexWidth] = "null"
							if _, ok := m.Schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
								m.Schema[coords[0]][coords[1]] = "null"
							}
							user.Murders++
							delete(m.Users, victimName.(string))
							m.Users[username] = user
							m.WriteToLog(strings.Split(username, "-")[0] + " убил " + strings.Split(victimName.(string), "-")[0])
							m.SendToClients(mutex)
							if conn, ok := m.Clients[victimName.(string)]; ok {
								conn.WriteJSON(map[string]bool{"dead": true})
							}
							break
						} else if _, ok := nextRect.(map[string]interface{})["tank"]; ok {
							m.Schema[indexHeight][indexWidth] = "null"
							if _, ok := m.Schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
								m.Schema[coords[0]][coords[1]] = "null"
							}
							m.SendToClients(mutex)
							break

						}
					} else {
						if _, ok := m.Schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
							m.Schema[coords[0]][coords[1]] = "null"
						}
						m.Schema[indexHeight][indexWidth] = rocketMap
						coords[arrLevel] = coords[arrLevel] + tmpVal
					}
				} else {
					if _, ok := m.Schema[coords[0]][coords[1]].(map[string]interface{})["tank"]; ok {
						m.Schema[coords[0]][coords[1]] = "null"
						m.SendToClients(mutex)
					}
					break
				}
			}
			m.SendToClients(mutex)
			time.Sleep(30 * time.Millisecond)
		}
		return
	} else {
		log.Printf("User %v is not found!", username)
		return
	}
}

func (m *Map) GetData() map[string]interface{} {
	data := make(map[string]interface{})
	data["map"] = m.Schema
	data["map_width"] = m.MapWidth
	data["map_height"] = m.MapHeight
	data["log"] = m.Log
	data["viewers_count"] = len(m.Clients)
	data["tanks_count"] = len(m.Users)
	data["rating"] = m.RatingRefresh()
	return data
}
