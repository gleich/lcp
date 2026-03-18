package auth

import "net/http"

func SetCorsPolicy(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	switch origin {
	// case "http://localhost:5173", "https://mattglei.ch", "https://lcp.mattglei.ch":
	case "https://mattglei.ch", "https://lcp.mattglei.ch":
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Vary", "Origin")
	default:
	}
}

func HandlePreflight(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		SetCorsPolicy(w, r)
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}
