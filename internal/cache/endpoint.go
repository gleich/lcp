package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mattglei.ch/timber"
)

type Response[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
}

func (c *Cache[T]) Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c.Mutex.RLock()
	err := json.NewEncoder(w).Encode(Response[T]{Data: c.Data, Updated: c.Updated})
	c.Mutex.RUnlock()
	if err != nil {
		responseError(w, err, "failed to write json data to request")
	}
}

func (c *Cache[T]) ServeStream(w http.ResponseWriter, r *http.Request) {
	// we globally set the write timeout to 20 seconds, but for SSE we want to disable this
	if rc := http.NewResponseController(w); rc != nil {
		err := rc.SetWriteDeadline(time.Time{})
		if err != nil {
			responseError(w, err, "failed to set write deadline to zero")
			return
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		msg := "failed to create flusher"
		timber.ErrorMsg(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// telling client how long to wait before reconnecting
	_, err := w.Write([]byte("retry: 5000\n\n"))
	if err != nil {
		responseError(w, err, "failed to write retry information")
		return
	}
	flusher.Flush()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done(): // client disconnected / request canceled
			return
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": heartbeat\n\n")
			if err != nil {
				responseError(w, err, "failed to write heartbeat")
			}
			flusher.Flush()
		}
	}
}

func responseError(w http.ResponseWriter, err error, msg string) {
	err = fmt.Errorf("%w %s", err, msg)
	timber.Error(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
