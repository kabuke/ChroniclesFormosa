package model

import "time"

// Village 莊頭/聚落模型 (對應 table: villages)
type Village struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"not null;size:100;index"`
	Level       int32     `gorm:"default:1"`
	FactionID   int32     `gorm:"index;default:0"` // 所屬派系 (0=無主, >0=被方佔領)
	Wood        int64     `gorm:"default:0"` // 經濟資源
	Food        int64     `gorm:"default:0"`
	Iron        int64     `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
