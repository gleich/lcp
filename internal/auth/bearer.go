package auth

import (
	"fmt"
	"net/http"

	"pkg.mattglei.ch/lcp-2/internal/secrets"
)

func IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", secrets.ENV.ValidToken) {
		http.Error(w, "Invalid bearer auth token", http.StatusUnauthorized)
		return false
	}
	return true
}
