package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
	pb "github.com/kabuke/ChroniclesFormosa/resource"
)

type ChatPanel struct {
	Visible    bool
	Messages   []*pb.ChatMessage
	MaxHistory int
	
	CurrentChannel pb.ChatChannelType
	Input          string
	
	// UI 座標
	X, Y, W, H float32
}

var GlobalChatPanel = &ChatPanel{
	Visible:    true,
	MaxHistory: 50,
	CurrentChannel: pb.ChatChannelType_CHANNEL_GLOBAL,
	W: 300,
	H: 200,
}

func (p *ChatPanel) AddMessage(msg *pb.ChatMessage) {
	p.Messages = append(p.Messages, msg)
	if len(p.Messages) > p.MaxHistory {
		p.Messages = p.Messages[1:]
	}
}

func (p *ChatPanel) Update() {
	if !p.Visible {
		return
	}

	_, sh := ebiten.WindowSize()
	p.X = 10
	p.Y = float32(sh) - p.H - 10

	// 切換頻道 (Tab)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		p.CurrentChannel = pb.ChatChannelType((int(p.CurrentChannel) + 1) % 4)
	}

	// 點擊輸入框喚起鍵盤
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if float32(mx) >= p.X && float32(mx) <= p.X+p.W &&
			float32(my) >= p.Y+p.H-30 && float32(my) <= p.Y+p.H {
			
			GlobalKeyboard.Input = p.Input
			GlobalKeyboard.Show()
			GlobalKeyboard.OnEnter = func(s string) {
				p.Input = s
				p.Send()
			}
		}
	}
}

func (p *ChatPanel) Send() {
	if p.Input == "" {
		return
	}
	if OnChatSubmit != nil {
		OnChatSubmit(p.CurrentChannel, p.Input)
	}
	p.Input = ""
}

var OnChatSubmit func(ch pb.ChatChannelType, content string)

func (p *ChatPanel) Draw(screen *ebiten.Image) {
	if !p.Visible {
		return
	}

	// 1. 繪製背景 (墨色半透明)
	DrawFilledRoundedRect(screen, p.X, p.Y, p.W, p.H, 8, color.RGBA{0, 0, 0, 160})
	
	// 2. 頻道標籤
	chName := "GLOBAL"
	var chColor color.Color = color.White
	switch p.CurrentChannel {
	case pb.ChatChannelType_CHANNEL_FACTION: 
		chName = "FACTION"
		chColor = ColorFactionMing
	case pb.ChatChannelType_CHANNEL_VILLAGE:
		chName = "VILLAGE"
		chColor = color.RGBA{100, 200, 100, 255}
	}
	
	text.Draw(screen, fmt.Sprintf("[%s]", chName), asset.DefaultFont, int(p.X)+10, int(p.Y)+20, chColor)

	// 3. 繪製歷史訊息 (倒序顯示最近 6 條)
	msgY := int(p.Y) + 45
	startIdx := len(p.Messages) - 6
	if startIdx < 0 { startIdx = 0 }
	
	for i := startIdx; i < len(p.Messages); i++ {
		m := p.Messages[i]
		var clr color.Color = color.White
		if m.Channel == pb.ChatChannelType_CHANNEL_FACTION { clr = ColorFactionMing }
		if m.Channel == pb.ChatChannelType_CHANNEL_VILLAGE { clr = color.RGBA{150, 255, 150, 255} }
		
		txt := fmt.Sprintf("%s: %s", m.Sender, m.Content)
		text.Draw(screen, txt, asset.DefaultFont, int(p.X)+10, msgY, clr)
		msgY += 22
	}

	// 4. 底部輸入提示
	inputBg := color.RGBA{255, 255, 255, 40}
	DrawFilledRoundedRect(screen, p.X+5, p.Y+p.H-30, p.W-10, 25, 4, inputBg)
	
	prompt := p.Input
	if prompt == "" { prompt = "Click to chat..." }
	text.Draw(screen, prompt, asset.DefaultFont, int(p.X)+15, int(p.Y+p.H)-12, color.RGBA{200, 200, 200, 200})
}
