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

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageurl"`
}

const port = ":1050"

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres:root@localhost/temp?sslmode=disable")
	fmt.Println("connected to database!")
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	// r.HandleFunc("/items", GetAllItems(db)).Methods("GET")
	r.HandleFunc("/items", createItemHandler(db)).Methods("POST")
	r.HandleFunc("/items", GetItems(db)).Methods("GET")

	// Set up the file server to serve files from the 'uploads' directory
	fileServer := http.FileServer(http.Dir("./uploads"))

	//r.PathPrefix("/uploads/") this means that any request that begins with /uploads/ will be handled by this route
	//http.StripPrefix basically removes the words "http://localhost:8020/uploads/imagename.jpg"
	//basically it will remove "http://localhost:8020" so that remaining path will be sent to fileserver
	//so that correct file is being served

	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", fileServer))

	log.Fatal(http.ListenAndServe(port, r))
}

func createItemHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		r.ParseMultipartForm(10 << 20)

		name := r.FormValue("name")
		description := r.FormValue("description")
		image, handler, err := r.FormFile("image")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error getting image file: %v", err)
			return
		}
		defer image.Close()
		allowedExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		filename := handler.Filename
		ext := filepath.Ext(filename)
		if !allowedExtensions[ext] {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid file format. Only JPG, JPEG, PNG files are allowed.")
			return
		}

		tempFile, err := os.CreateTemp("", "upload-*"+ext)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error creating temporary file: %v", err)
			return
		}

		defer tempFile.Close()
		io.Copy(tempFile, image)

		imageURL := tempFile.Name()
		filepath := fmt.Sprintf("./uploads/%s", handler.Filename)
		err = os.Rename(tempFile.Name(), filepath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error moving file to uploads directory: %v", err)
			return
		}

		imageURL = filepath

		//the newPath variable is created to hold the file path of the uploaded file
		// in the "uploads" directory. filepath.Join is used to join the "uploads" directory and
		// the file name to create the complete file path.
		// newPath := filepath.Join("uploads", filename)
		// err = os.Rename(tempFile.Name(), newPath)
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	fmt.Fprintf(w, "Error moving file to uploads directory: %v", err)
		// 	return
		// }
		// portwithoutcolun := strings.Replace(port, ":", "", 1)
		// imageURL := fmt.Sprintf("http://localhost:"+portwithoutcolun+"/uploads/%s", filename)
		// imageURL = fmt.Sprintf("http://%s/uploads/%s", portwithoutcolun, filename)
		if name == "" {
			apierror := APIError{
				Code:    http.StatusBadRequest,
				Message: "Product name is required",
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		stmt, err := db.Prepare(`INSERT INTO items (name, description, imageurl) VALUES ($1, $2, $3)`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w, "Error preparing SQL statement: %v", err)
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(name, description, imageURL)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error executing SQL statement: %v", err)
			return
		}

		id, _ := result.LastInsertId()

		item := Item{
			ID:          int(id),
			Name:        name,
			Description: description,
			ImageURL:    imageURL,
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Item created successfully")
		json.NewEncoder(w).Encode(item)
	}
}

func GetItems(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var items []Item
		rows, err := db.Query("SELECT id, name, description,imageurl FROM items")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			var item Item
			row_var := rows.Scan(&item.ID, &item.Name, &item.Description, &item.ImageURL)
			if row_var != nil {
				log.Fatal(row_var)
			}
			items = append(items, item)
		}
		json.NewEncoder(w).Encode(items)
	}
}
