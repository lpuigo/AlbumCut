# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

AlbumCut is a freshly scaffolded Go project (GoLand default template) with no real functionality yet — `main.go` only contains placeholder "Hello, gopher" sample code. There is no README, no package structure, and no tests. This file will need to be expanded once actual architecture exists.

## Commands

- Build: `go build ./...`
- Run: `go run main.go`
- Test: `go test ./...` (no tests exist yet)
- Format: `gofmt -l .`
- Vet: `go vet ./...`

## Module

- Module path: `AlbumCut` (go.mod), Go version 1.24.
