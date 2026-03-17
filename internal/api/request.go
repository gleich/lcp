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
func Request(client *http.Client, request *http.Request, cacheLogAttr timber.Attr) ([]byte, error) {
	var (
		url           = request.URL.String()
		start         = time.Now()
		resp, err     = client.Do(request)
		logAttributes = []timber.Attr{cacheLogAttr, timber.A("url", url)}
	)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			timber.Warning("connection timed out for", logAttributes...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, context.DeadlineExceeded) {
			timber.Warning("request timed out for", logAttributes...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			timber.Warning("unexpected EOF from", logAttributes...)
			return []byte{}, ErrWarning
		}
		if strings.Contains(err.Error(), "read: connection reset by peer") {
			timber.Warning("tcp connection reset by peer from", logAttributes...)
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
		timber.Warning(
			"non-200 status code",
			append(
				logAttributes,
				timber.A("code", resp.StatusCode),
			)...,
		)
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
			timber.Warning("reading body timed out for", logAttributes...)
			return []byte{}, ErrWarning
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			timber.Warning("unexpected EOF while reading body from", logAttributes...)
			return []byte{}, ErrWarning
		}
		return []byte{}, fmt.Errorf("reading response body for %s: %w", url, err)
	}

	timber.InfoSince(start, "made request", logAttributes...)
	return body, nil
}

// RequestJSON sends an HTTP request using the provided client, reads the response body, and
// unmarshals the JSON into a value of type T. It relies on Request to perform the HTTP call. In
// case of a request failure or JSON parsing error, it logs the relevant details and returns the
// error.
func RequestJSON[T any](
	client *http.Client,
	request *http.Request,
	cacheLogAttr timber.Attr,
) (T, error) {
	var data T

	body, err := Request(client, request, cacheLogAttr)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		timber.Debug(string(body))
		return data, fmt.Errorf("parsing json from %s: %w", request.URL.String(), err)
	}

	return data, nil
}
