package asset

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed font/* audio/* sprite/*
var assetsFS embed.FS

var (
	DefaultFont font.Face
)

// LoadAssets 初始化並載入核心資源 (如字體)
func LoadAssets() error {
	// 載入 Unifont 作為預設多語系字體
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

	return nil
}

// GetImage 載入指定的圖片並轉為 ebiten.Image
func GetImage(path string) (*ebiten.Image, error) {
	f, err := assetsFS.Open("sprite/" + path)
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

// GetFS 返回原始的 embed FS 以利進階調用
func GetFS() embed.FS {
	return assetsFS
}
