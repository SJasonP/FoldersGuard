# FoldersGuard Native Storage Format

This document describes the v1 native storage model.

## Storage Separation

FG native mode separates encrypted content from FG data.

- Encrypted content is the portable folder tree that users upload, download, move, and share.
- FG data is the private encrypted project database that stores metadata, name mappings, internal keys, and storage layout.

FG data is not embedded in the encrypted content tree. In v1, FG data is stored only in FG's data directory, but users may export it.

App id:

```text
com.SJasonP.FoldersGuard
```

FG data directory name:

```text
FoldersGuard
```

## Encrypted Content Layout

FG native mode stores encrypted content as a directory tree. The logical tree shape mirrors the original cleartext hierarchy, but visible names are UUID values.

There is no required FG metadata file inside the encrypted content tree.

FG reserved filenames and exported FG data do not need to be hidden. It is acceptable for an observer to know that FG is being used.

## FG Data Directory

FG data stores the metadata for one or more FG projects. V1 allows active FG data to live only in FG's data directory.

The platform-specific path is the host OS user configuration directory plus the `FoldersGuard` data directory name.

FG data is stored as SQLCipher-encrypted SQLite databases. FG data may be exported. Exported FG data is a database file that can be backed up or transferred.

## Encrypted SQLite Databases

FG uses SQLite for its internal data model.

The database is encrypted by SQLCipher. FG does not rely on hiding table names or database filenames for security.

Rationale:

- SQLite provides transactions for metadata changes.
- SQLite supports search, filtering, sorting, and indexing.
- SQLite supports incremental updates without rewriting the whole project metadata file.
- Whole-database encryption matches the user model: one password unlocks one project database or one share database.

Database types:

- Project database: the active database for one top-level folder, normally represented as `.fg` when exported.
- Exported project database: a portable copy of a project database.
- Share database: a database containing only the metadata and keys needed for one shared file or folder subtree, always represented as `.fgs`.

## File Extensions

FG database files use two public extensions.

### `.fg`

Use `.fg` only when the database represents a normal FG project with exactly one top-level object and that object is a directory.

### `.fgs`

Use `.fgs` for every database shape other than the normal `.fg` case.

This includes:

- Single-file databases.
- Multi-file databases.
- Multi-directory databases.
- Mixed file and directory databases.
- Databases with one top-level directory when they are produced by the sharing flow.

Sharing always produces `.fgs`.

## Project Database

Each top-level folder is represented by one SQLCipher-encrypted project database.

The project database contains all searchable and mutable FG data for that top-level folder:

- Project identity.
- Format and schema versions.
- Real names.
- Visible UUID names.
- Parent-child relationships.
- File and folder keys.
- Storage layout.
- Part layout.
- Integrity metadata.
- Operation plans.

The top-level folder must require a password. There is no passwordless project database mode.

## Share Database

A share database is generated from a project database.

It contains only the subset required to restore the shared file or folder subtree.

Share databases are rootless sets from the user's point of view. A `.fgs` database may contain one file, multiple files, one directory, multiple directories, or a mixed set of files and directories as top-level objects.

Internally, v1 stores these top-level objects under a virtual folder root so the same parent-child schema can be used. This virtual root is not restored as a user-visible directory and is not part of the shared content tree.

A share database may be password-protected or unprotected:

- Password-protected share database: the recipient must enter the share password.
- Unprotected share database: the database can be opened without a password and is a bearer secret that must be protected during transfer and storage.

The share database is sent together with the encrypted content it describes.

## Logical Schema

The v1 schema includes the following logical tables.

### meta

Stores database-level metadata.

```text
meta:
  key
  value
```

Required keys:

- `app_id`
- `format_version`
- `schema_version`
- `database_type`
- `project_id`
- `root_folder_id`
- `created_at`
- `updated_at`
- `crypto_suite`
- `content_crypto_suite`
- `database_crypto_suite`

### items

Stores logical files and folders.

```text
items:
  item_id
  parent_id
  item_type: file | folder
  visible_name: uuid
  real_name
  sort_name
  original_mode
  original_mod_time
  original_access_time
  original_birth_time
  windows_attributes
  metadata_capabilities
  created_at
  updated_at
  deleted_at
```

Notes:

- `real_name` is searchable only after the encrypted database is unlocked.
- `visible_name` maps to the UUID name in the encrypted content tree.
- `parent_id` records the logical tree.
- Metadata-only rename updates `real_name` and does not change `visible_name`.
- `created_at` and `updated_at` are FG metadata record times, not original filesystem timestamps.
- `original_mode` stores the original portable file mode bits.
- `original_mod_time` stores the original modification time.
- `original_access_time` stores the original access time when available.
- `original_birth_time` stores the original creation time when available.
- `windows_attributes` stores basic Windows file attributes when available.
- `metadata_capabilities` records which original filesystem metadata fields were captured for the item.

Original filesystem timestamps are stored in UTC with nanosecond precision. Host filesystems may round or truncate timestamps when FG restores them.

### folders

Stores folder-specific data.

```text
folders:
  folder_id
  folder_key
```

### files

Stores file-specific data.

```text
files:
  file_id
  file_key
  original_size
  content_algorithm
  storage_kind: single | split
```

### parts

Stores split layout. Non-split files do not have rows in this table.

```text
parts:
  part_id
  file_id
  part_index
  visible_name: uuid
  offset
  size
  integrity
```

### storage_objects

Stores expected encrypted content paths and integrity records.

```text
storage_objects:
  object_id
  item_id
  object_type: file | folder | part
  visible_path
  size
  integrity
```

### operation_plans

Stores manual content operation plans.

```text
operation_plans:
  plan_id
  status
  created_at
  updated_at
```

### operation_steps

Stores individual planned operations.

```text
operation_steps:
  step_id
  plan_id
  step_index
  operation_type
  source_visible_path
  target_visible_path
  expected_integrity
```

## Filesystem Metadata

FG data stores restorable filesystem metadata for every supported file and directory item.

Required restorable metadata:

- Item type.
- Real name.
- Parent-child structure.
- File size for files.
- Modification time.
- Access time when available.
- Creation time when available.
- Permission mode.
- Basic Windows file attributes when available.

Rules:

- Filesystem metadata is stored inside the SQLCipher-encrypted database.
- Metadata capture records the fields that were actually available on the source platform.
- Restore uses the recorded capability set and does not invent missing platform metadata.
- Directory metadata is restored after child entries are restored.
- Split files have one logical metadata record on the file item; individual parts do not have user-visible filesystem metadata.

## Internal Keys

FG still uses internal file and folder keys even though the database is encrypted as a whole.

Rules:

- Each file has one file key.
- Each folder has one folder key.
- File keys encrypt file content.
- Folder keys define the logical authorization boundary and support share database generation.
- The project database stores these keys inside the encrypted database.

The database password unlocks the project database through SQLCipher. Once the database is unlocked, FG can access internal keys and decrypt content according to the operation.

## V1 Encryption Suite

The v1 encryption suite protects project databases, share databases, file content, and part content.

Rules:

- Project databases and share databases use SQLCipher.
- Database passwords are processed by SQLCipher's password-based keying.
- File content and part content use AES-256-GCM.
- File keys are 256-bit random keys.
- Folder keys are 256-bit random keys.
- Nonces are generated so they are never reused with the same key.

## Password Model

Passwords are the only user-facing unlock mechanism.

Top-level project:

- One password unlocks the encrypted project database.
- The unlocked database contains the root folder and all descendants.

Share database:

- One optional password unlocks the encrypted share database.
- If no password is set, the share database can be opened without a password and is a bearer secret.

## Hidden Names

Visible names must not reveal real names.

Rules:

- Every encrypted file and directory name is a generated UUID.
- Real names are stored only inside encrypted project databases or share databases.
- UUID collisions must be handled before writing output.
- The same cleartext name does not need to map to the same UUID across separate encryption runs.

## Metadata-Only Rename

Renaming a protected file or folder does not require encrypted content.

Rules:

- The encrypted UUID name does not change.
- The real name in the project database changes.
- The parent-child relationship does not change.
- Integrity metadata for encrypted content does not need to be recalculated solely because of a rename.

## Manual Storage Operations

FG must support cases where encrypted content can only be changed manually by the user.

For operations that require content changes, FG may produce an operation plan:

```text
operation_plan:
  version
  operations:
    - upload encrypted object to UUID path
    - move encrypted UUID path to another UUID path
    - delete encrypted UUID path
```

When FG has direct access to encrypted content, it may execute the plan itself. When it does not, it should show the instructions to the user.

## File Encryption And Splitting

Each original file has one file key.

When a file is smaller than or equal to the configured maximum part size, it is stored as one encrypted object.

When a file is larger than the configured maximum part size, it is split into balanced parts.

### Balanced Splitting

FG native splitting is based on the maximum allowed part size.

```text
part_count = ceil(original_size / max_part_size)
```

The file is then split into parts that are as evenly sized as possible.

This avoids creating a nearly full leading part followed by a very small trailing part.

### Part Semantics

Parts are not independent files in the security model.

- A split file still has one file key.
- Parts are encrypted and authenticated as pieces of the same logical file.
- Parts are not independently shareable.
- Parts are not independently authorized.
- Decryption requires all parts in the recorded order.
- Any per-part authentication data exists only to detect corruption and tampering, not to create independent part-level access.

## Split Representation

FG data must represent split parts as one logical file.

Physical representation:

- A split logical file is stored as a directory named with the file UUID.
- That directory contains UUID-named part files.
- Part order, sizes, offsets, and integrity data are stored in FG data.
- Users treat the UUID directory as the encrypted representation of one original file.

This preserves the "one visible object per logical file" behavior in ordinary file managers and cloud drives. A cleartext file may become a UUID directory in encrypted storage only when splitting is required.

## Integrity

Native format must support tamper detection.

The implementation detects:

- Modified encrypted content.
- Missing parts.
- Reordered parts.
- FG database content modification.
- Wrong keys.
- Wrong algorithm identifiers.

## Versioning

All native storage data includes version information.

Required version and algorithm metadata:

- Native format version.
- FG database schema version.
- Encryption algorithm identifier.
- Crypto suite identifier.

The implementation rejects unsupported format versions.

## Filesystem Entry Scope

Only regular files and directories are represented in native storage.

Unsupported entries are ignored as if they do not exist:

- Symlinks.
- Sockets.
- FIFOs.
- Device files.
- Other special filesystem entries.

Unsupported entries are not recorded in the database.

Hard link relationships are not represented. Each hard link path is stored as a normal regular file entry.
