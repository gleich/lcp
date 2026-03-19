package util

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "strips query params",
			input: "https://example.com/image.jpg?token=abc&expires=123",
			want:  "https://example.com/image.jpg",
		},
		{
			name:  "strips fragment",
			input: "https://example.com/page#section",
			want:  "https://example.com/page",
		},
		{
			name:  "strips both query and fragment",
			input: "https://example.com/path?q=1#frag",
			want:  "https://example.com/path",
		},
		{
			name:  "no query or fragment unchanged",
			input: "https://example.com/image.jpg",
			want:  "https://example.com/image.jpg",
		},
		{
			name:  "apple music style url with permissions",
			input: "https://is1-ssl.mzstatic.com/image/thumb/abc123/source/100x100bb.jpg?a=1&b=2",
			want:  "https://is1-ssl.mzstatic.com/image/thumb/abc123/source/100x100bb.jpg",
		},
		{
			name:    "invalid url returns error",
			input:   "://bad-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.String() != tt.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}
