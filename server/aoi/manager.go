package aoi

import (
	"log"
	"sync"

	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// Manager 是一個 Singleton 協調器，用來路由坐標到具體的網格/區域(Hive)中。
type Manager struct {
	hives      map[string]*Hive
	playerHive map[string]string // 記錄哪一個玩家目前落在哪個大區 (Username -> HiveName)
	mu         sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			hives:      make(map[string]*Hive),
			playerHive: make(map[string]string),
		}
		// 預先建立 5 個獨立運行的戰區引擎
		instance.hives["Keelung"] = NewHive("Keelung")
		instance.hives["Taichung"] = NewHive("Taichung")
		instance.hives["Tainan"] = NewHive("Tainan")
		instance.hives["Taitung"] = NewHive("Taitung")
		instance.hives["Penghu"] = NewHive("Penghu")
	})
	return instance
}

// targetHive 以一個極致簡化的粗暴判斷方式，將台灣緯度切割為 5 大塊。
// X 為經度，Y 為緯度。
func targetHive(x, y float32) string {
	if x < 119.8 {
		return "Penghu" // 離島
	}
	if y >= 24.8 {
		return "Keelung" // 北部 (台北/基隆/宜蘭/桃園)
	}
	if y >= 23.8 {
		return "Taichung" // 中部 (新竹/苗栗/台中/彰化/南投)
	}
	if x >= 120.7 {
		return "Taitung" // 東部 (花蓮/台東)
	}
	return "Tainan" // 南部 (雲林/嘉義/台南/高雄/屏東)
}

// MovePlayer 是外面接收到 Envelope_AoiUpdate 時呼叫的。負責處理玩家的分區路由。
func (m *Manager) MovePlayer(username string, x, y float32, s *session.UserSession) {
	if username == "" {
		return
	}

	target := targetHive(x, y)
	
	m.mu.Lock()
	current, exists := m.playerHive[username]

	// 如果他原本隸屬的地區，跟最新座標的目標區不一樣 (發生跨區/初次上線)
	if !exists || current != target {
		// 把它從舊區拔除
		if exists {
			if h, ok := m.hives[current]; ok {
				h.RemovePlayer(username)
			}
		}
		// 加到新區
		m.playerHive[username] = target
		m.mu.Unlock() // Hive 本身有 mutex，走到這裡就能放開了
		
		if h, ok := m.hives[target]; ok {
			h.AddPlayer(username, x, y, s)
			log.Printf("[AOIManager] 🌍 Player '%s' migrated to Zone: %s (X=%.2f, Y=%.2f)", username, target, x, y)
		}
		return
	}
	m.mu.Unlock()

	// 沒有跨區，單純更新他目前在該區的座標位置即可
	if h, ok := m.hives[target]; ok {
		h.UpdatePosition(username, x, y)
	}
}

// RemovePlayer 負責在玩家斷線時清理其遺留在世界地圖上的殘留點
func (m *Manager) RemovePlayer(username string) {
	if username == "" {
		return
	}

	m.mu.Lock()
	current, exists := m.playerHive[username]
	if exists {
		delete(m.playerHive, username)
	}
	m.mu.Unlock()

	if exists {
		if h, ok := m.hives[current]; ok {
			h.RemovePlayer(username)
			log.Printf("[AOIManager] 💀 Player '%s' removed from Zone: %s", username, current)
		}
	}
}
