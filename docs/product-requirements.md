# FoldersGuard Product Requirements

## Overview

FoldersGuard, abbreviated as FG, is encryption software for protecting folders and the files inside them. The core product idea is that a folder is the natural unit for storage, download, upload, sharing, and authorization.

FG encrypts each file independently, while FG's own SQLCipher-encrypted database stores the metadata and internal keys needed to open folders and files. Users unlock FG data with passwords. If a user has the password for a folder, they can decrypt everything inside that folder recursively. If a user has a share database, they can decrypt only the content described by that share database. They cannot decrypt sibling folders, parent folders, or unrelated files unless access is explicitly shared.

FG creates one active project from one top-level folder at a time. The top-level folder is the root operating boundary for normal project encryption, project metadata operations, project verification, project decryption, and share generation. Share databases may describe a rootless shared payload instead of a normal project root.

Encrypted content and FG data are separate. The encrypted content tree contains the same logical folder and file structure as the original tree, except that visible names are UUID values. When a large file is split, all of its parts are treated as one logical file. FG data may live elsewhere and does not need to be stored beside the encrypted content. In v1, FG data is stored only in FG's data directory, but users may export it.

## Product Goals

- Protect the confidentiality of file contents.
- Use the v1 encryption suite for project databases, share databases, and encrypted content.
- Allow any encrypted file to be shared together with an FG share database that can restore only that file.
- Allow any encrypted folder to be shared together with an FG share database that can restore only that folder and its descendants.
- Preserve directory hierarchy in v1 so users can manually upload, download, and share a folder from normal file storage or cloud drives.
- Preserve entry count and structure for supported regular files and directories: with split parts treated as one logical file, the encrypted content tree matches the original tree's supported folder and file layout.
- Hide real file and directory names by replacing visible names with generated UUID values.
- Preserve the portable filesystem metadata users expect when restoring normal files and directories.
- Support large files by splitting them into balanced parts when needed.
- Require both sender and recipient to use FG.
- Provide a local desktop WebUI for normal interactive use.
- Provide a CLI for stable automation and scripting.
- Allow renaming files and directories by updating only FG data, without requiring access to encrypted content.
- For adding, moving, or deleting encrypted content in manually managed storage, generate clear operation instructions for the user.
- Use `.fg` for normal FG project databases that contain exactly one top-level object and that object is a directory.
- Use `.fgs` for share databases and all other database shapes.

## User Stories

### Encrypt A Folder

As a user, I want to encrypt a folder so that all files under it are protected, names are hidden, and the encrypted result can still be handled as a normal folder tree.

### Manage One FG Project

As a user, I want one top-level folder to behave like one FG project, with active project metadata tracked in FG's `FoldersGuard` data directory.

### Use A Local WebUI

As a user, I want to create, modify, decrypt, share, import, export, and delete FG projects through a local desktop WebUI without memorizing CLI commands.

### Automate With CLI

As a user, I want scriptable CLI commands for FG workflows so that automated jobs can use the same core behavior as the WebUI.

### Unlock A Top-Level Folder

As a user, I want the top-level FG folder to require a password so that opening the root grants access to all content inside it.

### Share A Folder

As a user, I want to send an encrypted folder and an FG-generated share database to another person so they can decrypt that folder and its descendants, without gaining access to anything outside that folder.

### Share A Single File

As a user, I want to send one encrypted file and an FG-generated share database to another person so they can decrypt only that file.

### Share Selected Items

As a user, I want to share any selected set of encrypted files and folders with one FG share database so the recipient can decrypt only that selected set.

### Hide Names

As a user, I want encrypted files and directories to use random UUID names so that someone browsing the encrypted storage cannot infer sensitive names.

### Rename Without Encrypted Content

As a user, I want to rename a protected file or folder in FG without needing the encrypted content to be present, because renaming changes only the encrypted name mapping stored in FG data.

### Restore Filesystem Metadata

As a user, I want restored files and folders to keep their original modification times, access times, creation times when supported, permissions, and basic Windows file attributes when supported.

### Generate Manual Storage Instructions

As a user, I want FG to tell me exactly how to upload, move, or delete encrypted objects when the encrypted content is stored somewhere FG cannot operate directly.

### Split Large Files

As a user, I want files larger than a configured maximum size to be split into multiple balanced parts so they can be stored on systems with file size limits.

## V1 Feature Set

### FG Native Mode

FG native mode is the only supported encryption mode.

- Every original file has its own file key.
- Every folder has its own folder key.
- The top-level folder must have a password.
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
- Unprotected share databases can be opened without a password and are bearer secrets: anyone with the encrypted content and share database can decrypt the shared content.
- FG data is separate from encrypted content.
- V1 stores FG data only in FG's data directory and supports exporting it.

### Filesystem Entry Policy

FG supports only regular files and directories.

- Symlinks, sockets, FIFOs, device files, and other special entries are ignored as if they do not exist.
- Unsupported entries are not represented in FG metadata and are not reported in normal command output.
- Hard link relationships are not preserved; each hard link path is processed as a normal regular file.

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

## Security Expectations

FG must protect file contents, file names, directory names, and internal key material from unauthorized readers.

FG v1 does not attempt to hide:

- The visible encrypted directory hierarchy.
- The number of visible entries in a directory.
- Approximate encrypted file or part sizes.
- File modification patterns observable through the storage provider.
- The fact that a folder is protected by FG.
- FG reserved data and share databases.
