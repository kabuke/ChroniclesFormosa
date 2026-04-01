package village

import (
	"errors"
	"fmt"

	"github.com/kabuke/ChroniclesFormosa/server/repo"
)

// ElectHeadman 推舉庄長
func ElectHeadman(username string, villageID int64) (string, error) {
	pRepo := repo.NewPlayerRepo()
	vRepo := repo.NewVillageRepo()

	p, err := pRepo.FindByUsername(username)
	if err != nil || p.VillageID != villageID {
		return "", errors.New("您不屬於此庄頭")
	}

	v, err := vRepo.FindByID(villageID)
	if err != nil { return "", err }

	// 簡單邏輯：第一個推舉的人如果目前無庄長則直接當選
	if v.Headman == "" {
		v.Headman = username
		p.VillageRole = 1 // RoleHeadman
		_ = vRepo.Update(v)
		_ = pRepo.Update(p)
		return fmt.Sprintf("眾望所歸！%s 成為了 %s 的新庄長。", p.Nickname, v.Name), nil
	}

	if v.Headman == username {
		return "", errors.New("您已經是庄長了")
	}

	return fmt.Sprintf("%s 提議推舉新庄長，但目前已有在任庄長。", p.Nickname), nil
}

// AssignDeputy 任命副庄長 (Phase 2 補強)
func AssignDeputy(headmanName, deputyName string, villageID int64) (string, error) {
	vRepo := repo.NewVillageRepo()
	pRepo := repo.NewPlayerRepo()

	v, _ := vRepo.FindByID(villageID)
	if v.Headman != headmanName {
		return "", errors.New("只有庄長能任命副手")
	}

	deputy, err := pRepo.FindByUsername(deputyName)
	if err != nil || deputy.VillageID != villageID {
		return "", errors.New("目標玩家不在此庄頭")
	}

	v.DeputyHeadman = deputyName
	deputy.VillageRole = 2 // RoleDeputy
	_ = vRepo.Update(v)
	_ = pRepo.Update(deputy)

	return fmt.Sprintf("%s 被任命為 %s 的副庄長。", deputy.Nickname, v.Name), nil
}

// ImpeachHeadman 彈劾庄長
func ImpeachHeadman(username string, villageID int64) (string, error) {
	vRepo := repo.NewVillageRepo()
	v, _ := vRepo.FindByID(villageID)

	if v.Loyalty > 30 {
		return "", errors.New("民心尚穩，彈劾未獲支持 (民忠需 < 30)")
	}

	oldHeadman := v.Headman
	v.Headman = ""
	v.DeputyHeadman = ""
	_ = vRepo.Update(v)

	return fmt.Sprintf("%s 的庄長 %s 因失去民心，已被族人罷免！", v.Name, oldHeadman), nil
}
