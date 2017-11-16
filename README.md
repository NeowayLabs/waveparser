# waveparser

Parses data from audio files

# Use

```
import "github.com/NeowayLabs/waveparser"

func main() {
    header, data, err := waveparser.LoadAudio("/path/to/audio.wav")
}
```
