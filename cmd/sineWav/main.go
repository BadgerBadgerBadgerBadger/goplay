package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

func main() {
	buf, err := createSineWaveWAV()
	if err != nil {
		panic(err)
	}

	// Write the WAV data to a file (optional)
	err = os.WriteFile("sine_wave.wav", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

func createSineWaveWAV() (*bytes.Buffer, error) {
	const (
		sampleRate = 44100
		duration   = 1.0
		frequency  = 440.0
	)

	numSamples := int(duration * sampleRate)
	samples := make([]int16, numSamples)
	for i := range samples {
		sample := 32767 * math.Sin(2*math.Pi*frequency*float64(i)/float64(sampleRate))
		samples[i] = int16(sample)
	}

	var buf bytes.Buffer
	err := writeWAV(&buf, samples, sampleRate)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func writeWAV(w io.Writer, samples []int16, sampleRate int) error {
	const (
		bitsPerSample = 16
		numChannels   = 1
	)

	dataSize := len(samples) * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	byteRate := sampleRate * blockAlign

	_, err := fmt.Fprint(w, "RIFF")
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(dataSize+36))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "WAVEfmt ")
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(16)) // Subchunk1Size
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint16(1)) // AudioFormat
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint16(numChannels))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(sampleRate))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(byteRate))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint16(blockAlign))
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint16(bitsPerSample))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, "data")
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(dataSize))
	if err != nil {
		return err
	}
	for _, sample := range samples {
		err = binary.Write(w, binary.LittleEndian, sample)
		if err != nil {
			return err
		}
	}

	return nil
}
