package disaster

import (
	"log"
	"math/rand"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// TriggerPlague 觸發瘟疫
func TriggerPlague() {
	log.Println("[Disaster] 🦠 Plague Outbreak!")

	db := database.GetDB()
	var villages []model.Village
	db.Find(&villages)

	// 隨機選一個庄頭爆發
	if len(villages) == 0 {
		return
	}
	v := &villages[rand.Intn(len(villages))]

	// 影響人口或民忠
	v.Loyalty -= 20
	if v.Loyalty < 0 { v.Loyalty = 0 }
	db.Save(v)

	env := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_Warning{
					Warning: &pb.DisasterWarning{
						Type:             pb.DisasterType_DISASTER_PLAGUE,
						Message:          "WARNING_PLAGUE|" + v.Name,
						CountdownSeconds: 0,
					},
				},
			},
		},
	}
	session.GetManager().AddToForwardQueue(env)

	go triggerReliefPhase([]int64{v.ID})
}
