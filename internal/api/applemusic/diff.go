package applemusic

import (
	"fmt"
	"time"

	"go.mattglei.ch/lcp/internal/cache"
	"go.mattglei.ch/lcp/internal/util"
	"go.mattglei.ch/lcp/pkg/lcp"
	"go.mattglei.ch/timber"
)

// We need a custom diff check function due to the fact that the apple music image service returns
// images with the permissions attached as url params. These change every time we make a request.
// This was causing false updates, so this will check for differences in everything,
// normalize the album art URL for the url without the permissions, and check to see if the
// permissions are expired.
func diff(c *cache.Cache[lcp.AppleMusicCache], old, new lcp.AppleMusicCache) (bool, error) {
	different, err := diffSongList(old.RecentlyPlayed, new.RecentlyPlayed)
	if err != nil {
		return false, fmt.Errorf("recently played diff check: %w", err)
	}
	if different {
		return true, nil
	}

	if len(old.Playlists) != len(new.Playlists) {
		return true, nil
	}
	for i, oldPlaylist := range old.Playlists {
		newPlaylist := new.Playlists[i]
		different, err = diffSongList(oldPlaylist.Tracks, newPlaylist.Tracks)
		if err != nil {
			return false, fmt.Errorf("checking track difference for %s: %w", oldPlaylist.Name, err)
		}
		if different {
			return true, nil
		}
		if oldPlaylist.Name != newPlaylist.Name ||
			oldPlaylist.LastModified != newPlaylist.LastModified ||
			oldPlaylist.URL != newPlaylist.URL ||
			oldPlaylist.SpotifyID != newPlaylist.SpotifyID ||
			oldPlaylist.ID != newPlaylist.ID {
			return true, nil
		}
	}

	return false, nil
}

func diffSongList(oldSongs, newSongs []lcp.AppleMusicSong) (bool, error) {
	if len(oldSongs) != len(newSongs) {
		return true, nil
	}

	for i, old := range oldSongs {
		new := newSongs[i]

		if old.Track != new.Track || old.Artist != new.Artist ||
			old.DurationInMillis != new.DurationInMillis ||
			old.URL != new.URL ||
			(old.PreviewAudioURL != nil && new.PreviewAudioURL != nil && *old.PreviewAudioURL != *new.PreviewAudioURL) {
			return true, nil
		}

		if old.AlbumArtURL != nil && new.AlbumArtURL != nil {
			if old.AlbumArtPermissionsExpiration != nil &&
				time.Now().After(old.AlbumArtPermissionsExpiration.Add(-1*time.Minute)) {
				var (
					oldURL = old.AlbumArtURL
					newURL = new.AlbumArtURL
				)
				if *oldURL == *newURL {
					timber.Warning("album art did not update even though it expired")
					timber.Warning("old url:", *oldURL)
					timber.Warning("new url:", *newURL)
				} else {
					return true, nil
				}
			}

			oldArtURL, err := util.NormalizeURL(*old.AlbumArtURL)
			if err != nil {
				return false, fmt.Errorf(
					"failed to normalize url for %s: %w",
					*old.AlbumArtURL,
					err,
				)
			}
			newArtURL, err := util.NormalizeURL(*new.AlbumArtURL)
			if err != nil {
				return false, fmt.Errorf(
					"failed to normalize url for %s: %w",
					*new.AlbumArtURL,
					err,
				)
			}

			if oldArtURL.String() != newArtURL.String() {
				return true, nil
			}
		}
	}

	return false, nil
}
