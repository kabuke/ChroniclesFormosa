package main

import (
	"encoding/binary"
	"math"
	"math/rand"
	"os"
)

const (
	SampleRate = 44100
	NumChannels = 1
	BitsPerSample = 16
)

func main() {
	generateWAV("client/asset/audio/sfx/sfx_earthquake.wav", 3.0, func(t float64) float64 {
		// Low frequency rumble (30Hz to 60Hz) + noise
		env := math.Exp(-t) // fade out
		f := 40.0 + 20.0*math.Sin(2*math.Pi*5*t)
		wave := math.Sin(2*math.Pi*f*t) * 0.5
		noise := (rand.Float64()*2 - 1) * 0.5
		return (wave + noise) * env * 0.8
	})

	generateWAV("client/asset/audio/sfx/sfx_typhoon.wav", 4.0, func(t float64) float64 {
		// Wind / rain white/pink noise
		env := 1.0
		if t < 0.5 { env = t / 0.5 } else if t > 3.5 { env = (4.0 - t) / 0.5 }
		noise := (rand.Float64()*2 - 1)
		lowpass := math.Sin(2*math.Pi*200*t) * 0.2 // A bit of hum
		return (noise*0.5 + lowpass) * env * 0.6
	})

	generateWAV("client/asset/audio/sfx/sfx_warning.wav", 1.5, func(t float64) float64 {
		// Bell or gong (high pitch, fast decay)
		env := math.Exp(-t * 3)
		f1 := math.Sin(2 * math.Pi * 600 * t)
		f2 := math.Sin(2 * math.Pi * 850 * t)
		return (f1 + f2) * 0.5 * env
	})

	generateWAV("client/asset/audio/sfx/sfx_relief_success.wav", 2.0, func(t float64) float64 {
		// Happy chord (C major arpeggio)
		env := 1.0
		if t > 1.5 { env = (2.0 - t) * 2 }
		c := math.Sin(2 * math.Pi * 523.25 * t)
		e := math.Sin(2 * math.Pi * 659.25 * t)
		g := math.Sin(2 * math.Pi * 783.99 * t)
		return (c + e + g) / 3.0 * env * 0.8
	})

	generateWAV("client/asset/audio/sfx/sfx_relief_fail.wav", 2.0, func(t float64) float64 {
		// Sad tone (pitch drop)
		env := math.Exp(-t * 2)
		freq := 300.0 - 100.0*t
		if freq < 50 { freq = 50 }
		return math.Sin(2*math.Pi*freq*t) * env
	})
}

func generateWAV(filename string, durationSec float64, synthFunc func(t float64) float64) {
	numSamples := int(durationSec * SampleRate)
	file, err := os.Create(filename)
	if err != nil { panic(err) }
	defer file.Close()

	// RIFF header
	file.WriteString("RIFF")
	dataSize := numSamples * NumChannels * BitsPerSample / 8
	binary.Write(file, binary.LittleEndian, uint32(36+dataSize))
	file.WriteString("WAVE")

	// fmt chunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))
	binary.Write(file, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(file, binary.LittleEndian, uint16(NumChannels))
	binary.Write(file, binary.LittleEndian, uint32(SampleRate))
	byteRate := SampleRate * NumChannels * BitsPerSample / 8
	binary.Write(file, binary.LittleEndian, uint32(byteRate))
	blockAlign := NumChannels * BitsPerSample / 8
	binary.Write(file, binary.LittleEndian, uint16(blockAlign))
	binary.Write(file, binary.LittleEndian, uint16(BitsPerSample))

	// data chunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(dataSize))

	for i := 0; i < numSamples; i++ {
		t := float64(i) / SampleRate
		val := synthFunc(t)
		if val > 1.0 { val = 1.0 }
		if val < -1.0 { val = -1.0 }
		sample := int16(val * 32767)
		binary.Write(file, binary.LittleEndian, sample)
	}
}
