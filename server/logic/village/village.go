package village

import (
	"errors"

	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
)

var (
	ErrNotLoggedIn     = errors.New("您必須先登入帳號才能執行操作")
	ErrVillageNotFound = errors.New("查無此莊頭")
	ErrPlayerNotFound  = errors.New("無法取得角色資料")
	ErrAlreadyInOther  = errors.New("您已是其他莊員，請先退出")
	ErrAlreadyInThis   = errors.New("您早已經是本莊頭的成員囉")
)

// GetVillageInfo 取出莊頭基本資料與即時動態人口，供網路層打包
func GetVillageInfo(villageID int64) (*model.Village, int32, error) {
	vRepo := repo.NewVillageRepo()
	village, err := vRepo.FindByID(villageID)
	if err != nil {
		return nil, 0, ErrVillageNotFound
	}

	pRepo := repo.NewPlayerRepo()
	population, _ := pRepo.CountByVillageID(villageID)

	return village, int32(population), nil
}

// JoinVillage 處理玩家申請加入莊頭的核心業務防呆與 DB 更新
func JoinVillage(username string, villageID int64) error {
	if username == "" {
		return ErrNotLoggedIn
	}

	vRepo := repo.NewVillageRepo()
	if _, err := vRepo.FindByID(villageID); err != nil {
		return ErrVillageNotFound
	}

	pRepo := repo.NewPlayerRepo()
	player, err := pRepo.FindByUsername(username)
	if err != nil {
		return ErrPlayerNotFound
	}

	if player.VillageID == villageID {
		return ErrAlreadyInThis
	}
	if player.VillageID != 0 {
		return ErrAlreadyInOther
	}

	// 滿足一切條件，更新歸屬
	player.VillageID = villageID
	return pRepo.Update(player)
}
