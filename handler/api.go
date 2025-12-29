// In handler/api.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/timsalokat/latios_proxy/db"
	"github.com/timsalokat/latios_proxy/middleware"
	"golang.org/x/time/rate"
)

// RegisterApiHandlers registers all the /latios-api endpoints.
func RegisterApiHandlers(router *http.ServeMux) {
	log.Println("[ROUTER] Setting up API routes...")

	loginLimiter := middleware.NewIPRateLimiter(rate.Every(time.Minute/5), 5)
	apiLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20)

	// Define your API routes here
	apiRoutes := map[string]http.Handler{
		"/latios-api/health": http.HandlerFunc(HealthCheckHandler),
		"/latios-api/login":  loginLimiter.RateLimitMiddleware(http.HandlerFunc(LoginHandler)),
		"/latios-api/routes": apiLimiter.RateLimitMiddleware(http.HandlerFunc(RoutesApiHandler)),
		"/latios-api/stats":  apiLimiter.RateLimitMiddleware(http.HandlerFunc(StatsApiHandler)),
		"/latios-api/logs":   apiLimiter.RateLimitMiddleware(http.HandlerFunc(LogsApiHandler)),
	}

	for path, handler := range apiRoutes {
		log.Printf("[ROUTER] Adding API path: %s\n", path)
		router.Handle(path, handler)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RoutesApiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// Should retrieve all routes
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

	// Create a new route
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

		db.AddRouteToCache(route)

		println("Created route:", route.Domain)
		w.WriteHeader(http.StatusCreated)

	// Delete route
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

		db.DeleteRouteFromCache(delBody.Domain)

		println("Deleted route:", delBody.Domain)
		w.WriteHeader(http.StatusOK)

	default:
		println("Received unsupported method:", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func StatsApiHandler(w http.ResponseWriter, r *http.Request) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var stats struct {
		TotalRequests         int64   `json:"total_requests"`
		TotalRequestsResolved int64   `json:"total_requests_resolved"`
		ServerErrorCount      int64   `json:"server_error_count"`
		ClientErrorCount      int64   `json:"client_error_count"`
		NotFoundCount         int64   `json:"not_found_count"`
		AvgLatency            float64 `json:"avg_latency_ms"`
	}

	err := db.Client.Model(&db.RequestLog{}).
		Where("timestamp > ?", thirtyDaysAgo).
		Select(`
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as total_requests_resolved,
			COUNT(CASE WHEN status_code >= 500 THEN 1 END) as server_error_count,
			COUNT(CASE WHEN status_code >= 400 AND status_code < 500 AND status_code != 404 THEN 1 END) as client_error_count,
			COUNT(CASE WHEN status_code = 404 THEN 1 END) as not_found_count,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms
		`).
		Scan(&stats).Error

	if err != nil {
		http.Error(w, "Failed to calculate stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func LogsApiHandler(w http.ResponseWriter, r *http.Request) {
	limit := 100

	pageStr := r.URL.Query().Get("page")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	var logs []db.RequestLog

	result := db.Client.Order("timestamp desc").Limit(limit).Offset(offset).Find(&logs)

	if result.Error != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(logs)
}
