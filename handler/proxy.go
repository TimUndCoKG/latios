package handler

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/timsalokat/latios_proxy/db"
)

//go:embed templates/404.html
var notFoundHTML string
var notFoundTemplate = template.Must(template.New("404").Parse(notFoundHTML))

func serveNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := struct {
		Host string
	}{
		Host: r.Host,
	}
	notFoundTemplate.Execute(w, data)
}

var logPrefix = "[PROXY] - "

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	var route db.Route

	log.Println(logPrefix + "ProxyHandler called for " + r.Host)

	// Find route
	result := db.Client.Where("domain = ?", host).First(&route)
	if result.Error != nil {
		serveNotFound(w, r)
		return
	}

	// Serve static file
	if route.IsStatic {
		// log.Println("Serving static route with path: " + route.TargetPath)
		http.StripPrefix("/", http.FileServer(http.Dir(route.TargetPath))).ServeHTTP(w, r)
		return
	}

	target, err := url.Parse(route.TargetPath)
	if err != nil {
		http.Error(w, "bad target", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Header setup
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		req.Host = r.Host
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)

		if r.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		} else {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
	}
	proxy.FlushInterval = -1 // for streaming and websocket support

	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Proxied-By", "Latios")

		// Log status code and headers
		// log.Printf("%sProxied response: %d %s", logPrefix, resp.StatusCode, resp.Status)
		// for k, v := range resp.Header {
		// 	log.Printf("%sHeader: %s=%v", logPrefix, k, v)
		// }

		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("%sproxy error: %v", logPrefix, e)
		http.Error(w, "proxy error", http.StatusBadGateway)
	}

	// log.Println(logPrefix + "Request proxied")
	proxy.ServeHTTP(w, r)
}
