# FoldersGuard

English | [简体中文](README.zh-CN.md)

> **Important notice:** All source code in this project was written by AI. This project makes no guarantee of security,
> cryptographic correctness, data durability, fitness for production use, or protection against data loss. Do not rely
> on
> it as the only protection for valuable, sensitive, or irreplaceable data.

FoldersGuard is an experimental desktop and CLI tool for protecting folders while keeping encrypted data practical to
move, upload, download, and share manually.

## Why FoldersGuard

FoldersGuard is built around one product idea: encrypted data should still behave like ordinary files and folders when
you need to store or move it.

The encrypted content remains a visible folder tree with UUID names. You can copy it to another disk, upload it to a
cloud drive, download only part of it, or send selected encrypted files and directories to someone else. The real names,
metadata, and keys needed to restore the data live separately in FoldersGuard's encrypted databases.

This means FoldersGuard can share encrypted files and directories without decrypting them first. A share database
describes exactly what the recipient is allowed to restore, so you can send one encrypted file, one encrypted folder, or
a mixed selection without exposing parent folders, siblings, or unrelated project data.

## Key Features

- Manual encrypted-data handling: encrypted output is a normal folder tree that can be copied, uploaded, downloaded,
  backed up, or shared with ordinary tools.
- Direct encrypted sharing: share encrypted files or directories without first creating a cleartext export.
- Share-scoped access: `.fgs` share databases contain only the metadata and keys needed for the selected files and
  folders.
- Integrity verification: verify encrypted content without decrypting it, and detect missing or tampered encrypted
  objects before restore or sharing.
- Hidden real names: visible encrypted file and directory names are UUID values.
- Separate metadata: FoldersGuard data is separate from encrypted content, so metadata-only changes such as renaming do
  not require the encrypted content to be present.
- Preserved folder hierarchy: the encrypted tree keeps the original logical structure, which makes manual storage
  workflows understandable.
- Large-file splitting: large files can be split into balanced parts while remaining one logical file inside
  FoldersGuard.
- Desktop and CLI workflows: the Wails WebUI is for interactive use; the CLI is for automation.

## Status

FoldersGuard is a work in progress. It currently includes a Go core, a CLI, a Wails desktop WebUI, SQLCipher-backed
`.fg` and `.fgs` databases, English and Simplified Chinese localization, and release scripts for signed and notarized
macOS builds.

The implementation should be treated as experimental software until independently reviewed and tested.

## Security Model

FoldersGuard v1 uses:

- SQLCipher for project and share databases.
- AES-256-GCM for encrypted file content.
- Random per-file and per-folder internal keys.
- Passwords as the user-facing unlock mechanism.
- UUID visible names for encrypted files and directories.

FoldersGuard aims to protect file contents, real names, directory metadata, and internal key material from unauthorized
readers.

FoldersGuard does not try to hide:

- The fact that FoldersGuard is being used.
- The visible encrypted directory hierarchy.
- The number of encrypted entries in a directory.
- Approximate encrypted file or part sizes.
- Storage-provider-visible modification patterns.

Again, no security guarantee is made. The code was AI-written and has not been presented here as audited cryptographic
software.

## Project Model

A normal FoldersGuard project is created from one top-level folder.

The encrypted content tree remains a normal folder tree that can be moved, uploaded, downloaded, or shared through
ordinary storage tools. Real names are replaced with UUID names. The metadata required to map those UUID names back to
real names is stored separately in FoldersGuard data.

In v1:

- `.fg` is used for normal project databases with exactly one top-level directory.
- `.fgs` is used for share databases and other share-scoped database shapes.
- Encrypted content and FoldersGuard metadata are separate.
- Active project data is stored in the user's FoldersGuard data directory.
- Share databases can describe one file, one folder, or a selected set of files and folders.

## Interfaces

### Desktop WebUI

The WebUI is the main interactive interface. It is built with Wails, React, and Ant Design.

It supports project creation, import, export, inspection, decryption, verification, deletion, sharing, share loading,
project browsing, and project modification workflows.

The WebUI also provides localized operation summaries and manual encrypted-content instructions for workflows where the
encrypted content is handled outside FoldersGuard.

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

## Releases

Release builds are published through GitHub Releases.

For v1.3.0, treat release artifacts as experimental. Test with copies of your data first, keep independent backups, and
verify encrypted content before relying on a restore or share workflow.

macOS release packages built with `make macos-release` are designed to be signed, notarized, and stapled before upload.
Windows builds must be produced with CGO and a working Windows-target C compiler so SQLCipher support is actually
included.

## Development And Builds

Requirements:

- Go matching the version declared in `go.mod`.
- Node.js and npm for the frontend.
- Wails v2 for desktop builds.
- CGO and a working target-platform C compiler for release builds with SQLCipher support.

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

FoldersGuard uses SQLCipher for encrypted project and share databases. SQLCipher is a CGO dependency, so real release
builds must use CGO and a working target-platform C compiler.

A build that produces an executable without working SQLCipher support should not be considered a complete or usable
FoldersGuard build.

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

Third-party components are licensed under their own license terms. See `THIRD-PARTY-NOTICES.md` before publishing or
redistributing release artifacts.
