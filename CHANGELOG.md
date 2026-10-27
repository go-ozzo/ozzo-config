# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

No version has been tagged yet; the first tagged release will be v1.0.0.

## [Unreleased]

### Added

- `go.mod`. The module path is unchanged: `github.com/go-ozzo/ozzo-config`.
- CI on GitHub Actions: tests with the race detector on Linux (oldstable, stable), Windows and macOS (stable) and on
  the minimum Go from `go.mod`, a linux/386 build, golangci-lint v2.15.0, gofmt and `go mod tidy` checks,
  govulncheck, and coverage on Codecov.
- Repository files: golangci-lint configuration, `SECURITY.md`, `CONTRIBUTING.md`, `CODEOWNERS`, Dependabot for
  GitHub Actions, `.gitattributes` keeping LF line endings.
- A test that pins the YAML semantics described in the README.

### Changed

- **Go 1.21 or newer is required** (the README used to say Go 1.2).
- `go.mod` requires the lowest dependency versions importers already use, so adding it raises no importer's
  versions: github.com/BurntSushi/toml v0.3.1, gopkg.in/yaml.v2 v2.2.8 (the first release without a known
  vulnerability) and github.com/hnakamur/jsonpreprocess at its last commit. No change in library behaviour.
- Doc comments of `GetInt`, `GetInt64`, `GetFloat` and `GetBool` name the right method.

### Removed

- Travis CI configuration (`.travis.yml`); CI runs on GitHub Actions.

[Unreleased]: https://github.com/go-ozzo/ozzo-config/commits/master
