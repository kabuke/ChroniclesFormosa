package main

import (
	"fmt"
	"log"

	"github.com/kabuke/ChroniclesFormosa/config"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/network"
	"github.com/xtaci/kcp-go/v5"
)

func main() {
	// 1. 載入配置檔
	if err := config.LoadConfig("config/config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 2. 初始化 Pure-Go SQLite 資料庫
	if err := database.InitDB(config.AppConfig.DatabasePath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	// 自動遷移 Schema
	database.GetDB().AutoMigrate(&model.Player{}, &model.Village{}, &model.SessionState{})

	// Seed Initial Villages (Phase 1 測試用)
	var villageCount int64
	database.GetDB().Model(&model.Village{}).Count(&villageCount)
	if villageCount == 0 {
		database.GetDB().Create([]model.Village{
			{Name: "打狗", Level: 1, FactionID: 2},
			{Name: "諸羅", Level: 1, FactionID: 1},
			{Name: "竹塹", Level: 1, FactionID: 0},
		})
		log.Println("[Database] 🛖 Seeded initial 3 villages (打狗, 諸羅, 竹塹).")
	}

	// 3. 建立 KCP Listener
	addr := fmt.Sprintf("%s:%d", config.AppConfig.ServerAddress, config.AppConfig.ServerPort)
	listener, err := kcp.ListenWithOptions(addr, nil, 10, 3)
	if err != nil {
		log.Fatalf("Failed to start server on %s: %v", addr, err)
	}

	log.Printf("==== Chronicles Formosa Server Started on KCP %s ====\n", addr)
	
	// 3. 阻塞等待連線
	for {
		conn, err := listener.AcceptKCP()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		
		log.Printf("[Listener] Accepted connection from %s\n", conn.RemoteAddr())
		// 非同步將連線拋給 Gateway 層處理明文握手與密文傳輸
		go network.HandleConnection(conn)
	}
}
