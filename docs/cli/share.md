# Share CLI Command

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
