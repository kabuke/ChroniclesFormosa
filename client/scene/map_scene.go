package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

const (
	TileSize    = 32
	ChunkWidth  = 20
	WorldChunks = 5 
)

type MapScene struct {
	manager  *SceneManager
	net      *network.NetworkClient
	camera   *Camera
	
	villages  []*pb.VillageSummary
	relations []*pb.DiplomacyRelation
	
	chunkCache [WorldChunks][WorldChunks]*ebiten.Image
}

func NewMapScene(m *SceneManager, net *network.NetworkClient) *MapScene {
	s := &MapScene{
		manager: m,
		net:     net,
		camera:  NewCamera(),
	}
	s.camera.X, s.camera.Y = 1600, 1600 // 初始對焦中心
	s.renderAllChunks()
	return s
}

func (s *MapScene) renderAllChunks() {
	for cx := 0; cx < WorldChunks; cx++ {
		for cy := 0; cy < WorldChunks; cy++ {
			img := ebiten.NewImage(ChunkWidth*TileSize, ChunkWidth*TileSize)
			for tx := 0; tx < ChunkWidth; tx++ {
				for ty := 0; ty < ChunkWidth; ty++ {
					tIdx := 17 // 海洋
					// 擴大島嶼：1,1 到 3,3 為草地
					if cx >= 1 && cx <= 3 && cy >= 1 && cy <= 3 {
						tIdx = 0 
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

func (s *MapScene) Update() error {
	s.camera.Update()
	return nil
}

func (s *MapScene) SyncVillages(list []*pb.VillageSummary) {
	s.villages = list
}

func (s *MapScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 20, 40, 255})

	// 1. 繪製緩存地形
	for cx := 0; cx < WorldChunks; cx++ {
		for cy := 0; cy < WorldChunks; cy++ {
			if s.chunkCache[cx][cy] == nil { continue }
			wx := float64(cx * ChunkWidth * TileSize)
			wy := float64(cy * ChunkWidth * TileSize)
			sx, sy := s.camera.WorldToScreen(wx, wy)
			
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s.camera.Zoom, s.camera.Zoom)
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(s.chunkCache[cx][cy], op)
		}
	}

	// 2. 繪製庄頭與名稱
	for _, v := range s.villages {
		wx := float64(v.X * TileSize)
		wy := float64(v.Y * TileSize)
		sx, sy := s.camera.WorldToScreen(wx, wy)
		
		roofs := []int{1180, 796, 988}
		idx := int(v.FactionId) - 1
		if idx < 0 { idx = 0 }
		
		img := asset.GetTile("core16", roofs[idx%len(roofs)])
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(s.camera.Zoom, s.camera.Zoom)
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(img, op)
			ui.DrawText(screen, v.Name, int(sx), int(sy)-5, color.White)
		}
	}

	// 3. 外交關係線
	s.drawDiplomacyLines(screen)
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
