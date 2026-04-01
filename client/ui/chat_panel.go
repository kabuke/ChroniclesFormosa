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
	if !p.Visible { return }

	_, sh := ebiten.WindowSize()
	p.X = 10
	p.Y = float32(sh) - p.H - 10

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		switch p.CurrentChannel {
		case pb.ChatChannelType_CHANNEL_GLOBAL:  p.CurrentChannel = pb.ChatChannelType_CHANNEL_VILLAGE
		case pb.ChatChannelType_CHANNEL_VILLAGE: p.CurrentChannel = pb.ChatChannelType_CHANNEL_REGION
		case pb.ChatChannelType_CHANNEL_REGION:  p.CurrentChannel = pb.ChatChannelType_CHANNEL_FACTION
		case pb.ChatChannelType_CHANNEL_FACTION: p.CurrentChannel = pb.ChatChannelType_CHANNEL_GLOBAL
		default: p.CurrentChannel = pb.ChatChannelType_CHANNEL_GLOBAL
		}
	}

	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if float32(mx) >= p.X && float32(mx) <= p.X+p.W &&
			float32(my) >= p.Y+p.H-30 && float32(my) <= p.Y+p.H {
			
			GlobalKeyboard.User = p.Input
			GlobalKeyboard.FocusIdx = 0
			GlobalKeyboard.Show(ModeChat)
			GlobalKeyboard.OnEnter = func() {
				p.Input = GlobalKeyboard.User
				p.Send()
				GlobalKeyboard.Hide()
			}
		}
	}
}

func (p *ChatPanel) Send() {
	if p.Input == "" { return }
	if OnChatSubmit != nil {
		OnChatSubmit(p.CurrentChannel, p.Input)
	}
	p.Input = ""
}

var OnChatSubmit func(ch pb.ChatChannelType, content string)

func (p *ChatPanel) Draw(screen *ebiten.Image) {
	if !p.Visible { return }

	DrawFilledRoundedRect(screen, p.X, p.Y, p.W, p.H, 8, color.RGBA{0, 0, 0, 160})
	
	chName := "GLOBAL"
	var chColor color.Color = color.White
	switch p.CurrentChannel {
	case pb.ChatChannelType_CHANNEL_FACTION: 
		chName = "FACTION"
		chColor = ColorFactionMing
	case pb.ChatChannelType_CHANNEL_VILLAGE:
		chName = "VILLAGE"
		chColor = color.RGBA{100, 200, 100, 255}
	case pb.ChatChannelType_CHANNEL_REGION:
		chName = "REGION"
		chColor = color.RGBA{100, 150, 255, 255}
	}
	text.Draw(screen, fmt.Sprintf("[%s]", chName), asset.DefaultFont, int(p.X)+10, int(p.Y)+20, chColor)

	msgY := int(p.Y) + 45
	startIdx := len(p.Messages) - 6
	if startIdx < 0 { startIdx = 0 }
	for i := startIdx; i < len(p.Messages); i++ {
		m := p.Messages[i]
		var clr color.Color = color.White
		if m.Channel == pb.ChatChannelType_CHANNEL_FACTION { clr = ColorFactionMing }
		if m.Channel == pb.ChatChannelType_CHANNEL_VILLAGE { clr = color.RGBA{150, 255, 150, 255} }
		if m.Channel == pb.ChatChannelType_CHANNEL_REGION { clr = color.RGBA{180, 200, 255, 255} }
		text.Draw(screen, fmt.Sprintf("%s: %s", m.Sender, m.Content), asset.DefaultFont, int(p.X)+10, msgY, clr)
		msgY += 22
	}

	inputBg := color.RGBA{255, 255, 255, 40}
	DrawFilledRoundedRect(screen, p.X+5, p.Y+p.H-30, p.W-10, 25, 4, inputBg)
	prompt := p.Input
	if prompt == "" { prompt = "Click to chat..." }
	text.Draw(screen, prompt, asset.DefaultFont, int(p.X)+15, int(p.Y+p.H)-12, color.RGBA{200, 200, 200, 200})
}
