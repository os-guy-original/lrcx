# srt2lrc

Convert SRT subtitle files to LRC music lyric files.

## Install

```sh
go install github.com/srt2lrc/srt2lrc/cmd/srt2lrc@latest
```

Or build from source:

```sh
make build
```

## Usage

**Pipe mode:**
```sh
cat file.srt | srt2lrc > file.lrc
```

**Explicit flags:**
```sh
srt2lrc -i file.srt -o file.lrc
```

**With time offset (milliseconds):**
```sh
srt2lrc -i file.srt -o file.lrc --offset=-500
```

**Version:**
```sh
srt2lrc --version
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

Multi-line subtitle blocks are merged with ` / ` as separator.

## Roadmap

- [ ] v0.2 — yt-dlp integration: fetch subtitles directly from a YouTube URL
- [ ] v0.3 — Enhanced LRC (word-level timestamps)
- [ ] v0.4 — Additional tag mappings (`<b>` → `*text*`, etc.)

## License

MIT
