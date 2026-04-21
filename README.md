> 🕯️ Turkey lost children to school shootings in Siverek and Kahramanmaraş. We are aware, and we grieve. Our deepest condolences to all the families.

# lrcx

**Download lyrics from YouTube and LRCLib, convert SRT/VTT subtitles to LRC format.**

## Why?

In my previous project [ncsdl](https://github.com/os-guy-original/ncsdl), most of the downloaded NCS tracks had no lyrics embedded. Even when a song had subtitles on YouTube, they came as SRT or VTT — not LRC. So I built this to fix that.

Now lrcx can fetch lyrics from multiple sources, making it easier to get synchronized lyrics for your music library.

## Features

- 📥 **Download lyrics from YouTube** - Fetch subtitles directly from YouTube URLs
- 🎵 **Download lyrics from LRCLib** - Search a community-sourced lyrics database
- 🔄 **Convert SRT/VTT to LRC** - Transform subtitle files into music lyric format
- 🤖 **Auto-generated captions** - Use YouTube's auto-generated subtitles when manual ones aren't available
- ⏱️ **Time offset** - Adjust timing for perfect sync
- ⚡ **Pipe-friendly** - Works seamlessly in shell pipelines

## Install

```sh
go install github.com/os-guy-original/lrcx/cmd/lrcx@latest
```

Or build from source:

```sh
make build
```

### Prerequisites

For YouTube subtitle downloading, you need [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed:

```sh
# macOS
brew install yt-dlp

# Linux
pip install yt-dlp
# or
wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O ~/.local/bin/yt-dlp
chmod +x ~/.local/bin/yt-dlp
```

## Usage

### Download from LRCLib

```sh
# Search by query (returns first match with synced lyrics)
lrcx lrclib "never gonna give you up"

# Search by artist and track (more precise)
lrcx lrclib --artist "Rick Astley" --track "Never Gonna Give You Up"

# Save to file
lrcx lrclib --artist "Artist" --track "Song" -o lyrics.lrc

# Interactive mode - choose from multiple results
lrcx lrclib --interactive "bohemian rhapsody"

# Get plain lyrics (no timestamps)
lrcx lrclib --plain --artist "Artist" --track "Song"

# With time offset (milliseconds)
lrcx lrclib --artist "Artist" --track "Song" --offset=-500
```

> **Note:** LRCLib searches are artist-aware. If you search with `--artist` and `--track`, results will be filtered to match the artist name. If no exact match is found, you'll be notified.

### Download from YouTube

```sh
# Basic download (outputs to stdout)
lrcx yt https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Save to file
lrcx yt https://www.youtube.com/watch?v=dQw4w9WgXcQ -o lyrics.lrc

# Interactive mode - choose from available subtitles
lrcx yt --interactive https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Use auto-generated captions (when manual subtitles aren't available)
lrcx yt --auto-captions https://www.youtube.com/watch?v=abc123

# With time offset (milliseconds)
lrcx yt https://www.youtube.com/watch?v=dQw4w9WgXcQ --offset=-500
```

### Convert Local Files

**Pipe mode:**
```sh
cat file.srt | lrcx > file.lrc
cat file.vtt | lrcx > file.lrc
```

**Explicit flags:**
```sh
lrcx -i file.srt -o file.lrc
lrcx -i file.vtt -o file.lrc
```

**With time offset:**
```sh
lrcx -i file.srt -o file.lrc --offset=-500
```

### Version

```sh
lrcx --version
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | stdin | Input SRT or VTT file |
| `-o` | stdout | Output LRC file |
| `--offset` | `0` | Time offset in milliseconds |
| `--interactive` | | Prompt for selection |
| `--verbose` | | Show output from yt-dlp |
| `--timeout` | `60s` | Timeout for network operations |
| `--version` | | Print version and exit |

### LRCLib-specific flags

| Flag | Description |
|------|-------------|
| `--artist` | Artist name for search |
| `--track` | Track name for search |
| `--plain` | Get plain lyrics only (no timestamps) |

### YouTube-specific flags

| Flag | Description |
|------|-------------|
| `--auto-captions` | Include auto-generated captions |

## Tag Handling

| Tag | Output |
|-----|--------|
| `<i>text</i>` | `_text_` |
| `<b>`, `<u>`, `<font ...>` | stripped (text preserved) |
| `{\an8}`, `{\pos(...)}` | stripped entirely |

Multi-line subtitle blocks are joined with a space into a single LRC line.

## Sources

| Source | Type | Quality | Notes |
|--------|------|---------|-------|
| LRCLib | Community-sourced LRC | High (synced) | Best quality, community maintained |
| YouTube | User-uploaded subtitles | Variable | Depends on video uploader |
| YouTube | Auto-generated captions | Good | Machine-generated, may have errors |

## Roadmap

- [x] v0.1 — SRT → LRC conversion
- [x] v0.1 — VTT → LRC conversion
- [x] v0.2 — YouTube subtitle download via yt-dlp
- [x] v0.2 — Auto-generated captions support
- [x] v0.2 — LRCLib integration
- [ ] v0.3 — Audio file embedding (MP3, FLAC, M4A)
- [ ] v0.4 — Enhanced LRC (word-level timestamps)
- [ ] v0.5 — Additional tag mappings (`<b>` → `*text*`, etc.)

## License

MIT
