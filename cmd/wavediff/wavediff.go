package main

import (
	"fmt"
	"os"

	"github.com/NeowayLabs/waveparser"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("usage: %s <wav file> <other wav file>\n", os.Args[0])
		return
	}

	wavpath1 := os.Args[1]
	wavpath2 := os.Args[2]

	wav1, err := waveparser.Load(wavpath1)
	abortonerr(err, "loading [%s]", wavpath1)

	wav2, err := waveparser.Load(wavpath2)
	abortonerr(err, "loading [%s]", wavpath2)

	if diffHeaders(wavpath1, wav1.Header, wavpath2, wav2.Header) {
		os.Exit(-1)
	}
}

func diffHeaders(
	wavpath1 string, h1 waveparser.WavHeader,
	wavpath2 string, h2 waveparser.WavHeader,
) bool {
	wroteHeader := false
	writeHeader := func() {
		if wroteHeader {
			return
		}
		wroteHeader = true
		fmt.Printf("\n[%s] header differs from [%s] header\n", wavpath1, wavpath2)
		fmt.Printf("[%s] values will be on the left, [%s] on the right\n\n", wavpath1, wavpath2)
	}
	writeDiff := func(f string, args ...interface{}) {
		writeHeader()
		fmt.Println(fmt.Sprintf(f, args...))
	}

	ident1 := h1.RIFFHdr.Ident
	ident2 := h2.RIFFHdr.Ident

	for i, b := range ident1 {
		if ident2[i] != b {
			writeDiff("RIFF Ident Byte[%d] differs: [%x] != [%x]", b, ident2[i])
		}
	}

	chunksize1 := h1.RIFFHdr.ChunkSize
	chunksize2 := h2.RIFFHdr.ChunkSize

	if chunksize1 != chunksize2 {
		writeDiff("ChunkSize: [%d] != [%d]", chunksize1, chunksize2)
	}

	ft1 := h1.RIFFHdr.FileType
	ft2 := h2.RIFFHdr.FileType

	for i, b := range ft1 {
		if ft2[i] != b {
			writeDiff("FileType Byte[%d] differs: [%x] != [%x]", b, ft2[i])
		}
	}

	cf1 := h1.RIFFChunkFmt
	cf2 := h2.RIFFChunkFmt

	if cf1.LengthOfHeader != cf2.LengthOfHeader {
		writeDiff("Length Of Header: [%d] != [%d]", cf1.LengthOfHeader, cf2.LengthOfHeader)
	}

	if cf1.AudioFormat != cf2.AudioFormat {
		writeDiff("Audio Format: [%d] != [%d]", cf1.AudioFormat, cf2.AudioFormat)
	}

	if cf1.NumChannels != cf2.NumChannels {
		writeDiff("Number Of Channels: [%d] != [%d]", cf1.NumChannels, cf2.NumChannels)
	}

	if cf1.SampleRate != cf2.SampleRate {
		writeDiff("Samplerate: [%d] != [%d]", cf1.SampleRate, cf2.SampleRate)
	}

	if cf1.BytesPerSec != cf2.BytesPerSec {
		writeDiff("Bytes Per Sec: [%d] != [%d]", cf1.BytesPerSec, cf2.BytesPerSec)
	}

	if cf1.BytesPerBloc != cf2.BytesPerBloc {
		writeDiff("Bytes Per Sec: [%d] != [%d]", cf1.BytesPerBloc, cf2.BytesPerBloc)
	}

	if cf1.BitsPerSample != cf2.BitsPerSample {
		writeDiff("Bits Per Sample: [%d] != [%d]", cf1.BitsPerSample, cf2.BitsPerSample)
	}

	if h1.FirstSamplePos != h2.FirstSamplePos {
		writeDiff("First Sample Position: [%d] != [%d]", h1.FirstSamplePos, h2.FirstSamplePos)
	}

	if h1.DataBlockSize != h2.DataBlockSize {
		writeDiff("Data Block Size: [%d] != [%d]", h1.DataBlockSize, h2.DataBlockSize)
	}

	return wroteHeader
}

func abortonerr(err error, f string, args ...interface{}) {
	if err == nil {
		return
	}

	panic(fmt.Sprintf("error: [%s] %s", err, fmt.Sprintf(f, args...)))
}
