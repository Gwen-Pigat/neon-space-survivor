package main

import (
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

const (
	sampleRate = 44100
)

var (
	audioContext *audio.Context
	laserBuf     []byte
	explosionBuf []byte
	hitBuf       []byte
	transitionBuf []byte

	// Music Management
	currentMusicPlayer *audio.Player
	nextMusicPlayer    *audio.Player
	currentStream      *mp3.Stream
	nextStream         *mp3.Stream

	musicState     int
	currentTrack   int
	isCrossfading  bool
	fadeAlpha      float64

	// Special Effects
	alarmPlayer *audio.Player
)

func InitAudio() {
	if audioContext != nil {
		return
	}
	audioContext = audio.NewContext(sampleRate)

	laserBuf = generateLaser()
	explosionBuf = generateExplosion()
	hitBuf = generateHit()
	transitionBuf = generateTransition()

	musicState = 0
	currentTrack = rand.Intn(2) + 1
	loadTrack(musicState, currentTrack, true)
}

func loadTrack(state int, track int, isCurrent bool) {
	path := ""
	switch state {
	case 0:
		path = "static/audio/main_theme_" + string(rune('0'+track)) + ".mp3"
	case 1:
		path = "static/audio/Start_" + string(rune('0'+track)) + ".mp3"
	case 2:
		path = "static/audio/End_" + string(rune('0'+track)) + ".mp3"
	}

	f, err := os.Open(path)
	if err != nil {
		return
	}
	s, err := mp3.DecodeWithSampleRate(sampleRate, f)
	if err != nil {
		return
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return
	}

	if isCurrent {
		if currentMusicPlayer != nil {
			currentMusicPlayer.Close()
		}
		currentMusicPlayer = p
		currentStream = s
		currentMusicPlayer.Play()
	} else {
		if nextMusicPlayer != nil {
			nextMusicPlayer.Close()
		}
		nextMusicPlayer = p
		nextStream = s
		nextMusicPlayer.Play()
		nextMusicPlayer.SetVolume(0)
	}
}

func UpdateMusic(timer int, started bool) {
	// Don't update normal music during Game Over or Victory states (States 3 and 4)
	if musicState >= 3 {
		return
	}

	minutes := timer / 3600
	newState := 0
	if started {
		if minutes >= 8 {
			newState = 2
		} else {
			newState = 1
		}
	}

	if newState != musicState {
		if newState == 2 {
			PlayTransition()
		}
		musicState = newState
		currentTrack = rand.Intn(2) + 1
		loadTrack(musicState, currentTrack, true)
		isCrossfading = false
		fadeAlpha = 0
		return
	}

	if currentMusicPlayer == nil {
		return
	}

	pos := currentMusicPlayer.Current().Seconds()
	length := float64(currentStream.Length()) / 4 / sampleRate
	fadeStart := length - 8.0

	if !isCrossfading && pos > fadeStart {
		isCrossfading = true
		fadeAlpha = 0
		nextTrack := 1
		if currentTrack == 1 {
			nextTrack = 2
		} else {
			nextTrack = 1
		}
		loadTrack(musicState, nextTrack, false)
		currentTrack = nextTrack
	}

	if isCrossfading {
		fadeAlpha += 1.0 / (60.0 * 5.0)
		if fadeAlpha >= 1.0 {
			fadeAlpha = 1.0
			isCrossfading = false
			if currentMusicPlayer != nil {
				currentMusicPlayer.Close()
			}
			currentMusicPlayer = nextMusicPlayer
			currentStream = nextStream
			nextMusicPlayer = nil
			nextStream = nil
		}
		if currentMusicPlayer != nil {
			currentMusicPlayer.SetVolume(1.0 - fadeAlpha)
		}
		if nextMusicPlayer != nil {
			nextMusicPlayer.SetVolume(fadeAlpha)
		}
	} else {
		currentMusicPlayer.SetVolume(1.0)
	}
}

func PlayBossAppears() {
	if audioContext == nil {
		return
	}
	f, _ := os.Open("static/audio/boss_appears.mp3")
	s, _ := mp3.DecodeWithSampleRate(sampleRate, f)
	p, _ := audioContext.NewPlayer(s)
	p.Play()
}

func StartAlarm() {
	if audioContext == nil || alarmPlayer != nil {
		return
	}
	f, err := os.Open("static/audio/Alarm, Boss Appears.mp3")
	if err != nil {
		return
	}
	s, _ := mp3.DecodeWithSampleRate(sampleRate, f)
	l := audio.NewInfiniteLoop(s, s.Length())
	alarmPlayer, _ = audioContext.NewPlayer(l)
	alarmPlayer.Play()
}

func StopAlarm() {
	if alarmPlayer != nil {
		alarmPlayer.Close()
		alarmPlayer = nil
	}
}

func PlayTransition() {
	if audioContext == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(transitionBuf)
	p.Play()
}

func PlayDefeat() {
	if audioContext == nil || musicState == 3 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil { currentMusicPlayer.Close() }
	if nextMusicPlayer != nil { nextMusicPlayer.Close() }
	musicState = 3

	// Sound
	f1, _ := os.Open("static/audio/lost_sound.mp3")
	s1, _ := mp3.DecodeWithSampleRate(sampleRate, f1)
	p1, _ := audioContext.NewPlayer(s1)
	p1.Play()

	// Song
	f2, _ := os.Open("static/audio/lost_song.mp3")
	s2, _ := mp3.DecodeWithSampleRate(sampleRate, f2)
	l2 := audio.NewInfiniteLoop(s2, s2.Length())
	currentMusicPlayer, _ = audioContext.NewPlayer(l2)
	currentMusicPlayer.Play()
}

func PlayVictory() {
	if audioContext == nil || musicState == 4 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil { currentMusicPlayer.Close() }
	if nextMusicPlayer != nil { nextMusicPlayer.Close() }
	musicState = 4

	// Sound
	f1, _ := os.Open("static/audio/success_sound.mp3")
	s1, _ := mp3.DecodeWithSampleRate(sampleRate, f1)
	p1, _ := audioContext.NewPlayer(s1)
	p1.Play()

	// Song
	f2, _ := os.Open("static/audio/success_song.mp3")
	s2, _ := mp3.DecodeWithSampleRate(sampleRate, f2)
	l2 := audio.NewInfiniteLoop(s2, s2.Length())
	currentMusicPlayer, _ = audioContext.NewPlayer(l2)
	currentMusicPlayer.Play()
}

func generateTransition() []byte {
	duration := 2.0
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		freq := 200.0 + (t/duration)*800.0
		v := math.Sin(2.0 * math.Pi * freq * t) * math.Sin(math.Pi * t / duration) * 0.4
		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}

func generateLaser() []byte {
	duration := 0.15
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		freq := 1200.0 * math.Exp(-t*15.0)
		v := math.Sin(2.0 * math.Pi * freq * t) * (1.0 - t/duration) * 0.3
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
		v := (rand.Float64()*2.0 - 1.0) * math.Exp(-t*5.0) * 0.4
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
		freq := 150.0
		v := math.Sin(2.0 * math.Pi * freq * t)
		if v > 0 { v = 0.5 } else { v = -0.5 }
		v *= math.Exp(-t * 10.0) * 0.5
		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}

func PlayLaser()     {}
func PlayExplosion() { if audioContext != nil { audioContext.NewPlayerFromBytes(explosionBuf).Play() } }
func PlayHit()       { if audioContext != nil { audioContext.NewPlayerFromBytes(hitBuf).Play() } }
