package waveparser

import (
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

func checkerr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func testParseWAV(t *testing.T, filename string) {
	r, err := os.Open(filename)
	checkerr(t, err)

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
		checkerr(t, err)

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

		if !reflect.DeepEqual(hdr, &expected) {
			t.Fatalf("WAV header differs: %#v != %#v", hdr, &expected)
		}
		return
	}

	t.Fatalf("no error file nor expected file found for input: %s", filename)
}
func TestParseWAV(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	checkerr(t, err)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "wav") {
			fname := file.Name()
			t.Run(fmt.Sprintf("header-%s", fname), func(t *testing.T) {
				testParseWAV(t, filepath.Join("testdata", fname))
			})
		}
	}
}
