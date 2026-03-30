package session

import (
	"sync"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// ChannelManager 負責管理全服頻道、莊頭頻道的訂閱狀態
type ChannelManager struct {
	// key: channel name (e.g., "#village_10"), value: map of session IDs
	channels map[string]map[string]struct{}
	mu       sync.RWMutex
}

var globalChannelManager *ChannelManager
var chanOnce sync.Once

func GetChannelManager() *ChannelManager {
	chanOnce.Do(func() {
		globalChannelManager = &ChannelManager{
			channels: make(map[string]map[string]struct{}),
		}
	})
	return globalChannelManager
}

func (cm *ChannelManager) Subscribe(sessionID, channel string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.channels[channel]; !ok {
		cm.channels[channel] = make(map[string]struct{})
	}
	cm.channels[channel][sessionID] = struct{}{}
}

func (cm *ChannelManager) Unsubscribe(sessionID, channel string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if subs, ok := cm.channels[channel]; ok {
		delete(subs, sessionID)
		if len(subs) == 0 {
			delete(cm.channels, channel)
		}
	}
}

// BroadcastToChannel 會將封包轉發對象限定於該頻道的訂閱者
func (cm *ChannelManager) BroadcastToChannel(channel string, env *pb.Envelope) {
	// TODO: Phase 3 聊天系統實作時，會與 SessionManager 的 Actor 深度整合
	// 1. 取得該頻道所有 sessionIDs
	// 2. 向 SessionManager 取得對應的 UserSession 並寫入 Queue
}
