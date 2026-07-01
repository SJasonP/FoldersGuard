# Encrypt CLI Command

## `fg encrypt`

Encrypts one cleartext top-level folder and creates one active FG project.

Usage:

```text
fg encrypt <source-folder> --content-out <encrypted-content-folder> --max-part-size <bytes> [--export <project.fg>] [--password-stdin | --password-env <name>] [--force] [--continue-on-error]
```

Arguments:

- `<source-folder>`: cleartext top-level folder to encrypt.
- `--content-out <encrypted-content-folder>`: encrypted content output directory.
- `--max-part-size <bytes>`: positive integer maximum part size used for balanced splitting.
- `--export <project.fg>`: optional exported copy of the created project database.
- `--continue-on-error`: record item-level failures and encrypt the remaining files instead of aborting on the first
  error. The default aborts on the first error.

Behavior:

- Creates one FG project in FG's data directory.
- Requires a project password.
- Scans regular files and directories under `<source-folder>`.
- Ignores unsupported filesystem entries as if they do not exist.
- Applies the noise file handling setting. By default, recognized noise files are ignored as if they do not exist and
  are not added to FG metadata.
- Captures restorable filesystem metadata for supported files and directories.
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
encrypted_files=<count>
failed_files=<count>
```

`database_export` is printed only when `--export` is used.

With `--continue-on-error`, `encrypted_files` is the count that succeeded and `failed_files` is the count that failed.
Each failed item is written to standard error as `failed_file=<visible id>`, and the command exits `1` when any file
failed. Only the non-secret visible file id is printed.
