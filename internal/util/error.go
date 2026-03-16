package util

import (
	"net/http"

	"go.mattglei.ch/timber"
)

func InternalServerError(w http.ResponseWriter, err error, cacheLogAttr timber.Attr, msg string) {
	timber.Error(err, msg, cacheLogAttr)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
