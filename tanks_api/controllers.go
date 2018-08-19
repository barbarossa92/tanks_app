package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func GetMapUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	users, err := json.Marshal(hashmap.Users)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(users)
}

func CreateTank(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		message := map[string]string{"message": "'username' field is required!"}
		errData, _ := json.Marshal(message)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errData)
		return
	}
	tank := hashmap.CreateTank(username, "bot", &mutex)
	tankData, err := json.Marshal(tank)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(tankData)
}
