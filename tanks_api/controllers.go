package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
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

func GetMapUsers(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	users, err := json.Marshal(hashmap.Users)
	if err != nil {
		log.Printf("[API] message: %v", err)
	}
	w.Write(users)
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
