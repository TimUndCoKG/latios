package handlers

import (
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/routes", HandleRoutes)
}

func HandleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		println("Received POST request")
		var route db.Route
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			println("Error decoding request body:", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		println("Decoded route:", route.Domain)
		result := db.DB.Create(&route)
		if result.Error != nil {
			println("Error creating route in DB:", result.Error.Error())
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
		println("Created route:", route.Domain)
		w.WriteHeader(http.StatusCreated)
	case http.MethodGet:
		println("Received GET request")
		var routes []db.Route
		result := db.DB.Find(&routes)
		if result.Error != nil {
			println("Error fetching routes from DB:", result.Error.Error())
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
		println("Fetched routes count:", len(routes))
		if err := json.NewEncoder(w).Encode(routes); err != nil {
			println("Error encoding response:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		println("Received unsupported method:", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
