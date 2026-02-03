package health

import (
	"encoding/json"
	"net/http"
	"time"

	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

var response []byte

func init() {
	data, err := json.Marshal(lcp.HealthStatus{
		Ok:      true,
		UpSince: time.Now(),
	})
	if err != nil {
		timber.Fatal(err, "failed to set health check response")
		return
	}
	response = data
}

func Endpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(response)
	if err != nil {
		util.InternalServerError(w, err)
		return
	}
}
