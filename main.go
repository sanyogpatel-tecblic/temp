package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
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

	if err != nil {
		log.Fatal(err)
	}

	// Create a new router using the Gorilla/mux package
	r := mux.NewRouter()

	// Define the handler function for the POST /products endpoint
	// Define the handler function for the POST /products endpoint
	// Define the handler function for the POST /products endpoint
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
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

		// Generate a unique filename for the image based on the current time and file extension
		ext := filepath.Ext(imageFile.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		// Create a new file to save the image
		outFile, err := os.Create(filepath.Join("uploads", filename))
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
		result, err := db.Exec("INSERT INTO items (name, description, imageurl) VALUES ($1, $2, $3)", name, description, filename)
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
		product := &Product{
			ID:          int(id),
			Name:        name,
			Description: description,
			ImageURL:    fmt.Sprintf("/uploads/%s", filename),
		}
		json.NewEncoder(w).Encode(product)
	}).Methods("POST")

	// Serve static files from the uploads directory
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Start the HTTP server
	log.Println("Listening on :7080...")
	err = http.ListenAndServe(":7080", r)
	if err != nil {
		log.Fatal(err)
	}
}
