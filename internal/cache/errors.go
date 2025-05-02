package cache

import "errors"

// AppleMusicUnknownSongError is an error that occurs in the Apple Music API in which a given song
// is basically empty. This causes the cache to update with invalid data for a song that doesn't
// even exist. To prevent this problem the song is checked on each update cycle and is logged as a
// warning if encountered. It is an expected error that we can do nothing about.
var AppleMusicUnknownSongError = errors.New(
	"non-critical error with unknown song appearing in data",
)
