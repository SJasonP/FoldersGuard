# FoldersGuard Product Requirements

## Overview

FoldersGuard, abbreviated as FG, is encryption software for protecting folders and the files inside them. The core
product idea is that a folder is the natural unit for storage, download, upload, sharing, and authorization.

FG encrypts each file independently, while FG's own SQLCipher-encrypted database stores the metadata and internal keys
needed to open folders and files. Users unlock project and share databases with passwords. A project password unlocks
the whole active project database. A share database unlocks only the content described by that share database. Users
cannot decrypt sibling folders, parent folders, or unrelated files unless access is explicitly shared.

FG creates one active project from one top-level folder at a time. The top-level folder is the root operating boundary
for normal project encryption, project metadata operations, project verification, project decryption, and share
generation. Share databases may describe a rootless shared payload instead of a normal project root.

Encrypted content and FG data are separate. The encrypted content tree contains the same logical folder and file
structure as the original tree, except that visible names are UUID values. When a large file is split, all of its parts
are treated as one logical file. FG data may live elsewhere and does not need to be stored beside the encrypted content.
In v1, FG data is stored only in FG's data directory, but users may export it.

## Product Goals

- Protect the confidentiality of file contents.
- Use the v1 encryption suite for project databases, share databases, and encrypted content.
- Allow any encrypted file to be shared together with an FG share database that can restore only that file.
- Allow any encrypted folder to be shared together with an FG share database that can restore only that folder and its
  descendants.
- Preserve directory hierarchy in v1 so users can manually upload, download, and share a folder from normal file storage
  or cloud drives.
- Preserve entry count and structure for supported regular files and directories: with split parts treated as one
  logical file, the encrypted content tree matches the original tree's supported folder and file layout.
- Hide real file and directory names by replacing visible names with generated UUID values.
- Preserve the portable filesystem metadata users expect when restoring normal files and directories.
- Support large files by splitting them into balanced parts when needed.
- Require both sender and recipient to use FG.
- Provide a local desktop WebUI for normal interactive use.
- Provide WebUI localization for American English and Simplified Chinese, with an extensible structure for adding more
  languages.
- Provide a complete WebUI dark theme and automatically follow system light or dark appearance by default.
- Provide a CLI for stable automation and scripting.
- Allow renaming files and directories by updating only FG data, without requiring access to encrypted content.
- For adding, moving, or deleting encrypted content in manually managed storage, generate clear operation instructions
  for the user.
- Use `.fg` for normal FG project databases that contain exactly one top-level object and that object is a directory.
- Use `.fgs` for share databases and all other database shapes.

## User Stories

### Encrypt A Folder

As a user, I want to encrypt a folder so that all files under it are protected, names are hidden, and the encrypted
result can still be handled as a normal folder tree.

### Manage One FG Project

As a user, I want one top-level folder to behave like one FG project, with active project metadata tracked in FG's
`FoldersGuard` data directory.

### Use A Local WebUI

As a user, I want to create, modify, decrypt, share, import, export, and delete FG projects through a local desktop
WebUI without memorizing CLI commands.

### Automate With CLI

As a user, I want scriptable CLI commands for FG workflows so that automated jobs can use the same core behavior as the
WebUI.

### Unlock A Top-Level Folder

As a user, I want the top-level FG folder to require a password so that opening the root grants access to all content
inside it.

### Share A Folder

As a user, I want to send an encrypted folder and an FG-generated share database to another person so they can decrypt
that folder and its descendants, without gaining access to anything outside that folder.

### Share A Single File

As a user, I want to send one encrypted file and an FG-generated share database to another person so they can decrypt
only that file.

### Share Selected Items

As a user, I want to share any selected set of encrypted files and folders with one FG share database so the recipient
can decrypt only that selected set.

### Hide Names

As a user, I want encrypted files and directories to use random UUID names so that someone browsing the encrypted
storage cannot infer sensitive names.

### Rename Without Encrypted Content

As a user, I want to rename a protected file or folder in FG without needing the encrypted content to be present,
because renaming changes only the encrypted name mapping stored in FG data.

### Restore Filesystem Metadata

As a user, I want restored files and folders to keep their original modification times, access times, creation times
when supported, permissions, and basic Windows file attributes when supported.

### Generate Manual Storage Instructions

As a user, I want FG to tell me exactly how to upload, move, or delete encrypted objects when the encrypted content is
stored somewhere FG cannot operate directly.

### Split Large Files

As a user, I want files larger than a configured maximum size to be split into multiple balanced parts so they can be
stored on systems with file size limits.

## V1 Feature Set

### FG Native Mode

FG native mode is the only supported encryption mode.

- Every original file has its own file key.
- Every folder has its own folder key.
- The top-level project database must have a password.
- FG data stores encrypted metadata and child keys for folders and files.
- FG data is stored as a SQLCipher-encrypted SQLite database.
- Real file and directory names are stored only inside encrypted FG databases or share databases.
- Visible file and directory names are UUID values.
- Directory hierarchy is preserved.
- Portable filesystem metadata is preserved for supported regular files and directories.
- With split files treated as one logical file, encrypted output preserves the original folder and file structure.
- Large files use FG balanced splitting.
- File parts are storage fragments of one logical file, not independently encrypted files.
- Share databases are FG proprietary SQLCipher-encrypted databases containing a share-scoped data subset.
- Share generation always creates `.fgs` databases.
- Share databases may be password-protected or unprotected.
- Unprotected share databases can be opened without a password and are bearer secrets: anyone with the encrypted content
  and share database can decrypt the shared content.
- FG data is separate from encrypted content.
- V1 stores FG data only in FG's data directory and supports exporting it.

### Filesystem Entry Policy

FG supports only regular files and directories.

- Symlinks, sockets, FIFOs, device files, and other special entries are ignored as if they do not exist.
- Unsupported entries are not represented in FG metadata and are not reported in normal command output.
- Hard link relationships are not preserved; each hard link path is processed as a normal regular file.
- FG provides a noise file handling setting for platform-generated metadata files such as `.DS_Store`, AppleDouble
  `._*` files, `Thumbs.db`, `ehthumbs.db`, `desktop.ini`, `.Spotlight-V100`, `.Trashes`, and `.fseventsd`.
- Noise file handling defaults to ignore everywhere, meaning recognized noise files are treated as absent during source
  scanning, project creation, project add, encrypted content matching, verification, decryption, share restore, source
  cleanup, and output-folder emptiness checks.
- Under ignore everywhere, recognized noise files are not user content. They may be removed as incidental cleanup when FG
  removes or replaces a containing directory, but they are not represented in FG metadata or normal operation output.
- Users may instead choose to ignore recognized noise files only during verification and matching, or to not ignore them
  at all.

### Filesystem Metadata Policy

FG preserves normal restorable metadata for supported files and directories.

Required metadata:

- File and directory type.
- Real name.
- Parent-child structure.
- File size.
- Modification time.
- Access time.
- Creation time when the host platform and filesystem support it.
- Permission mode.
- Basic Windows file attributes when the host platform supports them.

Restore rules:

- File content authenticity has priority over metadata restoration.
- Directory metadata is restored after child entries are restored.
- Platform or filesystem limitations may reduce timestamp precision.

## Planned Feature Set (v1.2–v1.6)

The planned feature set extends v1.1 with reliability and security hardening for large, valuable data.

Versioning approach:

- Each feature ships as its own minor release, starting at v1.2, in the dependency order of the subsections below.
- None of these features changes the storage format, so the storage format version is unchanged and every release stays
  data-compatible with v1.
- The target version numbers are the current plan and may shift.
- Items marked planned are specified here to be built against and are not yet implemented.

### Metadata-Database Backup

**Status: Planned for v1.2 — not yet implemented.**

- FG automatically snapshots a project database before destructive operations: applying changes, deleting a project, and
  changing a password.
- Backups are stored in a managed location under FG's data directory with bounded rotation.
- Backups are encrypted SQLCipher databases and are stored with the same file restrictions as the live database.
- The backup retention limit is configurable in Settings.
- FG can restore a project database from a backup.

### Password Change

**Status: Planned for v1.3 — not yet implemented.**

- A project (`.fg`) or share (`.fgs`) database password can be changed without re-encrypting content.
- Internal per-file and per-folder content keys are unchanged; no encrypted object is rewritten.
- Changing a password verifies the old password and confirms the new password before completing.
- The change is crash-safe: FG backs up the database, re-keys a copy, confirms it opens under the new password, then
  replaces the live database.
- Changing a share password protects only future copies; share databases already distributed are unaffected.
- Password change is available in both the CLI and the WebUI.

### Resumable Operations

**Status: Planned for v1.4 — not yet implemented.**

- Long-running content operations can be re-run after an interruption and continue instead of restarting from the
  beginning.
- Encryption skips an encrypted object whose visible path already exists and passes integrity verification; a present
  but corrupt object is rewritten.
- Decryption and restore skip an output file that already exists and matches the expected content.
- A source file is deleted only after its encrypted object is confirmed complete, even across a resumed run.
- Progress counts already-completed work as processed at the start of a resumed run so totals stay accurate.
- The user chooses between resuming and starting fresh.
- Resuming verifies an existing object before skipping it; a faster skip-by-presence option may be offered.

### Partial-Failure Tolerance

**Status: Planned for v1.5 — not yet implemented.**

- Content operations support an optional continue-on-error mode. The default remains abort on the first error.
- In continue-on-error mode, file-level failures are recorded and the operation processes the remaining items.
- The result reports the count and the list of failed items with reasons; sensitive values stay hidden.
- A source file is never deleted when its own encryption failed.
- Fatal conditions, such as the output disk being full or a database error, still abort the operation regardless of
  mode.

### Parallel Encryption

**Status: Planned for v1.6 — not yet implemented.**

- File encryption may process multiple files concurrently with a bounded worker pool.
- Concurrency defaults to a value derived from the host CPU count and is configurable.
- Within-file chunk streaming is unchanged; concurrency is across files, not within a file.
- Byte-weighted progress remains accurate and monotonic under concurrency.
- Source-file deletion and folder-creation ordering remain correct under concurrency.
- A failure in one worker stops the operation cleanly, unless continue-on-error mode is enabled.

## Security Expectations

FG must protect file contents, file names, directory names, and internal key material from unauthorized readers.

FG v1 does not attempt to hide:

- The visible encrypted directory hierarchy.
- The number of visible entries in a directory.
- Approximate encrypted file or part sizes.
- File modification patterns observable through the storage provider.
- The fact that a folder is protected by FG.
- FG reserved data and share databases.
