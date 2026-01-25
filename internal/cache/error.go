package cache

import (
	"errors"

	"go.mattglei.ch/lcp/internal/api"
)

var ExpectedErrors = []error{
	api.ErrWarning,
	ErrSteamOwnedGamesEmpty,
}

// ErrSteamOwnedGamesEmpty is an error when a song returned from the Steam API fails to load and
// returns an empty list of owned games. This is an expected error that we should be able to handle.
var ErrSteamOwnedGamesEmpty = errors.New(
	"empty owned games",
)
