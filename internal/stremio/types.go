package stremio

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
	Resources      []string       `json:"resources"`
	IdPrefixes     []string       `json:"idPrefixes"`
	BehaviourHints BehaviourHints `json:"behaviourHints"`
}
