package handler

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/timsalokat/latios_proxy/db"
)

var logPrefix = "proxy-logger - "

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	var route db.Route

	log.Println(logPrefix + "ProxyHandler called for " + r.Host)

	// Find route
	result := db.Client.Where("domain = ?", host).First(&route)
	if result.Error != nil {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	// Serve static file
	if route.IsStatic {
		http.StripPrefix("/", http.FileServer(http.Dir(route.TargetPath))).ServeHTTP(w, r)
		return
	}

	target, err := url.Parse(route.TargetPath)
	if err != nil {
		http.Error(w, "bad target", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1 // for streaming support

	proxy.ModifyResponse = func(resp *http.Response) error {
		// Add a header
		resp.Header.Set("X-Proxied-By", "Latios")

		// Log status code and headers
		log.Printf("%sProxied response: %d %s", logPrefix, resp.StatusCode, resp.Status)
		for k, v := range resp.Header {
			log.Printf("%sHeader: %s=%v", logPrefix, k, v)
		}

		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%sproxy error: %v", logPrefix, e)
		http.Error(w, "proxy error", http.StatusBadGateway)
	}

	log.Println(logPrefix + "Request proxied")
	proxy.ServeHTTP(w, r)
}
