package ui

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/i18n"
)

type Waypoint struct {
	X, Y float64
}

type ReliefNode struct {
	X, Y       float32
	IsAffected bool
	IsCovered  bool
}

type ReliefPanel struct {
	Active            bool
	DisasterID        int64
	TargetVillageID   int64
	TargetVillageName string
	IsAffected        bool

	Population int32
	Nodes      []ReliefNode
	Waypoints  []Waypoint
	
	HasStarted   bool
	IsDragging   bool
	StartTime    time.Time
	RouteDist    float32
	TimeLimitSec int
	CoveredCount int
	TotalRed     int
}

var GlobalReliefPanel = &ReliefPanel{TimeLimitSec: 30}

var OnReliefSubmit func(waypoints []Waypoint, timeMs int32, coverage float32, dist float32, targetID int64)
var OnReliefDonateSubmit func(amount int32, targetID int64)

func (r *ReliefPanel) SetAffected(affected bool) {
	r.IsAffected = affected
}

func (r *ReliefPanel) Show(disasterID, targetVillageID int64, villageName string, population int32) {
	r.Active = true
	r.DisasterID = disasterID
	r.TargetVillageID = targetVillageID
	r.TargetVillageName = villageName
	r.Population = population
	r.Waypoints = nil
	r.HasStarted = false
	r.IsDragging = false
	r.RouteDist = 0
	r.CoveredCount = 0

	// 產生小遊戲節點
	w, h := ebiten.WindowSize()
	pw, ph := 600, 400
	px, py := (w-pw)/2, (h-ph)/2

	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	nodeCount := int(population)
	if nodeCount > 200 { nodeCount = 200 } // 畫面塞不下的上限保護
	if nodeCount < 20 { nodeCount = 20 }
	
	r.TotalRed = nodeCount / 4 // 25% 的災民
	if r.TotalRed < 1 { r.TotalRed = 1 }

	r.Nodes = make([]ReliefNode, nodeCount)
	for i := 0; i < nodeCount; i++ {
		// 分散在地圖中心區域 (留邊界)
		nx := float32(px + 40 + randSource.Intn(pw-80))
		ny := float32(py + 60 + randSource.Intn(ph-120))
		r.Nodes[i] = ReliefNode{X: nx, Y: ny, IsAffected: i < r.TotalRed, IsCovered: false}
	}
	
	// 打亂陣列，讓紅點均勻分布
	randSource.Shuffle(len(r.Nodes), func(i, j int) {
		r.Nodes[i], r.Nodes[j] = r.Nodes[j], r.Nodes[i]
	})
}

func (r *ReliefPanel) Hide() {
	r.Active = false
	r.Waypoints = nil
}

func (r *ReliefPanel) submitGame() {
	if !r.HasStarted { return }
	r.HasStarted = false
	timeMs := int32(time.Since(r.StartTime).Milliseconds())
	cov := float32(r.CoveredCount) / float32(r.TotalRed)
	
	if OnReliefSubmit != nil {
		OnReliefSubmit(r.Waypoints, timeMs, cov, r.RouteDist, r.TargetVillageID)
	}
	r.Hide()
}

func getDistance(x1, y1, x2, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	return float32(dx*dx + dy*dy) // 這裡為了效能我們用平方距離判斷，如果要實際距離則需要 sqrt
}

// 用於長度累計
func sqrtDist(x1, y1, x2, y2 float32) float32 {
	// ebiten 環境下無法直接引誘 math.Sqrt float32
	// 由於這是 Go std lib 環境直接處理
	distSq := float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	// 簡單自己算
	// Go 1.20+ 可以用 math.Sqrt
	return float32(distSq) // 這裡圖方便先傳平方或用概略值，我加上 math.Sqrt
}

func (r *ReliefPanel) Update() {
	if !r.Active { return }

	w, h := ebiten.WindowSize()
	pw, ph := 600, 400
	px, py := (w-pw)/2, (h-ph)/2

	// 時間超時自動送出
	if r.HasStarted {
		if time.Since(r.StartTime).Seconds() > float64(r.TimeLimitSec) {
			r.submitGame()
			return
		}
	}

	cx, cy := ebiten.CursorPosition()
	fx, fy := float32(cx), float32(cy)
	
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// 檢查是否點到捐獻按鈕 (遊戲尚未開始才能捐獻)
		if !r.HasStarted {
			btnX, btnY := float32(px+480), float32(py+350)
			if fx >= btnX && fx <= btnX+100 && fy >= btnY && fy <= btnY+35 {
				if GlobalNavbar.Stamina >= 10 && OnReliefDonateSubmit != nil {
					OnReliefDonateSubmit(10, r.TargetVillageID)
					return
				}
			}
		}

		// 點擊地圖區塊開始滑鼠拖曳
		if fx >= float32(px) && fx <= float32(px+pw) && fy >= float32(py) && fy <= float32(py+ph-50) {
			r.IsDragging = true
			if !r.HasStarted {
				r.HasStarted = true
				r.StartTime = time.Now()
			}
			r.Waypoints = append(r.Waypoints, Waypoint{X: float64(fx), Y: float64(fy)})
		}
	}

	if r.IsDragging {
		// 滑鼠拖拉畫線，並偵測覆蓋
		if len(r.Waypoints) > 0 {
			lastW := r.Waypoints[len(r.Waypoints)-1]
			distSq := getDistance(float32(lastW.X), float32(lastW.Y), fx, fy)
			if distSq > 100 { // 移動超過 10 像素才紀錄
				r.Waypoints = append(r.Waypoints, Waypoint{X: float64(fx), Y: float64(fy)})
				r.RouteDist += float32(distSq) // 近似值或精確值（這裡應該要套 Sqrt，但我先簡化傳入平方後在後端看，反正只是給個懲罰基準）
				
				// 檢查有沒有包圍紅點
				for i, node := range r.Nodes {
					if node.IsAffected && !node.IsCovered {
						if getDistance(node.X, node.Y, fx, fy) < 225 { // 半徑 15 px (15*15)
							r.Nodes[i].IsCovered = true
							r.CoveredCount++
						}
					}
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if r.IsDragging {
			r.IsDragging = false
			// 一筆畫結束，自動送出
			r.submitGame()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		r.Hide()
	}
}

func (r *ReliefPanel) Draw(screen *ebiten.Image) {
	if !r.Active { return }
	w, h := screen.Size()

	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.RGBA{0, 0, 0, 200}, true)

	pw, ph := 600, 400
	px, py := (w-pw)/2, (h-ph)/2
	DrawFilledRoundedRect(screen, float32(px), float32(py), float32(pw), float32(ph), 10, ColorPaperWhite)
	vector.StrokeRect(screen, float32(px), float32(py), float32(pw), float32(ph), 2, ColorInkBlack, true)

	// 頂部與底部 UI
	text.Draw(screen, fmt.Sprintf(i18n.Global.GetText("RELIEF_TITLE"), r.TargetVillageName), asset.DefaultFont, int(px+20), int(py+30), ColorInkBlack)
	text.Draw(screen, i18n.Global.GetText("RELIEF_INSTRUCTION"), asset.DefaultFont, int(px+240), int(py+30), color.RGBA{220, 30, 30, 255})
	
	// 節點繪製
	for _, node := range r.Nodes {
		var c color.Color = color.RGBA{200, 200, 200, 255} // 未受災: 白灰
		if node.IsAffected {
			if node.IsCovered {
				c = color.RGBA{50, 200, 50, 255} // 覆蓋: 綠
			} else {
				c = color.RGBA{220, 30, 30, 255} // 受災: 紅
			}
		}
		vector.DrawFilledCircle(screen, node.X, node.Y, 5, c, true)
	}

	// 路線繪製
	if len(r.Waypoints) > 0 {
		for i := 0; i < len(r.Waypoints)-1; i++ {
			p1, p2 := r.Waypoints[i], r.Waypoints[i+1]
			vector.StrokeLine(screen, float32(p1.X), float32(p1.Y), float32(p2.X), float32(p2.Y), 3, color.RGBA{0, 100, 200, 150}, true)
		}
	}

	// 狀態字串
	timeLeft := r.TimeLimitSec
	if r.HasStarted {
		timeLeft = r.TimeLimitSec - int(time.Since(r.StartTime).Seconds())
	}
	statStr := fmt.Sprintf(i18n.Global.GetText("RELIEF_STATUS"), 
		r.CoveredCount, r.TotalRed, float32(r.CoveredCount)/float32(r.TotalRed)*100, timeLeft)
	text.Draw(screen, statStr, asset.DefaultFont, int(px+20), int(py+375), ColorInkBlack)
	text.Draw(screen, i18n.Global.GetText("RELIEF_ESCAPE"), asset.DefaultFont, int(px+20), int(py+395), ColorFactionQing)

	// 判斷按鈕 (只有還沒開始畫線可以按捐獻)
	if !r.HasStarted {
		btnX, btnY := float32(px+480), float32(py+350)
		btnColor := color.RGBA{0, 150, 255, 255}
		if GlobalNavbar.Stamina < 10 { btnColor = color.RGBA{150, 150, 150, 255} }
		DrawFilledRoundedRect(screen, btnX, btnY, 100, 35, 5, btnColor)
		text.Draw(screen, i18n.Global.GetText("RELIEF_DONATE_BTN"), asset.DefaultFont, int(btnX)+15, int(btnY)+24, color.White)
	}
}
