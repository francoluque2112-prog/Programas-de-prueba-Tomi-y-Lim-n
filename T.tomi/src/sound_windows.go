package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"syscall"
	"unsafe"
)

var (
	winmmDLL      = syscall.NewLazyDLL("winmm.dll")
	procPlaySound = winmmDLL.NewProc("PlaySoundW")
)

const (
	SND_SYNC   = 0x0000
	SND_ASYNC  = 0x0001
	SND_MEMORY = 0x0004
)

// generateWAV generates a standard 16-bit 22050Hz mono PCM WAV in memory
func generateWAV(sampleRate int, durationSec float64, freqFn func(t float64) float64, ampFn func(t float64) float64) []byte {
	numSamples := int(float64(sampleRate) * durationSec)
	var buf bytes.Buffer

	// WAV Header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+numSamples*2))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16)) // Subchunk1Size
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // AudioFormat (PCM)
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // NumChannels (1 mono)
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2)) // ByteRate
	binary.Write(&buf, binary.LittleEndian, uint16(2))            // BlockAlign
	binary.Write(&buf, binary.LittleEndian, uint16(16))           // BitsPerSample
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(numSamples*2))

	phase := 0.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		freq := freqFn(t)
		amp := ampFn(t)
		phase += 2.0 * math.Pi * freq / float64(sampleRate)

		// Soft clipping / harmonic warmth
		val := math.Sin(phase) + 0.2*math.Sin(phase*2.0)
		sample := int16(val * amp * 30000.0)
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	return buf.Bytes()
}

var (
	soundBark    []byte
	soundPickup  []byte
	soundCuna    []byte
	soundWin     []byte
	soundLose    []byte
)

func initSounds() {
	// 1. Bark / Wakeup Sound (two quick woofs)
	soundBark = generateWAV(22050, 0.35, func(t float64) float64 {
		if t < 0.15 {
			return 280.0 + 80.0*math.Sin(t*40.0)
		}
		return 350.0 + 100.0*math.Sin(t*40.0)
	}, func(t float64) float64 {
		if t < 0.15 {
			return math.Exp(-t * 25.0)
		}
		t2 := t - 0.15
		return math.Exp(-t2 * 20.0)
	})

	// 2. Bone pickup sound (cute bubble pop)
	soundPickup = generateWAV(22050, 0.18, func(t float64) float64 {
		return 400.0 + t*1200.0
	}, func(t float64) float64 {
		return math.Exp(-t * 18.0)
	})

	// 3. Bone delivered to Cuna (bright bell chime)
	soundCuna = generateWAV(22050, 0.45, func(t float64) float64 {
		if t < 0.2 {
			return 659.25 // E5
		}
		return 880.0 // A5
	}, func(t float64) float64 {
		if t < 0.2 {
			return 0.8 * math.Exp(-t*10.0)
		}
		return 0.9 * math.Exp(-(t-0.2)*8.0)
	})

	// 4. Win fanfare (tada! C-E-G-C high)
	soundWin = generateWAV(22050, 0.8, func(t float64) float64 {
		if t < 0.2 {
			return 523.25 // C5
		} else if t < 0.4 {
			return 659.25 // E5
		} else if t < 0.6 {
			return 783.99 // G5
		}
		return 1046.50 // C6
	}, func(t float64) float64 {
		return 0.8 * (1.0 - t/0.8)
	})

	// 5. Lose / Catch sound (crunch / chomp)
	soundLose = generateWAV(22050, 0.5, func(t float64) float64 {
		return 180.0 - t*120.0 + 30.0*math.Sin(t*100.0)
	}, func(t float64) float64 {
		return 0.9 * math.Exp(-t*6.0)
	})
}

func playSoundBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	procPlaySound.Call(
		uintptr(unsafe.Pointer(&data[0])),
		0,
		uintptr(SND_MEMORY|SND_ASYNC),
	)
}

func PlayBarkSound()   { playSoundBytes(soundBark) }
func PlayPickupSound() { playSoundBytes(soundPickup) }
func PlayCunaSound()   { playSoundBytes(soundCuna) }
func PlayWinSound()    { playSoundBytes(soundWin) }
func PlayLoseSound()   { playSoundBytes(soundLose) }
