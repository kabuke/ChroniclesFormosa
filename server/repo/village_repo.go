package repo

import (
	"errors"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"gorm.io/gorm"
)

var ErrVillageNotFound = errors.New("village not found")

type villageRepoImpl struct{}

func NewVillageRepo() VillageRepo {
	return &villageRepoImpl{}
}

func (r *villageRepoImpl) Create(village *model.Village) error {
	return database.GetDB().Create(village).Error
}

func (r *villageRepoImpl) FindByID(id int64) (*model.Village, error) {
	var v model.Village
	err := database.GetDB().First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrVillageNotFound
	}
	return &v, err
}

func (r *villageRepoImpl) FindAll() ([]*model.Village, error) {
	var villages []*model.Village
	err := database.GetDB().Find(&villages).Error
	return villages, err
}

func (r *villageRepoImpl) Update(village *model.Village) error {
	return database.GetDB().Save(village).Error
}
