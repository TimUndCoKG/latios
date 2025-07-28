package handler

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/timsalokat/latios_proxy/db"
)

var prefix = "proxy-logger - "

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	var route db.Route

	log.Println(prefix + "ProxyHandler called for " + r.Host)

	result := db.Client.Where("domain = ?", host).First(&route)
	if result.Error != nil {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	if route.Static {
		http.StripPrefix("/", http.FileServer(http.Dir(route.StaticPath))).ServeHTTP(w, r)
		return
	}

	target, err := url.Parse(route.Target)
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
		log.Printf("%sProxied response: %d %s", prefix, resp.StatusCode, resp.Status)
		for k, v := range resp.Header {
			log.Printf("%sHeader: %s=%v", prefix, k, v)
		}

		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%sproxy error: %v", prefix, e)
		http.Error(w, "proxy error", http.StatusBadGateway)
	}

	log.Println(prefix + "Request proxied")
	proxy.ServeHTTP(w, r)
}
