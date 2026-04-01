package village

import (
	"errors"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
)

var (
	ErrNotLoggedIn     = errors.New("您必須先登入帳號才能執行操作")
	ErrVillageNotFound = errors.New("找不到指定的莊頭")
	ErrAlreadyInThis   = errors.New("您早已經是本莊頭的成員囉")
	ErrAlreadyInOther  = errors.New("您已經加入其他莊頭了，請先退出")
)

// GetVillageInfo 獲取莊頭詳情與當前人口
func GetVillageInfo(villageID int64) (*model.Village, int32, error) {
	vRepo := repo.NewVillageRepo()
	pRepo := repo.NewPlayerRepo()

	v, err := vRepo.FindByID(villageID)
	if err != nil {
		return nil, 0, err
	}

	pop, _ := pRepo.CountByVillageID(villageID)
	return v, int32(pop), nil
}

// GetAllVillages 獲取所有莊頭的簡要清單 (Phase 2 加入莊頭用)
func GetAllVillages() ([]*pb.VillageSummary, error) {
	vRepo := repo.NewVillageRepo()
	pRepo := repo.NewPlayerRepo()

	villages, err := vRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var list []*pb.VillageSummary
	for _, v := range villages {
		pop, _ := pRepo.CountByVillageID(v.ID)
		list = append(list, &pb.VillageSummary{
			VillageId:  v.ID,
			Name:       v.Name,
			Level:      v.Level,
			Population: int32(pop),
			FactionId:  v.FactionID,
			X:          v.X,
			Y:          v.Y,
		})
	}
	return list, nil
}

// JoinVillage 讓玩家加入莊頭
func JoinVillage(username string, villageID int64) error {
	vRepo := repo.NewVillageRepo()
	v, err := vRepo.FindByID(villageID)
	if err != nil || v == nil {
		return ErrVillageNotFound
	}

	pRepo := repo.NewPlayerRepo()
	player, err := pRepo.FindByUsername(username)
	if err != nil {
		return repo.ErrPlayerNotFound
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

// GetVillageMembers 獲取庄頭內所有成員資訊
func GetVillageMembers(villageID int64) ([]*pb.VillageMember, error) {
	pRepo := repo.NewPlayerRepo()
	players, err := pRepo.FindByVillageID(villageID)
	if err != nil {
		return nil, err
	}

	var members []*pb.VillageMember
	for _, p := range players {
		members = append(members, &pb.VillageMember{
			Username: p.Username,
			Nickname: p.Nickname,
			Role:     p.VillageRole,
		})
	}
	return members, nil
}
