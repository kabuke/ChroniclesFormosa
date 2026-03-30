package session

import (
	"sync"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// AOIManager 負責管理萬人地圖的視野(Area of Interest)廣播
type AOIManager struct {
	// key: tileID (座標換算區塊), value: map of session IDs
	tiles map[int64]map[string]struct{}
	mu    sync.RWMutex
}

var globalAOIManager *AOIManager
var aoiOnce sync.Once

func GetAOIManager() *AOIManager {
	aoiOnce.Do(func() {
		globalAOIManager = &AOIManager{
			tiles: make(map[int64]map[string]struct{}),
		}
	})
	return globalAOIManager
}

func (am *AOIManager) EnterTile(sessionID string, tileID int64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if _, ok := am.tiles[tileID]; !ok {
		am.tiles[tileID] = make(map[string]struct{})
	}
	am.tiles[tileID][sessionID] = struct{}{}
}

func (am *AOIManager) LeaveTile(sessionID string, tileID int64) {
	am.mu.Lock()
	defer am.mu.Unlock()
	if subs, ok := am.tiles[tileID]; ok {
		delete(subs, sessionID)
		if len(subs) == 0 {
			delete(am.tiles, tileID)
		}
	}
}

// BroadcastToTile 會將封包推入，僅傳送給九宮格內的玩家
func (am *AOIManager) BroadcastToTile(tileID int64, env *pb.Envelope) {
	// TODO: Phase 3 地圖系統實作
	// 1. 查詢該 tileID 及周圍鄰近 tile 的 sessionIDs
	// 2. 向 SessionManager 取得對應的 UserSession 並分發 AOIUpdate
}
