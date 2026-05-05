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

FG validates SQLCipher database opens by querying `sqlite_master`. This ensures a wrong password or incompatible database fails during open instead of failing later during metadata access.

Database passwords are currently provided to SQLCipher's password-based keying. FG does not currently derive SQLCipher raw keys itself.

## Password Handling

The current CLI development interface reads the project password from `FG_PASSWORD`.

This avoids placing the password directly in command-line arguments. It is not the final interactive password input model.

## Content Encryption

File content and split part content use AES-256-GCM.

Each file has a random 256-bit file key. Split parts of one logical file use the same file key and are authenticated with part-specific associated data.

## Verification Coverage

Current tests verify:

- SQLCipher database files do not expose a plaintext SQLite header.
- SQLCipher database files do not contain plaintext sample metadata.
- Wrong database passwords fail during open.
- SQLCipher database file permissions are restricted to owner read/write.
- Passwords containing double quotes can create and reopen a database.
