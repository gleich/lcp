package lcp

type HealthStatus struct {
	Ok bool `json:"ok"`
}

func FetchHealthStatus(client *Client) (HealthStatus, error) {
	resp, err := fetch[HealthStatus](client, "health")
	if err != nil {
		return HealthStatus{}, err
	}
	return resp, nil
}
