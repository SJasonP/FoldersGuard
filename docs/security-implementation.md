# FoldersGuard Security Implementation

This document records implementation-level security decisions for v1.

## SQLCipher Databases

FG project and share databases use SQLCipher.

Current database opening policy:

- SQLCipher driver: `github.com/mutecomm/go-sqlcipher/v4`.
- Cipher page size: `4096`.
- Foreign-key enforcement: enabled.
- Journal mode: `DELETE`.
- Secure delete: enabled.
- Maximum open connections per database handle: `1`.
- Database files are restricted to owner read/write permissions where supported by the host filesystem.

FG validates SQLCipher database opens by querying `sqlite_master`. This ensures a wrong password or incompatible
database fails during open instead of failing later during metadata access.

Database passwords are currently provided to SQLCipher's password-based keying. FG does not currently derive SQLCipher
raw keys itself.

## Password Handling

Passwords are never accepted as positional arguments or flag values.

Interactive password prompts use hidden input and write prompts to stderr, so stdout remains reserved for command
output.

Project password creation and password-protected share creation require confirmation.

Automation can provide passwords through stdin or through an explicitly named environment variable. This avoids placing
passwords directly in command-line arguments.

## Content Encryption

File content and split part content use AES-256-GCM.

Each file has a random 256-bit file key. Split parts of one logical file use the same file key and are authenticated
with part-specific associated data.

## Filesystem Metadata

FG stores captured filesystem metadata inside the SQLCipher-encrypted project or share database.

Restore policy:

- FG records which metadata fields were captured and restores only those fields.
- Directory metadata is restored after child files and folders are restored.

FG does not rely on filesystem metadata for cryptographic authentication. Content authentication and database
authentication remain separate from metadata restoration.

## Verification Coverage

Current tests verify:

- SQLCipher database files do not expose a plaintext SQLite header.
- SQLCipher database files do not contain plaintext sample metadata.
- Wrong database passwords fail during open.
- SQLCipher database file permissions are restricted to owner read/write.
- Passwords containing double quotes can create and reopen a database.

## Password Change

Planned; not yet implemented.

FG changes a database password by re-keying the SQLCipher database, not by re-encrypting content.

- The old password is verified by opening the database before any change is made.
- Re-keying applies to a backup copy of the database, not to the live database, so an interruption cannot corrupt the
  only working copy.
- The sequence is: back up the database, re-key the copy, validate that the copy opens under the new password by
  querying `sqlite_master`, then atomically replace the live database.
- The re-keyed database keeps owner-only read/write permissions.
- Internal per-file and per-folder content keys are unchanged, so no encrypted content object is rewritten.
- A share database password change affects only future copies of the share database; already distributed copies are
  independent and unaffected.

## Database Backups

Planned; not yet implemented.

- Backups are byte copies of the encrypted SQLCipher database and contain key material.
- Backups are stored with the same owner-only read/write restriction as the live database.
- Backups never expose a plaintext SQLite header or plaintext metadata, because they are copies of the encrypted
  database.
- Backup retention is bounded and configurable.

## Encryption Concurrency

Planned; not yet implemented.

- Parallel file encryption does not weaken the encryption suite. Each file uses an independent random 256-bit key, and
  each object uses per-object nonces that are never reused with the same key.
- Concurrent encryption shares no AEAD cipher state across files, so concurrency does not affect nonce uniqueness or
  authentication.
