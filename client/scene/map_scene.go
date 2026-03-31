package scene

import (
	"fmt"
	"image/color"

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
	Type int32
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
		showGrid: true,
	}

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
					tType := int32(0)
					if cx == 0 || cy == 0 || cx == WorldChunks-1 || cy == WorldChunks-1 {
						tType = 2
					}
					if (tx+cx*ChunkWidth)%7 == 0 && (ty+cy*ChunkWidth)%5 == 0 {
						tType = 1
					}
					c.Tiles[tx][ty] = Tile{Type: tType}
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
	// 修正：按 G 切換時，強制所有 Chunk 重繪
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
	screen.Fill(color.RGBA{10, 40, 80, 255})

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
	
	camInfo := fmt.Sprintf("Cam: %.1f, %.1f\nZoom: x%.1f\nGrid: %v", s.camera.X, s.camera.Y, s.camera.Zoom, s.showGrid)
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
			clr := color.RGBA{50, 150, 50, 255}
			switch t.Type {
			case 1: clr = color.RGBA{100, 100, 100, 255}
			case 2: clr = color.RGBA{30, 80, 160, 255}
			}
			tx, ty := float32(x*TileSize), float32(y*TileSize)
			vector.DrawFilledRect(c.Image, tx, ty, TileSize, TileSize, clr, true)
			
			// 只有開啟 showGrid 時才繪製瓦片格線
			if s.showGrid {
				vector.StrokeRect(c.Image, tx, ty, TileSize, TileSize, 1, color.RGBA{0, 0, 0, 40}, true)
			}
		}
	}
	// Chunk 邊框 (常駐，顏色變淡)
	vector.StrokeRect(c.Image, 0, 0, ChunkWidth*TileSize, ChunkWidth*TileSize, 1, color.RGBA{255, 255, 255, 60}, true)
}

func (s *MapScene) OnEnter() {
	ui.GlobalNavbar.SceneName = "World Map"
}

func (s *MapScene) OnLeave() {}
