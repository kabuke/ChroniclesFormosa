package village

import (
	"testing"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestElectHeadman(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Player{}, &model.Village{})

	v := model.Village{ID: 1, Name: "選舉庄"}
	db.Create(&v)
	p := model.Player{Username: "voter1", VillageID: 1, Nickname: "大庄家"}
	db.Create(&p)

	// 1. 第一次推舉
	msg, err := ElectHeadman("voter1", 1)
	if err != nil { t.Fatal(err) }
	t.Log(msg)

	var updatedV model.Village
	db.First(&updatedV, 1)
	if updatedV.Headman != "voter1" { t.Error("當選人錯誤") }

	// 2. 重複推舉
	_, err = ElectHeadman("voter1", 1)
	if err == nil { t.Error("重複推舉應報錯") }
}

func TestImpeachHeadman(t *testing.T) {
	database.InitDB(":memory:")
	db := database.GetDB()
	db.AutoMigrate(&model.Player{}, &model.Village{})

	v := model.Village{ID: 2, Name: "民變庄", Headman: "tyrant", Loyalty: 50}
	db.Create(&v)

	// 1. 民忠高時彈劾失敗
	_, err := ImpeachHeadman("rebel", 2)
	if err == nil { t.Error("民忠 > 30 應無法彈劾") }

	// 2. 民忠低時彈劾成功
	db.Model(&v).Update("loyalty", 20)
	msg, err := ImpeachHeadman("rebel", 2)
	if err != nil { t.Fatal(err) }
	t.Log(msg)

	var updatedV model.Village
	db.First(&updatedV, 2)
	if updatedV.Headman != "" { t.Error("庄長應被罷免") }
}
