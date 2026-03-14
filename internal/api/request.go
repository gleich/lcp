package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"go.mattglei.ch/tlog"
)

// ErrWarning indicates that a non-critical error occurred during a request. Although the error
// prevents the cache from being updated, it is expected under certain transient conditions (for
// example, a 502 Gateway error) that are beyond our control. Such errors warrant only a warning
// rather than a full failure.
var ErrWarning = errors.New("non-critical error encountered during request")

// Request sends an HTTP request using the provided client with a 1-minute timeout and returns
// the response body as a byte slice. It handles common transient network errors—including timeouts,
// unexpected EOFs, and TCP connection resets—by logging warnings and returning a non-critical
// WarningError. Non-2xx HTTP responses are also treated as warnings.
func Request(task tlog.Task, client *http.Client, request *http.Request) ([]byte, error) {
	task, start := task.Extend("request").Start()
	var (
		url       = request.URL.String()
		path      = request.URL.Path
		resp, err = client.Do(request)
		logCtx    = []any{"path", path}
	)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			task.Warn("connection timed out", logCtx...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, context.DeadlineExceeded) {
			task.Warn("request timed out", logCtx...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			task.Warn("unexpected EOF", logCtx...)
			return []byte{}, ErrWarning
		}
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			task.Warn("tcp connection reset by peer", logCtx...)
			return []byte{}, ErrWarning
		}
		return []byte{}, fmt.Errorf("sending request to %s: %w", url, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		task.Warn(
			"non 200 status code returned",
			append([]any{"code", resp.StatusCode}, logCtx...)...)
		return []byte{}, ErrWarning
	} else if resp.StatusCode == http.StatusNoContent {
		return []byte{}, fmt.Errorf(
			"%d status no content returned when content is expected from %s",
			resp.StatusCode,
			url,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			task.Warn("reading body timed out", logCtx...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			task.Warn("unexpected EOF while ready body", logCtx...)
			return []byte{}, ErrWarning
		}
		return []byte{}, fmt.Errorf("reading response body for %s: %w", url, err)
	}

	task.InfoSince("request successful", start, "path", path)
	return body, nil
}

// RequestJSON sends an HTTP request using the provided client, reads the response body, and
// unmarshals the JSON into a value of type T. It relies on Request to perform the HTTP call. In
// case of a request failure or JSON parsing error, it logs the relevant details and returns the
// error.
func RequestJSON[T any](task tlog.Task, client *http.Client, request *http.Request) (T, error) {
	var data T

	body, err := Request(task, client, request)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		task.Debug("JSON error response", "body", string(body))
		return data, fmt.Errorf("parsing json from %s: %w", request.URL.String(), err)
	}

	return data, nil
}
