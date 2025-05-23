package db

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/glebarez/sqlite"
	gorm_seeder "github.com/kachit/gorm-seeder"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MakeDB(config *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	if os.Getenv("TEST") != "" {
		slog.Info("Using in-memory database for testing")
		db, err = gorm.Open(sqlite.Open(""), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("could not open in-memory database: %v", err)
		}
	} else {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s dbname=%s password=%s",
			config.Postgres.Host,
			config.Postgres.Port,
			config.Postgres.User,
			config.Postgres.Database,
			config.Postgres.Password,
		)

		slog.Info("Connecting to postgres", "dsn", dsn)
		pg := postgres.Open(dsn)
		slog.Info("Opening gorm database connection")

		db, err = gorm.Open(pg, &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("could not open database: %v", err)
		}

		slog.Info("Gorm database connection opened")
	}

	err = db.AutoMigrate(&models.AppSettings{}, &models.User{}, &models.Tunnel{})
	if err != nil {
		return nil, fmt.Errorf("could not migrate database: %v", err)
	}

	// Grab the first (and only) AppSettings record. If that record doesn't exist, create it.
	var appSettings models.AppSettings
	result := db.First(&appSettings)
	if result.RowsAffected == 0 {
		slog.Debug("App settings entry doesn't exist, creating it")
		// The record doesn't exist, so create it
		appSettings = models.AppSettings{
			HasSeeded: false,
		}
		err := db.Create(&appSettings).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create app settings: %v", err)
		}
		slog.Debug("App settings entry created")
	}

	// If the record exists and HasSeeded is true, then we don't need to seed the database.
	if !appSettings.HasSeeded {
		usersSeeder := models.NewUsersSeeder(gorm_seeder.SeederConfiguration{Rows: models.UserSeederRows}, config)
		seedersStack := gorm_seeder.NewSeedersStack(db)
		seedersStack.AddSeeder(&usersSeeder)

		// Apply seed
		err = seedersStack.Seed()
		if err != nil {
			return nil, fmt.Errorf("failed to seed database: %v", err)
		}
		appSettings.HasSeeded = true
		err := db.Save(&appSettings).Error
		if err != nil {
			return nil, fmt.Errorf("failed to save app settings: %v", err)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(runtime.GOMAXPROCS(0))
	const connsPerCPU = 10
	sqlDB.SetMaxOpenConns(runtime.GOMAXPROCS(0) * connsPerCPU)
	const maxIdleTime = 10 * time.Minute
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	return db, nil
}
