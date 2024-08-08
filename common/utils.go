package common

var langMap = map[string]string{
	"Bosanski": "bos",
	"Hrvatski": "hrv",
	"Srpski":   "srb",
	"Cirilica": "cpb",
	"English":  "eng",
}

func ConvertLangToISO(language string) string {
	code, ok := langMap[language]
	if !ok {
		return ""
	}

	return code
}

func GetLanguagesToQuery() []string {
	keys := make([]string, len(langMap))

	i := 0
	for k := range langMap {
		keys[i] = k
		i++
	}

	return keys
}
