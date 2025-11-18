package cache

import (
	"errors"

	"go.mattglei.ch/lcp/internal/apis"
)

var ExpectedErrors = []error{
	apis.ErrWarning,
	ErrAppleMusicNoArtwork,
	ErrAppleMusicNoArtwork,
}

// ErrAppleMusicNoArtwork is an error when a song returned from the Apple Music API fails to load
// and returns a empty URL. This is an expected error that we should be able to handle.
var ErrAppleMusicNoArtwork = errors.New(
	"artwork failed to load for a given song",
)

// ErrSteamOwnedGamesEmpty is an error when a song returned from the Steam API fails to load and
// returns an empty list of owned games. This is an expected error that we should be able to handle.
var ErrSteamOwnedGamesEmpty = errors.New(
	"empty owned games",
)
