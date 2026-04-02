package disaster

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "github.com/kabuke/ChroniclesFormosa/resource"
	"github.com/kabuke/ChroniclesFormosa/server/session"
)

var (
	InGameTime  time.Time
	TimeRatio   float64 = 1217.5 // 1 real sec = ~20 in-game minutes
	SeasonStart time.Time
)

// StartDisasterTimer 啟動天災排程引擎與時間流速
func StartDisasterTimer() {
	// 台灣三國誌第一賽季起始時間: 現實的 4 月 2 日 00:00:00 (以當前年份計算)
	now := time.Now()
	SeasonStart = time.Date(now.Year(), 4, 2, 0, 0, 0, 0, time.Local)
	
	// 如果當前時間早於 4/2，則往前推一年作為賽季起點
	if now.Before(SeasonStart) {
		SeasonStart = time.Date(now.Year()-1, 4, 2, 0, 0, 0, 0, time.Local)
	}

	// 遊戲內起始時間: 1600 年 1 月 1 日 00:00:00
	baseInGameTime := time.Date(1600, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 每 1 秒 (現實時間) 推進時間
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			// 計算經過的現實時間
			elapsedReal := time.Since(SeasonStart).Seconds()
			if elapsedReal < 0 { elapsedReal = 0 }
			
			// 換算遊戲內時間
			InGameTime = baseInGameTime.Add(time.Duration(elapsedReal * TimeRatio) * time.Second)
			
			advanceTime()
		}
	}()
	log.Printf("[Disaster] 🌪️ Time Engine Started. Season 1 Start: %s", SeasonStart.Format("2006-01-02"))
}

func advanceTime() {

	month := InGameTime.Month()
	hour := InGameTime.Hour()
	year := InGameTime.Year()

	var timeOfDay string
	switch {
	case hour >= 6 && hour < 12:
		timeOfDay = "早上"
	case hour >= 12 && hour < 18:
		timeOfDay = "中午/下午"
	case hour >= 18 && hour < 24:
		timeOfDay = "晚上"
	default:
		timeOfDay = "凌晨"
	}

	if year > 1900 {
		return // 遊戲時間結束或進入下一個賽季
	}

	// 廣播時間同步
	timeSyncEnv := &pb.Envelope{
		Payload: &pb.Envelope_TimeSync{
			TimeSync: &pb.TimeSync{
				Year:      int32(year),
				Month:     int32(month),
				Day:       int32(InGameTime.Day()),
				TimeOfDay: timeOfDay,
			},
		},
	}
	session.GetManager().AddToForwardQueue(timeSyncEnv)

	evaluateDisaster(month, timeOfDay)
}

func evaluateDisaster(month time.Month, timeOfDay string) {
	// 機率計算: 一個賽季 90 天，對應遊戲內 300 年
	// 每現實秒約前進 20 分鐘
	// 每年 = 25920 真實秒
	probHarmlessEQ := 500.0 / 25920.0
	probHarmfulEQ := 4.0 / 25920.0
	probTyphoon := 0.0
	if month >= 5 && month <= 10 {
		probTyphoon = 7.0 / 12960.0 // 集中在半年內
	}

	r := rand.Float64()

	// 幫助玩家觀察時間流逝
	if InGameTime.Day() == 1 && InGameTime.Hour() == 0 && InGameTime.Minute() < 30 {
		log.Printf("[TimeEngine] 🕰️ 遊戲時間: %04d年 %02d月 %s", InGameTime.Year(), month, timeOfDay)
	}

	// 1. 判定颱風
	if probTyphoon > 0 && r < probTyphoon {
		msg := fmt.Sprintf("【%04d年%02d月 %s】海上颱風警報！強烈颱風即將登陸，請盡速儲備物資並準備救災。", InGameTime.Year(), month, timeOfDay)
		triggerWarning(pb.DisasterType_DISASTER_TYPHOON, msg)
		time.AfterFunc(30*time.Second, func() {
			TriggerTyphoon()
		})
		return
	}

	// 2. 判定有害地震
	if r < probTyphoon+probHarmfulEQ {
		msg := fmt.Sprintf("【%04d年%02d月 %s】地牛翻身預警！預計不久將發生強烈有感地震，請各庄頭做好準備。", InGameTime.Year(), month, timeOfDay)
		triggerWarning(pb.DisasterType_DISASTER_EARTHQUAKE, msg)
		time.AfterFunc(30*time.Second, func() {
			TriggerEarthquake(true) // 有害
		})
		return
	}

	// 3. 判定無害地震
	if r < probTyphoon+probHarmfulEQ+probHarmlessEQ {
		TriggerEarthquake(false) // 無害，不預警，純搖晃動畫
	}
}

func triggerWarning(disasterType pb.DisasterType, msg string) {
	log.Printf("[Disaster] ⚠️ Warning: %s", msg)
	
	env := &pb.Envelope{
		Payload: &pb.Envelope_Disaster{
			Disaster: &pb.DisasterAction{
				Action: &pb.DisasterAction_Warning{
					Warning: &pb.DisasterWarning{
						Type:             disasterType,
						Message:          msg,
						CountdownSeconds: 3600, // 遊戲內 1 小時
					},
				},
			},
		},
	}
	
	session.GetManager().AddToForwardQueue(env)
}
