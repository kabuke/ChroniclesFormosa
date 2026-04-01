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

	X           int32     `gorm:"default:0"` // 世界座標
	Y           int32     `gorm:"default:0"`

	// 社會與族群屬性 (Phase 2)
	PopMinNan     int32 `gorm:"default:0"` // 閩南人口
	PopHakka      int32 `gorm:"default:0"` // 客家人口
	PopIndigenous int32 `gorm:"default:0"` // 原民人口
	
	TensionValue int32 `gorm:"default:0"`  // 0 (和平) ~ 100 (爆發械鬥)
	Stability    int32 `gorm:"default:100"`// 治安/穩定度 (0~100)
	Loyalty      int32 `gorm:"default:80"` // 民忠 (0~100)

	Headman      string `gorm:"size:64"`   // 當前庄長 Username
	DeputyHeadman string `gorm:"size:64"`  // 當前副庄長 (Phase 2)
	Soldiers     int32  `gorm:"default:0"` // 庄頭武力 (傭兵/族勇)

	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
