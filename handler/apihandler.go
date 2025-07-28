package handler

import (
	"encoding/json"
	"net/http"

	"github.com/timsalokat/latios_proxy/db"
)

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		println("Received GET request")
		var routes []db.Route
		result := db.Client.Find(&routes)
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

	case http.MethodPost:
		println("Received POST request")
		var route db.Route
		if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
			println("Error decoding request body:", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		println("Decoded route:", route.Domain)
		result := db.Client.Create(&route)
		if result.Error != nil {
			println("Error creating route in DB:", result.Error.Error())
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
		println("Created route:", route.Domain)
		w.WriteHeader(http.StatusCreated)

	case http.MethodDelete:
		println("Received DELETE request")
		type DeleteBody struct {
			Domain string
		}

		var delBody DeleteBody
		if err := json.NewDecoder(r.Body).Decode(&delBody); err != nil {
			println("Error decoding request body:", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		println("Decoded route for deletion:", delBody.Domain)
		result := db.Client.Where("domain = ?", delBody.Domain).Delete(&db.Route{})
		if result.Error != nil {
			println("Error deleting route in DB:", result.Error.Error())
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
		delete(db.MemoryRoutes, delBody.Domain)
		println("Deleted route:", delBody.Domain)
		w.WriteHeader(http.StatusOK)

	default:
		println("Received unsupported method:", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
