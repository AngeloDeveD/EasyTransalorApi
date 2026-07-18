package game

import (
	"log"
	"os"
	"strings"
	"time"
)

// Запуск фонового процесса чистки отклонённых файлов
func StartCleaner(repo GameRepository) {
	//Запуск горутины, чтобы не блокировать основной сервер
	go func() {
		ticket := time.NewTicker(24 * time.Hour) //Срабатывает раз в 24 часа
		defer ticket.Stop()

		cleanupRejectedFiles(repo)

		for range ticket.C {
			cleanupRejectedFiles(repo)
		}
	}()
	log.Println("Фоновый чистильщик запущен. Срабатыввает раз в 24 часа")
}

func cleanupRejectedFiles(repo GameRepository) {
	//Поиск переводов, которые отклонили 7 дней назад
	cards, err := repo.GetOldRejectedTranslations(7)
	if err != nil {
		log.Printf("Ошибка при поискке старых отклоненных переводов: %v", err)
		return
	}

	if len(cards) == 0 {
		return
	}
	log.Printf("Найдено %d старых отклонённых переводов для удаления.", len(cards))

	for _, card := range cards {
		//Удаление файла с самого диска
		if card.UrlToDownload != "" {
			filePath := strings.Replace(card.UrlToDownload, "/static/", "uploads/", 1)
			err := os.Remove(filePath)
			if err != nil && !os.IsNotExist(err) {
				log.Printf("Не удалось найти файл %s: %v", filePath, err)
				continue
			}
		}

		//Удаление записи из БД
		err := repo.DeleteRejectedTranslation(card.ID)
		if err != nil {
			log.Printf("не удалось удалить запить перевода ID %d из БД: %v", card.ID, err)
		} else {
			log.Printf("Удалён отклонённый перевод ID %d и его файл.", card.ID)
		}
	}
}
