package util

import (
	"net/http"

	"go.mattglei.ch/lcp/internal/tasks"
)

func InternalServerError(w http.ResponseWriter, err error) {
	tasks.Endpoint.Error(err, err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
