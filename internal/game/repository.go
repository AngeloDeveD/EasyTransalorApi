package game

type GameRepository interface {
	GetAllCards() ([]GameCard, error)
	//В будущем GetCardsByID	(id int) (GameInfo, error)
	GetAllGamesInfo() ([]GameInfo, error)
	//CreateCard(card GameCard) error
}

type InMemoryGameRepo struct{}

func (r *InMemoryGameRepo) GetAllCards() ([]GameCard, error) {
	// Старые захардкоженные данные
	gameCard = []GameCard{
		{
			ID:      1,
			Title:   "Игра номер 1",
			IconUrl: "source",
			GameId:  1,
		},
	}
	return gameCard, nil
}

func (r *InMemoryGameRepo) GetAllGamesInfo() ([]GameInfo, error) {
	gameInfo = []GameInfo{
		{
			ID:      1,
			Title:   "Игра номер 1",
			IconUrl: "Url1",
			TranslateCards: []TranslateCard{
				{
					ID:            1,
					AuthorName:    "Васька 1",
					Source:        "url",
					Version:       1.0,
					PercentReady:  0.0,
					UrlToDownload: "url",
					FileSize:      0.0,
				},
			},
		},
		{
			ID:      1,
			Title:   "Игра номер 1",
			IconUrl: "Url1",
			TranslateCards: []TranslateCard{
				{
					ID:            1,
					AuthorName:    "Васька 1",
					Source:        "url",
					Version:       1.0,
					PercentReady:  0.0,
					UrlToDownload: "url",
					FileSize:      0.0,
				},
			},
		},
	}

	return gameInfo, nil
}
