package waveparser

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type wavExpectedHeader struct {
	RIFFHeader struct {
		Ident     string
		ChunkSize uint32
		FileType  string
	}
	RIFFChunkFmt   RiffChunkFmt
	FirstSamplePos uint32
	DataBlockSize  uint32
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func newWaveFloat(data []byte) Wav {
	var wav Wav
	wav.Header.RIFFChunkFmt.AudioFormat = WaveFormatIEEEFloat
	wav.Header.RIFFChunkFmt.NumChannels = 1
	wav.Data = data
	return wav
}

func TestFloatSamplesMustBeNormalized(t *testing.T) {

	type tcase struct {
		name    string
		samples []float32
		success bool
	}

	tcases := []tcase{
		tcase{
			name: "validValues",
			samples: []float32{
				-1.0,
				-0.99,
				0,
				0.99,
				1.0,
			},
			success: true,
		},
		tcase{
			name:    "firstBellowRange",
			samples: []float32{-1.01, -0.99, 0},
			success: false,
		},
		tcase{
			name:    "secondBellowRange",
			samples: []float32{-0.99, -1.01, 0},
			success: false,
		},
		tcase{
			name:    "lastBellowRange",
			samples: []float32{-1.00, -0.99, -1.01},
			success: false,
		},
		tcase{
			name:    "firstAboveRange",
			samples: []float32{1.01, 0.99, 0},
			success: false,
		},
		tcase{
			name:    "secondAboveRange",
			samples: []float32{0.99, 1.01, 0},
			success: false,
		},
		tcase{
			name:    "lastAboveRange",
			samples: []float32{1.00, 0.99, 1.01},
			success: false,
		},
	}

	for _, tcase := range tcases {
		t.Run(tcase.name, func(t *testing.T) {
			data := &bytes.Buffer{}
			err := binary.Write(data, binary.LittleEndian, tcase.samples)
			assertNoError(t, err)
			wav := newWaveFloat(data.Bytes())
			_, err = wav.Float32LESamples()
			if tcase.success {
				assertNoError(t, err)
			} else {
				assertError(t, err)
			}
		})
	}
}

func testParseWAV(t *testing.T, filename string) {

	r, err := os.Open(filename)
	assertNoError(t, err)

	hdr, hdrerr := parseHeader(r)

	ext := filepath.Ext(filename)
	noext := strings.TrimSuffix(filename, ext)
	expectedHdrFile := noext + ".hdr.expected"
	errFile := noext + ".err"

	expectedErr, err := ioutil.ReadFile(errFile)
	if err == nil {
		if hdrerr == nil {
			t.Fatalf("Expected error: %s but ran successfully...", string(expectedErr))
		}
		if hdrerr.Error() != string(expectedErr) {
			t.Fatalf("Error differs: '%s' != '%s'", hdrerr, string(expectedErr))
		}
		return
	} else if hdrerr != nil {
		t.Fatalf("Error: %s", hdrerr)
	}

	expectedHdrContent, err := ioutil.ReadFile(expectedHdrFile)
	if err == nil {
		var expectedHdr wavExpectedHeader
		err = json.Unmarshal(expectedHdrContent, &expectedHdr)
		assertNoError(t, err)

		expected := WavHeader{
			RIFFHdr: RiffHeader{
				ChunkSize: expectedHdr.RIFFHeader.ChunkSize,
			},
			RIFFChunkFmt:   expectedHdr.RIFFChunkFmt,
			FirstSamplePos: expectedHdr.FirstSamplePos,
			DataBlockSize:  expectedHdr.DataBlockSize,
		}

		// adjust expected file because JSON spec do not support char/runes
		expected.RIFFHdr.Ident[0] = expectedHdr.RIFFHeader.Ident[0]
		expected.RIFFHdr.Ident[1] = expectedHdr.RIFFHeader.Ident[1]
		expected.RIFFHdr.Ident[2] = expectedHdr.RIFFHeader.Ident[2]
		expected.RIFFHdr.Ident[3] = expectedHdr.RIFFHeader.Ident[3]

		expected.RIFFHdr.FileType[0] = expectedHdr.RIFFHeader.FileType[0]
		expected.RIFFHdr.FileType[1] = expectedHdr.RIFFHeader.FileType[1]
		expected.RIFFHdr.FileType[2] = expectedHdr.RIFFHeader.FileType[2]
		expected.RIFFHdr.FileType[3] = expectedHdr.RIFFHeader.FileType[3]

		if !reflect.DeepEqual(hdr, expected) {
			t.Fatalf("WAV header differs:\n\n%#v\n\n!=\n\n%#v\n", hdr, expected)
		}
		return
	}

	t.Fatalf("no error file nor expected file found for input: %s", filename)
}

func TestParseWAVHeaders(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	assertNoError(t, err)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "wav") {
			fname := file.Name()
			t.Run(fmt.Sprintf("header-%s", fname), func(t *testing.T) {
				testParseWAV(t, filepath.Join("testdata", fname))
			})
		}
	}
}

func TestSignedInt16LittleEndianSamples(t *testing.T) {

	testSamplesRetrieve(t, "sint16le", func(wav *Wav) *bytes.Buffer {
		samples, err := wav.Int16LESamples()
		assertNoError(t, err)

		gotbuf := &bytes.Buffer{}
		err = binary.Write(gotbuf, binary.LittleEndian, samples)
		assertNoError(t, err)

		return gotbuf
	})
}

func TestFloat32LittleEndianSamples(t *testing.T) {

	testSamplesRetrieve(t, "float32le", func(wav *Wav) *bytes.Buffer {
		samples, err := wav.Float32LESamples()
		assertNoError(t, err)

		gotbuf := &bytes.Buffer{}
		err = binary.Write(gotbuf, binary.LittleEndian, samples)
		assertNoError(t, err)
		return gotbuf
	})
}

type SamplesRetriever func(*Wav) *bytes.Buffer

func testSamplesRetrieve(t *testing.T, audioname string, retrieveSamples SamplesRetriever) {

	wav, err := Load(fmt.Sprintf("testdata/audios/%s.wav", audioname))
	assertNoError(t, err)

	samples := retrieveSamples(wav)

	expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/audios/%s.raw", audioname))
	assertNoError(t, err)

	assertBytesEqual(t, expected, samples.Bytes())
}

func assertBytesEqual(t *testing.T, expected []byte, got []byte) {
	if len(expected) != len(got) {
		t.Fatalf("expected len[%d] != got len[%d]", len(expected), len(got))
	}

	for i, expectedByte := range expected {
		gotByte := got[i]
		if expectedByte != gotByte {
			t.Fatalf("got wrong byte at index[%d] expected[%d] got[%d]", i, expectedByte, gotByte)
		}
	}
}
