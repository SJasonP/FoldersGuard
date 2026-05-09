# FoldersGuard CLI

This document defines the v1 command-line interface for `foldersguard` and `fg`.

The CLI is the stable automation interface for FG. CLI semantics must remain predictable and scriptable.

Command examples use `fg` as the short form. Every example can replace `fg` with `foldersguard` without changing behavior.

## Global Rules

The primary executable name is `foldersguard`.

The short executable alias is `fg`.

FG is distributed as one primary executable named `foldersguard`.

The short alias `fg` is a filesystem link to `foldersguard`, not a separate executable build.

`foldersguard` and `fg` provide the same commands, flags, validation rules, output format, and exit codes.

Normal command output goes to stdout.

Errors go to stderr and use the invoked executable name as the prefix:

```text
<foldersguard|fg>: <message>
```

Successful commands exit with status code `0`.

Failed commands exit with status code `1`.

Passwords, internal file keys, and folder keys must never be printed in command output, logs, or errors.

All paths are local filesystem paths.

FG supports only regular files and directories. Symlinks, sockets, device files, FIFOs, and other special entries are unsupported.

## Password Input

Passwords are never accepted as positional arguments or flag values.

Default interactive behavior:

- Commands that need a password prompt with hidden input.
- Project password creation prompts require confirmation.
- Password-protected share creation prompts require confirmation.
- Password confirmation is not required when opening an existing database.
- Unprotected `.fgs` share databases can be opened by read/restore commands without a password prompt.

Automation behavior:

- `--password-stdin` reads one password from stdin.
- `--password-env <name>` reads one password from the named environment variable.
- `--share-password-stdin` reads one share password from stdin when creating a password-protected share database.
- `--share-password-env <name>` reads one share password from the named environment variable.

Rules:

- Empty passwords are rejected for `.fg` project databases.
- Password-protected `.fgs` share databases reject empty passwords.
- Unprotected `.fgs` share databases require an explicit `--no-share-password` flag at creation time.
- Environment-variable password input is intended for automation.
- Interactive and environment-variable password input can collect both the project password and share password in one command.
- A single command must not combine `--password-stdin` and `--share-password-stdin`, because stdin can provide only one password value.

## Project References

Commands that operate on active project data use a project reference.

A project reference is a project id.

FG opens the matching active project database from FG's data directory.

Exported `.fg` files are not active databases. They are accepted only by `fg import` as input and by `fg export` as output.

Commands that explicitly accept share databases also accept a path to an `.fgs` share database.

Share databases are accepted by read/restore commands only: `fg decrypt`, `fg inspect`, and `fg verify`.

Project editing, planning, and share creation commands require a project id and reject external `.fg` and `.fgs` database paths.

## Output And Overwrite Rules

Commands do not overwrite existing output files or non-empty output directories unless `--force` is provided.

`--force` never disables authentication, password checks, or output path safety checks.

Commands that restore plaintext must reject restored paths that escape the requested output directory.

Commands that write encrypted content must reject content output paths inside the cleartext source folder.

Commands that create databases must reject database output paths inside the cleartext source folder.

## Database Extensions

`.fg` is used for normal project databases with exactly one top-level object and that object is a directory.

`.fgs` is used for share databases and all other database shapes.

Sharing always creates `.fgs`.

## Command Index

General:

- [`fg help`](cli/general.md#fg-help)
- [`fg version`](cli/general.md#fg-version)

Project lifecycle:

- [`fg encrypt`](cli/encrypt.md#fg-encrypt)
- [`fg decrypt`](cli/read-restore.md#fg-decrypt)
- [`fg inspect`](cli/read-restore.md#fg-inspect)
- [`fg verify`](cli/read-restore.md#fg-verify)
- [`fg export`](cli/export-import.md#fg-export)
- [`fg import`](cli/export-import.md#fg-import)

Metadata operations:

- [`fg rename`](cli/metadata.md#fg-rename)
- [`fg add`](cli/metadata.md#fg-add)
- [`fg move`](cli/metadata.md#fg-move)
- [`fg remove`](cli/metadata.md#fg-remove)

Sharing:

- [`fg share`](cli/share.md#fg-share)

Planning:

- [`fg plan encrypt`](cli/plan.md#fg-plan-encrypt)
- [`fg plan add`](cli/plan.md#fg-plan-add)
- [`fg plan move`](cli/plan.md#fg-plan-move)
- [`fg plan remove`](cli/plan.md#fg-plan-remove)
