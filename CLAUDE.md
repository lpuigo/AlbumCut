# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

AlbumCut is a Go CLI tool that splits a single MP3 file (a full album
recording) into individual per-track MP3 files, based on a text file of
timestamps. Splitting uses `ffmpeg -c copy` (stream copy, no re-encode) to
preserve original audio quality. Full functional spec: `Docs/split-album-spec.md`.

## Design rules

@~/.claude/design-rules.md

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

Layered, per [design-rules.md](~/.claude/design-rules.md): CLI → pipeline
(orchestration) → domain (track/segment/naming) → infra (ffmpeg/textfile),
dependencies flow one way only, downward. `main.go` at the module root is
a thin entrypoint: it just calls `cli.Run(os.Args[1:], os.Stdout)` and
handles the exit code (`errors.Is(err, flag.ErrHelp)` short-circuits to
exit 0; any other non-nil error prints `erreur: ...` and exits 1).

- **`internal/cli`** (route/handler layer) — `Run(args, out)` is the
  top-level pipeline driver: arg parsing (`parseArgs`), file-existence
  checks (`checkFileExists`), prerequisite check, then it drives
  `pipeline.BuildPlan`/`pipeline.SplitTrack` and writes all progress/report
  output to the injected `io.Writer` (`out`) — no business logic lives
  here, only orchestration + presentation. `-h`/`--help` prints `helpText`
  and returns `flag.ErrHelp` unwrapped so `main()` can special-case it.
- **`internal/pipeline`** (service/orchestration layer) — `BuildPlan` reads
  the timestamps file, probes real MP3 duration, and builds segments;
  `SplitTrack` splits one segment and flags a warning if the resulting
  track is shorter than `ShortTrackWarningThreshold`. Returns plain data,
  prints nothing. Depends on the domain and infra packages below; nothing
  depends on it except `internal/cli`.
- **`internal/track`** (domain) — `ParseTracks([]string) ([]Track, error)`
  parses/validates timestamp lines (pure, no I/O) into `[]Track`.
  Validation happens fully before any splitting: malformed lines,
  non-strictly-increasing timestamps, and no-valid-line input all return a
  `*ParseError` carrying line number + line content + reason.
- **`internal/segment`** (domain) — `BuildSegments` turns `[]track.Track` +
  total duration + an `endPadding` (seconds, from `-end-padding`) into
  `[]Segment` (start/end per track). Every non-last track's end is pushed
  forward by `endPadding` (capped at total duration) to avoid cutting its
  tail too short; the last track is never padded. Rejects a last track
  whose timestamp is at or past the real file duration (would produce an
  empty/zero-length file — this is a hard error, not a warning).
- **`internal/naming`** (domain) — `OutputPath` builds
  `NN. AlbumTitle_TrackTitle.mp3` in the source file's directory; sanitizes
  filesystem-forbidden characters (`/ \ : * ? " < > |`) but deliberately
  preserves spaces in track titles.
- **`internal/duration`** (domain helper) — `Format(seconds int) string`
  renders `HH:MM:SS`; shared by `track` (error messages), `segment` (error
  messages) and `cli` (progress/report output).
- **`internal/infra/ffmpeg`** (infra) — `CheckAvailable` (PATH check for
  `ffmpeg`/`ffprobe`), `ProbeDuration` (shells out to `ffprobe` for the
  real total duration of the source MP3), and `SplitSegment(mp3Path,
  start, end, outputPath)` (invokes `ffmpeg -ss -to -c copy`,
  output-seeking after `-i` for better cut accuracy than input-seeking).
  Takes plain `float64`/`string` parameters rather than a domain
  `segment.Segment`, so this layer never imports the domain layer
  (no upward dependency).
- **`internal/infra/textfile`** (infra) — `ReadLines(path)` reads a text
  file into `[]string`, one line per entry, no transformation; used by
  `pipeline` before handing the lines to `track.ParseTracks`.
- **`internal/testutil`** — test-only fixtures shared by integration tests
  that need a real ffmpeg (`RequireFFmpeg`, `GenerateTestMP3`); not part of
  the runtime binary's dependency graph.
- **`integration_test.go`** (module root) — end-to-end tests against real
  ffmpeg-generated fixture MP3s (`sine=` lavfi source), driving the full
  pipeline through `cli.Run`.

## Known deliberate scope decisions

These were explicitly descoped/decided during initial implementation —
don't reintroduce without checking with the user first:

- **No re-encode fallback.** If `ffmpeg -c copy` produces an imprecise cut,
  there is currently no automatic detection or re-encoding fallback. Only
  stream-copy splitting is implemented. May be revisited later if it proves
  necessary in practice.
- **No ID3 tag handling** on output files (out of scope per spec).
- **Arguments are positional**, with one optional flag: `albumcut
  [-end-padding=N] <album.mp3> <tracks.txt>`. Standard `flag` package
  semantics apply — flags must precede positional args.
- **`-h`/`--help`** prints `helpText` (in `main.go`) and exits 0. Detected
  via `errors.Is(err, flag.ErrHelp)`, which `parseArgs` returns unwrapped
  (unlike other parse errors) so `main()` can special-case it and skip the
  `erreur:` prefix/exit-1 path.
- **ffmpeg is a required, unmocked dependency** — tests shell out to a real
  `ffmpeg`/`ffprobe` rather than mocking `exec.Command`.
