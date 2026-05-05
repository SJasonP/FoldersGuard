# Export And Import CLI Commands

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
