package metrics

import "testing"

func TestNormalizeLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"", "xx"},
		{"  ", "xx"},
		{"zh-CN", "zh"},
		{"zh-Hans", "zh"},
		{"sv-SE", "sv"},
		{"de-DE", "de"},
		{"de-at", "de"},
		{"en-US", "en"},
		{"eng-us", "en"},
		{"DE", "de"},
		{"Fr", "fr"},
		{"en", "en"},
		{"arabic", "ar"},
		{"cat", "ca"},
		{"cz", "cs"},
		{"eus", "eu"},
		{"fil", "tl"},
		{"jp", "ja"},
		{"jp-JP", "ja"},
		{"in", "id"},
		{"und", "xx"},
		{"emoji", "xx"},
		{"both", "xx"},
		{"wuu", "zh"},
		{"yue", "zh"},
		{"Angika", "xx"},
		{"tlh", "xx"},
		{"not-a-real-language-tag", "xx"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeLanguage(tt.in); got != tt.want {
				t.Fatalf("NormalizeLanguage(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
