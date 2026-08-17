package game

import (
	"errors"

	"time"

	"gorm.io/gorm"
)

type GameRepository interface {
	GetAllCards() ([]GameCard, error)
	GetAllGamesInfo() ([]GameInfo, error)
	GetTranslationByID(id int) (TranslateCard, error)
	CreateNewGame(GameCard, GameInfo) error
	AddTranslation(int, TranslateCard) error
	CheckCreatedGame(int) error
	GetGameInfoById(int) (GameInfo, error)
	DeleteGame(gameId int) error
	DeleteTranslation(gameId int, translationId int) error
	GetOldRejectedTranslations(daysOld int) ([]TranslateCard, error)
	DeleteRejectedTranslation(id int) error
	GetModerationQueue(limit int, offset int) ([]TranslateCard, int, error)
	ApproveTranslation(translationId int) error
	RejectTranslation(translationId int) error
	ChangeStatusTranslation(translationId int, status string) error
	UpdateScanResult(transID int, status string, details string) error
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

func (r *SqliteGameRepo) GetTranslationByID(id int) (TranslateCard, error) {
	var card TranslateCard
	err := r.db.First(&card, id).Error
	if err != nil {
		return TranslateCard{}, err
	}
	return card, nil
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

func (r *SqliteGameRepo) CheckCreatedGame(gameId int) error {
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

func (r *SqliteGameRepo) AddTranslation(gameId int, newTranslateCard TranslateCard) error {
	newTranslateCard.GameInfoID = int(gameId)

	if err := r.db.Create(&newTranslateCard).Error; err != nil {
		return errors.New("Ошибка добавления перевода в игру. Возможно такой игры не существует.")
	}

	return nil
}

func (r *SqliteGameRepo) GetGameInfoById(gameId int) (GameInfo, error) {
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

func (r *SqliteGameRepo) DeleteGame(gameId int) error {
	tx := r.db.Begin()

	//Удаление всех переводов игры
	if err := tx.Where("game_info_id = ?", gameId).Delete(&TranslateCard{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Удаление GameInfo
	if err := tx.Delete(&GameInfo{}, gameId).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Удаление GameCard
	if err := tx.Where("game_id = ?", gameId).Delete(&GameCard{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *SqliteGameRepo) DeleteTranslation(gameId int, translationId int) error {
	result := r.db.Where("id = ? AND game_info_id = ?", translationId, gameId).Delete(&TranslateCard{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Перевод либо не найден либо не принадлежит этой игре")
	}

	return nil
}

func (r *SqliteGameRepo) GetOldRejectedTranslations(daysOld int) ([]TranslateCard, error) {
	var cards []TranslateCard
	threshold := time.Now().AddDate(0, 0, -daysOld)

	//Поиск переводов со статусом rejected
	err := r.db.Where("status = ? AND created_at < ?", "rejected", threshold).Find(&cards).Error
	return cards, err
}

func (r *SqliteGameRepo) DeleteRejectedTranslation(id int) error {
	return r.db.Unscoped().Delete(&TranslateCard{}, id).Error
}

func (r *SqliteGameRepo) GetModerationQueue(limit int, offset int) ([]TranslateCard, int, error) {
	var translations []TranslateCard
	var totalCount int

	err := r.db.Model(&TranslateCard{}).Where("status = ?", "pending").Limit(limit).Offset(offset).Order("id desc").Find(&translations).Error

	return translations, totalCount, err
}

func (r *SqliteGameRepo) ApproveTranslation(translationId int) error {
	result := r.db.Model(&TranslateCard{}).Where("id = ?", translationId).Update("status", "approved")
	if result.RowsAffected == 0 {
		return errors.New("Перевод не найден")
	}

	return result.Error
}

func (r *SqliteGameRepo) RejectTranslation(translationId int) error {
	result := r.db.Model(&TranslateCard{}).Where("id = ?", translationId).Update("status", "reject")
	if result.RowsAffected == 0 {
		return errors.New("Перевод не найден")
	}

	return result.Error
}

func (r *SqliteGameRepo) ChangeStatusTranslation(translationId int, status string) error {
	result := r.db.Model(&TranslateCard{}).Where("id = ?", translationId).Update("status", status)
	if result.RowsAffected == 0 {
		return errors.New("Перевод не найден")
	}

	return result.Error
}

func (r *SqliteGameRepo) UpdateScanResult(transID int, status string, details string) error {
	result := r.db.Model(&TranslateCard{}).Where("id = ?", transID).Updates(map[string]interface{}{
		"status":       status,
		"scan_details": details,
	})

	if result.RowsAffected == 0 {
		return errors.New("Перевод не найден")
	}
	return result.Error
}

/*Для тестов*/

func (r *InMemoryGameRepo) GetAllCards() ([]GameCard, error) {
	// Старые захардкоженные данные
	return r.gameCard, nil
}

func (r *InMemoryGameRepo) GetAllGamesInfo() ([]GameInfo, error) {
	return r.gameInfo, nil
}

func (r *InMemoryGameRepo) GetTranslationByID(id int) (TranslateCard, error) {
	for _, game := range r.gameInfo {
		for _, card := range game.TranslateCards {
			if card.ID == id {
				return card, nil
			}
		}
	}
	return TranslateCard{}, errors.New("перевод не найден")
}

func (r *InMemoryGameRepo) CreateNewGame(newGameCard GameCard, newGameInfo GameInfo) error {
	r.gameInfo = append(r.gameInfo, newGameInfo)
	r.gameCard = append(r.gameCard, newGameCard)

	return nil
}

func (r *InMemoryGameRepo) CheckCreatedGame(gameId int) error {
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

func (r *InMemoryGameRepo) GetGameInfoById(gameId int) (GameInfo, error) {
	for _, game := range r.gameInfo {
		if game.ID == int(gameId) {
			return game, nil
		}
	}
	return GameInfo{}, errors.New("Игра не найдена")
}

func (r *InMemoryGameRepo) AddTranslation(gameId int, newTranslateCard TranslateCard) error {
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

func (r *InMemoryGameRepo) DeleteGame(gameId int) error {
	// Простая логика удаления для тестов
	for i, game := range r.gameInfo {
		if game.ID == int(gameId) {
			r.gameInfo = append(r.gameInfo[:i], r.gameInfo[i+1:]...)
			return nil
		}
	}
	return errors.New("игра не найдена")
}

func (r *InMemoryGameRepo) DeleteTranslation(gameId int, translationId int) error {
	for i, game := range r.gameInfo {
		if game.ID == int(gameId) {
			for j, card := range game.TranslateCards {
				if card.ID == translationId {
					r.gameInfo[i].TranslateCards = append(r.gameInfo[i].TranslateCards[:j], r.gameInfo[i].TranslateCards[j+1:]...)
					return nil
				}
			}
		}
	}
	return errors.New("перевод не найден")
}

func (r *InMemoryGameRepo) GetModerationQueue(limit int, offset int) ([]TranslateCard, int, error) {
	var queue []TranslateCard
	for _, game := range r.gameInfo {
		for _, card := range game.TranslateCards {
			if card.Status == "pending" {
				queue = append(queue, card)
			}
		}
	}
	return queue, len(queue), nil
}

func (r *InMemoryGameRepo) ApproveTranslation(translationId int) error { return nil }
func (r *InMemoryGameRepo) RejectTranslation(translationId int) error  { return nil }

// GetOldRejectedTranslations в тестах отдаёт все отклонённые переводы
// (без учёта возраста — в памяти нет реальных дат создания).
func (r *InMemoryGameRepo) GetOldRejectedTranslations(daysOld int) ([]TranslateCard, error) {
	var cards []TranslateCard
	for _, game := range r.gameInfo {
		for _, card := range game.TranslateCards {
			if card.Status == "rejected" {
				cards = append(cards, card)
			}
		}
	}
	return cards, nil
}

// DeleteRejectedTranslation удаляет перевод по id из памяти.
func (r *InMemoryGameRepo) DeleteRejectedTranslation(id int) error {
	for i := range r.gameInfo {
		for j, card := range r.gameInfo[i].TranslateCards {
			if card.ID == id {
				r.gameInfo[i].TranslateCards = append(r.gameInfo[i].TranslateCards[:j], r.gameInfo[i].TranslateCards[j+1:]...)
				return nil
			}
		}
	}
	return errors.New("перевод не найден")
}

func (r *InMemoryGameRepo) UpdateScanResult(transID int, status string, details string) error {
	for i, game := range r.gameInfo {
		for j, card := range game.TranslateCards {
			if card.ID == transID {
				r.gameInfo[i].TranslateCards[j].Status = status
				r.gameInfo[i].TranslateCards[j].ScanDetails = details
				return nil
			}
		}
	}
	return errors.New("перевод не найден")
}
