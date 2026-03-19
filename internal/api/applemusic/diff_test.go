package applemusic

import (
	"testing"
	"time"

	"go.mattglei.ch/lcp/pkg/lcp"
)

func ptr[T any](v T) *T { return &v }

func TestDiffSongList(t *testing.T) {
	tests := []struct {
		name    string
		old     []lcp.AppleMusicSong
		new     []lcp.AppleMusicSong
		want    bool
		wantErr bool
	}{
		{
			name: "identical songs",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			want: false,
		},
		{
			name: "different track name",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{{Track: "Song B", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			want: true,
		},
		{
			name: "different artist",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist B", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			want: true,
		},
		{
			name: "different duration",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 2000, URL: "https://music.apple.com/a"}},
			want: true,
		},
		{
			name: "different url",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/b"}},
			want: true,
		},
		{
			name: "different list length",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "Artist A", DurationInMillis: 1000, URL: "https://music.apple.com/a"}},
			new:  []lcp.AppleMusicSong{},
			want: true,
		},
		{
			name: "empty lists are identical",
			old:  []lcp.AppleMusicSong{},
			new:  []lcp.AppleMusicSong{},
			want: false,
		},
		{
			name: "preview audio url changed",
			old:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u", PreviewAudioURL: ptr("https://preview.com/old.m4a")}},
			new:  []lcp.AppleMusicSong{{Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u", PreviewAudioURL: ptr("https://preview.com/new.m4a")}},
			want: true,
		},
		{
			name: "album art url same base different query params - not changed",
			old: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL: ptr("https://is1-ssl.mzstatic.com/image/thumb/abc/source/100x100bb.jpg?token=old"),
			}},
			new: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL: ptr("https://is1-ssl.mzstatic.com/image/thumb/abc/source/100x100bb.jpg?token=new"),
			}},
			want: false,
		},
		{
			name: "album art url base path changed",
			old: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL: ptr("https://is1-ssl.mzstatic.com/image/thumb/old/source/100x100bb.jpg?token=x"),
			}},
			new: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL: ptr("https://is1-ssl.mzstatic.com/image/thumb/new/source/100x100bb.jpg?token=x"),
			}},
			want: true,
		},
		{
			name: "expired permissions with different url - changed",
			old: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL:                   ptr("https://is1-ssl.mzstatic.com/image/thumb/old/100x100bb.jpg"),
				AlbumArtPermissionsExpiration: ptr(time.Now().Add(-10 * time.Minute)), // already expired
			}},
			new: []lcp.AppleMusicSong{{
				Track: "Song A", Artist: "A", DurationInMillis: 1000, URL: "u",
				AlbumArtURL: ptr("https://is1-ssl.mzstatic.com/image/thumb/new/100x100bb.jpg"),
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := diffSongList(tt.old, tt.new)
			if (err != nil) != tt.wantErr {
				t.Fatalf("diffSongList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("diffSongList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	song := func(track, artist string) lcp.AppleMusicSong {
		return lcp.AppleMusicSong{Track: track, Artist: artist, DurationInMillis: 1000, URL: "https://music.apple.com/x"}
	}
	playlist := func(name string, tracks ...lcp.AppleMusicSong) lcp.AppleMusicPlaylist {
		return lcp.AppleMusicPlaylist{
			Name:         name,
			Tracks:       tracks,
			LastModified: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			URL:          "https://music.apple.com/playlist",
			ID:           "pl-1",
		}
	}

	tests := []struct {
		name    string
		old     lcp.AppleMusicCache
		new     lcp.AppleMusicCache
		want    bool
		wantErr bool
	}{
		{
			name: "identical caches",
			old:  lcp.AppleMusicCache{RecentlyPlayed: []lcp.AppleMusicSong{song("A", "B")}, Playlists: []lcp.AppleMusicPlaylist{playlist("Mix", song("A", "B"))}},
			new:  lcp.AppleMusicCache{RecentlyPlayed: []lcp.AppleMusicSong{song("A", "B")}, Playlists: []lcp.AppleMusicPlaylist{playlist("Mix", song("A", "B"))}},
			want: false,
		},
		{
			name: "recently played changed",
			old:  lcp.AppleMusicCache{RecentlyPlayed: []lcp.AppleMusicSong{song("A", "B")}},
			new:  lcp.AppleMusicCache{RecentlyPlayed: []lcp.AppleMusicSong{song("C", "D")}},
			want: true,
		},
		{
			name: "playlist count changed",
			old:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Mix")}},
			new:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Mix"), playlist("Chill")}},
			want: true,
		},
		{
			name: "playlist name changed",
			old:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Mix")}},
			new:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Chill")}},
			want: true,
		},
		{
			name: "playlist track changed",
			old:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Mix", song("A", "B"))}},
			new:  lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{playlist("Mix", song("C", "D"))}},
			want: true,
		},
		{
			name: "playlist last modified changed",
			old: lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{
				{Name: "Mix", LastModified: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), ID: "pl-1"},
			}},
			new: lcp.AppleMusicCache{Playlists: []lcp.AppleMusicPlaylist{
				{Name: "Mix", LastModified: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), ID: "pl-1"},
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// c is not used by the diff function so nil is safe
			got, err := diff(nil, tt.old, tt.new)
			if (err != nil) != tt.wantErr {
				t.Fatalf("diff() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("diff() = %v, want %v", got, tt.want)
			}
		})
	}
}
