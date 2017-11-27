package waveparser

import (
	"bytes"
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

	Wav struct {
		Header *WavHeader
		Data   []byte
	}
)

func Load(audiofile string) (*Wav, error) {
	f, err := os.Open(audiofile)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	hdr, err := parseHeader(f)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return &Wav{
		Header: hdr,
		Data:   data,
	}, nil
}

func (w *Wav) Int16LESamples() ([]int16, error) {
	// TODO: validate using header
	audio := []int16{}
	for i := 0; i < len(w.Data)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(w.Data[i : i+2]))
		audio = append(audio, sample)
	}
	return audio, nil
}

func (w *Wav) Float32LESamples() ([]float32, error) {
	// TODO: validate using header
	const typesize = 4

	audio := []float32{}
	reader := bytes.NewBuffer(w.Data)
	var err error

	for err == nil {
		var sample float32
		err = binary.Read(reader, binary.LittleEndian, &sample)
		if err == nil {
			audio = append(audio, sample)
		}
	}

	if err != io.EOF {
		return nil, fmt.Errorf("error[%s] loading audio as float32 samples", err)
	}

	return audio, nil
}

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
