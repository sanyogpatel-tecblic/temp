package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func main() {
	// Set up database connection
	db, err := sql.Open("postgres", "user=postgres password=root dbname=temp sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set up Gin router
	router := gin.Default()

	// Handle POST request to create a new product
	router.POST("/products", func(c *gin.Context) {
		// Get form data from request PostForm is used to get form data
		name := c.PostForm("name")
		description := c.PostForm("description")
		//FORM_FILE is used to get image that is uploaded by us
		imageFile, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save image to disk
		imagePath := "uploads/" + imageFile.Filename
		err = c.SaveUploadedFile(imageFile, imagePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Insert new product into database
		var id int
		err = db.QueryRow("INSERT INTO items (name, description, imageurl) VALUES ($1, $2, $3) RETURNING id", name, description, imagePath).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return response with new product data
		product := Product{id, name, description, imagePath}
		c.JSON(http.StatusCreated, product)
	})

	router.GET("/products", func(c *gin.Context) {
		// Query all products from database
		rows, err := db.Query("SELECT id, name, description, imageurl FROM items")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		// Create list of products with image paths
		var products []Product
		for rows.Next() {
			var id int
			var name, description, image string
			err = rows.Scan(&id, &name, &description, &image)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			product := Product{id, name, description, image}
			products = append(products, product)
		}

		// Serve list of products as JSON response
		c.JSON(http.StatusOK, products)
	})

	// Serve static files (uploaded images)
	router.Static("/uploads", "./uploads")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8010"
	}
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal(err)
	}
}
