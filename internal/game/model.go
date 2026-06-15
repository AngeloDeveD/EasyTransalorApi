package game

type GameCard struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	IconUrl string `json:"iconUrl"`
	GameId  int    `json:"gameId"`
}

type GameInfo struct {
	ID             int             `json:"id"`
	Title          string          `json:"title"`
	IconUrl        string          `json:"iconUrl"`
	TranslateCards []TranslateCard `json:"translateCards"`
}

type TranslateCard struct {
	ID            int     `json:"id"`
	AuthorName    string  `json:"authorName"`
	Source        string  `json:"source"`
	Version       float64 `json:"version"`
	PercentReady  float64 `json:"percentReady"`
	UrlToDownload string  `json:"urlToDownload"`
	FileSize      float64 `json:"fileSize"`
}
