package model

import "time"

// DiplomacyRelation 記錄庄頭間或玩家間的外交關係
type DiplomacyRelation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	SourceID  int64     `gorm:"index"` // 發起方 (VillageID 或 PlayerID)
	TargetID  int64     `gorm:"index"` // 接受方
	Type      int32     // 0=結盟, 1=聯姻, 2=拜把, 3=理番
	ExpiredAt *time.Time // 有效期限 (nil 為永久)
	CreatedAt time.Time
}
