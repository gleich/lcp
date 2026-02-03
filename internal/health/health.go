package health

import (
	"encoding/json"
	"net/http"
	"time"

	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/lcp/pkg/lcp"
)

var upSince = time.Now()

func Endpoint(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(lcp.HealthStatus{
		Ok:      true,
		UpSince: upSince,
	})
	if err != nil {
		util.InternalServerError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		util.InternalServerError(w, err)
		return
	}
}
