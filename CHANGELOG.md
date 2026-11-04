# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); this project uses [semantic versioning](https://semver.org/).

## [Unreleased]

### Changed

- Toolchains: Python 3.15, Java 25 LTS, Go 1.27, Rust 1.99 (edition 2024), Node 26 LTS; .NET stays on 10 (pinned via `global.json`). CI actions moved to their latest major versions.
- Test/tooling libraries: xUnit v3 (4.x, on Microsoft.Testing.Platform), JUnit 6, Maven Surefire 3.6, TypeScript 7, Vitest 5.

## [0.1.0] - 2026-08-09

First public release. schema diffing that classifies changes safe/breaking and fails the PR on a break, in Python, C#, and Java.

[0.1.0]: https://github.com/parag-labs/schema-guard/releases/tag/v0.1.0
