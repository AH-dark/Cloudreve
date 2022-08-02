package model

import (
	"gorm.io/driver/sqlite"
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMigration(t *testing.T) {
	asserts := assert.New(t)
	conf.DatabaseConfig.Type = "sqlite3"
	DB, _ = gorm.Open(sqlite.Open("file::memory:?cache=shared"))

	asserts.NotPanics(func() {
		migration()
	})
	conf.DatabaseConfig.Type = "mysql"
	DB = mockDB
}
