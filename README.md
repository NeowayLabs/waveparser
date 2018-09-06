# waveparser

Parses data from audio files

# Use

```
import "github.com/NeowayLabs/waveparser"

func main() {
    wav, err := waveparser.LoadAudio("/path/to/audio.wav")
}
```

# Wave Diff

There is also a tool that helps you to check differences on the header
of wave files, just run:

```
go install github.com/NeowayLabs/waveparser/cmd/wavediff
```

And then:

```
wavediff <wavfile1> <wavfile2>
```

It will only compare the header, not the audio contents. It has been useful
to debug problems when tools like **file** and **ffmpeg** indicates that files
have the same type (samplerate, endianess, etc) but in the end one of them
does not work properly on some tools (like audacity, happened to me =().