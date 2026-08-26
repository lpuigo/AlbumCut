# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

AlbumCut is a Go CLI tool that splits a single MP3 file (a full album
recording) into individual per-track MP3 files, based on a text file of
timestamps. Splitting uses `ffmpeg -c copy` (stream copy, no re-encode) to
preserve original audio quality. Full functional spec: `Docs/split-album-spec.md`.

## Commands

- Build: `go build -o albumcut.exe .`
- Run: `go run . <album.mp3> <tracks.txt>`
- Test (all): `go test ./...`
- Test (single): `go test -run TestName ./...`
- Vet: `go vet ./...`

`ffmpeg`/`ffprobe` must be on `PATH`. Integration tests (`integration_test.go`)
call real `ffmpeg`/`ffprobe` (no mocking) and self-skip via `t.Skip` if the
tools aren't found on `PATH` — they are not gated behind a build tag, just
run with the rest of the suite.

## Architecture

Single `main` package, one file per concern, wired together by `run()` in
`main.go` (the actual `main()` just calls `run(os.Args[1:])` and handles
exit code — `run` returns a plain `error` and is what tests call directly):

1. `main.go` — CLI arg parsing (`parseArgs`), file-existence checks, ffmpeg/
   ffprobe presence check (`checkExternalTools`, via `exec.LookPath`), and
   the top-level pipeline in `run()`.
2. `timestamps.go` — parses/validates the tracks text file into `[]Track`.
   Validation happens fully before any splitting: malformed lines,
   non-strictly-increasing timestamps, and empty/all-blank files all return
   a `*ParseError` carrying line number + line content + reason.
3. `ffprobe.go` — `ProbeDuration` shells out to `ffprobe` to get the real
   total duration of the source MP3.
4. `segments.go` — `BuildSegments` turns `[]Track` + total duration into
   `[]Segment` (start/end per track). Rejects a last track whose timestamp
   is at or past the real file duration (would produce an empty/zero-length
   file — this is a hard error, not a warning).
5. `naming.go` — `OutputPath` builds `NN_AlbumTitle_TrackTitle.mp3` in the
   source file's directory; sanitizes filesystem-forbidden characters
   (`/ \ : * ? " < > |`) but deliberately preserves spaces in track titles.
6. `split.go` — `SplitSegment` invokes `ffmpeg -ss -to -c copy` per track
   (output-seeking, after `-i`, for better cut accuracy than input-seeking).
7. `integration_test.go` — end-to-end tests against real ffmpeg-generated
   fixture MP3s (`sine=` lavfi source), including the full `run()` pipeline.

## Known deliberate scope decisions

These were explicitly descoped/decided during initial implementation —
don't reintroduce without checking with the user first:

- **No re-encode fallback.** If `ffmpeg -c copy` produces an imprecise cut,
  there is currently no automatic detection or re-encoding fallback. Only
  stream-copy splitting is implemented. May be revisited later if it proves
  necessary in practice.
- **No ID3 tag handling** on output files (out of scope per spec).
- **No CLI flags** — arguments are strictly positional:
  `albumcut <album.mp3> <tracks.txt>`.
- **ffmpeg is a required, unmocked dependency** — tests shell out to a real
  `ffmpeg`/`ffprobe` rather than mocking `exec.Command`.
