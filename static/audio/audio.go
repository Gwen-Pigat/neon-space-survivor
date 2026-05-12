package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	sampleRate = 44100
)

var (
	audioContext *audio.Context
	laserBuf     []byte
	explosionBuf []byte
	hitBuf       []byte
)

func InitAudio() {
	if audioContext != nil {
		return
	}
	audioContext = audio.NewContext(sampleRate)

	laserBuf = generateLaser()
	explosionBuf = generateExplosion()
	hitBuf = generateHit()
}

func generateLaser() []byte {
	duration := 0.15
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		// Quick pitch drop
		freq := 1200.0 * math.Exp(-t*15.0)
		v := math.Sin(2.0 * math.Pi * freq * t)
		v *= (1.0 - t/duration) // Fade out
		v *= 0.3                // Volume

		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}

func generateExplosion() []byte {
	duration := 0.5
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		// White noise
		v := rand.Float64()*2.0 - 1.0
		// Low pass filter approximation: favor lower frequencies as it fades
		v *= math.Exp(-t * 5.0)
		v *= 0.4 // Volume

		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}

func generateHit() []byte {
	duration := 0.2
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		// Square-ish wave for "crunchy" feel
		freq := 150.0
		v := math.Sin(2.0 * math.Pi * freq * t)
		if v > 0 {
			v = 0.5
		} else {
			v = -0.5
		}
		v *= math.Exp(-t * 10.0)
		v *= 0.5 // Volume

		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}

func PlayLaser() {
	if audioContext == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(laserBuf)
	p.Play()
}

func PlayExplosion() {
	if audioContext == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(explosionBuf)
	p.Play()
}

func PlayHit() {
	if audioContext == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(hitBuf)
	p.Play()
}
