# lrcx

**Multi-purpose LRC tool.**

Download lyrics, convert subtitles, sync your music library.

## Features

| Feature | Description |
|---------|-------------|
| **LRCLib** | Download synced lyrics from community database |
| **YouTube** | Extract subtitles from YouTube videos |
| **Convert** | Transform SRT/VTT files to LRC format |
| **Interactive** | Select from multiple results |
| **Time offset** | Adjust timing for perfect sync |

## Install

```sh
go install github.com/os-guy-original/lrcx/cmd/lrcx@latest
```

## Usage

```sh
# LRCLib
lrcx lrclib "song name"
lrcx lrclib --artist "Artist" --track "Song"
lrcx lrclib --interactive "bohemian rhapsody"

# YouTube (requires yt-dlp)
lrcx yt https://youtube.com/watch?v=...
lrcx yt --interactive https://youtube.com/watch?v=...
lrcx yt --auto-captions https://youtube.com/watch?v=...

# Convert files
cat file.srt | lrcx > file.lrc
lrcx -i file.vtt -o file.lrc
```

## Help

```sh
lrcx lrclib --help
lrcx yt --help
```

## Roadmap

- [x] SRT/VTT → LRC conversion
- [x] YouTube subtitle download
- [x] LRCLib integration
- [ ] Audio file embedding (MP3, FLAC, M4A)

## License

MIT
