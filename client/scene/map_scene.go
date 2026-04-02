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
	
	// 西岸控制線 (控制「肚子」的突出度與離島間距)
	leftPts := []edge{
		{2,  60}, // 北尖
		{10, 54}, // 台北西
		{25, 40}, // 桃園/新竹
		{40, 26}, // 苗栗/台中
		{55, 18}, // 彰化/南投西
		{65, 15}, // 嘉義/雲林 (西岸最凸出點)
		{75, 17}, // 台南/安平 (確保 25, 70 在陸地)
		{85, 24}, // 高雄/打狗 (確保 28, 85 在陸地)
		{90, 30}, // 屏東
		{95, 38}, // 瑯嶠南
		{99, 43}, // 鵝鑾鼻
	}

	// 東岸控制線 (控制東岸直線度)
	rightPts := []edge{
		{2,  65},  // 北尖
		{10, 71},  // 基隆 (確保 65, 10 在陸地)
		{20, 73},  // 宜蘭 (確保 75, 25 緊貼邊緣)
		{55, 74},  // 花蓮 (確保 70, 55 在陸地)
		{85, 66},  // 卑南 (確保 60, 85 在陸地)
		{92, 58},  // 恆春北
		{96, 52},  // 瑯嶠東
		{99, 45},  // 鵝鑾鼻
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

	// 海岸線微擾雜訊
	lJitter := 1.2 * (math.Sin(sy*0.8) + math.Cos(sx*0.6))
	rJitter := 0.5 * (math.Sin(sy*1.1) + math.Cos(sx*0.9))

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
	// 中央山脈 (精確縮減寬度，呈現脊樑感)
	if posRatio > 0.82 && posRatio < 0.94 {
		return asset.TileMountain
	}

	return asset.TileGrass
}

// isPenghu 確保澎湖與本島有絕對隔離空間
func isPenghu(gx, gy int) bool {
	sx := float64(gx) / 5.0
	sy := float64(gy) / 5.0

	// 澎湖中心 (5, 50)，限定在 X:12 之前，確保與本島 X:15 有深水區
	if sx < 1 || sx > 12 || sy < 43 || sy > 57 {
		return false
	}

	centers := []struct{ x, y, r float64 }{
		{4, 50, 3.2},
		{9, 52, 1.8},
		{6, 46, 1.2},
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

func (s *MapScene) Update() error {
	s.camera.Update()
	
	// Phase 3 Visual Effects Update
	ui.GlobalScreenShake.Update()
	ui.GlobalExplosion.Update()
	ui.GlobalTyphoon.Update()
	
	// Update Relief Panel if active
	if ui.GlobalReliefPanel.Active {
		ui.GlobalReliefPanel.Update()
	}

	// Handle Debug Buttons
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
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
			shakeScreen.DrawImage(img, op)
			ui.DrawText(shakeScreen, v.Name, int(sx), int(sy)-5, color.White)
		}
	}

	// 3. 外交關係線
	s.drawDiplomacyLines(shakeScreen)

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
