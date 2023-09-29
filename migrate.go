package db

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func migrate(config *Config) error {
	migrator := gormigrate.New(i, gormigrate.DefaultOptions, config.Migrations)

	// 1. If there is no existing database, run auto migrations.
	migrator.InitSchema(func(db *gorm.DB) error {
		return autoMigrate(config)
	})

	// 2. Run all the manual migrations.
	// This is a no-op if (1) was run.
	if err := migrator.Migrate(); err != nil {
		return err
	}

	// 3. Run auto migrations to handle new columns and other auto-migratable schema changes.
	// This is a no-op if (1) was run.
	return autoMigrate(config)
}

func autoMigrate(config *Config) error {
	for _, model := range config.Models {
		if err := i.AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
