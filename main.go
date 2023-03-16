package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Items struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageurl"`
}

func main() {

	// Open the SQLite database
	db, err := sql.Open("postgres", "postgresql://postgres:root@localhost/temp?sslmode=disable")
	fmt.Println("connected to database!")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new router using the Gorilla/mux package
	r := mux.NewRouter()

	// Define the handler function for the POST /products endpoint
	r.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form data from the request
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get the product data from the form data
		name := r.FormValue("name")
		description := r.FormValue("description")
		imageFile, _, err := r.FormFile("image")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer imageFile.Close()

		// Generate a unique filename for the image
		rand.Seed(time.Now().UnixNano())
		randString := strconv.FormatInt(rand.Int63(), 16)
		fileName := time.Now().Format("2006-01-02T15-04-05.999999999") + "-" + randString

		// Create a new file to save the image
		outFile, err := os.Create(filepath.Join("uploads", fileName))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		// Copy the image data to the new file
		_, err = io.Copy(outFile, imageFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Save the product data to the database
		result, err := db.Exec("INSERT INTO items (name, description, imageurl) VALUES ($1, $2, $3)", name, description, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the ID of the newly inserted product
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the response with the new product data
		items := &Items{
			ID:          int(id),
			Name:        name,
			Description: description,
			ImageURL:    fmt.Sprintf("/uploads/%s", fileName),
		}
		json.NewEncoder(w).Encode(items)
	}).Methods("POST")

	// Serve static files from the uploads directory
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Start the HTTP server
	log.Println("Server listening on :8080...")
	http.ListenAndServe(":8080", r)
}
