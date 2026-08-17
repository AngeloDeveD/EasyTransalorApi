package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSqlRepositoryCreateAndList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Event{}))

	repo := NewSqlRepository(db)

	require.NoError(t, repo.Create(&Event{
		ActorID:    10,
		ActorRole:  "moderator",
		Action:     "translation.approve",
		TargetType: TargetTranslation,
		TargetID:   42,
		IP:         "127.0.0.1",
		UserAgent:  "test-agent",
	}))

	events, total, err := repo.List(20, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, events, 1)
	assert.Equal(t, "translation.approve", events[0].Action)
	assert.Equal(t, 42, events[0].TargetID)
	assert.Equal(t, "127.0.0.1", events[0].IP)
}
