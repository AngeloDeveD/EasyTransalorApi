// Package scanner — HTTP-клиент к python-сканеру архивов.
// После загрузки перевода Go шлёт сюда POST /scan; вердикт возвращается
// позже асинхронно на POST /api/internal/scan-result.
package scanner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Client отправляет архивы на антивирусную проверку сканеру.
type Client struct {
	scanURL     string // напр. http://scanner:8000/scan
	internalKey string // общий секрет для заголовка X-Internal-Key
	fileRoot    string // корень, по которому загрузки видны сканеру (напр. /app/uploads)
	http        *http.Client
}

// NewClient создаёт клиент сканера.
func NewClient(scanURL, internalKey, fileRoot string) *Client {
	return &Client{
		scanURL:     scanURL,
		internalKey: internalKey,
		fileRoot:    fileRoot,
		http:        &http.Client{Timeout: 15 * time.Second},
	}
}

type scanRequest struct {
	TransID  int    `json:"transId"`
	FilePath string `json:"filePath"`
}

// Scan уведомляет сканер о новом архиве асинхронно (в отдельной горутине),
// чтобы не блокировать ответ пользователю. Вердикт придёт позже колбэком.
func (c *Client) Scan(transID int, fileURL string) {
	go func() {
		if err := c.send(transID, fileURL); err != nil {
			log.Printf("scanner: не удалось отправить архив на проверку (transId=%d): %v", transID, err)
		}
	}()
}

func (c *Client) send(transID int, fileURL string) error {
	body, err := json.Marshal(scanRequest{
		TransID:  transID,
		FilePath: c.filePath(fileURL),
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.scanURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.internalKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сканер ответил статусом %d", resp.StatusCode)
	}
	return nil
}

// filePath переводит публичный file_url (static/files/{gameid}/{uuid}.ext)
// в путь, по которому сканер видит файл внутри своего контейнера
// (fileRoot/files/{gameid}/{uuid}.ext, forward-slash — сканер всегда на linux).
func (c *Client) filePath(fileURL string) string {
	rel := strings.TrimPrefix(fileURL, "/")
	rel = strings.TrimPrefix(rel, "static/")
	return strings.TrimRight(c.fileRoot, "/") + "/" + rel
}
