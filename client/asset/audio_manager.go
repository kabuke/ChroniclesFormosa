package asset

import (
	"bytes"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	sampleRate = 44100
)

var (
	AudioContext *audio.Context
	bgmPlayer    *audio.Player
	currentBGM   string
)

func InitAudio() {
	AudioContext = audio.NewContext(sampleRate)
}

func PlayBGM(name string) {
	if currentBGM == name { return }

	var path string
	if strings.HasPrefix(name, "bgm_") {
		path = "audio/bgm/" + name
	} else if strings.HasPrefix(name, "sfx_") {
		path = "audio/sfx/" + name
	} else {
		path = "audio/bgm/" + name // 預設
	}

	data, err := assetsFS.ReadFile(path)
	if err != nil {
		log.Printf("[Audio] Failed to load %s: %v", path, err)
		return
	}

	var s io.ReadSeeker
	if strings.HasSuffix(name, ".mp3") {
		s, err = mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	} else {
		s, err = vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	}

	if err != nil {
		log.Printf("[Audio] Decode error %s: %v", name, err)
		return
	}

	loop := audio.NewInfiniteLoop(s.(interface {
		io.ReadSeeker
		Length() int64
	}), s.(interface {
		io.ReadSeeker
		Length() int64
	}).Length())
	
	if bgmPlayer != nil { bgmPlayer.Close() }
	bgmPlayer, err = AudioContext.NewPlayer(loop)
	if err == nil {
		bgmPlayer.Play()
		currentBGM = name
	}
}

func PlaySFX(name string) {
	path := "audio/sfx/" + name
	data, err := assetsFS.ReadFile(path)
	if err != nil { return }

	var s io.ReadSeeker
	if filepath.Ext(name) == ".mp3" {
		s, _ = mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	} else if filepath.Ext(name) == ".wav" {
		s, _ = wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	} else {
		s, _ = vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	}

	p, err := AudioContext.NewPlayer(s)
	if err == nil { p.Play() }
}
