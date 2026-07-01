# Read And Restore CLI Commands

## `fg decrypt`

Decrypts encrypted content using an active project or share database.

Usage:

```text
fg decrypt <project-ref> --content <encrypted-content-folder> --out <output-folder> [--password-stdin | --password-env <name>] [--force] [--resume] [--continue-on-error]
```

Arguments:

- `<project-ref>`: project id or `.fgs` share database.
- `--content <encrypted-content-folder>`: encrypted content folder.
- `--out <output-folder>`: restored plaintext output folder.
- `--resume`: continue an interrupted decryption, keeping the existing output and restoring only files that are missing
  or the wrong size. Mutually exclusive with `--force`.
- `--continue-on-error`: record item-level failures and restore the remaining files instead of aborting on the first
  error. The default aborts on the first error.

Behavior:

- Opens the active project database from FG's data directory or opens a `.fgs` share database directly.
- Restores real names from FG metadata.
- Restores captured filesystem metadata for supported files and directories.
- Reads encrypted content from `--content`.
- Applies the noise file handling setting while matching encrypted content. By default, recognized noise files are
  ignored as if they do not exist.
- Authenticates encrypted file objects and split parts before committing restored plaintext files.
- Writes restored plaintext under `--out`.

Validation:

- Password-protected databases require a password.
- Unprotected `.fgs` share databases can be opened without a password flag.
- Exported `.fg` databases must be imported before use.
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
failed_files=<count>
```

With `--continue-on-error`, `restored_files` is the count that succeeded and `failed_files` is the count that failed.
Each failed item is written to standard error as `failed_file=<visible id>`, and the command exits `1` when any file
failed. Only the non-secret visible file id is printed.

## `fg inspect`

Displays FG metadata without decrypting file content.

Usage:

```text
fg inspect <project-ref> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the active project database from FG's data directory or opens a `.fgs` share database directly.
- Unprotected `.fgs` share databases can be opened without a password flag.
- Exported `.fg` databases must be imported before use.
- Prints project id, database type, project name, root item, format version, created time, updated time, item counts,
  file counts, folder counts, and part counts.
- Does not require encrypted content to be present.
- Does not decrypt file content.

Output:

```text
project_id=<uuid>
database_type=<project|share>
project_name=<name>
root_folder_id=<uuid>
root_name=<name>
format_version=<version>
created_at=<timestamp>
updated_at=<timestamp>
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

- Opens the active project database from FG's data directory or opens a `.fgs` share database directly.
- Unprotected `.fgs` share databases can be opened without a password flag.
- Exported `.fg` databases must be imported before use.
- Checks that required encrypted content paths exist.
- Authenticates encrypted file objects and split parts.
- Reports missing, extra, or tampered content.
- Applies the noise file handling setting. By default, recognized noise files are ignored as if they do not exist and
  are not reported as extra content.
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
