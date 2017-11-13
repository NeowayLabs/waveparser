package waveparser

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type (
	riffHeader struct {
		Ident     [4]byte // RIFF
		ChunkSize uint32
		FileType  [4]byte // WAVE
	}

	riffChunkFmt struct {
		LengthOfHeader uint32
		AudioFormat    uint16
		NumChannels    uint16
		SampleRate     uint32
		BytesPerSec    uint32
		BytesPerBloc   uint16
		BitsPerSample  uint16
	}

	WavHeader struct {
		RIFFHdr      riffHeader
		RIFFChunkFmt riffChunkFmt

		FirstSamplePos uint32 // position of start of sample data
		DataBlockSize  uint32 // size of sample block (PCM data)
	}
)

func (hdr *WavHeader) String() string {
	strs := []string{
		"=== RIFF Header ===",
		fmt.Sprintf("RIFF Ident: %s", string(hdr.RIFFHdr.Ident[:])),
		fmt.Sprintf("RIFF Size: %d bytes", hdr.RIFFHdr.ChunkSize),
		fmt.Sprintf("File type: %s", string(hdr.RIFFHdr.FileType[:])),
		"=== Fmt ===",
		fmt.Sprintf("Audio format: %d", hdr.RIFFChunkFmt.AudioFormat),
		fmt.Sprintf("Number of channels: %d", hdr.RIFFChunkFmt.NumChannels),
		fmt.Sprintf("Sample rate: %d", hdr.RIFFChunkFmt.SampleRate),
		fmt.Sprintf("Bytes/seconds: %d", hdr.RIFFChunkFmt.BytesPerSec),
		fmt.Sprintf("Bytes/block: %d", hdr.RIFFChunkFmt.BytesPerBloc),
		fmt.Sprintf("Bits/sample: %d", hdr.RIFFChunkFmt.BitsPerSample),
	}
	return strings.Join(strs, "\n")
}

func abortonerr(err error, op string) {
	if err != nil {
		fmt.Printf("%s: fatal error: %s\n", op, err)
		os.Exit(1)
	}
}

func parseRIFFHeader(r io.Reader) (*riffHeader, error) {
	var hdr riffHeader
	err := binary.Read(r, binary.LittleEndian, &hdr)
	if err != nil {
		return nil, err
	}
	if string(hdr.Ident[:]) != "RIFF" {
		return nil, fmt.Errorf("Invalid RIFF identification: %s", string(hdr.Ident[:]))
	}
	return &hdr, nil
}

func parseHeader(r io.ReadSeeker) (*WavHeader, error) {
	riffhdr, err := parseRIFFHeader(r)
	if err != nil {
		return nil, err
	}

	// FMT chunk
	var chunk [4]byte
	var chunkFmt riffChunkFmt

	err = binary.Read(r, binary.LittleEndian, &chunk)
	if err != nil {
		return nil, err
	}
	if string(chunk[:]) != "fmt " {
		return nil, fmt.Errorf("Unexpected chunk type: %s", string(chunk[:]))
	}

	err = binary.Read(r, binary.LittleEndian, &chunkFmt)
	if err != nil {
		return nil, err
	}

	if chunkFmt.AudioFormat != 1 {
		return nil, fmt.Errorf("Isn't an audio format: format[%d]", chunkFmt.AudioFormat)
	}

	if chunkFmt.LengthOfHeader != 16 {
		var extraparams uint16
		// Get extra params size
		if err = binary.Read(r, binary.LittleEndian, &extraparams); err != nil {
			return nil, fmt.Errorf("error getting extra fmt params: %s", err)
		}
		// Skip
		if _, err = r.Seek(int64(extraparams), os.SEEK_CUR); err != nil {
			return nil, fmt.Errorf("error skipping extra params: %s", err)
		}
	}

	var chunkSize uint32

	for string(chunk[:]) != "data" {
		// Read chunkID
		err = binary.Read(r, binary.BigEndian, &chunk)
		if err != nil {
			return nil, fmt.Errorf("Expected data chunkid: %s", err)
		}

		err = binary.Read(r, binary.LittleEndian, &chunkSize)
		if err != nil {
			return nil, fmt.Errorf("Expected data chunkSize: %s", err)
		}

		// ignores LIST chunkIDs (unused for now)
		if string(chunk[:]) != "data" {
			if _, err = r.Seek(int64(chunkSize), os.SEEK_CUR); err != nil {
				return nil, err
			}
		}
	}

	pos, _ := r.Seek(0, os.SEEK_CUR)
	return &WavHeader{
		RIFFHdr:      *riffhdr,
		RIFFChunkFmt: chunkFmt,

		FirstSamplePos: uint32(pos),
		DataBlockSize:  uint32(chunkSize),
	}, nil
}

func LoadAudio(audiofile string) (*WavHeader, []int16) {
	audio := []int16{}
	f, err := os.Open(audiofile)
	abortonerr(err, "opening audio file")
	defer f.Close()

	hdr, err := parseHeader(f)
	abortonerr(err, "parsing WAV header")

	// header already skipped

	data, err := ioutil.ReadAll(f)
	abortonerr(err, "opening audio file")

	for i := 0; i < len(data)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(data[i : i+2])
		audio = append(audio, int16(sample))
	}

	return hdr, audio
}
