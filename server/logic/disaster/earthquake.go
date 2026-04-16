package disaster

import (
	"log"
	"math"
	"math/rand"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

var TriggerEarthquake = func(isHarmful bool) {
	db := database.GetDB()
	var villages []model.Village
	db.Find(&villages)

	if len(villages) == 0 {
		return
	}

	center := villages[rand.Intn(len(villages))]
	magnitude := rand.Float32()*2.0 + 1.0 // 1.0~3.0 (無害地震，輕微晃動)

	if isHarmful {
		log.Println("[Disaster] 🌋 Harmful Earthquake Triggered!")
		magnitude = rand.Float32()*4.0 + 3.0 // 3.0 ~ 7.0
	} else {
		// log.Println("[Disaster] 〰️ Harmless Earthquake (Minor Tremor).")
	}

	affected := make([]int64, 0)

	for i := range villages {
		v := &villages[i]

		dx := float64(v.X - center.X)
		dy := float64(v.Y - center.Y)
		distance := math.Sqrt(dx*dx + dy*dy)

		radius := float64(magnitude * 3.0)

		if distance <= radius {
			attenuation := 1.0 - (distance / radius)
			if attenuation < 0 {
				attenuation = 0
			}
			localMag := float32(float64(magnitude) * attenuation)

			if isHarmful && localMag >= 1.0 {
				applyEarthquakeDamage(v, localMag)
				db.Save(v)
				affected = append(affected, v.ID)
			} else if !isHarmful {
				affected = append(affected, v.ID) // 讓客戶端產生動畫
			}
		}
	}
	// 廣播災情
	env := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_Earthquake{
					Earthquake: &pb.EarthquakeNotify{
						EpicenterTileId:  int64(center.X*1000 + center.Y),
						Magnitude:        magnitude,
						AffectedVillages: affected,
						EpicenterName:    center.Name,
					},
				},
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)

	if isHarmful {
		go triggerReliefPhase(affected)
	}
}

func applyEarthquakeDamage(v *model.Village, localMag float32) {

	foodLossRate := 0.0
	loyaltyLoss := int32(0)

	if localMag < 3.0 {
		foodLossRate = 0.05
		loyaltyLoss = 5
	} else if localMag < 5.0 {
		foodLossRate = 0.15
		loyaltyLoss = 10
	} else if localMag < 7.0 {
		foodLossRate = 0.30
		loyaltyLoss = 15
	} else {
		foodLossRate = 0.40
		loyaltyLoss = 20
	}

	v.Food = int64(float64(v.Food) * (1.0 - foodLossRate))
	v.Wood = int64(float64(v.Wood) * (1.0 - foodLossRate))

	v.Loyalty -= loyaltyLoss
	if v.Loyalty < 0 {
		v.Loyalty = 0
	}

	if localMag >= 5.0 && rand.Float32() < 0.2 {
		log.Printf("[Disaster] 🌋 庄頭 %s 附近的地形發生了永久性改變！", v.Name)
	}
}

func triggerReliefPhase(affected []int64) {
	time.Sleep(5 * time.Second)

	env := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_ReliefStart{
					ReliefStart: &pb.ReliefGameStart{
						DisasterId:       time.Now().Unix(),
						AffectedVillages: affected,
					},
				},
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)
	log.Println("[Disaster] 🚑 Relief Phase Started.")
}
