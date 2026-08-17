package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunAppliesInitialSchemaOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, Run(db, "sqlite"))
	require.NoError(t, Run(db, "sqlite"))

	var count int64
	require.NoError(t, db.Model(&SchemaMigration{}).Where("id = ?", "0001_initial_schema").Count(&count).Error)
	assert.Equal(t, int64(1), count)

	for _, table := range []string{
		"users",
		"warnings",
		"game_infos",
		"game_cards",
		"translate_cards",
		"notifications",
		"chat_messages",
		"events",
	} {
		assert.True(t, db.Migrator().HasTable(table), "missing table %s", table)
	}

	assert.True(t, db.Migrator().HasIndex("users", "idx_users_nickname"))
	assert.True(t, db.Migrator().HasIndex("translate_cards", "idx_translate_cards_archive_hash"))
	assert.True(t, db.Migrator().HasIndex("events", "idx_events_action"))
}
