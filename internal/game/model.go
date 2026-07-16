package game

type GameCard struct {
	ID      int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Title   string `json:"title"`
	IconUrl string `json:"iconUrl"`
	GameId  int    `json:"gameId" gorm:"unique"`
}

type GameInfo struct {
	ID             int             `json:"id" gorm:"primaryKey;autoIncrement"`
	Title          string          `json:"title"`
	IconUrl        string          `json:"iconUrl"`
	TranslateCards []TranslateCard `json:"translateCards" gorm:"foreignKey:GameInfoID"`
}

type TranslateCard struct {
	ID            int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	AuthorName    string  `json:"authorName"`
	AuthorId      int     `json:"authoreId"`
	Source        string  `json:"source"`
	Version       float64 `json:"version"`
	PercentReady  float64 `json:"percentReady"`
	UrlToDownload string  `json:"urlToDownload"`
	FileSize      float64 `json:"fileSize"`
	Status        string  `json:"status" gorm:"default:pending"` //pending, approved, rejected
	GameInfoID    int     `json:"-"`
}

type CreateGameRequest struct {
	Title string `json:"title"`
}

type CreateTraslateRequest struct {
	AuthorName   string  `json:"authorName"`
	Source       string  `json:"source"`
	Version      float64 `json:"version"`
	PercentReady float64 `json:"percentReady"`
}
