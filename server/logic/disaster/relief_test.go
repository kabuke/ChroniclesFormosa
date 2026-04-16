package disaster

import (
	"testing"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/database"
	"github.com/kabuke/ChroniclesFormosa/server/model"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

// makeTestSession 建立一個已注冊到 SessionManager 的測試 Session
func makeTestSession(id string, username string, villageID int64) *session.UserSession {
	sm := session.GetManager()
	s := sm.CreateSession(id, []byte("test-secret"))
	s.Username = username
	s.VillageID = villageID
	return s
}

// findReliefResult 從 Outbox 找到最後一個 ReliefResult
func findReliefResult(s *session.UserSession) *pb.ReliefResult {
	for i := len(s.Outbox) - 1; i >= 0; i-- {
		if r := s.Outbox[i].GetDisaster().GetReliefResult(); r != nil {
			return r
		}
	}
	return nil
}

func TestReliefScoringAndReward(t *testing.T) {
	db := database.GetDB()
	db.Exec("DELETE FROM villages")
	db.Exec("DELETE FROM players")

	db.Create(&model.Player{Username: "tester_rel", Nickname: "Tester"})

	// ── 1. PERFECT grade ──────────────────────────────────────────────
	db.Create(&model.Village{ID: 101, Food: 1000, Loyalty: 50, Headman: "tester_rel"})

	sess1 := makeTestSession("rel-sess-1", "tester_rel", 101)

	_, err := HandleReliefRouteSubmit(sess1, &pb.ReliefRouteSubmit{
		TargetVillageId: 101,
		CoveragePercent: 1.0,   // score = 100
		TimeTakenMs:     10000, // < 15s → no penalty
		RouteDistance:   1000,  // < 1500 → no penalty
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	res1 := findReliefResult(sess1)
	if res1 == nil {
		t.Fatal("Expected ReliefResult in outbox, got nil")
	}
	if res1.Score != 100 || res1.Grade != pb.ReliefGrade_GRADE_PERFECT {
		t.Errorf("Expected PERFECT/100, got %v/%d", res1.Grade, res1.Score)
	}

	var v1 model.Village
	db.First(&v1, 101)
	if v1.Food != 2000 { // 1000 + 100*10 = 2000
		t.Errorf("Expected food 2000, got %d", v1.Food)
	}
	if v1.Loyalty != 70 { // 50 + 20
		t.Errorf("Expected loyalty 70, got %d", v1.Loyalty)
	}

	// ── 2. GOOD grade ─────────────────────────────────────────────────
	db.Create(&model.Village{ID: 102, Food: 1000, Loyalty: 50, Headman: "tester_rel"})

	sess2 := makeTestSession("rel-sess-2", "tester_rel", 102)

	_, _ = HandleReliefRouteSubmit(sess2, &pb.ReliefRouteSubmit{
		TargetVillageId: 102,
		CoveragePercent: 0.7,   // score = 70 → no penalties → GOOD
		TimeTakenMs:     15000,
		RouteDistance:   1500,
	})

	res2 := findReliefResult(sess2)
	if res2 == nil {
		t.Fatal("Expected ReliefResult for GOOD, got nil")
	}
	if res2.Grade != pb.ReliefGrade_GRADE_GOOD {
		t.Errorf("Expected GOOD grade, got %v", res2.Grade)
	}

	var v2 model.Village
	db.First(&v2, 102)
	expectedFood2 := int64(1000 + (70*10)/2) // 1350
	if v2.Food != expectedFood2 {
		t.Errorf("Expected food %d, got %d", expectedFood2, v2.Food)
	}

	// ── 3. FAIL grade ─────────────────────────────────────────────────
	db.Create(&model.Village{ID: 103, Food: 1000, Loyalty: 10, Headman: "tester_rel"})

	sess3 := makeTestSession("rel-sess-3", "tester_rel", 103)

	_, _ = HandleReliefRouteSubmit(sess3, &pb.ReliefRouteSubmit{
		TargetVillageId: 103,
		CoveragePercent: 0.2,   // 20 - penalty(45) - penalty(30) → 0
		TimeTakenMs:     30000,
		RouteDistance:   3000,
	})

	res3 := findReliefResult(sess3)
	if res3 == nil {
		t.Fatal("Expected ReliefResult for FAIL, got nil")
	}
	if res3.Grade != pb.ReliefGrade_GRADE_FAIL || res3.Score != 0 {
		t.Errorf("Expected FAIL/0, got %v/%d", res3.Grade, res3.Score)
	}

	var v3 model.Village
	db.First(&v3, 103)
	if v3.Loyalty != 0 { // 10 - 20 → clamped to 0
		t.Errorf("Expected loyalty 0, got %d", v3.Loyalty)
	}
}
