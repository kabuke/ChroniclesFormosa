package model

import (
	"time"
)

// Player 玩家領域模型 (對應 table: players)
type Player struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"uniqueIndex;not null;size:64"`
	PasswordHash string    `gorm:"not null;size:255"` // SHA-256
	FactionID    int32     `gorm:"index;default:0"`   // 隸屬陣營 (0=未入)
	VillageID    int64     `gorm:"index;default:0"`   // 所屬莊頭ID
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	LastLoginAt  time.Time
}
