package audit

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	TargetUser        = "user"
	TargetTranslation = "translation"
)

type Event struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ActorID    int       `json:"actorId" gorm:"index"`
	ActorRole  string    `json:"actorRole"`
	Action     string    `json:"action" gorm:"index"`
	TargetType string    `json:"targetType" gorm:"index"`
	TargetID   int       `json:"targetId" gorm:"index"`
	Details    string    `json:"details,omitempty"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"userAgent"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Repository interface {
	Create(event *Event) error
	List(limit int, offset int) ([]Event, int64, error)
}

type SqlRepository struct {
	db *gorm.DB
}

func NewSqlRepository(db *gorm.DB) *SqlRepository {
	return &SqlRepository{db: db}
}

func (r *SqlRepository) Create(event *Event) error {
	return r.db.Create(event).Error
}

func (r *SqlRepository) List(limit int, offset int) ([]Event, int64, error) {
	var events []Event
	var total int64

	if err := r.db.Model(&Event{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&events).Error
	return events, total, err
}

type NoopRepository struct{}

func (NoopRepository) Create(event *Event) error { return nil }
func (NoopRepository) List(limit int, offset int) ([]Event, int64, error) {
	return []Event{}, 0, nil
}

func NewEventFromContext(c *gin.Context, action string, targetType string, targetID int, details string) *Event {
	event := &Event{
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}

	if actorID, exists := c.Get("userID"); exists {
		if id, ok := actorID.(int); ok {
			event.ActorID = id
		}
	}

	if role, exists := c.Get("role"); exists {
		if value, ok := role.(string); ok {
			event.ActorRole = value
		}
	}

	return event
}
