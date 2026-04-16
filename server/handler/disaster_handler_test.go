package handler

import (
	"testing"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

func TestReliefDonateFlow(t *testing.T) {
	db := database.GetDB()
	db.Exec("DELETE FROM players")
	db.Exec("DELETE FROM villages")

	db.Create(&model.Player{Username: "donate_user", Nickname: "Donator", Stamina: 100})
	// Assume stamina system initialized player stamina to max
	db.Create(&model.Village{ID: 10, Name: "Target", Food: 100})

	sess := &session.UserSession{
		SessionID: "test-session-donate",
		Username:  "donate_user",
	}

	req := &pb.DisasterAction{
		Action: &pb.DisasterAction_ReliefDonate{
			ReliefDonate: &pb.ReliefDonateReq{
				TargetVillageId: 10,
				ResourceAmount:  10,
			},
		},
	}

	HandleDisasterAction(req, sess)

	// Since HandleReliefDonate doesn't queue any response directly through returning it,
	// we just read the DB to verify.
	var v model.Village
	db.First(&v, 10)
	if v.Food != 120 { // 100 + 10 * 2 (1 stamina = 2 food)
		t.Errorf("Expected village food 120, got %d", v.Food)
	}
	if v.Loyalty != 81 { // base 80 + 1
		t.Errorf("Expected village loyalty 81, got %d", v.Loyalty)
	}
}

func TestReliefRouteSubmitFlow(t *testing.T) {
	db := database.GetDB()
	db.Exec("DELETE FROM villages")
	db.Create(&model.Village{ID: 11, Name: "MyVillage", Headman: "route_user", Food: 1000, Loyalty: 50})

	sess := &session.UserSession{
		SessionID: "test-session-route",
		Username:  "route_user",
		VillageID: 11,
	}

	req := &pb.DisasterAction{
		Action: &pb.DisasterAction_ReliefRoute{
			ReliefRoute: &pb.ReliefRouteSubmit{
				TargetVillageId: 11,
				CoveragePercent: 0.9,
				TimeTakenMs:     12000,
				RouteDistance:   1200,
			},
		},
	}

	// Make sure the session manager knows about the session for broadcast
	sm := session.GetManager()
	sm.AddSession(sess)
	// We wait slightly to allow registration, but for a synchronous test, we can just manipulate it.
	// Actually no need, because HandleReliefRouteSubmit uses broadcastReliefResult which looks it up.
	// Let's rely on Outbox testing or sleep.

	HandleDisasterAction(req, sess)
}
