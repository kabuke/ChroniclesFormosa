package asset

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed font/* audio/* sprite/* tilemap/*
var assetsFS embed.FS

var (
	DefaultFont font.Face
	// Tilesets 存儲不同來源的瓦片集 [name]map[index]*ebiten.Image
	Tilesets map[string]map[int]*ebiten.Image
)

// LoadAssets 初始化並載入核心資源
func LoadAssets() error {
	// 1. 載入 Unifont
	fontData, err := assetsFS.ReadFile("font/unifont-16.0.04.otf")
	if err != nil {
		return fmt.Errorf("failed to read font: %v", err)
	}

	tt, err := opentype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	const dpi = 72
	DefaultFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    16,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %v", err)
	}

	// 2. 初始化所有 Tilemaps
	Tilesets = make(map[string]map[int]*ebiten.Image)
	
	// 載入 16px 核心資源並放大 2 倍 (方案 A)
	// 這是目前的最高優先級圖源
	if err := loadTileset("core16", "tilemap/spr_tileset_sunnysideworld_16px.png", 16, 2.0); err != nil {
		return err
	}

	// 其他備用圖源
	_ = loadTileset("forest", "tilemap/spr_tileset_sunnysideworld_forest_32px.png", 32, 1.0)
	_ = loadTileset("terrain2", "tilemap/Tilesets/Tileset-Terrain2.png", 32, 1.0)

	return nil
}

func loadTileset(name string, path string, tileSize int, scale float64) error {
	f, err := assetsFS.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	srcImg := ebiten.NewImageFromImage(img)
	w, h := srcImg.Size()
	
	cols := w / tileSize
	rows := h / tileSize

	Tilesets[name] = make(map[int]*ebiten.Image)
	index := 0
	
	destSize := int(float64(tileSize) * scale)

	for r := 0; rows > r; r++ {
		for c := 0; cols > c; c++ {
			sx := c * tileSize
			sy := r * tileSize
			
			rect := image.Rect(sx, sy, sx+tileSize, sy+tileSize)
			sub := srcImg.SubImage(rect).(*ebiten.Image)
			
			if scale != 1.0 {
				// 確保放大至精確的 destSize (如 32x32)
				scaled := ebiten.NewImage(destSize, destSize)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(scale, scale)
				// 使用預設的 Nearest Filter 保持像素銳利
				scaled.DrawImage(sub, op)
				Tilesets[name][index] = scaled
			} else {
				Tilesets[name][index] = sub
			}
			index++
		}
	}
	return nil
}

// GetTile 從指定的 Tileset 獲取瓦片
func GetTile(set string, index int) *ebiten.Image {
	if s, ok := Tilesets[set]; ok {
		return s[index]
	}
	return nil
}

// GetImage 載入指定的圖片
func GetImage(path string) (*ebiten.Image, error) {
	fullPath := path
	if !strings.HasPrefix(path, "sprite/") {
		fullPath = "sprite/" + path
	}
	f, err := assetsFS.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// GetAudioData 讀取音訊原始資料
func GetAudioData(path string) ([]byte, error) {
	return assetsFS.ReadFile("audio/" + path)
}

// WalkSprites 遍歷指定目錄下的所有圖片
func WalkSprites(root string, fn func(path string, img *ebiten.Image) error) error {
	return fs.WalkDir(assetsFS, "sprite/"+root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".png") {
			return err
		}
		img, err := GetImage(path)
		if err != nil {
			return err
		}
		return fn(path, img)
	})
}
