package db

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/glebarez/sqlite"
	gorm_seeder "github.com/kachit/gorm-seeder"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func MakeDB(config *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error
	if os.Getenv("TEST") != "" {
		fmt.Println("Using in-memory database for testing")
		db, err = gorm.Open(sqlite.Open(""), &gorm.Config{})
		if err != nil {
			fmt.Printf("Could not open database: %v\n", err)
			os.Exit(1)
		}
	} else {
		db, err = gorm.Open(postgres.Open(config.PostgresDSN), &gorm.Config{})
		if err != nil {
			fmt.Printf("Could not open database: %v\n", err)
			os.Exit(1)
		}
		if config.OTLPEndpoint != "" {
			if err = db.Use(otelgorm.NewPlugin()); err != nil {
				fmt.Printf("Could not trace database: %v\n", err)
				os.Exit(1)
			}
		}
	}

	err = db.AutoMigrate(&models.AppSettings{}, &models.User{}, &models.Tunnel{})
	if err != nil {
		fmt.Printf("Could not migrate database: %v\n", err)
		os.Exit(1)
	}

	// Grab the first (and only) AppSettings record. If that record doesn't exist, create it.
	var appSettings models.AppSettings
	result := db.First(&appSettings)
	if result.RowsAffected == 0 {
		if config.Debug {
			fmt.Println("App settings entry doesn't exist, creating it")
		}
		// The record doesn't exist, so create it
		appSettings = models.AppSettings{
			HasSeeded: false,
		}
		err := db.Create(&appSettings).Error
		if err != nil {
			fmt.Printf("Failed to create app settings: %v\n", err)
			os.Exit(1)
		}
		if config.Debug {
			fmt.Println("App settings saved")
		}
	}

	// If the record exists and HasSeeded is true, then we don't need to seed the database.
	if !appSettings.HasSeeded {
		usersSeeder := models.NewUsersSeeder(gorm_seeder.SeederConfiguration{Rows: models.UserSeederRows}, config)
		seedersStack := gorm_seeder.NewSeedersStack(db)
		seedersStack.AddSeeder(&usersSeeder)

		// Apply seed
		err = seedersStack.Seed()
		if err != nil {
			fmt.Printf("Failed to seed database: %v\n", err)
			os.Exit(1)
		}
		appSettings.HasSeeded = true
		err := db.Save(&appSettings).Error
		if err != nil {
			fmt.Printf("Failed to save app settings: %v\n", err)
			os.Exit(1)
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	sqlDB.SetMaxIdleConns(runtime.GOMAXPROCS(0))
	const connsPerCPU = 10
	sqlDB.SetMaxOpenConns(runtime.GOMAXPROCS(0) * connsPerCPU)
	const maxIdleTime = 10 * time.Minute
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	return db
}
