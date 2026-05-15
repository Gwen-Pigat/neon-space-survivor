package main

import (
	"bytes"
	"fmt"
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
	audioContext    *audio.Context
	laserBuf        []byte
	explosionBuf    []byte
	hitBuf          []byte
	transitionBuf   []byte
	bossAppearsBuf  []byte
	lostSoundBuf    []byte
	successSoundBuf []byte
	// Music Management
	currentMusicPlayer *audio.Player
	nextMusicPlayer    *audio.Player
	currentStream      *mp3.Stream
	nextStream         *mp3.Stream
	musicState         int
	currentTrack       int
	isCrossfading      bool
	fadeAlpha          float64
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
	bossAppearsBuf = loadSoundFile("static/audio/enemies/boss/pop.mp3")
	lostSoundBuf = loadSoundFile("static/audio/global/losses/sound.mp3")
	successSoundBuf = loadSoundFile("static/audio/global/success/sound.mp3")
	ResetMusic()
}
func loadSoundFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}
func ResetMusic() {
	if audioContext == nil {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
	}

	musicState = 0
	currentTrack = rand.Intn(2) + 1
	loadTrack(musicState, currentTrack, true)
	isCrossfading = false
	fadeAlpha = 0
}
func loadTrack(state int, track int, isCurrent bool) {
	path := ""
	switch state {
	case 0:
		path = fmt.Sprintf("static/audio/global/theme/theme_%d.mp3", track)
	case 1:
		path = fmt.Sprintf("static/audio/global/home/home_%d.mp3", track)
	case 2:
		path = fmt.Sprintf("static/audio/global/segment_last/segment_last_%d.mp3", track)
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
	if audioContext == nil || bossAppearsBuf == nil {
		return
	}
	s, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(bossAppearsBuf))
	if err != nil {
		return
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return
	}
	p.Play()
}
func StartAlarm() {
	if audioContext == nil || alarmPlayer != nil {
		return
	}
	f, err := os.Open("static/audio/global/alarms/alarm_1.mp3")
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
func PlayOverdrive() {
	if audioContext == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(transitionBuf)
	p.SetVolume(0.8)
	p.Play()
}
func PlayDefeat() {
	if audioContext == nil || musicState == 3 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
	}
	musicState = 3
	// Sound
	if lostSoundBuf != nil {
		s1, _ := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(lostSoundBuf))
		p1, _ := audioContext.NewPlayer(s1)
		p1.Play()
	}
	// Song
	f2, err := os.Open("static/audio/global/losses/song.mp3")
	if err == nil {
		s2, _ := mp3.DecodeWithSampleRate(sampleRate, f2)
		l2 := audio.NewInfiniteLoop(s2, s2.Length())
		currentMusicPlayer, _ = audioContext.NewPlayer(l2)
		currentMusicPlayer.Play()
	}
}
func PlayVictory() {
	if audioContext == nil || musicState == 4 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
	}
	musicState = 4
	// Sound
	if successSoundBuf != nil {
		s1, _ := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(successSoundBuf))
		p1, _ := audioContext.NewPlayer(s1)
		p1.Play()
	}
	// Song
	f2, err := os.Open("static/audio/global/success/song.mp3")
	if err == nil {
		s2, _ := mp3.DecodeWithSampleRate(sampleRate, f2)
		l2 := audio.NewInfiniteLoop(s2, s2.Length())
		currentMusicPlayer, _ = audioContext.NewPlayer(l2)
		currentMusicPlayer.Play()
	}
}
func generateTransition() []byte {
	duration := 2.0
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		freq := 200.0 + (t/duration)*800.0
		v := math.Sin(2.0*math.Pi*freq*t) * math.Sin(math.Pi*t/duration) * 0.4
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
		v := math.Sin(2.0*math.Pi*freq*t) * (1.0 - t/duration) * 0.3
		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}
func generateExplosion() []byte {
	duration := 0.3
	numSamples := int(sampleRate * duration)
	buf := make([]byte, numSamples*4)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		// Reduced volume and faster decay for a softer "pop" sound
		v := (rand.Float64()*2.0 - 1.0) * math.Exp(-t*10.0) * 0.15
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
		if v > 0 {
			v = 0.5
		} else {
			v = -0.5
		}
		v *= math.Exp(-t*10.0) * 0.5
		s := int16(v * 32767)
		buf[4*i] = byte(s)
		buf[4*i+1] = byte(s >> 8)
		buf[4*i+2] = byte(s)
		buf[4*i+3] = byte(s >> 8)
	}
	return buf
}
func PlayLaser() {}
func PlayExplosion() {
	if audioContext != nil {
		audioContext.NewPlayerFromBytes(explosionBuf).Play()
	}
}
func PlayHit() {
	if audioContext != nil {
		audioContext.NewPlayerFromBytes(hitBuf).Play()
	}
}
