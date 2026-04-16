package main

import (
	"fmt"
	"log"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/logic/disaster"
	"github.com/kabuke/ChroniclesFormosa/server/logic/faction"
	"github.com/kabuke/ChroniclesFormosa/server/logic/social"
	"github.com/kabuke/ChroniclesFormosa/server/logic/stamina"
	"github.com/kabuke/ChroniclesFormosa/server/logic/village"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/network"
	"github.com/xtaci/kcp-go/v5"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化資料庫 (路徑與之前 cmd/server/main.go 保持一致)
	dbPath := "./data/formosa.db"
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("[Server] Database init failed: %v", err)
	}
	db := database.GetDB()
	
	// 2. 自動遷移模型
	db.AutoMigrate(&model.Player{}, &model.Village{}, &model.DiplomacyRelation{}, &model.SessionState{})

	// 3. 植入種子資料
	SeedVillages(db)

	// 4. 啟動背景引擎
	go village.StartEconomyEngine()
	go social.StartTensionEngine()
	go stamina.StartStaminaTicker()
	disaster.StartDisasterTimer()
	
	// 5. 陣營平衡初始化
	faction.RebalanceFactions()

	// 6. 啟動 KCP 監聽
	addr := ":8999"
	listener, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	if err != nil {
		log.Fatalf("[Server] KCP Listen failed: %v", err)
	}

	log.Printf("[Server] 🇹🇼 Chronicles Formosa KCP Started at %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		kcpConn := conn.(*kcp.UDPSession)
		kcpConn.SetStreamMode(true)
		kcpConn.SetWindowSize(1024, 1024)
		kcpConn.SetNoDelay(1, 10, 2, 1)
		
		// 處理連線
		go network.HandleConnection(kcpConn)
	}
}

// SeedVillages 植入台灣十六庄預設資料
func SeedVillages(db *gorm.DB) {
	var count int64
	db.Model(&model.Village{}).Count(&count)
	
	if count < 16 {
		log.Println("[Database] 🛖 Villages count incorrect, re-seeding all with fixed IDs...")
		db.Exec("DELETE FROM villages")
		db.Exec("DELETE FROM sqlite_sequence WHERE name='villages'")

		villages := []model.Village{
			{ID: 1, Name: "雞籠 (Keelung)", X: 64, Y: 10, FactionID: 1, Level: 1},
			{ID: 2, Name: "艋舺 (Bangka)", X: 58, Y: 15, FactionID: 1, Level: 1},
			{ID: 3, Name: "竹塹 (Hsinchu)", X: 48, Y: 25, FactionID: 0, Level: 1},
			{ID: 4, Name: "大甲 (Dajia)", X: 42, Y: 35, FactionID: 0, Level: 1},
			{ID: 5, Name: "鹿港 (Lukang)", X: 36, Y: 45, FactionID: 2, Level: 1},
			{ID: 6, Name: "北港 (Beigang)", X: 35, Y: 52, FactionID: 2, Level: 1},
			{ID: 7, Name: "諸羅 (Chiayi)", X: 40, Y: 55, FactionID: 1, Level: 1},
			{ID: 8, Name: "府城 (Anping)", X: 39, Y: 65, FactionID: 1, Level: 2},
			{ID: 9, Name: "打狗 (Dagao)", X: 43, Y: 75, FactionID: 2, Level: 1},
			{ID: 10, Name: "阿猴 (Pingtung)", X: 47, Y: 80, FactionID: 3, Level: 1},
			{ID: 11, Name: "澎湖 (Penghu)", X: 15, Y: 53, FactionID: 1, Level: 1},
			{ID: 12, Name: "蛤仔難 (Yilan)", X: 65, Y: 20, FactionID: 0, Level: 1},
			{ID: 13, Name: "卑南 (Taitung)", X: 56, Y: 70, FactionID: 3, Level: 1},
			{ID: 14, Name: "花蓮港 (Hualien)", X: 65, Y: 45, FactionID: 3, Level: 1},
			{ID: 15, Name: "瑯嶠 (Hengchun)", X: 51, Y: 95, FactionID: 0, Level: 1},
			{ID: 16, Name: "埔里社 (Puli)", X: 50, Y: 45, FactionID: 0, Level: 1},
		}

		for _, v := range villages {
			db.Create(&v)

			// 植入 10 名假村民作為測試對象
			for i := 1; i <= 10; i++ {
				botName := fmt.Sprintf("bot_v%d_%02d", v.ID, i)
				bot := model.Player{
					Username:    botName,
					PasswordHash: "dummy", // 機器人不需登入
					Nickname:    fmt.Sprintf("村民%d號", i),
					FactionID:   v.FactionID,
					VillageID:   int64(v.ID),
					VillageRole: 0,
				}
				db.Where(model.Player{Username: botName}).FirstOrCreate(&bot)
			}
		}
		
		db.Exec("UPDATE players SET village_id = 0 WHERE username NOT LIKE 'bot_%'")
		log.Println("[Database] 🛡️ All players' village assignments reset (except bots) for re-testing.")
	}
}
