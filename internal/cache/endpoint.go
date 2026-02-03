package cache

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mattglei.ch/lcp/internal/auth"
	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/timber"
)

func (c *Cache[T]) Endpoints(mux *http.ServeMux) {
	mux.HandleFunc(fmt.Sprintf("GET /%s", c.instance), c.Serve)
	mux.HandleFunc(fmt.Sprintf("POST /%s/stream", c.instance), c.ServeStream)
}

func (c *Cache[T]) Serve(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAuthorized(w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	c.Mutex.RLock()
	data, err := c.MarshalResponse(c)
	if err != nil {
		err = fmt.Errorf("creating endpoint data: %w", err)
		util.InternalServerError(w, err)
		return
	}
	c.Mutex.RUnlock()
	_, err = w.Write([]byte(data))
	if err != nil {
		err = fmt.Errorf("writing data to request: %w", err)
		util.InternalServerError(w, err)
		return
	}
}

func (c *Cache[T]) ServeStream(w http.ResponseWriter, r *http.Request) {
	// we globally set the write timeout to 20 seconds, but for SSE we want to disable this
	if rc := http.NewResponseController(w); rc != nil {
		err := rc.SetWriteDeadline(time.Time{})
		if err != nil {
			err = fmt.Errorf("setting writing deadline to zero: %w", err)
			util.InternalServerError(w, err)
			return
		}
	}

	origin := r.Header.Get("Origin")
	switch origin {
	// case "http://localhost:5173", "https://mattglei.ch", "https://lcp.mattglei.ch":
	case "https://mattglei.ch", "https://lcp.mattglei.ch":
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	default:
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		err := errors.New("creating flusher")
		util.InternalServerError(w, err)
		return
	}

	// telling client how long to wait before reconnecting
	_, err := w.Write([]byte("retry: 5000\n\n"))
	if err != nil {
		err = fmt.Errorf("writing retry information: %w", err)
		util.InternalServerError(w, err)
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
				err = fmt.Errorf("writing retry information: %w", err)
				util.InternalServerError(w, err)
				return
			}
			flusher.Flush()
		case frame, ok := <-channel:
			if !ok {
				timber.ErrorMsg("failed to get data from channel for update")
			}
			_, err = fmt.Fprintf(w, "event: message\ndata: %s\n\n", frame)
			if err != nil {
				err = fmt.Errorf("writing data: %w", err)
				util.InternalServerError(w, err)
				return
			}
			flusher.Flush()
		}
	}
}
