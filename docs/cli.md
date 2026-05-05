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

Passwords, internal file keys, folder keys, and decrypted metadata secrets must never be printed in command output, logs, or errors.

All paths are local filesystem paths.

FG supports only regular files and directories. Symlinks, sockets, device files, FIFOs, and other special entries are unsupported.

## Password Input

Passwords are never accepted as positional arguments or flag values.

Default interactive behavior:

- Commands that need a password prompt with hidden input.
- Password creation prompts require confirmation.
- Password confirmation is not required when opening an existing database.

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
- A single command must not combine `--password-stdin` and `--share-password-stdin`.

## Project References

Commands that operate on project data use a project reference.

A project reference may be:

- A project id.
- A path to an exported `.fg` project database.

When a command receives a project id, FG opens the matching project database from FG's data directory.

When a command receives a database path, FG opens that database file directly.

Commands that explicitly accept share databases also accept a path to an `.fgs` share database.

Share databases are accepted by read/restore commands only: `fg decrypt`, `fg inspect`, and `fg verify`.

Project editing and share creation commands reject `.fgs` inputs.

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

## Command Summary

General:

- `fg help`
- `fg version`
- `fg schema`

Project lifecycle:

- `fg encrypt`
- `fg decrypt`
- `fg inspect`
- `fg verify`
- `fg export`
- `fg import`

Metadata operations:

- `fg rename`
- `fg add`
- `fg move`
- `fg remove`

Sharing:

- `fg share`

Planning:

- `fg plan encrypt`
- `fg plan add`
- `fg plan move`
- `fg plan remove`

## `fg help`

Prints CLI usage.

Usage:

```text
foldersguard help
foldersguard -h
foldersguard --help
fg help
fg -h
fg --help
```

Running `foldersguard` or `fg` without arguments is equivalent to the matching help command.

## `fg version`

Prints the application id and native format version.

Usage:

```text
fg version
```

Output:

```text
app_id=com.SJasonP.FoldersGuard
format_version=fg-native-v1
```

## `fg schema`

Prints the FG database schema version.

Usage:

```text
fg schema
```

Output:

```text
schema_version=<number>
```

## `fg encrypt`

Encrypts one cleartext top-level folder and creates one active FG project.

Usage:

```text
fg encrypt <source-folder> --content-out <encrypted-content-folder> --max-part-size <bytes> [--export <project.fg>] [--password-stdin | --password-env <name>] [--force]
```

Arguments:

- `<source-folder>`: cleartext top-level folder to encrypt.
- `--content-out <encrypted-content-folder>`: encrypted content output directory.
- `--max-part-size <bytes>`: positive integer maximum part size used for balanced splitting.
- `--export <project.fg>`: optional exported copy of the created project database.

Behavior:

- Creates one FG project in FG's data directory.
- Requires a project password.
- Scans regular files and directories under `<source-folder>`.
- Skips unsupported filesystem entries and reports them.
- Generates UUID visible names for encrypted files and directories.
- Generates random internal file and folder keys.
- Encrypts each file independently.
- Splits files larger than `--max-part-size` using balanced splitting.
- Writes encrypted content to `--content-out`.
- Writes active FG data to FG's data directory.
- If `--export` is provided, writes an exported `.fg` project database.

Validation:

- `<source-folder>` must be a regular directory.
- `--content-out` must not be inside `<source-folder>`.
- `--export`, when provided, must use `.fg`.
- `--export`, when provided, must not be inside `<source-folder>`.
- Existing output paths require `--force`.

Output:

```text
project_id=<uuid>
root_folder_id=<uuid>
content_output=<path>
database_export=<path>
items=<count>
folders=<count>
files=<count>
parts=<count>
storage_objects=<count>
skipped=<count>
skipped_entry=<path> reason=<reason>
```

`database_export` is printed only when `--export` is used.

`skipped_entry` lines are printed only when entries were skipped.

## `fg decrypt`

Decrypts encrypted content using a project or share database.

Usage:

```text
fg decrypt <project-ref> --content <encrypted-content-folder> --out <output-folder> [--password-stdin | --password-env <name>] [--force]
```

Arguments:

- `<project-ref>`: project id, exported `.fg`, or `.fgs` share database.
- `--content <encrypted-content-folder>`: encrypted content folder.
- `--out <output-folder>`: restored plaintext output folder.

Behavior:

- Opens the project or share database.
- Restores real names from FG metadata.
- Reads encrypted content from `--content`.
- Authenticates encrypted file objects and split parts before committing restored plaintext files.
- Writes restored plaintext under `--out`.

Validation:

- Password-protected databases require a password.
- `--content` must exist and be a directory.
- `--out` must not be inside `--content`.
- Restored paths must remain inside `--out`.
- Existing output paths require `--force`.

Output:

```text
project_id=<uuid>
output=<path>
folders=<count>
files=<count>
parts=<count>
restored_files=<count>
```

## `fg inspect`

Displays FG metadata without decrypting file content.

Usage:

```text
fg inspect <project-ref> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project or share database.
- Prints project id, database type, root item, format version, schema version, item counts, file counts, folder counts, part counts, and skipped-entry records stored in metadata.
- Does not require encrypted content to be present.
- Does not decrypt file content.

Output:

```text
project_id=<uuid>
database_type=<project|share>
root_folder_id=<uuid>
root_name=<name>
format_version=<version>
schema_version=<number>
items=<count>
folders=<count>
files=<count>
parts=<count>
storage_objects=<count>
```

## `fg verify`

Verifies database and encrypted content consistency.

Usage:

```text
fg verify <project-ref> --content <encrypted-content-folder> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project or share database.
- Checks that required encrypted content paths exist.
- Authenticates encrypted file objects and split parts.
- Reports missing, extra, or tampered content.
- Does not write plaintext output.

Output:

```text
project_id=<uuid>
checked_objects=<count>
missing_objects=<count>
tampered_objects=<count>
extra_objects=<count>
status=<ok|failed>
```

## `fg export`

Exports an active project database from FG's data directory.

Usage:

```text
fg export <project-id> --out <project.fg> [--password-stdin | --password-env <name>] [--force]
```

Behavior:

- Opens the active project database.
- Writes an exported `.fg` project database.
- Does not require encrypted content to be present.
- Does not decrypt file content.

Validation:

- `--out` must use `.fg`.
- Existing output paths require `--force`.

Output:

```text
project_id=<uuid>
database_output=<path>
```

## `fg import`

Imports an exported project database into FG's data directory.

Usage:

```text
fg import <project.fg> [--password-stdin | --password-env <name>] [--force]
```

Behavior:

- Opens and validates the exported `.fg` project database.
- Adds it to FG's data directory as an active project.
- Does not require encrypted content to be present.
- Does not decrypt file content.

Validation:

- Input must use `.fg`.
- The database must represent a normal project database.
- Existing active project id conflicts require `--force`.

Output:

```text
project_id=<uuid>
imported=true
```

## `fg rename`

Renames a file or folder in FG metadata.

Usage:

```text
fg rename <project-ref> <item-path> <new-name> [--password-stdin | --password-env <name>]
```

Arguments:

- `<item-path>`: real-name path inside the FG project.
- `<new-name>`: new file or folder name, not a path.

Behavior:

- Updates only FG metadata.
- Does not require encrypted content to be present.
- Does not change encrypted UUID paths.
- Does not recalculate content integrity solely because of rename.

Validation:

- `<new-name>` must be a single filesystem name.
- `<new-name>` must not be empty.
- `<new-name>` must not contain path separators.
- `<new-name>` must not be `.` or `..`.
- The destination sibling name must not already exist.

Output:

```text
project_id=<uuid>
item_id=<uuid>
old_name=<name>
new_name=<name>
content_operations=0
```

## `fg add`

Adds cleartext content to an existing project.

Usage:

```text
fg add <project-ref> <source-path> <target-folder-path> --staging-content <folder> [--content <encrypted-content-folder>] [--password-stdin | --password-env <name>] [--force]
```

Behavior:

- Scans `<source-path>`.
- Encrypts new content into `--staging-content`.
- Updates FG metadata with new items, keys, and storage objects.
- Produces storage operation instructions telling the user where to upload or move staged encrypted content.
- If `--content` is provided, FG applies the storage operations directly.
- If `--content` is omitted, FG writes only staged encrypted content and metadata changes, then prints the storage operations for manual execution.

Output:

```text
project_id=<uuid>
operation_plan_id=<uuid>
staging_content=<path>
operations=<count>
operation=<upload|move|delete> source=<path> target=<path>
```

## `fg move`

Moves an item within FG metadata and produces any required storage operation plan.

Usage:

```text
fg move <project-ref> <item-path> <target-folder-path> [--content <encrypted-content-folder>] [--password-stdin | --password-env <name>]
```

Behavior:

- Updates parent-child metadata.
- Preserves internal file and folder keys.
- Produces storage operation instructions if encrypted content paths must move.
- If `--content` is provided, FG applies the storage operations directly.
- If `--content` is omitted, FG updates metadata and prints the storage operations for manual execution.

Output:

```text
project_id=<uuid>
operation_plan_id=<uuid>
operations=<count>
operation=<upload|move|delete> source=<path> target=<path>
```

## `fg remove`

Removes an item from a project.

Usage:

```text
fg remove <project-ref> <item-path> [--content <encrypted-content-folder>] [--password-stdin | --password-env <name>] [--force]
```

Behavior:

- Marks or removes the item in FG metadata according to the storage format rules.
- Produces delete instructions for encrypted content.
- If `--content` is provided, FG deletes encrypted content directly.
- If `--content` is omitted, FG updates metadata and prints the delete operations for manual execution.
- Does not expose sibling or parent keys.

Output:

```text
project_id=<uuid>
operation_plan_id=<uuid>
operations=<count>
operation=delete target=<path>
```

## `fg share`

Creates a share database for one file or folder.

Usage:

```text
fg share <project-ref> <item-path> --content <encrypted-content-folder> --out-content <folder> --out-database <share.fgs> [--share-password-stdin | --share-password-env <name> | --no-share-password] [--password-stdin | --password-env <name>] [--force]
```

Arguments:

- `<item-path>`: file or folder path inside the project.
- `--content <encrypted-content-folder>`: encrypted content root for the source project.
- `--out-content <folder>`: output folder containing only encrypted content needed for the share.
- `--out-database <share.fgs>`: output share database.

Behavior:

- Opens the source project database.
- Selects only metadata and keys required for `<item-path>`.
- Copies or stages the encrypted content needed for the selected file or folder.
- Writes a `.fgs` share database.
- Does not grant access to parent folders, sibling files, sibling folders, or unrelated content.

Password behavior:

- If `--share-password-stdin` or `--share-password-env` is used, the share database is password-protected.
- If `--no-share-password` is used, the share database is unprotected and must be treated as a bearer secret.
- Exactly one share password mode must be selected.

Validation:

- `--out-database` must use `.fgs`.
- Existing output paths require `--force`.

Output:

```text
project_id=<uuid>
share_id=<uuid>
share_database=<path>
share_content=<path>
items=<count>
files=<count>
folders=<count>
parts=<count>
password_protected=<true|false>
```

## `fg plan encrypt`

Prints an encryption plan without writing encrypted content or FG databases.

Usage:

```text
fg plan encrypt <source-folder> --max-part-size <bytes>
```

Behavior:

- Scans regular files and directories under `<source-folder>`.
- Calculates file splitting and encrypted storage object counts.
- Reports unsupported filesystem entries that would be skipped.
- Does not generate durable project data.
- Does not write encrypted content.
- Does not create or update FG databases.

Output:

```text
items=<count>
folders=<count>
files=<count>
parts=<count>
storage_objects=<count>
skipped=<count>
skipped_entry=<path> reason=<reason>
```

## `fg plan add`

Prints the storage operations that would be required to add cleartext content.

Usage:

```text
fg plan add <project-ref> <source-path> <target-folder-path> --staging-content <folder> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project database.
- Scans `<source-path>`.
- Uses `--staging-content` to calculate planned staged encrypted content paths.
- Calculates the metadata and encrypted storage changes that would be required.
- Does not write staged encrypted content.
- Does not update FG metadata.
- Does not create an operation plan record.

Output:

```text
project_id=<uuid>
operations=<count>
operation=<upload|move|delete> source=<path> target=<path>
```

## `fg plan move`

Prints the storage operations that would be required to move an item.

Usage:

```text
fg plan move <project-ref> <item-path> <target-folder-path> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project database.
- Calculates the metadata and encrypted storage changes that would be required.
- Does not update FG metadata.
- Does not move encrypted content.
- Does not create an operation plan record.

Output:

```text
project_id=<uuid>
operations=<count>
operation=<upload|move|delete> source=<path> target=<path>
```

## `fg plan remove`

Prints the storage operations that would be required to remove an item.

Usage:

```text
fg plan remove <project-ref> <item-path> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project database.
- Calculates the metadata and encrypted storage changes that would be required.
- Does not update FG metadata.
- Does not delete encrypted content.
- Does not create an operation plan record.

Output:

```text
project_id=<uuid>
operations=<count>
operation=delete target=<path>
```
