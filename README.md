> 🕯️ Turkey lost children to school shootings in Siverek and Kahramanmaraş. We are aware, and we grieve. Our deepest condolences to all the families.

# lrcx

## Why?

In my previous project [ncsdl](https://github.com/os-guy-original/ncsdl), most of the downloaded NCS tracks had no lyrics embedded. Even when a song had subtitles on YouTube, they came as SRT or VTT — not LRC. So I built this to fix that :/

Convert SRT and VTT subtitle files to LRC music lyric files.

## Install

```sh
go install github.com/os-guy-original/lrcx/cmd/lrcx@latest
```

Or build from source:

```sh
make build
```

## Usage

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

**With time offset (milliseconds):**
```sh
lrcx -i file.srt -o file.lrc --offset=-500
```

**Fetch from YouTube (beta):**
```sh
lrcx --beta-feature=yt --interactive https://www.youtube.com/watch?v=...
```

**Version:**
```sh
lrcx --version
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | stdin | Input SRT or VTT file |
| `-o` | stdout | Output LRC file |
| `--offset` | `0` | Time offset in milliseconds |
| `--beta-feature` | | Enable a beta feature (e.g. `yt`) |
| `--interactive` | | Prompt for subtitle selection (yt) |
| `--verbose` | | Show output from third-party tools |
| `--version` | | Print version and exit |

## Tag Handling

| Tag | Output |
|-----|--------|
| `<i>text</i>` | `_text_` |
| `<b>`, `<u>`, `<font ...>` | stripped (text preserved) |
| `{\an8}`, `{\pos(...)}` | stripped entirely |

Multi-line subtitle blocks are joined with a space into a single LRC line.

## Roadmap

- [x] v0.1 — SRT → LRC conversion
- [x] v0.1 — VTT → LRC conversion
- [x] v0.2 — yt-dlp integration: fetch subtitles directly from a YouTube URL
- [ ] v0.3 — Enhanced LRC (word-level timestamps)
- [ ] v0.4 — Additional tag mappings (`<b>` → `*text*`, etc.)

## License

MIT
