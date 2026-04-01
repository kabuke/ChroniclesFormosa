package stamina

import (
	"log"
	"time"

	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

// StartStaminaTicker 啟動全服玩家精力恢復計時器
func StartStaminaTicker() {
	// 每 5 分鐘恢復 5 點
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			RestoreAll(5)
		}
	}()
	log.Println("[Stamina] ⚡ Recovery Ticker Started.")
}

func RestoreAll(amount int32) {
	db := database.GetDB()
	// 原子累加：stamina = MIN(100, stamina + amount)
	db.Model(&model.Player{}).Where("stamina < ?", 100).Update("stamina", database.GetDB().Raw("MIN(100, stamina + ?)", amount))
	
	sm := session.GetManager()
	sessions := sm.GetAllSessions()
	for _, s := range sessions {
		if s.Username == "" { continue }
		var p model.Player
		db.Where("username = ?", s.Username).First(&p)
		SyncStamina(s, &p)
	}
}

func ConsumeStamina(p *model.Player, amount int32) bool {
	if p.Stamina < amount { return false }
	p.Stamina -= amount
	return true
}

func SyncStamina(s *session.UserSession, p *model.Player) {
	env := &pb.Envelope{
		Payload: &pb.Envelope_Stamina{
			Stamina: &pb.StaminaUpdate{
				Current: p.Stamina,
				Max:     100,
			},
		},
	}
	s.QueueMessage(env)
	if s.TriggerFlush != nil { go s.TriggerFlush() }
}
