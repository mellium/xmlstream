# Changelog

All notable changes to this project will be documented in this file.

## v0.15.4 — 2023-01-11

### Fixed

- `InnerElement` now works correctly when `io.EOF` is returned at the same time
  as the `EndElement` token.


## v0.15.3 — 2021-09-26

### Fixed

- `Iter` now returns all child elements
- `MultiReader` no longer panics on a nil reader


## v0.15.2 — 2021-02-14

### Added

- The `InnerElement` transformer
- The `InsertFunc` transformer
- The `Insert` transformer


## v0.15.1 — 2020-11-24

### Added

- The `TokenWriteFlusher` interface


## v0.15.0 — 2020-03-17

### Added

- A mechanism for iterating over child elements
- The `DecodeCloser` interface
- The `DecodeEncoder` interface
- The `Decoder` interface
- The `EncodeCloser` interface


### Changed

- Bump the language version to Go 1.13


### Fixed

- Previously errors or infinite loops could occur in Copy, MultiReader, and
  ReadAll when an underlying xml.Decoder's Token method returned "nil, nil"


## v0.14.0 — 2019-08-08

### Changed

- Only return the token from `Token` once, then return `io.EOF`


## v0.13.6 — 2019-08-06

### Added

- The `ReadAll` function


## v0.13.5 — 2019-07-23

### Added

- The `Encoder` interface
- The `TokenReadEncoder` interface


## v0.13.4 — 2019-07-21

### Added

- The `NopCloser` function
- The `TokenWriteFlushCloser` interface
