package scene

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
	"github.com/kabuke/ChroniclesFormosa/client/network"
	"github.com/kabuke/ChroniclesFormosa/client/ui"
)

const (
	TileSize    = 32
	ChunkWidth  = 20
	WorldChunks = 5
)

type Tile struct {
	Type       int32
	Tileset    string
	TileIndex  int
	OverlaySet string
	OverlayIdx int
}

type Chunk struct {
	ID    int
	X, Y  int
	Tiles [ChunkWidth][ChunkWidth]Tile
	Image *ebiten.Image
	Dirty bool
}

type MapScene struct {
	manager  *SceneManager
	net      *network.NetworkClient
	camera   *Camera
	chunks   [WorldChunks][WorldChunks]*Chunk
	showGrid bool
}

func NewMapScene(m *SceneManager, net *network.NetworkClient) *MapScene {
	s := &MapScene{
		manager:  m,
		net:      net,
		camera:   NewCamera(),
		showGrid: false,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for cx := 0; cx < WorldChunks; cx++ {
		for cy := 0; cy < WorldChunks; cy++ {
			c := &Chunk{
				ID:    cx*WorldChunks + cy,
				X:     cx,
				Y:     cy,
				Dirty: true,
			}
			for tx := 0; tx < ChunkWidth; tx++ {
				for ty := 0; ty < ChunkWidth; ty++ {
					tType := int32(0) // 草地
					tSet := "core16"
					tIdx := 0 // 基礎草地
					
					oSet := ""
					oIdx := -1

					// 1. 基礎地形判定
					if cx == 0 || cy == 0 || cx == WorldChunks-1 || cy == WorldChunks-1 {
						tType = 2
						tSet = "core16"
						tIdx = 17 // 深藍純海洋 (四格連在一起的那種質感)
					} else {
						// 2. 隨機生成物件
						prob := r.Float32()
						if prob < 0.03 {
							oSet = "core16"
							// 隨機選一種陣營屋頂 (紅/綠/橘)
							roofs := []int{1180, 796, 988}
							oIdx = roofs[r.Intn(len(roofs))]
						} else if prob < 0.06 {
							oSet = "core16"
							oIdx = 304 // 農田/作物
						} else if prob < 0.09 {
							oSet = "core16"
							oIdx = 1328 // 礦石
						} else if prob < 0.15 {
							oSet = "core16"
							oIdx = 48 // 樹木
						}
					}
					
					c.Tiles[tx][ty] = Tile{
						Type: tType, 
						Tileset: tSet, 
						TileIndex: tIdx,
						OverlaySet: oSet,
						OverlayIdx: oIdx,
					}
				}
			}
			s.chunks[cx][cy] = c
		}
	}
	return s
}

func (s *MapScene) Update() error {
	s.camera.Update()
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.manager.SwitchTo("Login")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		s.showGrid = !s.showGrid
		for cx := 0; cx < WorldChunks; cx++ {
			for cy := 0; cy < WorldChunks; cy++ {
				s.chunks[cx][cy].Dirty = true
			}
		}
	}
	return nil
}

func (s *MapScene) Draw(screen *ebiten.Image) {
	// 背景底色改為深藍
	screen.Fill(color.RGBA{10, 30, 60, 255})

	centerX := int(s.camera.X / (ChunkWidth * TileSize))
	centerY := int(s.camera.Y / (ChunkWidth * TileSize))

	for cx := centerX - 1; cx <= centerX+1; cx++ {
		for cy := centerY - 1; cy <= centerY+1; cy++ {
			if cx < 0 || cy < 0 || cx >= WorldChunks || cy >= WorldChunks {
				continue
			}
			chunk := s.chunks[cx][cy]
			s.drawChunk(screen, chunk)
		}
	}

	vector.DrawFilledRect(screen, 10, 40, 240, 140, color.RGBA{0, 0, 0, 180}, true)
	text.Draw(screen, "MAP MODE\n[RightDrag] Move\n[Wheel] Zoom\n[G] Grid\n[ESC] Logout", asset.DefaultFont, 20, 65, color.White)
	
	camInfo := fmt.Sprintf("Cam: %.1f, %.1f\nZoom: x%.1f", s.camera.X, s.camera.Y, s.camera.Zoom)
	text.Draw(screen, camInfo, asset.DefaultFont, 20, 145, color.RGBA{200, 200, 100, 255})
}

func (s *MapScene) drawChunk(screen *ebiten.Image, c *Chunk) {
	if c.Image == nil || c.Dirty {
		if c.Image == nil {
			c.Image = ebiten.NewImage(ChunkWidth*TileSize, ChunkWidth*TileSize)
		}
		s.renderChunkToImage(c)
		c.Dirty = false
	}

	worldX := float64(c.X * ChunkWidth * TileSize)
	worldY := float64(c.Y * ChunkWidth * TileSize)
	sx, sy := s.camera.WorldToScreen(worldX, worldY)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(s.camera.Zoom, s.camera.Zoom)
	op.GeoM.Translate(sx, sy)
	screen.DrawImage(c.Image, op)
}

func (s *MapScene) renderChunkToImage(c *Chunk) {
	c.Image.Fill(color.Transparent)
	for x := 0; x < ChunkWidth; x++ {
		for y := 0; y < ChunkWidth; y++ {
			t := c.Tiles[x][y]
			tx, ty := float64(x*TileSize), float64(y*TileSize)
			
			// 1. 繪製基礎地形
			tileImg := asset.GetTile(t.Tileset, t.TileIndex)
			if tileImg != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(tx, ty)
				c.Image.DrawImage(tileImg, op)
			}

			// 2. 繪製疊加物件 (全部來自 core16 且已確保是 32x32)
			if t.OverlaySet != "" && t.OverlayIdx != -1 {
				oImg := asset.GetTile(t.OverlaySet, t.OverlayIdx)
				if oImg != nil {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(tx, ty)
					c.Image.DrawImage(oImg, op)
				}
			}
			
			if s.showGrid {
				vector.StrokeRect(c.Image, float32(tx), float32(ty), TileSize, TileSize, 1, color.RGBA{0, 0, 0, 40}, true)
			}
		}
	}
	vector.StrokeRect(c.Image, 0, 0, ChunkWidth*TileSize, ChunkWidth*TileSize, 1, color.RGBA{255, 255, 255, 60}, true)
}

func (s *MapScene) OnEnter() {
	ui.GlobalNavbar.SceneName = "World Map"
}

func (s *MapScene) OnLeave() {}
