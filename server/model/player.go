package model

import (
	"time"
)

// Player 玩家領域模型 (對應 table: players)
type Player struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"uniqueIndex;not null;size:64"`
	PasswordHash string    `gorm:"not null;size:255"` // SHA-256
	Nickname     string    `gorm:"size:64"`
	FactionID    int32     `gorm:"index;default:0"`   // 隸屬陣營 (0=未入)
	VillageID    int64     `gorm:"index;default:0"`   // 所屬莊頭ID
	
	// 精力值系統 (Phase 2)
	Stamina      int32     `gorm:"default:100"`       // 當前精力 (上限 100)
	LastRegenAt  time.Time `gorm:"autoCreateTime"`    // 上次恢復時間

	// 庄頭職位 (Phase 2)
	// 0=普通族民, 1=庄長, 2=墾首, 3=武師, 4=商賈
	VillageRole  int32     `gorm:"default:0"`

	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	LastLoginAt  time.Time
}
