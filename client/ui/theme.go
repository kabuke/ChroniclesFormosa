package ui

import (
	"image/color"
)

// ThemeManager 管理全局 UI 主題
type ThemeManager struct {
	IsNightMode bool
}

var GlobalTheme = &ThemeManager{
	IsNightMode: false,
}

// 宣紙風格配色
var (
	ColorPaperWhite = color.RGBA{0xF5, 0xE6, 0xC8, 0xFF} // 宣紙米白 #F5E6C8
	ColorInkBlack   = color.RGBA{0x2C, 0x18, 0x10, 0xFF} // 墨黑 #2C1810
	ColorInkPale    = color.RGBA{0x8B, 0x73, 0x55, 0xFF} // 淡墨 #8B7355

	// 夜間/戰鬥模式
	ColorNightDark = color.RGBA{0x1A, 0x12, 0x0B, 0xFF} // 深墨 #1A120B
	ColorNightGold = color.RGBA{0x8B, 0x69, 0x14, 0xFF} // 暗金 #8B6914
)

// 陣營配色
var (
	ColorFactionQing    = color.RGBA{0xDA, 0xA5, 0x20, 0xFF} // 清軍：金黃 #DAA520
	ColorFactionMing    = color.RGBA{0x2E, 0x8B, 0x57, 0xFF} // 義軍：墨綠 #2E8B57
	ColorFactionNative  = color.RGBA{0x2F, 0x4F, 0x4F, 0xFF} // 原民：藏青 #2F4F4F
)

// GetBackgroundColor 根據模式返回背景色
func (m *ThemeManager) GetBackgroundColor() color.Color {
	if m.IsNightMode {
		return ColorNightDark
	}
	return ColorPaperWhite
}

// GetForegroundColor 根據模式返回前景/文字色
func (m *ThemeManager) GetForegroundColor() color.Color {
	if m.IsNightMode {
		return color.White
	}
	return ColorInkBlack
}
