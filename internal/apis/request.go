package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.mattglei.ch/timber"
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
func Request(logPrefix string, client *http.Client, req *http.Request) ([]byte, error) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Minute)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			timber.Warning(logPrefix, "connection timed out for", req.URL.Path)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, context.DeadlineExceeded) {
			timber.Warning(logPrefix, "request timed out for", req.URL.Path)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			timber.Warning(logPrefix, "unexpected EOF from", req.URL.Path)
			return []byte{}, ErrWarning
		}
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			timber.Warning(logPrefix, "tcp connection reset by peer from", req.URL.Path)
			return []byte{}, ErrWarning
		}
		return []byte{}, fmt.Errorf("%w sending request to %s failed", err, req.URL.String())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("%w reading response body failed", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		timber.Warning(
			logPrefix,
			resp.StatusCode,
			fmt.Sprintf("(%s)", strings.ToLower(http.StatusText(resp.StatusCode))),
			"from",
			req.URL.String(),
		)
		return []byte{}, ErrWarning
	}

	err = resp.Body.Close()
	if err != nil {
		return []byte{}, fmt.Errorf("%w failed to close response body", err)
	}

	return body, nil
}

// RequestJSON sends an HTTP request using the provided client, reads the response body, and
// unmarshals the JSON into a value of type T. It relies on Request to perform the HTTP call. In
// case of a request failure or JSON parsing error, it logs the relevant details and returns the
// error.
func RequestJSON[T any](logPrefix string, client *http.Client, req *http.Request) (T, error) {
	var data T

	body, err := Request(logPrefix, client, req)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		timber.Debug(string(body))
		return data, fmt.Errorf("%w failed to parse json", err)
	}

	return data, nil
}
