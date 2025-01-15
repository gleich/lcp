package apis

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"pkg.mattglei.ch/timber"
)

var WarningError = errors.New("Warning error when trying to make request. Ignore error.")

// sends a given http.Request and will unmarshal the JSON from the response body and return that as the given type.
func SendRequest[T any](req *http.Request) (T, error) {
	var zeroValue T // to be used as "nil" when returning errors
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		timber.Error(err, "sending request failed")
		return zeroValue, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			timber.Warning("Unexpected EOF from", req.URL.String())
			return zeroValue, WarningError
		}
		var netErr *net.OpError
		if errors.As(err, &netErr) && netErr.Err != nil &&
			netErr.Err.Error() == "connection reset by peer" {
			timber.Warning("TCP connection reset by peer from", req.URL.String())
			return zeroValue, WarningError
		}

		timber.Error(err, "reading response body failed")
		return zeroValue, err
	}
	if resp.StatusCode != http.StatusOK {
		timber.Warning(resp.StatusCode, "returned from", req.URL.String())
		return zeroValue, WarningError
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		timber.Error(err, "failed to parse json")
		timber.Debug(string(body))
		return zeroValue, err
	}

	return data, nil
}
