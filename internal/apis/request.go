package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.mattglei.ch/timber"
)

// WarningError indicates that a non-critical error occurred during a request. Although the error
// prevents the cache from being updated, it is expected under certain transient conditions (for
// example, a 502 Gateway error) that are beyond our control. Such errors warrant only a warning
// rather than a full failure.
var WarningError = errors.New("non-critical error encountered during request")

// Sends a given http.Request and unmarshal the JSON from the response body and return that as
// the given type. Handles common errors like 502, unexpected EOFs, and timeouts.
func Request[T any](logPrefix string, client *http.Client, req *http.Request) (T, error) {
	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Minute)
	defer cancel()
	req = req.WithContext(ctx)

	var zeroValue T // to be used as "nil" when returning errors
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timber.Warning(logPrefix, "request timed out for", req.URL.String())
			return zeroValue, WarningError
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			timber.Warning(logPrefix, "unexpected EOF from", req.URL.String())
			return zeroValue, WarningError
		}
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			timber.Warning(logPrefix, "tcp connection reset by peer from", req.URL.String())
			return zeroValue, WarningError
		}
		return zeroValue, fmt.Errorf("%w sending request failed", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroValue, fmt.Errorf("%w reading response body failed", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		timber.Warning(logPrefix, resp.StatusCode, "->", req.URL.Path)
		return zeroValue, WarningError
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		timber.Debug(string(body))
		return zeroValue, fmt.Errorf("%w failed to parse json", err)
	}

	return data, nil
}
