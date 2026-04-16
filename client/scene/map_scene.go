package scene

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

const (
	TileSize     = 32
	ChunkWidth   = 20
	WorldChunksX = 25 // 500 tiles
	WorldChunksY = 25 // 500 tiles
)

type MapScene struct {
	manager  *SceneManager
	net      *network.NetworkClient
	camera   *Camera

	villages  []*pb.VillageSummary
	relations []*pb.DiplomacyRelation

	chunkCache [WorldChunksX][WorldChunksY]*ebiten.Image
}

func NewMapScene(m *SceneManager, net *network.NetworkClient) *MapScene {
	s := &MapScene{
		manager: m,
		net:     net,
		camera:  NewCamera(),
	}
	// 初始相機對焦於地圖正中央 (500*32/2 = 8000)
	s.camera.X, s.camera.Y = 8000, 8000
	s.renderAllChunks()
	return s
}

func (s *MapScene) renderAllChunks() {
	for cx := 0; cx < WorldChunksX; cx++ {
		for cy := 0; cy < WorldChunksY; cy++ {
			img := ebiten.NewImage(ChunkWidth*TileSize, ChunkWidth*TileSize)
			// 背景藍色 (深海)
			img.Fill(color.RGBA{15, 45, 80, 255})
			for tx := 0; tx < ChunkWidth; tx++ {
				for ty := 0; ty < ChunkWidth; ty++ {
					gx := cx*ChunkWidth + tx
					gy := cy*ChunkWidth + ty

					tIdx := asset.TileWater

					// 高解析度地形判定 (對應 500x500)
					if isPenghu(gx, gy) {
						tIdx = asset.TileGrass
					} else {
						terrainType := getTaiwanTerrain(gx, gy)
						if terrainType != -1 {
							tIdx = terrainType
						}
					}

					tileImg := asset.GetTile("core16", tIdx)
					if tileImg != nil {
						op := &ebiten.DrawImageOptions{}
						op.GeoM.Translate(float64(tx*TileSize), float64(ty*TileSize))
						img.DrawImage(tileImg, op)
					}
				}
			}
			s.chunkCache[cx][cy] = img
		}
	}
}

// getTaiwanTerrain 採用衛星圖 1:1 輪廓重構演算法
func getTaiwanTerrain(gx, gy int) int {
	sx := float64(gx) / 5.0
	sy := float64(gy) / 5.0

	// 衛星圖邊界數據點 (Y, X_Coordinate)
	// 這些點位是根據您的庄頭座標 (0-100) 與衛星圖輪廓精確匹配得出的
	type edge struct{ y, x float64 }
	
	// 高精度西岸控制線 (完全以衛星圖每5格描點，去除不自然的圓角)
	leftPts := []edge{
		{2, 63}, {5, 61}, {10, 58}, {15, 53}, 
		{20, 48}, {25, 45}, {30, 42}, {35, 39}, 
		{40, 36}, {45, 34}, {50, 32}, {55, 33}, 
		{60, 35}, {65, 37}, {70, 39}, {75, 41}, 
		{80, 44}, {85, 47}, {90, 49}, {95, 51}, {99, 52},
	}

	// 高精度東岸控制線
	rightPts := []edge{
		{2, 64}, {5, 67}, {10, 68}, {15, 70}, 
		{20, 71}, {25, 71}, {30, 72}, {35, 72}, 
		{40, 71}, {45, 71}, {50, 70}, {55, 68}, 
		{60, 67}, {65, 66}, {70, 64}, {75, 61}, 
		{80, 58}, {85, 56}, {90, 54}, {95, 53}, {99, 53},
	}

	if sy < leftPts[0].y || sy > leftPts[len(leftPts)-1].y {
		return -1
	}

	// 插值計算
	lerp := func(pts []edge) float64 {
		for i := 0; i < len(pts)-1; i++ {
			p1, p2 := pts[i], pts[i+1]
			if sy >= p1.y && sy <= p2.y {
				t := (sy - p1.y) / (p2.y - p1.y)
				return p1.x + t*(p2.x-p1.x)
			}
		}
		return pts[len(pts)-1].x
	}

	rawL := lerp(leftPts)
	rawR := lerp(rightPts)

	// 拔除大幅度雜訊，採用極少量微擾以維持精密描出來的形狀
	lJitter := 0.2*math.Cos(sy*2.0)
	rJitter := 0.2*math.Sin(sy*2.0)

	lEdge := rawL + lJitter
	rEdge := rawR + rJitter

	if sx < lEdge || sx > rEdge {
		return -1
	}

	// 內部分佈比例
	totalW := rEdge - lEdge
	if totalW <= 0 { return -1 }
	posRatio := (sx - lEdge) / totalW

	// 地形分佈 (偏東的山脈)
	if posRatio < 0.05 || posRatio > 0.95 {
		return asset.TileSand
	}
	// 中央山脈 (精確縮減寬度，呈現脊樑感，東移)
	if posRatio > 0.60 && posRatio < 0.85 {
		return asset.TileMountain
	}

	return asset.TileGrass
}

// isPenghu 確保澎湖與本島有絕對隔離空間
func isPenghu(gx, gy int) bool {
	sx := float64(gx) / 5.0
	sy := float64(gy) / 5.0

	// 澎湖區域判定，更貼近真實澎湖群島位置與散佈
	if sx < 5 || sx > 25 || sy < 40 || sy > 65 {
		return false
	}

	centers := []struct{ x, y, r float64 }{
		{15, 53, 2.0}, // 馬公/湖西
		{14, 50, 1.2}, // 白沙
		{11, 48, 0.8}, // 吉貝
		{12, 55, 1.0}, // 望安
		{10, 57, 0.6}, // 七美
	}
	for _, c := range centers {
		dx := sx - c.x
		dy := sy - c.y
		if dx*dx+dy*dy <= c.r*c.r {
			return true
		}
	}
	return false
}

type GlobalDisasterState struct {
	Active           bool
	Type             pb.DisasterType
	EpicenterTileID  int64
	EpicenterName    string
	PathTiles        []int64
	AffectedVillages map[int64]bool
	Intensity        float32
	Magnitude        float32
	AnimTimer        float64
}

var CurrentDisaster = &GlobalDisasterState{AffectedVillages: make(map[int64]bool)}

func (s *MapScene) Update() error {
	s.camera.Update()
	
	if CurrentDisaster.Active {
		CurrentDisaster.AnimTimer += 1.0 / 60.0
	}

	// Phase 3 Visual Effects Update
	ui.GlobalScreenShake.Update()
	ui.GlobalExplosion.Update()
	ui.GlobalTyphoon.Update()
	
	// Update Relief Panel if active
	if ui.GlobalReliefPanel.Active {
		ui.GlobalReliefPanel.Update()
		return nil // Block other clicks if panel is active
	}

	// Handle Map Village Clicks for Relief
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		worldX, worldY := s.camera.ScreenToWorld(float64(mx), float64(my))
		
		// Check Village clicks
		for _, v := range s.villages {
			vx := float64(v.X) * 5.0 * TileSize
			vy := float64(v.Y) * 5.0 * TileSize
			if math.Abs(worldX-vx) < 30 && math.Abs(worldY-vy) < 30 {
				if CurrentDisaster.Active && CurrentDisaster.AffectedVillages[v.VillageId] {
					// Open Relief Panel for this village
					ui.GlobalReliefPanel.Show(CurrentDisaster.EpicenterTileID, v.VillageId, v.Name, v.Population)
					return nil
				}
			}
		}

		// Handle Debug Buttons
		if my >= 40 && my <= 70 {
			disasterType := -1
			if mx >= 10 && mx <= 110 {
				disasterType = 0
			} else if mx >= 120 && mx <= 220 {
				disasterType = 1
			} else if mx >= 230 && mx <= 330 {
				disasterType = 2
			}
			
			if disasterType != -1 && s.net != nil {
				s.net.SendEnvelope(&pb.Envelope{
					Payload: &pb.Envelope_Disaster{
						Disaster: &pb.DisasterAction{
							Action: &pb.DisasterAction_DebugTrigger{
								DebugTrigger: &pb.DebugTriggerReq{
									DisasterType: int32(disasterType),
								},
							},
						},
					},
				})
			}
		}
	}

	return nil
}

func (s *MapScene) SyncVillages(list []*pb.VillageSummary) {
	s.villages = list
}

func (s *MapScene) Draw(screen *ebiten.Image) {
	// Create a temporary screen for shake
	shakeScreen := ebiten.NewImage(screen.Size())

	// 1. 繪製緩存地圖區塊 (只繪製視野內的)
	for cx := 0; cx < WorldChunksX; cx++ {
		for cy := 0; cy < WorldChunksY; cy++ {
			img := s.chunkCache[cx][cy]
			if img == nil {
				continue
			}
			wx := float64(cx * ChunkWidth * TileSize)
			wy := float64(cy * ChunkWidth * TileSize)
			sx, sy := s.camera.WorldToScreen(wx, wy)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s.camera.Zoom, s.camera.Zoom)
			op.GeoM.Translate(sx, sy)
			shakeScreen.DrawImage(img, op)
		}
	}

	// 2. 繪製庄頭
	for _, v := range s.villages {
		vx := float64(v.X) * 5.0 * TileSize
		vy := float64(v.Y) * 5.0 * TileSize
		
		sx, sy := s.camera.WorldToScreen(vx, vy)

		// 繪製建築圖示
		roofs := []int{1180, 796, 988}
		idx := int(v.FactionId) - 1
		if idx < 0 { idx = 0 }
		
		img := asset.GetTile("core16", roofs[idx%len(roofs)])
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s.camera.Zoom, s.camera.Zoom)
			op.GeoM.Translate(sx, sy)
			
			// Highlight affected villages
			if CurrentDisaster.Active && CurrentDisaster.AffectedVillages[v.VillageId] {
				// Blink red tint
				if int(CurrentDisaster.AnimTimer*2)%2 == 0 {
					op.ColorScale.ScaleWithColor(color.RGBA{255, 100, 100, 255})
				}
				ui.DrawText(shakeScreen, "⚠️ "+v.Name, int(sx)-10, int(sy)-15, color.RGBA{255, 50, 50, 255})
			} else {
				ui.DrawText(shakeScreen, v.Name, int(sx), int(sy)-5, color.White)
			}
			shakeScreen.DrawImage(img, op)
		}
	}

	// 3. 外交關係線
	s.drawDiplomacyLines(shakeScreen)

	// Draw Epicenter or Typhoon Path
	if CurrentDisaster.Active {
		if CurrentDisaster.Type == pb.DisasterType_DISASTER_EARTHQUAKE {
			// Epicenter rings
			ex := float64(CurrentDisaster.EpicenterTileID / 1000) * 5.0 * TileSize
			ey := float64(CurrentDisaster.EpicenterTileID % 1000) * 5.0 * TileSize
			sx, sy := s.camera.WorldToScreen(ex, ey)
			radius := float32(math.Mod(CurrentDisaster.AnimTimer*50, 150))
			vector.StrokeCircle(shakeScreen, float32(sx), float32(sy), radius, 2, color.RGBA{255, 50, 50, 150}, true)
			vector.StrokeCircle(shakeScreen, float32(sx), float32(sy), radius+20, 2, color.RGBA{255, 100, 100, 100}, true)
		} else if CurrentDisaster.Type == pb.DisasterType_DISASTER_TYPHOON {
			// Draw Typhoon path as a swirling vortex moving along the path
			if len(CurrentDisaster.PathTiles) > 0 {
				pathIdx := int(CurrentDisaster.AnimTimer) % len(CurrentDisaster.PathTiles)
				ex := float64(CurrentDisaster.PathTiles[pathIdx] / 1000) * 5.0 * TileSize
				ey := float64(CurrentDisaster.PathTiles[pathIdx] % 1000) * 5.0 * TileSize
				sx, sy := s.camera.WorldToScreen(ex, ey)
				radius := float32(100.0 + math.Sin(CurrentDisaster.AnimTimer*5)*20.0)
				vector.DrawFilledCircle(shakeScreen, float32(sx), float32(sy), radius, color.RGBA{100, 150, 255, 50}, true)
				vector.StrokeCircle(shakeScreen, float32(sx), float32(sy), radius*0.5, 5, color.RGBA{200, 220, 255, 100}, true)
			}
		}
	}

	// 4. Phase 3 粒子特效
	ui.GlobalExplosion.Draw(shakeScreen)
	ui.GlobalTyphoon.Draw(shakeScreen)

	// 套用震動並畫回主螢幕
	ui.GlobalScreenShake.Apply(shakeScreen)
	screen.DrawImage(shakeScreen, nil)

	// UI 面板
	if ui.GlobalReliefPanel.Active {
		ui.GlobalReliefPanel.Draw(screen)
	}

	// 繪製測試天災按鈕 (Debug Buttons)
	ui.DrawFilledRoundedRect(screen, 10, 40, 100, 30, 5, color.RGBA{100, 150, 100, 200})
	ui.DrawText(screen, "無害地震", 25, 46, color.White)
	
	ui.DrawFilledRoundedRect(screen, 120, 40, 100, 30, 5, color.RGBA{200, 100, 100, 200})
	ui.DrawText(screen, "有害地震", 135, 46, color.White)
	
	ui.DrawFilledRoundedRect(screen, 230, 40, 100, 30, 5, color.RGBA{100, 100, 200, 200})
	ui.DrawText(screen, "颱  風", 255, 46, color.White)
}

func (s *MapScene) drawDiplomacyLines(screen *ebiten.Image) {
	vMap := make(map[int64]*pb.VillageSummary)
	for _, v := range s.villages { vMap[v.VillageId] = v }

	for _, rel := range s.relations {
		v1, ok1 := vMap[rel.SourceId]
		v2, ok2 := vMap[rel.TargetId]
		if ok1 && ok2 {
			sx1, sy1 := s.camera.WorldToScreen(float64(v1.X*TileSize+16), float64(v1.Y*TileSize+16))
			sx2, sy2 := s.camera.WorldToScreen(float64(v2.X*TileSize+16), float64(v2.Y*TileSize+16))
			clr := color.RGBA{0, 255, 0, 150}
			if rel.Type == 1 { clr = color.RGBA{255, 105, 180, 150} }
			vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 3, clr, true)
		}
	}
}

var OnRequestVillages func()

func (s *MapScene) OnEnter() {
	ui.GlobalNavbar.SceneName = "World Map"
	if OnRequestVillages != nil {
		OnRequestVillages() // 這裡應僅用於背景渲染，不帶 UI 狀態
	}
}
func (s *MapScene) OnLeave() {}
