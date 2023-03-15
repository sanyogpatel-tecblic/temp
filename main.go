package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/sanyogpatel-tecblic/API-simple/go/pkg/mod/github.com/gorilla/mux@v1.8.0"
)

type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/items", createItemHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", r))
}
