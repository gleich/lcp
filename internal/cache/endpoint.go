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
		err = fmt.Errorf("%w failed to write json data to request", err)
		timber.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Cache[T]) ServeStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		err := "failed to create flusher"
		timber.ErrorMsg(err)
		http.Error(w, err, http.StatusInternalServerError)
	}

	// telling client how long to wait before reconnecting
	_, _ = w.Write([]byte("retry: 5000\n\n"))
	flusher.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done(): // client disconnected / request canceled
			return
		case <-ticker.C:
			// NOTE: SSE events need a blank line after the data block
			fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("Event %d", i+1))
			flusher.Flush()
		}
	}

	// Optionally send a final comment, then exit
	_, _ = w.Write([]byte(": done\n\n"))
	flusher.Flush()
}
