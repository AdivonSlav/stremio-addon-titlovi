package stremio

type UserConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CatalogItem struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SubtitleItem struct {
	Id   string `json:"id"`
	Url  string `json:"url"`
	Lang string `json:"lang"`
}

type SubtitlesResponse struct {
	Subtitles []*SubtitleItem `json:"subtitles"`
}

type BehaviourHints struct {
	Configurable          bool `json:"configurable"`
	ConfigurationRequired bool `json:"configurationRequired"`
}

type Manifest struct {
	Id             string         `json:"id"`
	Version        string         `json:"version"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Types          []string       `json:"types"`
	Resources      []string       `json:"resources,omitempty"`
	Catalogs       []CatalogItem  `json:"catalogs"`
	IdPrefixes     []string       `json:"idPrefixes,omitempty"`
	Logo           string         `json:"logo,omitempty"`
	BehaviourHints BehaviourHints `json:"behaviourHints,omitempty"`
}
