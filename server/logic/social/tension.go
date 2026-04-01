package social

import (
	"errors"
	"fmt"
	"log"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/repo"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// HandleStabilityOperation 處理庄長的維穩操作邏輯
func HandleStabilityOperation(username string, villageID int64, opType pb.StabilityOpType) (string, error) {
	vRepo := repo.NewVillageRepo()
	v, err := vRepo.FindByID(villageID)
	if err != nil || v == nil { return "", errors.New("找不到指定的庄頭") }

	// 🇹🇼 直接檢查庄頭表中的 Headman，確保推舉成功後權限立即生效
	if v.Headman != username {
		return "", errors.New("權限不足：僅庄長可發起社會維穩操作")
	}

	switch opType {
	case pb.StabilityOpType_OP_BANQUET:
		if v.Food < 50 { return "", errors.New("糧食不足 (需 50)") }
		v.Food -= 50
		v.TensionValue -= 15
		if v.TensionValue < 0 { v.TensionValue = 0 }
		_ = vRepo.Update(v)
		broadcastTension(v)
		return fmt.Sprintf("「大辦筵席！」%s 庄長設宴款待，緊張氣氛大為緩和。", v.Name), nil
	case pb.StabilityOpType_OP_RITUAL:
		if v.Wood < 50 { return "", errors.New("木材不足 (需 50)") }
		v.Wood -= 50
		v.Loyalty += 10
		if v.Loyalty > 100 { v.Loyalty = 100 }
		_ = vRepo.Update(v)
		return fmt.Sprintf("「祭祀先祖！」%s 庄長焚香祭拜，聚落民忠提升。", v.Name), nil
	default:
		return "", errors.New("未知的操作類型")
	}
}

func broadcastTension(v *model.Village) {
	level := "PEACE"
	if v.TensionValue > 80 { level = "RIOT" } else if v.TensionValue > 60 { level = "TENSE" } else if v.TensionValue > 30 { level = "UNEASY" }
	
	env := &pb.Envelope{
		Payload: &pb.Envelope_Tension{
			Tension: &pb.TensionUpdate{
				VillageId: v.ID, TensionValue: v.TensionValue, VisualLevel: level,
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)
}

func StartTensionEngine() {
	ticker := time.NewTicker(30 * time.Second)
	log.Println("[TensionEngine] 🧨 Ethnic Tension Engine Started. Tick: 30s.")
	go func() {
		for range ticker.C {
			processTensionTick()
		}
	}()
}

func processTensionTick() {
	db := database.GetDB()
	var villages []model.Village
	db.Find(&villages)

	for _, v := range villages {
		totalPop := float64(v.PopMinNan + v.PopHakka + v.PopIndigenous)
		if totalPop == 0 { continue }

		// 計算人口異質度：比例越平均，張力基礎越高
		p1, p2, p3 := float64(v.PopMinNan)/totalPop, float64(v.PopHakka)/totalPop, float64(v.PopIndigenous)/totalPop
		diversity := 1.0 - (p1*p1 + p2*p2 + p3*p3) 
		
		delta := int32(diversity * 10) // 最大 +6~7 每 Tick
		
		// 糧食壓力
		if v.Food < int64(totalPop/2) { delta += 5 }
		
		v.TensionValue += delta
		if v.TensionValue >= 100 {
			v.TensionValue = 100
			triggerRiot(&v)
		}
		db.Save(&v)
		broadcastTension(&v)
	}
}

func triggerRiot(v *model.Village) {
	msg := fmt.Sprintf("【分類械鬥】%s 爆發族群衝突！人口受損，資源遭劫。", v.Name)
	v.PopMinNan = int32(float64(v.PopMinNan) * 0.9)
	v.PopHakka = int32(float64(v.PopHakka) * 0.9)
	v.PopIndigenous = int32(float64(v.PopIndigenous) * 0.9)
	v.Food /= 2
	
	env := &pb.Envelope{
		Payload: &pb.Envelope_Chat{
			Chat: &pb.ChatMessage{Channel: pb.ChatChannelType_CHANNEL_GLOBAL, Sender: "廟口說書人", Content: msg},
		},
	}
	session.GetManager().AddToForwardQueue(env)
}
