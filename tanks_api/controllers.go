package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func UsernameRequiredDecorator(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			username := r.FormValue("username")
			if username == "" {
				message := map[string]string{"message": "'username' field is required!"}
				errData, _ := json.Marshal(message)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(errData)
				return
			}
			f(w, r)
		}
	}
}

func CreateTank(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	usernameFinal := username + "-" + RandStringBytes(6)
	tank := hashmap.CreateTank(usernameFinal, "bot", &mutex)
	tankData, err := json.Marshal(tank)
	if err != nil {
		log.Printf("[API] message: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(tankData)
}

func DeleteTank(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	status, err := hashmap.DeleteTank(username, &mutex)
	message := make(map[string]string)
	if !status {
		log.Printf("[API] message: %v", err)
		message["message"] = err
		messageData, _ := json.Marshal(message)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(messageData)
		return
	}
	message["message"] = "Tank is successfully deleted."
	messageData, _ := json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(messageData)
}

func MoveTank(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	route := r.FormValue("route")
	data := make(map[string]interface{})
	if route == "" {
		data["message"] = "'route' field is required!"
		errData, _ := json.Marshal(data)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errData)
		return
	}
	status, err := hashmap.StepUser(username, route, &mutex)
	if !status {
		data["message"] = err
		errData, _ := json.Marshal(data)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errData)
		return
	}
	data["message"] = "Tank is successfully move " + route
	data["currentCoords"] = hashmap.Users[username].Coords
	finalData, _ := json.Marshal(data)
	w.WriteHeader(http.StatusOK)
	w.Write(finalData)
}

func RocketFire(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	go hashmap.RocketFire(username, &mutex)
	data := map[string]string{"message": "Fire"}
	finalData, _ := json.Marshal(data)
	w.WriteHeader(http.StatusOK)
	w.Write(finalData)
}

func GetMapInfo(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["map_height"] = hashmap.MapHeight
	data["map_width"] = hashmap.MapWidth
	data["tanks"] = len(hashmap.Users)
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func GetWallsCoords(w http.ResponseWriter, r *http.Request) {
	walls := map[string][][2]int{"walls": hashmapWalls}
	wallsData, _ := json.Marshal(walls)
	w.WriteHeader(http.StatusOK)
	w.Write(wallsData)
}

func GetEnemies(w http.ResponseWriter, r *http.Request) {
	queries := mux.Vars(r)
	username, ok := queries["username"]
	if !ok {
		err := map[string]string{"message": "'username' query is required!"}
		errData, _ := json.Marshal(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errData)
		return
	}
	var enemies []map[string]interface{}
	for _, v := range hashmap.Users {
		if v.Name != username {
			enemyMap := make(map[string]interface{})
			clearUsername := strings.Split(v.Name, "-")[0]
			e, _ := json.Marshal(&v)
			err := json.Unmarshal(e, &enemyMap)
			if err != nil {
				log.Printf("[API] message: %v", err)
			}
			enemyMap["name"] = clearUsername
			enemies = append(enemies, enemyMap)
		}
	}
	enemiesData, _ := json.Marshal(enemies)
	w.WriteHeader(http.StatusOK)
	w.Write(enemiesData)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	queries := mux.Vars(r)
	username, ok := queries["username"]
	if !ok {
		err := map[string]string{"message": "'username' query is required!"}
		errData, _ := json.Marshal(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errData)
		return
	}
	user := make(map[string]interface{})
	for k, v := range hashmap.Users {
		if strings.HasPrefix(k, username) {
			u, _ := json.Marshal(&v)
			err := json.Unmarshal(u, &user)
			if err != nil {
				log.Printf("[API] message: %", err)
			}
		}
	}
	if len(user) == 0 {
		user["is_dead"] = true
	} else {
		user["is_dead"] = false
	}
	delete(user, "name")
	userData, _ := json.Marshal(user)
	w.WriteHeader(http.StatusOK)
	w.Write(userData)
}
