package auth

import (
	"fmt"
	"net/http"
	"strings"

	"pkg.mattglei.ch/lcp-2/internal/secrets"
)

func IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	validTokens := strings.Fields(secrets.ENV.ValidTokens)

	givenToken := r.Header.Get("Authorization")
	authorized := false
	for _, token := range validTokens {
		if givenToken == fmt.Sprintf("Bearer %s", token) {
			authorized = true
			break
		}
	}

	if !authorized {
		http.Error(w, "Invalid bearer auth token", http.StatusUnauthorized)
	}
	return authorized
}
