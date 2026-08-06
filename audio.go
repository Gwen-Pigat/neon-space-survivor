package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sync"

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

	// In-Memory Audio File Cache
	musicFileCache = make(map[string][]byte)

	// Beat Energy Map & Mutex for Async Background Analysis
	beatEnergyMap   = make(map[string][]float32)
	beatEnergyMutex sync.RWMutex
	currentTrackPath string

	// Music Management
	currentMusicPlayer      *audio.Player
	nextMusicPlayer         *audio.Player
	currentStream           *mp3.Stream
	nextStream              *mp3.Stream
	musicState              int
	currentTrack            int
	isCrossfading           bool
	fadeAlpha               float64
	crossfadeDurationFrames float64 = 300.0 // Default 5s crossfade (60fps)

	// Special Effects
	alarmPlayer    *audio.Player
	isMusicStopped bool
)

func InitAudio() {
	if audioContext != nil {
		return
	}
	audioContext = audio.NewContext(sampleRate)
	preloadAudio()

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
	if data, ok := musicFileCache[path]; ok {
		return data
	}
	data, err := assetsFS.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil
		}
	}
	musicFileCache[path] = data
	return data
}

func ensureBeatAnalysis(path string, data []byte) {
	beatEnergyMutex.RLock()
	_, ok := beatEnergyMap[path]
	beatEnergyMutex.RUnlock()
	if ok || len(data) == 0 {
		return
	}
	go func() {
		energies := analyzeBeatEnergy(path, data)
		if energies != nil {
			beatEnergyMutex.Lock()
			beatEnergyMap[path] = energies
			beatEnergyMutex.Unlock()
		}
	}()
}

func analyzeBeatEnergy(path string, mp3Data []byte) []float32 {
	if len(mp3Data) == 0 {
		return nil
	}
	s, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(mp3Data))
	if err != nil {
		return nil
	}

	pcmData := make([]byte, s.Length())
	n, _ := io.ReadFull(s, pcmData)
	if n <= 0 {
		return nil
	}

	// 50 analysis frames per second
	samplesPerChunk := sampleRate / 50
	bytesPerChunk := samplesPerChunk * 4
	numChunks := len(pcmData) / bytesPerChunk
	if numChunks <= 0 {
		return nil
	}

	rawEnergies := make([]float64, numChunks)
	for c := 0; c < numChunks; c++ {
		offset := c * bytesPerChunk
		chunkBytes := pcmData[offset : offset+bytesPerChunk]

		var sumSq float64
		for i := 0; i < len(chunkBytes)-3; i += 4 {
			sampleL := int16(chunkBytes[i]) | (int16(chunkBytes[i+1]) << 8)
			val := float64(sampleL) / 32768.0
			sumSq += val * val
		}
		rawEnergies[c] = math.Sqrt(sumSq / float64(samplesPerChunk))
	}

	// Compute moving average and detect beat energy peaks above baseline
	pulses := make([]float32, numChunks)
	windowSize := 50
	for c := 0; c < numChunks; c++ {
		start := c - windowSize/2
		if start < 0 {
			start = 0
		}
		end := c + windowSize/2
		if end > numChunks {
			end = numChunks
		}
		var avg float64
		for w := start; w < end; w++ {
			avg += rawEnergies[w]
		}
		avg /= float64(end - start)

		diff := rawEnergies[c] - avg
		if diff > 0 && avg > 0.005 {
			peak := diff / avg
			if peak > 0.2 {
				pulseVal := (peak - 0.2) * 1.8
				if pulseVal > 1.0 {
					pulseVal = 1.0
				}
				pulses[c] = float32(pulseVal)
			}
		}
	}

	return pulses
}

func GetBeatPulse() float64 {
	if audioContext == nil || currentMusicPlayer == nil || currentTrackPath == "" {
		return 0.0
	}
	beatEnergyMutex.RLock()
	energies, ok := beatEnergyMap[currentTrackPath]
	beatEnergyMutex.RUnlock()
	if !ok || len(energies) == 0 {
		return 0.0
	}
	pos := currentMusicPlayer.Current().Seconds()
	frameIdx := int(pos * 50.0)
	if frameIdx < 0 || frameIdx >= len(energies) {
		return 0.0
	}
	return float64(energies[frameIdx])
}

func preloadAudio() {
	paths := []string{
		"static/audio/global/theme/theme_1.mp3",
		"static/audio/global/theme/theme_2.mp3",
		"static/audio/global/home/home_1.mp3",
		"static/audio/global/home/home_2.mp3",
		"static/audio/global/segment_last/segment_last_1.mp3",
		"static/audio/global/segment_last/segment_last_2.mp3",
		"static/audio/global/theme/wave_mode.mp3",
		"static/audio/global/alarms/alarm_1.mp3",
		"static/audio/global/losses/song.mp3",
		"static/audio/global/losses/sound.mp3",
		"static/audio/global/success/song.mp3",
		"static/audio/global/success/sound.mp3",
		"static/audio/enemies/boss/pop.mp3",
	}
	for _, p := range paths {
		loadSoundFile(p)
	}
}

func ResetMusic() {
	if audioContext == nil {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
		currentMusicPlayer = nil
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
		nextMusicPlayer = nil
	}

	musicState = 0
	currentTrack = rand.Intn(2) + 1
	loadTrack(musicState, currentTrack, true)
	isCrossfading = false
	fadeAlpha = 0
	isMusicStopped = false
}

func StopMusic() {
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
		currentMusicPlayer = nil
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
		nextMusicPlayer = nil
	}
	isMusicStopped = true
}

func StartMusic() {
	isMusicStopped = false
}

var (
	nextTrackPath string
)

func loadTrack(state int, track int, isCurrent bool) {
	if audioContext == nil {
		return
	}
	path := ""
	switch state {
	case 0:
		path = fmt.Sprintf("static/audio/global/theme/theme_%d.mp3", track)
	case 1:
		path = fmt.Sprintf("static/audio/global/home/home_%d.mp3", track)
	case 2:
		path = fmt.Sprintf("static/audio/global/segment_last/segment_last_%d.mp3", track)
	case 5:
		path = "static/audio/global/theme/wave_mode.mp3"
	}

	data := loadSoundFile(path)
	if data == nil {
		return
	}

	ensureBeatAnalysis(path, data)

	s, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
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
		currentTrackPath = path
		currentMusicPlayer.SetVolume(1.0)
		currentMusicPlayer.Play()
	} else {
		if nextMusicPlayer != nil {
			nextMusicPlayer.Close()
		}
		nextMusicPlayer = p
		nextStream = s
		nextTrackPath = path
		nextMusicPlayer.SetVolume(0.0)
		nextMusicPlayer.Play()
	}
}

func UpdateMusic(timer int, started bool, isHardMode bool) {
	if audioContext == nil {
		return
	}
	// Don't update normal music during Game Over or Victory states (States 3 and 4)
	if musicState >= 3 {
		return
	}

	minutes := timer / 3600
	newState := 0
	if started {
		if isHardMode {
			newState = 5 // Hard Mode track (wave_mode.mp3)
		} else if minutes >= 8 {
			newState = 2
		} else {
			newState = 1
		}
	}

	// 1. Handle State Changes (e.g. Menu State 0 -> Wave Mode State 5 or Game State 1)
	if newState != musicState {
		if isMusicStopped {
			return
		}
		if newState == 2 {
			PlayTransition()
		}
		musicState = newState
		currentTrack = rand.Intn(2) + 1

		// Immediately start new mode track (e.g. wave_mode.mp3) so main menu song stops cleanly
		loadTrack(musicState, currentTrack, true)
		isCrossfading = false
		fadeAlpha = 0.0
		return
	}

	if currentMusicPlayer == nil || isMusicStopped {
		return
	}

	// 2. Continuous Track Loop / Crossfade when track is finishing
	if currentStream != nil && !isCrossfading {
		pos := currentMusicPlayer.Current().Seconds()
		length := float64(currentStream.Length()) / 4.0 / float64(sampleRate)
		fadeStart := length - 5.0 // Start crossfade 5 seconds before track ends

		if fadeStart > 0 && pos >= fadeStart {
			isCrossfading = true
			fadeAlpha = 0.0
			crossfadeDurationFrames = 300.0 // 5 second smooth crossfade for looping tracks

			nextTrack := 1
			if currentTrack == 1 {
				nextTrack = 2
			} else {
				nextTrack = 1
			}
			loadTrack(musicState, nextTrack, false)
			currentTrack = nextTrack
		}
	}

	// 3. Process Active Crossfade Volumes
	if isCrossfading {
		fadeAlpha += 1.0 / crossfadeDurationFrames
		if fadeAlpha >= 1.0 {
			fadeAlpha = 1.0
			isCrossfading = false
			if currentMusicPlayer != nil {
				currentMusicPlayer.Close()
			}
			currentMusicPlayer = nextMusicPlayer
			currentStream = nextStream
			currentTrackPath = nextTrackPath
			if currentMusicPlayer != nil {
				currentMusicPlayer.SetVolume(1.0)
			}
			nextMusicPlayer = nil
			nextStream = nil
		} else {
			if currentMusicPlayer != nil {
				currentMusicPlayer.SetVolume(1.0 - fadeAlpha)
			}
			if nextMusicPlayer != nil {
				nextMusicPlayer.SetVolume(fadeAlpha)
			}
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
	data := loadSoundFile("static/audio/global/alarms/alarm_1.mp3")
	if data == nil {
		return
	}
	s, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return
	}
	l := audio.NewInfiniteLoop(s, s.Length())
	p, err := audioContext.NewPlayer(l)
	if err != nil {
		return
	}
	alarmPlayer = p
	alarmPlayer.Play()
}

func StopAlarm() {
	if alarmPlayer != nil {
		alarmPlayer.Close()
		alarmPlayer = nil
	}
}

func PlayTransition() {
	if audioContext == nil || transitionBuf == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(transitionBuf)
	if p != nil {
		p.Play()
	}
}

func PlayOverdrive() {
	if audioContext == nil || transitionBuf == nil {
		return
	}
	p := audioContext.NewPlayerFromBytes(transitionBuf)
	if p != nil {
		p.SetVolume(0.8)
		p.Play()
	}
}

func PlayDefeat() {
	if audioContext == nil || musicState == 3 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
		currentMusicPlayer = nil
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
		nextMusicPlayer = nil
	}
	musicState = 3
	// Sound
	if lostSoundBuf != nil {
		if s1, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(lostSoundBuf)); err == nil {
			if p1, err := audioContext.NewPlayer(s1); err == nil {
				p1.Play()
			}
		}
	}
	// Song
	if data := loadSoundFile("static/audio/global/losses/song.mp3"); data != nil {
		if s2, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data)); err == nil {
			if l2 := audio.NewInfiniteLoop(s2, s2.Length()); l2 != nil {
				if p2, err := audioContext.NewPlayer(l2); err == nil {
					currentMusicPlayer = p2
					currentMusicPlayer.Play()
				}
			}
		}
	}
}

func PlayVictory() {
	if audioContext == nil || musicState == 4 {
		return
	}
	StopAlarm()
	if currentMusicPlayer != nil {
		currentMusicPlayer.Close()
		currentMusicPlayer = nil
	}
	if nextMusicPlayer != nil {
		nextMusicPlayer.Close()
		nextMusicPlayer = nil
	}
	musicState = 4
	// Sound
	if successSoundBuf != nil {
		if s1, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(successSoundBuf)); err == nil {
			if p1, err := audioContext.NewPlayer(s1); err == nil {
				p1.Play()
			}
		}
	}
	// Song
	if data := loadSoundFile("static/audio/global/success/song.mp3"); data != nil {
		if s2, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data)); err == nil {
			if l2 := audio.NewInfiniteLoop(s2, s2.Length()); l2 != nil {
				if p2, err := audioContext.NewPlayer(l2); err == nil {
					currentMusicPlayer = p2
					currentMusicPlayer.Play()
				}
			}
		}
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
	if audioContext != nil && explosionBuf != nil {
		p := audioContext.NewPlayerFromBytes(explosionBuf)
		if p != nil {
			p.Play()
		}
	}
}
func PlayHit() {
	if audioContext != nil && hitBuf != nil {
		p := audioContext.NewPlayerFromBytes(hitBuf)
		if p != nil {
			p.Play()
		}
	}
}
