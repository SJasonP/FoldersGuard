# Share CLI Command

## `fg share`

Creates a share database for selected files and folders.

Usage:

```text
fg share <project-id> <item-path>... --out-database <share.fgs> [--share-password-stdin | --share-password-env <name> | --no-share-password] [--password-stdin | --password-env <name>] [--force]
```

Arguments:

- `<item-path>...`: one or more file or folder paths inside the project.
- `<project-id>`: active project id in FG's data directory.
- `--out-database <share.fgs>`: output share database.

Behavior:

- Opens the source project database from FG's data directory.
- Selects only metadata and keys required for the selected item paths.
- Supports a single file, multiple files, a single folder, multiple folders, or a mixed top-level set.
- Includes captured filesystem metadata for selected files and folder subtrees.
- Writes a `.fgs` share database.
- Does not copy, stage, upload, move, or delete encrypted content.
- Prints encrypted content location mappings the user must provide with the share.
- Does not grant access to parent folders, sibling files, sibling folders, or unrelated content unless those items are explicitly selected.

Password behavior:

- By default, FG prompts for a share password and creates a password-protected share database.
- If `--share-password-stdin` or `--share-password-env` is used, the share database is password-protected.
- If `--no-share-password` is used, the share database is unprotected and must be treated as a bearer secret.
- `--no-share-password` must be explicit; FG never creates an unprotected share database by default.

Validation:

- `--out-database` must use `.fgs`.
- At least one `<item-path>` is required.
- Exported `.fg` databases must be imported before sharing from them.
- Existing output paths require `--force`.

Output:

```text
project_id=<uuid>
share_id=<uuid>
share_database=<path>
items=<count>
files=<count>
folders=<count>
parts=<count>
password_protected=<true|false>
content_location source=<encrypted-visible-path> target=<share-visible-path>
```
