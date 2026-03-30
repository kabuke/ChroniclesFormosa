package repo

import (
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"gorm.io/gorm/clause"
)

type sessionRepoImpl struct{}

func NewSessionRepo() SessionRepo {
	return &sessionRepoImpl{}
}

func (r *sessionRepoImpl) UpsertBatch(states []*model.SessionState) error {
	if len(states) == 0 {
		return nil
	}
	
	// 利用 GORM 的 Clause() 實現 ON CONFLICT 批量插入與更新
	return database.GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}}, // 以 PK SessionID 為準
		DoUpdates: clause.AssignmentColumns([]string{
			"username", "last_ack", "next_seq", "max_client_seq", "updated_at", // SharedSecret 一開始交換完就不會變
		}),
	}).Create(states).Error
}

func (r *sessionRepoImpl) LoadActive() ([]*model.SessionState, error) {
	var states []*model.SessionState
	// 只有在一小時內活動的 session 才有資格被載回記憶體供 Resume
	cutoff := time.Now().Add(-60 * time.Minute)
	err := database.GetDB().Where("updated_at > ?", cutoff).Find(&states).Error
	return states, err
}

func (r *sessionRepoImpl) DeleteExpired() error {
	// 保留 24 小時的殘影，超過的實體刪除以節省 DB 空間
	cutoff := time.Now().Add(-24 * time.Hour)
	return database.GetDB().Where("updated_at < ?", cutoff).Delete(&model.SessionState{}).Error
}
