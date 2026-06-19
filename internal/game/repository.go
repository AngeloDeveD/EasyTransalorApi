package game

import (
	"errors"

	"gorm.io/gorm"
)

type GameRepository interface {
	GetAllCards() ([]GameCard, error)
	//В будущем GetCardsByID	(id int) (GameInfo, error)
	GetAllGamesInfo() ([]GameInfo, error)
	CreateNewGame(GameCard, GameInfo) error
	AddTranslation(int64, TranslateCard) error
	CheckCreatedGame(int64) error
	GetGameInfoById(int64) (GameInfo, error)
}

type SqliteGameRepo struct {
	db *gorm.DB
}

type InMemoryGameRepo struct {
	gameInfo []GameInfo
	gameCard []GameCard
}

func NewSqlGameRepo(db *gorm.DB) *SqliteGameRepo {
	return &SqliteGameRepo{db: db}
}

func NewInMemoryGameRepo() *InMemoryGameRepo {
	return &InMemoryGameRepo{
		gameInfo: []GameInfo{
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
		},
		gameCard: []GameCard{
			{
				ID:      1,
				Title:   "Игра номер 1",
				IconUrl: "source",
				GameId:  1,
			},
		},
	}
}

/*Для работы с БД*/
func (r *SqliteGameRepo) GetAllCards() ([]GameCard, error) {
	var cards []GameCard
	result := r.db.Find(&cards)
	return cards, result.Error
}

func (r *SqliteGameRepo) GetAllGamesInfo() ([]GameInfo, error) {
	var games []GameInfo
	result := r.db.Preload("TranslateCards").Find(&games)
	return games, result.Error
}

func (r *SqliteGameRepo) CreateNewGame(newGameCard GameCard, newGameInfo GameInfo) error {
	tx := r.db.Begin()

	//Сохранение GameCard
	if err := tx.Create(&newGameCard).Error; err != nil {
		tx.Rollback()
		return err
	}

	//сохранение GameInfo
	if err := tx.Create(&newGameInfo).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *SqliteGameRepo) CheckCreatedGame(gameId int64) error {
	var gameInfo GameInfo
	if err := r.db.First(&gameInfo, gameId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Такой игры не существует")
		} else {
			return err
		}
	}

	return nil
}

func (r *SqliteGameRepo) AddTranslation(gameId int64, newTranslateCard TranslateCard) error {
	newTranslateCard.GameInfoID = int(gameId)

	if err := r.db.Create(&newTranslateCard).Error; err != nil {
		return errors.New("Ошибка добавления перевода в игру. Возможно такой игры не существует.")
	}

	return nil
}

func (r *SqliteGameRepo) GetGameInfoById(gameId int64) (GameInfo, error) {
	var gameInfo GameInfo

	err := r.db.Preload("TranslateCards").First(&gameInfo, gameId).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GameInfo{}, errors.New("Игра не найдена")
		}

		return GameInfo{}, err
	}

	return gameInfo, nil
}

/*Для тестов*/

func (r *InMemoryGameRepo) GetAllCards() ([]GameCard, error) {
	// Старые захардкоженные данные
	return r.gameCard, nil
}

func (r *InMemoryGameRepo) GetAllGamesInfo() ([]GameInfo, error) {
	return r.gameInfo, nil
}

func (r *InMemoryGameRepo) CreateNewGame(newGameCard GameCard, newGameInfo GameInfo) error {
	r.gameInfo = append(r.gameInfo, newGameInfo)
	r.gameCard = append(r.gameCard, newGameCard)

	return nil
}

func (r *InMemoryGameRepo) CheckCreatedGame(gameId int64) error {
	found := false

	for i := range r.gameInfo {
		if r.gameInfo[i].ID == int(gameId) {
			found = true
			break
		}
	}

	if !found {
		return errors.New("Игра была не найдена!")
	}
	return nil
}

func (r *InMemoryGameRepo) AddTranslation(gameId int64, newTranslateCard TranslateCard) error {
	status := false
	for i := range r.gameInfo {
		if r.gameInfo[i].ID == int(gameId) {
			r.gameInfo[i].TranslateCards = append(r.gameInfo[i].TranslateCards, newTranslateCard)
			status = true
			break
		}
	}

	if !status {
		return errors.New("Ошибка добавления перевода в игру. Возможно такой игры не существует.")
	}

	return nil
}
