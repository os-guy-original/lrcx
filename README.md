> 🕯️ Turkey lost children to school shootings in Siverek and Kahramanmaraş. We are aware, and we grieve. Our deepest condolences to all the families.

# lrcx

Convert SRT subtitle files to LRC music lyric files.

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
```

**Explicit flags:**
```sh
lrcx -i file.srt -o file.lrc
```

**With time offset (milliseconds):**
```sh
lrcx -i file.srt -o file.lrc --offset=-500
```

**Version:**
```sh
lrcx --version
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | stdin | Input SRT file |
| `-o` | stdout | Output LRC file |
| `--offset` | `0` | Time offset in milliseconds |
| `--version` | | Print version and exit |

## SRT Tag Handling

| Tag | Output |
|-----|--------|
| `<i>text</i>` | `_text_` |
| `<b>`, `<u>`, `<font ...>` | stripped (text preserved) |
| `{\an8}`, `{\pos(...)}` | stripped entirely |

Multi-line subtitle blocks are joined with a space into a single LRC line.

## Roadmap

- [ ] v0.2 — yt-dlp integration: fetch subtitles directly from a YouTube URL
- [ ] v0.3 — Enhanced LRC (word-level timestamps)
- [ ] v0.4 — Additional tag mappings (`<b>` → `*text*`, etc.)

## License

MIT
