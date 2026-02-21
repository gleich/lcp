package util

import "net/http"

func SetCorsPolicy(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	switch origin {
	// case "http://localhost:5173", "https://mattglei.ch", "https://lcp.mattglei.ch":
	case "https://mattglei.ch", "https://lcp.mattglei.ch":
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	default:
	}
}
