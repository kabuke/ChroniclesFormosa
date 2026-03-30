package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerPort          int    `json:"server_port"`
	ServerAddress       string `json:"server_address"`
	DatabasePath        string `json:"database_path"`
	AppWindowSize       int    `json:"app_window_size"`
	MaxPendingQueueSize int    `json:"max_pending_queue_size"`
	KcpNoDelay          int    `json:"kcp_nodelay"`
	KcpInterval         int    `json:"kcp_interval"`
	KcpResend           int    `json:"kcp_resend"`
	KcpNc               int    `json:"kcp_nc"`
}

var AppConfig *Config

// LoadConfig 讀取設定檔並優先考慮 Command Line Arguments (--port, --db)
func LoadConfig(configPath string) error {
	cfg := &Config{
		// 預設值
		ServerPort:          8999,
		ServerAddress:       "0.0.0.0",
		DatabasePath:        "formosa.db",
		AppWindowSize:       128,
		MaxPendingQueueSize: 1024,
		KcpNoDelay:          1,
		KcpInterval:         20,
		KcpResend:           2,
		KcpNc:               1,
	}

	// 1. 嘗試從 JSON 檔案載入
	if file, err := os.Open(configPath); err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return fmt.Errorf("解析 config.json 失敗: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("開啟 config.json 錯誤: %w", err)
	}

	// 2. 解析命令列參數 (覆寫優先級最高)
	var portFlag int
	var dbFlag string

	// 防止重複定義 flag (有助於 testing 或熱重載)
	if flag.Lookup("port") == nil {
		flag.IntVar(&portFlag, "port", 0, "Server port (overrides config.json)")
	}
	if flag.Lookup("db") == nil {
		flag.StringVar(&dbFlag, "db", "", "Database path (overrides config.json)")
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	if portFlag != 0 {
		cfg.ServerPort = portFlag
	}
	if dbFlag != "" {
		cfg.DatabasePath = dbFlag
	}

	AppConfig = cfg
	return nil
}
