package cache

import "errors"

// ErrAppleMusicNoArtwork is an error when a song returned from the Apple Music API fails to load
// and returns a empty URL. This is an expected error that we should be able to handle.
var ErrAppleMusicNoArtwork = errors.New(
	"artwork failed to load for a given song",
)
