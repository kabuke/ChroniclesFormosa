package village

import (
	"errors"
	"fmt"

	"github.com/kabuke/ChroniclesFormosa/server/repo"
)

// AssignRole 提供庄長任命其他玩家為墾首(2)、武師(3)、商賈(4) 或解任(0) 的功能
func AssignRole(operatorUsername, targetUsername string, villageID int64, targetRole int32) (string, error) {
	vRepo := repo.NewVillageRepo()
	pRepo := repo.NewPlayerRepo()

	v, err := vRepo.FindByID(villageID)
	if err != nil {
		return "", errors.New("庄頭不存在")
	}

	// 權限檢查：操作者必須是本庄的庄長
	if v.Headman != operatorUsername {
		return "", errors.New("只有庄長可以進行人事任命")
	}

	// 目標檢查：目標不能是自己 (若要辭職需走卸任流程，暫不開放)
	if targetUsername == operatorUsername {
		return "", errors.New("庄長不能任命或解任自己")
	}

	// 目標玩家檢查
	targetPlayer, err := pRepo.FindByUsername(targetUsername)
	if err != nil || targetPlayer.VillageID != villageID {
		return "", errors.New("目標玩家不在此庄頭")
	}

	// 防呆：不可任命其他玩家為庄長 (角色 1)
	if targetRole == 1 {
		return "", errors.New("不能直接指派庄長職位")
	}

	// 儲存新的職位
	targetPlayer.VillageRole = targetRole
	if err := pRepo.Update(targetPlayer); err != nil {
		return "", errors.New("更新職位失敗")
	}

	roleName := "普通族民"
	switch targetRole {
	case 2:
		roleName = "墾首"
	case 3:
		roleName = "武師"
	case 4:
		roleName = "商賈"
	}

	if targetRole == 0 {
		return fmt.Sprintf("庄長已解除 %s 的職位，貶為平民。", targetPlayer.Nickname), nil
	}

	return fmt.Sprintf("%s 拔擢了 %s，任命其為本庄的【%s】！", operatorUsername, targetPlayer.Nickname, roleName), nil
}
