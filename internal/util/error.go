package util

import (
	"net/http"

	"go.mattglei.ch/timber"
)

func InternalServerError(w http.ResponseWriter, err error) {
	timber.Error(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
