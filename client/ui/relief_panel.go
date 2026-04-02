package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type Waypoint struct {
	X, Y float64
}

type ReliefPanel struct {
	Active     bool
	DisasterID int64
	Waypoints  []Waypoint
	IsAffected bool
}

var GlobalReliefPanel = &ReliefPanel{}

var OnReliefSubmit func(waypoints []Waypoint)

func (r *ReliefPanel) SetAffected(affected bool) {
	r.IsAffected = affected
}

func (r *ReliefPanel) Show(disasterID int64) {
	r.Active = true
	r.DisasterID = disasterID
	r.Waypoints = nil
}

func (r *ReliefPanel) Hide() {
	r.Active = false
	r.Waypoints = nil
}

func (r *ReliefPanel) Update() {
	if !r.Active { return }

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		// Only add waypoint if it's within the map area (roughly)
		// For simplicity in this panel, just record screen coords
		r.Waypoints = append(r.Waypoints, Waypoint{X: float64(cx), Y: float64(cy)})
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if OnReliefSubmit != nil {
			OnReliefSubmit(r.Waypoints)
		}
		r.Hide()
	}
	
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		r.Hide()
	}
}

func (r *ReliefPanel) Draw(screen *ebiten.Image) {
	if !r.Active { return }
	w, h := screen.Size()

	// Dark overlay
	vector.DrawFilledRect(screen, 0, 0, float32(w), float32(h), color.RGBA{0, 0, 0, 200}, true)

	// Panel BG
	pw, ph := 600, 400
	px, py := (w-pw)/2, (h-ph)/2
	DrawFilledRoundedRect(screen, float32(px), float32(py), float32(pw), float32(ph), 10, ColorPaperWhite)
	vector.StrokeRect(screen, float32(px), float32(py), float32(pw), float32(ph), 2, ColorInkBlack, true)

	text.Draw(screen, "【救災調度】", asset.DefaultFont, px+250, py+40, ColorInkBlack)
	text.Draw(screen, "點擊滑鼠左鍵在地圖上繪製牛車運送路線", asset.DefaultFont, px+50, py+80, ColorInkBlack)
	text.Draw(screen, fmt.Sprintf("已規劃節點數: %d", len(r.Waypoints)), asset.DefaultFont, px+50, py+120, ColorInkBlack)
	
	text.Draw(screen, "按 ENTER 提交路線，按 ESC 取消", asset.DefaultFont, px+50, py+350, ColorFactionQing)

	// Draw lines connecting waypoints
	if len(r.Waypoints) > 0 {
		for i := 0; i < len(r.Waypoints)-1; i++ {
			p1, p2 := r.Waypoints[i], r.Waypoints[i+1]
			vector.StrokeLine(screen, float32(p1.X), float32(p1.Y), float32(p2.X), float32(p2.Y), 3, color.RGBA{0, 255, 0, 255}, true)
		}
		// Draw nodes
		for _, p := range r.Waypoints {
			vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), 5, color.RGBA{255, 0, 0, 255}, true)
		}
	}
}
