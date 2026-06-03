package metrics

import "strings"

const unknownLanguage = "xx"

// languageAliases maps non-standard language tags (after lowercasing and
// optional region stripping) to ISO 639-1-style two-letter codes.
var languageAliases = map[string]string{
	"arabic": "ar",
	"ars":    "ar",
	"ast":    "es",
	"astu":   "es",
	"bas":    "eu",
	"cat":    "ca",
	"cz":     "cs",
	"eng":    "en",
	"eus":    "eu",
	"fil":    "tl",
	"gal":    "gl",
	"gsw":    "de",
	"in":     "id",
	"jp":     "ja",
	"sco":    "en",
	"wuu":    "zh",
	"yue":    "zh",
}

// NormalizeLanguage collapses Bluesky language tags to a stable two-letter
// label for metrics. Region subtags (e.g. zh-CN, sv-SE) are dropped; longer
// non-standard names are mapped via languageAliases; anything else unknown
// becomes unknownLanguage.
func NormalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return unknownLanguage
	}

	if i := strings.IndexByte(lang, '-'); i >= 0 {
		lang = lang[:i]
	}

	if code, ok := languageAliases[lang]; ok {
		return code
	}
	if len(lang) <= 2 {
		return lang
	}
	return unknownLanguage
}
