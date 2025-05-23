package models

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	gorm_seeder "github.com/kachit/gorm-seeder"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey" binding:"required"`
	Username  string         `json:"username" gorm:"uniqueIndex" binding:"required"`
	Password  string         `json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u User) TableName() string {
	return "users"
}

func UserExists(db *gorm.DB, user User) (bool, error) {
	var count int64
	err := db.Model(&User{}).Where("ID = ?", user.ID).Limit(1).Count(&count).Error
	return count > 0, err
}

func UserIDExists(db *gorm.DB, id uint) (bool, error) {
	var count int64
	err := db.Model(&User{}).Where("ID = ?", id).Limit(1).Count(&count).Error
	return count > 0, err
}

func FindUserByID(db *gorm.DB, id uint) (User, error) {
	var user User
	err := db.First(&user, id).Error
	return user, err
}

func ListUsers(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Order("id asc").Find(&users).Error
	return users, err
}

func CountUsers(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&User{}).Count(&count).Error
	return int(count), err
}

type UsersSeeder struct {
	gorm_seeder.SeederAbstract
	config *config.Config
}

const UserSeederRows = 1

func NewUsersSeeder(cfg gorm_seeder.SeederConfiguration, config *config.Config) UsersSeeder {
	return UsersSeeder{gorm_seeder.NewSeederAbstract(cfg), config}
}

func (s *UsersSeeder) Seed(db *gorm.DB) error {
	pass := s.config.InitialAdminUserPassword
	if pass == "" {
		slog.Error("Initial admin user password not set, using auto-generated password")
		const randLen = 15
		const randNums = 4
		const randSpecial = 2
		var err error
		pass, err = utils.RandomPassword(randLen, randNums, randSpecial)
		if err != nil {
			return fmt.Errorf("password generation failed: %w", err)
		}
	}
	var users = []User{
		{
			ID:       0,
			Username: "admin",
			Password: utils.HashPassword(pass, s.config.PasswordSalt),
		},
	}
	slog.Error("!#!#!#!#!# Initial admin user password #!#!#!#!#!", "password", pass)
	return db.CreateInBatches(users, s.Configuration.Rows).Error
}

func (s *UsersSeeder) Clear(db *gorm.DB) error {
	return db.Delete(&User{ID: 0}).Error
}

func DeleteUser(db *gorm.DB, id uint) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		tx.Unscoped().Delete(&User{ID: id})
		return nil
	})
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}
