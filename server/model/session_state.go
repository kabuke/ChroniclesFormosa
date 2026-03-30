package model

import "time"

// SessionState 用於保存已斷線但尚未 GC 的連線狀態，以利重連時無縫接軌 (對應 table: session_states)
type SessionState struct {
	SessionID    string    `gorm:"primaryKey;size:64"`
	Username     string    `gorm:"index;size:64"` // 若未登入可能為空
	SharedSecret []byte    `gorm:"type:blob"`     // ECDH Secret (儲存 bytes)
	LastAck      uint64
	NextSeq      uint64
	MaxClientSeq uint64
	UpdatedAt    time.Time `gorm:"autoUpdateTime;index"` // 輔助 DB 端的冷資料清理
}
