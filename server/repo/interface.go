package repo

import "github.com/kabuke/ChroniclesFormosa/server/model"

// PlayerRepo 定義玩家帳號的持久化介面
type PlayerRepo interface {
	Create(player *model.Player) error
	FindByUsername(username string) (*model.Player, error)
	FindByVillageID(villageID int64) ([]*model.Player, error)
	Update(player *model.Player) error
	CountByVillageID(villageID int64) (int64, error)
}

// VillageRepo 定義莊頭層級的持久化介面
type VillageRepo interface {
	Create(village *model.Village) error
	FindByID(id int64) (*model.Village, error)
	FindAll() ([]*model.Village, error)
	Update(village *model.Village) error
}

// SessionRepo 定義網路中斷接續狀態的持久化介面
type SessionRepo interface {
	// UpsertBatch 執行批次寫入 (Insert ... On Conflict Do Update) 供高併發定時存檔
	UpsertBatch(states []*model.SessionState) error
	// LoadAll 讀出所有未過期的狀態，以利 Server 崩潰重啟時接關 (Resume)
	LoadActive() ([]*model.SessionState, error)
	// DeleteExpired 清理很久之前的 Session (如超過 24 Hr)
	DeleteExpired() error
}
