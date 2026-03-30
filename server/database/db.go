package database

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	globalDB *gorm.DB
	once     sync.Once
)

// InitDB 根據傳入的 dbPath 建立資料夾並初始化 GORM SQLite 連線
func InitDB(dbPath string) error {
	var initErr error
	once.Do(func() {
		// 確保 DB 目錄存在
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				initErr = err
				return
			}
		}

		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn), // 測試期間可考慮使用 Info
		})
		if err != nil {
			initErr = err
			return
		}
		
		globalDB = db
		log.Printf("[Database] Pure-Go SQLite initialized at: %s", dbPath)
	})
	return initErr
}

// GetDB 取得 Global DB 實例 (使用前必須呼叫 InitDB)
func GetDB() *gorm.DB {
	if globalDB == nil {
		log.Fatal("[Database] Fatal: GetDB() called before InitDB()")
	}
	return globalDB
}
