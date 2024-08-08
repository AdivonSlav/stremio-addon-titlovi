package titlovi

type LoginData struct {
	Username       string `json:"UserName"`
	UserId         int64  `json:"UserId"`
	Token          string `json:"Token"`
	ExpirationDate string `json:"ExpirationDate"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SubtitleData struct {
	Id    int64  `json:"Id"`
	Title string `json:"Title"`
	Link  string `json:"Link"`
	Lang  string `json:"Lang"`
	Type  int64  `json:"Type"`
}

type SubtitleDataResponse struct {
	Subtitles []SubtitleData `json:"SubtitleResults"`
}
