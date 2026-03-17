package cache

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mattglei.ch/lcp/internal/auth"
	"go.mattglei.ch/lcp/internal/util"
)

func (c *Cache[T]) Endpoints(mux *http.ServeMux) {
	mux.Handle(fmt.Sprintf("GET /%s", c.instance), c)
	mux.HandleFunc(fmt.Sprintf("POST /%s/stream", c.instance), c.ServeStream)
}

func (c *Cache[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	c.Mutex.RLock()
	data, err := c.MarshalResponse(c)
	if err != nil {
		util.InternalServerError(w, err, c.LogAttr, "failed to marshal endpoint data")
		return
	}
	c.Mutex.RUnlock()
	_, err = w.Write([]byte(data))
	if err != nil {
		util.InternalServerError(w, err, c.LogAttr, "failed to write data to request")
		return
	}
}

func (c *Cache[T]) ServeStream(w http.ResponseWriter, r *http.Request) {
	auth.SetCorsPolicy(w, r)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		util.InternalServerError(
			w,
			errors.New("creating flusher"),
			c.LogAttr,
			"failed to create flusher",
		)
		return
	}

	// telling client how long to wait before reconnecting
	_, err := w.Write([]byte("retry: 5000\n\n"))
	if err != nil {
		util.InternalServerError(w, err, c.LogAttr, "failed to write reconnection time")
		return
	}
	flusher.Flush()

	// add connection to connections pool
	channel := make(chan string, 8)
	c.connectionsMutex.Lock()
	c.connections[channel] = struct{}{}
	c.connectionsMutex.Unlock()

	// remove connection from connection pool when done
	defer func() {
		c.connectionsMutex.Lock()
		delete(c.connections, channel)
		c.connectionsMutex.Unlock()
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": heartbeat\n\n")
			if err != nil {
				util.InternalServerError(w, err, c.LogAttr, "failed to write heartbeat")
				return
			}
			flusher.Flush()
		case frame, ok := <-channel:
			if !ok {
				return
			}
			_, err = fmt.Fprintf(w, "event: message\ndata: %s\n\n", frame)
			if err != nil {
				util.InternalServerError(w, err, c.LogAttr, "writing data")
				return
			}
			flusher.Flush()
		}
	}
}
