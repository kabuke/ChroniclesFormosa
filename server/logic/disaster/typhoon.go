package disaster

import (
	"log"
	"math/rand"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

var TriggerTyphoon = func() {
	log.Println("[Disaster] 🌀 Typhoon Landed!")

	db := database.GetDB()
	var villages []model.Village
	db.Find(&villages)

	// 模擬颱風路徑：從東岸到西岸或沿海
	// 這裡簡單隨機選 3-6 個庄頭作為受災區
	affectedCount := rand.Intn(4) + 3
	if affectedCount > len(villages) {
		affectedCount = len(villages)
	}

	// 隨機打亂 villages 來選取受災庄頭
	rand.Shuffle(len(villages), func(i, j int) {
		villages[i], villages[j] = villages[j], villages[i]
	})

	affected := make([]int64, 0)
	var pathTiles []int64

	// 颱風影響受災區農業，農田產量歸零 (扣除現有糧食的一半，模擬災情)
	for i := 0; i < affectedCount; i++ {
		v := &villages[i]
		v.Food = v.Food / 2
		db.Save(v)
		affected = append(affected, v.ID)
		pathTiles = append(pathTiles, int64(v.X*1000 + v.Y))
	}

	intensity := rand.Float32()*3.0 + 1.0 // 1-4 級

	env := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_Typhoon{
					Typhoon: &pb.TyphoonNotify{
						PathTiles: pathTiles,
						Intensity: intensity,
						Phase:     pb.TyphoonPhase_PHASE_LANDING,
						AffectedVillages: affected,
						PathDesc: "由東部海域登陸，橫掃全島",
					},
				},
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)

	// 觸發救災階段
	go triggerReliefPhase(affected)

	// 依照遊戲時間流速，颱風經過約需 1~2 個遊戲天 (1440~2880遊戲分鐘)
	// 根據 TimeRatio: 1 實秒 = 20 遊戲分 -> 72~144 實秒
	durationSecs := 72 + rand.Intn(73)
	time.AfterFunc(time.Duration(durationSecs)*time.Second, func() {
		log.Printf("[Disaster] ☀️ Typhoon dissipated after %d real seconds.", durationSecs)
		clearEnv := &pb.Envelope{
			Payload: &pb.Envelope_Disaster{
				Disaster: &pb.DisasterAction{
					Action: &pb.DisasterAction_Typhoon{
						Typhoon: &pb.TyphoonNotify{
							Intensity: 0.0, // 強度 0 作為清除天氣動畫的信號
							PathDesc:  "颱風已經遠離，天氣放晴",
						},
					},
				},
			},
		}
		session.GetManager().AddToForwardQueue(clearEnv)
	})
}
