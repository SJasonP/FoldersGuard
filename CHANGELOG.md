# Changelog

All notable changes to FoldersGuard are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com), and this project follows
[Semantic Versioning](https://semver.org).

**Data-format compatibility:** every 1.x release uses the same storage format (`fg-native-v1`) and reads and writes the
same `.fg` / `.fgs` databases and encrypted content. No 1.x release changes the on-disk format.

FoldersGuard is experimental software and makes no guarantee of security, cryptographic correctness, or data durability.
Do not rely on it as the only protection for valuable, sensitive, or irreplaceable data.

## [1.4.0] - 2026-06-26

### Added

- Resumable decryption. `fg decrypt --resume`, and a Resume option in the Decrypt Project and Decrypt Share dialogs,
  continue an interrupted decryption: the existing output is kept and only files that are missing or the wrong size are
  restored. Resume is off by default and is mutually exclusive with force overwrite.
- Resumable encryption as a core primitive: the encryptor can skip encrypted objects that already exist and, optionally,
  pass integrity verification, rewriting missing or corrupt objects. It is not exposed as a `create` flag, because each
  create generates a fresh project with new keys and visible names.

## [1.3.0] - 2026-06-25

### Added

- Change a project or share password without re-encrypting content. The database is re-keyed; internal per-file and
  per-folder content keys are unchanged, so no encrypted object is rewritten.
- Desktop: a "Change Password" action for projects, and "Change Share Password" for password-protected shares (which
  protects only future copies of the share).
- CLI: `fg passwd <project-id>` and `fg passwd --share <share.fgs>`.
- Password change is crash-safe: it verifies the current password, backs up the database, re-keys a copy, confirms it
  opens under the new password, then atomically replaces the live database.

## [1.2.0] - 2026-06-25

### Added

- Project-database backups. The `.fg` database is automatically snapshotted before destructive operations — applying
  changes, deleting a project, and changing a password — with a configurable retention limit.
- Desktop: a "Restore Database Backup" action, and a "Database backups to keep" setting.
- CLI: `fg backups list` and `fg backups restore`.

### Fixed

- Adding files or folders during a project modification now honors the source-cleanup setting: when set to delete, the
  cleartext source is removed after the add is committed, matching project creation.

### Changed

- Red (danger) button styling is reserved for irreversible, data-losing actions (delete project, discard unapplied
  changes, the delete confirmation). Decrypt, share-creation, and restore-backup confirmations are no longer styled red.

## [1.1.0] - 2026-06-24

### Added

- Reliable, byte-weighted progress for long-running operations (create, decrypt, verify, export, import, share, apply
  changes), showing the current phase, processed and total bytes, throughput, and estimated time remaining.
- Configurable noise-file handling for platform-generated metadata files (`.DS_Store`, AppleDouble `._*`, `Thumbs.db`,
  `ehthumbs.db`, `desktop.ini`, `.Spotlight-V100`, `.Trashes`, `.fseventsd`): ignore everywhere (default), ignore only
  during verification and matching, or do not ignore.
- Updated application icon.

### Changed

- Decryption and verification stream content instead of loading whole files into memory, so very large files no longer
  spike memory usage.
- Operations cannot be cancelled. While one runs, the project list is locked and closing the window or quitting the app
  is blocked, with a warning that forcing a quit is at the user's own risk.

## [1.0.0] - 2026-05-13

### Added

- Initial release. Encrypt a folder into a portable encrypted tree with UUID names, while the real names, metadata, and
  keys live in separate SQLCipher-encrypted `.fg` databases.
- AES-256-GCM encryption for file content, with random per-file and per-folder keys and password-based database unlock.
- Direct encrypted sharing through scoped `.fgs` share databases, password-protected or unprotected.
- Integrity verification without decryption, large-file balanced splitting, and metadata-only operations such as rename.
- A Wails desktop WebUI and a CLI (`foldersguard` / `fg`), with English and Simplified Chinese localization.

[1.4.0]: https://github.com/SJasonP/FoldersGuard/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/SJasonP/FoldersGuard/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/SJasonP/FoldersGuard/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/SJasonP/FoldersGuard/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/SJasonP/FoldersGuard/releases/tag/v1.0.0
