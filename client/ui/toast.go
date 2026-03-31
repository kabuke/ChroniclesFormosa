package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/kabuke/ChroniclesFormosa/client/asset"
)

type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

type Toast struct {
	Message   string
	Type      ToastType
	ExpiresAt time.Time
}

type ToastManager struct {
	toasts []*Toast
}

var GlobalToastManager = &ToastManager{}

func (m *ToastManager) Show(msg string, t ToastType, duration time.Duration) {
	m.toasts = append(m.toasts, &Toast{
		Message:   msg,
		Type:      t,
		ExpiresAt: time.Now().Add(duration),
	})
}

func (m *ToastManager) Success(msg string) { m.Show(msg, ToastSuccess, 3*time.Second) }
func (m *ToastManager) Error(msg string)   { m.Show(msg, ToastError, 5*time.Second) }
func (m *ToastManager) Warning(msg string) { m.Show(msg, ToastWarning, 4*time.Second) }
func (m *ToastManager) Info(msg string)    { m.Show(msg, ToastInfo, 3*time.Second) }

func (m *ToastManager) Update() {
	now := time.Now()
	active := m.toasts[:0]
	for _, t := range m.toasts {
		if now.Before(t.ExpiresAt) {
			active = append(active, t)
		}
	}
	m.toasts = active
}

func (m *ToastManager) Draw(screen *ebiten.Image) {
	w, h := screen.Size()
	toastH := 35
	padding := 10

	for i, t := range m.toasts {
		bgClr := color.RGBA{40, 40, 40, 220}
		switch t.Type {
		case ToastSuccess: bgClr = color.RGBA{46, 139, 87, 220}
		case ToastError:   bgClr = color.RGBA{178, 34, 34, 220}
		case ToastWarning: bgClr = color.RGBA{218, 165, 32, 220}
		}

		tw := 320.0
		tx := (float32(w) - float32(tw)) / 2
		ty := float32(h) - float32(i+1)*float32(toastH+padding) - 50

		DrawFilledRoundedRect(screen, tx, ty, float32(tw), float32(toastH), 5, bgClr)
		text.Draw(screen, t.Message, asset.DefaultFont, int(tx)+15, int(ty)+24, color.White)
	}
}
