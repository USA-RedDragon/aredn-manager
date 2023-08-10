package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Supernode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	MeshName  string         `json:"mesh_name" binding:"required"`
	IPs       []string       `json:"ips" binding:"required"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func SupernodeIDExists(db *gorm.DB, id uint) (bool, error) {
	var count int64
	err := db.Model(&Supernode{}).Where("ID = ?", id).Limit(1).Count(&count).Error
	return count > 0, err
}

func FindSupernodeByID(db *gorm.DB, id uint) (Supernode, error) {
	var supernode Supernode
	err := db.First(&supernode, id).Error
	return supernode, err
}

func ListSupernodes(db *gorm.DB) ([]Supernode, error) {
	var supernodes []Supernode
	err := db.Find(&supernodes).Error
	return supernodes, err
}

func CountSupernodes(db *gorm.DB) (int, error) {
	var count int64
	err := db.Model(&Supernode{}).Count(&count).Error
	return int(count), err
}

func DeleteSupernode(db *gorm.DB, id uint) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		tx.Unscoped().Delete(&Supernode{ID: id})
		return nil
	})
	if err != nil {
		fmt.Printf("Error deleting supernode: %v\n", err)
		return err
	}
	return nil
}
