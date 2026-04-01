package repo

import (
	"errors"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"gorm.io/gorm"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
	ErrPlayerExists   = errors.New("player already exists (unique constraint failed)")
)

type playerRepoImpl struct{}

func NewPlayerRepo() PlayerRepo {
	return &playerRepoImpl{}
}

func (r *playerRepoImpl) Create(player *model.Player) error {
	err := database.GetDB().Create(player).Error
	if err != nil {
		// 未來若需要可以靠字串 mapping 區分 Unique Violation
		return err
	}
	return nil
}

func (r *playerRepoImpl) FindByUsername(username string) (*model.Player, error) {
	var p model.Player
	err := database.GetDB().Where("username = ?", username).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPlayerNotFound
	}
	return &p, err
}

func (r *playerRepoImpl) FindByVillageID(villageID int64) ([]*model.Player, error) {
	var players []*model.Player
	err := database.GetDB().Where("village_id = ?", villageID).Find(&players).Error
	return players, err
}

func (r *playerRepoImpl) Update(player *model.Player) error {
	return database.GetDB().Save(player).Error
}

func (r *playerRepoImpl) CountByVillageID(villageID int64) (int64, error) {
	var count int64
	err := database.GetDB().Model(&model.Player{}).Where("village_id = ?", villageID).Count(&count).Error
	return count, err
}
