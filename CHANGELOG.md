# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Breaking

- Dropped support for versions of Go before 1.13


### Added

- A mechanism for iterating over child elements
- The `DecodeCloser` interface
- The `DecodeEncoder` interface
- The `Decoder` interface
- The `EncodeCloser` interface


## Fixed

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
