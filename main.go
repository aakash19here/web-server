package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

/*
	Mux - Request Multiplexer. Control traffic to specific endpoint or handler functions
	ResponseWriter - Contructing response for the client
	Request - Body , Header etc of the req
	Fprintf - Allows us to writer anything to a ResponseWriter
	To make the creation thread safe and to avoid race conditions we use mutex
*/

type User struct {
	Name string `json:"name"`
}

var cacheMutex sync.RWMutex

var userCache = make(map[int]User)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello World")
	})

	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUserById)
	mux.HandleFunc("GET /users", getUsers)
	mux.HandleFunc("DELETE /users/{id}", deleteUserById)

	fmt.Println("Server listening to :8080")

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Println("server error:", err)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(&userCache)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getUserById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	user, ok := userCache[id]

	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func deleteUserById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, ok := userCache[id]

	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(userCache, id)

	w.WriteHeader(http.StatusNoContent)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	userCache[len(userCache)+1] = user

	w.WriteHeader(http.StatusNoContent)
}
