package lcp

import "time"

type HealthStatus struct {
	Ok      bool      `json:"ok"`
	UpSince time.Time `json:"up_since"`
}
