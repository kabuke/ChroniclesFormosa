package disaster

import (
	"os"
	"testing"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
)

func TestMain(m *testing.M) {
	_ = database.InitDB(":memory:")
	database.GetDB().AutoMigrate(&model.Player{}, &model.Village{})
	
	code := m.Run()
	os.Exit(code)
}
