# Read And Restore CLI Commands

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
- Unprotected `.fgs` share databases can be opened without a password flag.
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
- Unprotected `.fgs` share databases can be opened without a password flag.
- Prints project id, database type, root item, format version, schema version, item counts, file counts, folder counts, and part counts.
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
- Unprotected `.fgs` share databases can be opened without a password flag.
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
