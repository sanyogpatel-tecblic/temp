package consuming

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
)

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageurl"`
}

func MakeGet() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Make an HTTP request to the API
		resp, err := http.Get("http://localhost:8020/items")
		if err != nil {
			http.Error(w, "Error making request to API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		// Unmarshal the response data
		var items []Item
		err = json.Unmarshal(body, &items)
		if err != nil {
			http.Error(w, "Error parsing response data", http.StatusInternalServerError)
			return
		}

		// Create an HTML template
		tmpl := template.Must(template.ParseFiles("./templates/home.page.tmpl"))

		// Pass the extracted data to the HTML template
		err = tmpl.Execute(w, items)
		if err != nil {
			http.Error(w, "Error rendering HTML template", http.StatusInternalServerError)
			return
		}
	})

}
