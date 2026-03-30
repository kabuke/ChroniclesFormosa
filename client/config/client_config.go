package config

// ClientConfig 定義客戶端運行時的初始設定值
type ClientConfig struct {
	ServerAddress string
	ScreenWidth   int
	ScreenHeight  int
}

// AppConfig 為全域單例設定，載入後只讀
var AppConfig *ClientConfig

// LoadConfig 讀取設定 (目前直接寫死，未來可從 client.json 讀取)
func LoadConfig() {
	AppConfig = &ClientConfig{
		ServerAddress: "127.0.0.1:8999", // 連接剛才寫好伺服器的本機測試
		ScreenWidth:   1280,             // 固定基礎繪製解析度 16:9
		ScreenHeight:  720,
	}
}
