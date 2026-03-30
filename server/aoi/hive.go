package aoi

import (
	"sync"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// Hive 代表在地圖切割上的一個大範圍分流 (例如: 北台灣、中台灣)
// Hive 會管理自己底下的所有玩家，並將該區的移動頻繁廣播給此區所有人
type Hive struct {
	Name    string
	players map[string]*PlayerEntity
	mu      sync.RWMutex
}

// PlayerEntity 在地圖上移動的世界物件
type PlayerEntity struct {
	Username string
	X        float32
	Y        float32
	Session  *session.UserSession
	Dirty    bool // 是否有新的移動待廣播
}

// NewHive 建立一個並行安全的區域伺服節點
func NewHive(name string) *Hive {
	h := &Hive{
		Name:    name,
		players: make(map[string]*PlayerEntity),
	}
	// 每隔 100ms 執行一次打包與廣播 (相當於 10 TPS 伺服器幀率)
	go h.broadcastLoop()
	return h
}

// AddPlayer 新增或覆寫玩家的實體參考
func (h *Hive) AddPlayer(username string, x, y float32, s *session.UserSession) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.players[username] = &PlayerEntity{
		Username: username,
		X:        x,
		Y:        y,
		Session:  s,
		Dirty:    true, // 剛加進來，廣播一次位置，讓別人看見
	}
	// log.Printf("[Hive:%s] Player %s entered the zone. X:%.2f, Y:%.2f", h.Name, username, x, y)
}

// RemovePlayer 玩家斷線或離開此區
func (h *Hive) RemovePlayer(username string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.players[username]; ok {
		delete(h.players, username)
		// log.Printf("[Hive:%s] Player %s left the zone.", h.Name, username)
	}
}

// UpdatePosition 只更新玩家座標
func (h *Hive) UpdatePosition(username string, x, y float32) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if p, ok := h.players[username]; ok {
		// 如果沒有移動，就不重發 (節省頻寬)
		if p.X != x || p.Y != y {
			p.X = x
			p.Y = y
			p.Dirty = true
		}
	}
}

// broadcastLoop 是單一一個地區內專屬的廣播網路，每秒 10 幀。
// 這個專屬 Loop 確保我們不會因為其它地區的卡頓而影響到這裡的廣播效能。
func (h *Hive) broadcastLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // 10 TPS
	defer ticker.Stop()

	for range ticker.C {
		h.broadcastTick()
	}
}

func (h *Hive) broadcastTick() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.players) == 0 {
		return // 此區無人，什麼都不做
	}

	// 1. 蒐集本幀(Tick)內所有移動過的(Dirty)實體
	var updates []*pb.AOIBroadcast_EntityPos
	for _, p := range h.players {
		if p.Dirty {
			updates = append(updates, &pb.AOIBroadcast_EntityPos{
				Username: p.Username,
				X:        p.X,
				Y:        p.Y,
			})
			p.Dirty = false // 清洗標記
		}
	}

	if len(updates) == 0 {
		return // 大家都掛機，沒人動，就省頻寬
	}

	// 2. 組裝 Protobuf: AOIBroadcast (包含這一幀所有有動靜的人的座標)
	syncMsg := &pb.Envelope{
		Payload: &pb.Envelope_AoiBroadcast{
			AoiBroadcast: &pb.AOIBroadcast{
				Entities: updates,
			},
		},
	}

	// 3. 把打包好的一包更新資料，派發給這個大區塊裡所有的 UserSession
	for _, p := range h.players {
		if p.Session != nil {
			p.Session.QueueMessage(syncMsg)
			if p.Session.TriggerFlush != nil {
				p.Session.TriggerFlush()
			}
		}
	}
}
