# FoldersGuard

English | [简体中文](README.zh-CN.md)

> **Important notice:** All source code in this project was written by AI. This project makes no guarantee of security, cryptographic correctness, data durability, fitness for production use, or protection against data loss. Do not rely on it as the only protection for valuable, sensitive, or irreplaceable data.

FoldersGuard is an experimental desktop and CLI tool for protecting folders by separating encrypted content from encrypted metadata.

The core idea is simple: a folder is the natural unit for storage, sharing, authorization, backup, upload, and restore. FoldersGuard encrypts file contents, hides real file and directory names behind UUID names, and stores the metadata needed to restore those files in SQLCipher-encrypted databases.

## Status

FoldersGuard is a work in progress.

It currently includes:

- A Go core for scanning, planning, encrypting, restoring, verifying, sharing, and modifying protected folder projects.
- A CLI named `foldersguard`, with `fg` intended as its short alias.
- A Wails desktop WebUI backed by the same Go core.
- A React and Ant Design frontend with English and Simplified Chinese localization.
- SQLCipher-backed `.fg` project databases and `.fgs` share databases.
- Support for split large files, preserved directory hierarchy, UUID visible names, and portable filesystem metadata.

The implementation should be treated as experimental software until independently reviewed and tested.

## Security Model

FoldersGuard v1 uses:

- SQLCipher for project and share databases.
- AES-256-GCM for encrypted file content.
- Random per-file and per-folder internal keys.
- Passwords as the user-facing unlock mechanism.
- UUID visible names for encrypted files and directories.

FoldersGuard aims to protect file contents, real names, directory metadata, and internal key material from unauthorized readers.

FoldersGuard does not try to hide:

- The fact that FoldersGuard is being used.
- The visible encrypted directory hierarchy.
- The number of encrypted entries in a directory.
- Approximate encrypted file or part sizes.
- Storage-provider-visible modification patterns.

Again, no security guarantee is made. The code was AI-written and has not been presented here as audited cryptographic software.

## Project Model

A normal FoldersGuard project is created from one top-level folder.

The encrypted content tree remains a normal folder tree that can be moved, uploaded, downloaded, or shared through ordinary storage tools. Real names are replaced with UUID names. The metadata required to map those UUID names back to real names is stored separately in FoldersGuard data.

In v1:

- `.fg` is used for normal project databases with exactly one top-level directory.
- `.fgs` is used for share databases and other share-scoped database shapes.
- Encrypted content and FoldersGuard metadata are separate.
- Active project data is stored in the user's FoldersGuard data directory.
- Share databases can describe one file, one folder, or a selected set of files and folders.

## Interfaces

### Desktop WebUI

The WebUI is the main interactive interface. It is built with Wails, React, and Ant Design.

It supports project creation, import, export, inspection, decryption, verification, deletion, sharing, share loading, project browsing, and project modification workflows.

### CLI

The CLI is intended for automation and repeatable workflows.

The primary executable name is:

```text
foldersguard
```

The short alias is:

```text
fg
```

Main command groups include:

```text
fg encrypt
fg decrypt
fg inspect
fg verify
fg export
fg import
fg share
fg rename
fg add
fg move
fg remove
fg plan encrypt
```

See `docs/cli.md` and the files under `docs/cli/` for the CLI specification.

## Repository Layout

```text
.
├── cmd/foldersguard/        CLI entrypoint
├── internal/app/            Application service layer used by the WebUI
├── internal/cli/            Cobra CLI commands
├── internal/content/        Content encryption and restore logic
├── internal/crypto/         Cryptographic primitives used by the project
├── internal/db/             SQLite and SQLCipher database opening logic
├── internal/format/         Format constants and extension rules
├── internal/fsmeta/         Filesystem metadata capture and restore helpers
├── internal/fswalk/         Filesystem scanning
├── internal/model/          Core data structures and split planning
├── internal/project/        Project planning, execution, restore, and verify logic
├── internal/storage/        Database schema and metadata persistence
├── frontend/                React WebUI
├── docs/                    Product, architecture, CLI, WebUI, and storage docs
└── scripts/                 Build helper scripts
```

## Development

Requirements:

- Go matching the version declared in `go.mod`.
- Node.js and npm for the frontend.
- Wails v2 for desktop builds.
- A working C compiler when building with SQLCipher support.

Run Go tests:

```text
go test ./...
```

Build the frontend:

```text
npm --prefix frontend run build
```

Build the CLI:

```text
make build
```

Build the Wails desktop app:

```text
wails build
```

## SQLCipher And CGO

FoldersGuard uses SQLCipher for encrypted project and share databases. SQLCipher is a CGO dependency, so real release builds must use CGO and a working target-platform C compiler.

A build that produces an executable without working SQLCipher support should not be considered a complete or usable FoldersGuard build.

For Windows AMD64 build notes, see:

```text
docs/build.md
scripts/build-windows-amd64.ps1
```

## Documentation

Important documents:

- `docs/product-requirements.md`
- `docs/architecture.md`
- `docs/storage-format.md`
- `docs/security-implementation.md`
- `docs/cli.md`
- `docs/webui.md`
- `docs/build.md`

## License

FoldersGuard's own source code is licensed under the MIT License. See `LICENSE`.

Third-party components are licensed under their own license terms. See `THIRD-PARTY-NOTICES.md` before publishing or redistributing release artifacts.
