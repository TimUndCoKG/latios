package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/timsalokat/latios_proxy/db"
)

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	var route db.Route
	result := db.DB.Where("domain = ?", host).First(&route)
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
		resp.Header.Set("X-Proxied-By", "GoProxy")
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("proxy error: %v", e)
		http.Error(w, "proxy error", http.StatusBadGateway)
	}

	// Support WebSocket by hijacking the connection (automatically handled by ReverseProxy)
	proxy.ServeHTTP(w, r)
}
