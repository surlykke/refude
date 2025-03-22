package translate

import (
	"os"
	"regexp"
)

var lcMessage string
var lcMessagePattern = regexp.MustCompile(`([^_.@]+)(_[^.@]+)?(\.[^@]+)?(@.*)?`) // 1: language, 2: country, 3: encoding, 4: modifier
var lcMatchers []string
var lang string

func init() {
	if os.Getenv("LC_ALL") != "" {
		lcMessage = os.Getenv("LC_ALL")
	} else if os.Getenv("LC_MESSAGE") != "" {
		lcMessage = os.Getenv("LC_MESSAGE")
	} else {
		lcMessage = os.Getenv("LANG")
	}

	if m := lcMessagePattern.FindStringSubmatch(lcMessage); m != nil {
		var lang = m[1]
		var country = m[2]
		var modifier = m[4]

		if country != "" && modifier != "" {
			lcMatchers = []string{
				lang + country + modifier,
				lang + country,
				lang + modifier,
				lang,
			}
		} else if country != "" {
			lcMatchers = []string{
				lang + country,
				lang,
			}
		} else if modifier != "" {
			lcMatchers = []string{
				lang + modifier,
				lang,
			}
		} else {
			lcMatchers = []string{lang}
		}
	} else {
		lcMatchers = []string{}
	}

	if len(lcMatchers) > 0 {
		lang = lcMatchers[len(lcMatchers)-1]
	}
}

func Text(text string) string {
	if lang != "" {
		if m, ok := translations[lang]; ok {
			if translation, ok := m[text]; ok {
				return translation
			}
		}
	}
	return text
}

func Texts(texts []string) []string {
	var translated = make([]string, len(texts), len(texts))
	for i, text := range texts {
		translated[i] = Text(text)
	}
	return translated
}

func LocaleMatch(loc string) bool {
	for _, lm := range lcMatchers {
		if loc == lm {
			return true
		}
	}
	return false
}
