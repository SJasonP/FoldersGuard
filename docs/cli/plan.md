# Planning CLI Commands

## `fg plan encrypt`

Prints an encryption plan without writing encrypted content or FG databases.

Usage:

```text
fg plan encrypt <source-folder> --max-part-size <bytes>
```

Behavior:

- Scans regular files and directories under `<source-folder>`.
- Calculates file splitting and encrypted storage object counts.
- Calculates captured filesystem metadata fields.
- Ignores unsupported filesystem entries as if they do not exist.
- Applies the noise file handling setting. By default, recognized noise files are ignored as if they do not exist.
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
```

## `fg plan add`

Prints the storage operations that would be required to add cleartext content.

Usage:

```text
fg plan add <project-id> <source-path> <target-folder-path> --staging-content <folder> --max-part-size <bytes> [--password-stdin | --password-env <name>]
```

Behavior:

- Opens the project database.
- Scans `<source-path>`.
- Applies the noise file handling setting. By default, recognized noise files are ignored as if they do not exist.
- Calculates captured filesystem metadata fields for the added files and directories.
- Uses `--max-part-size` to calculate native balanced splitting for newly added files.
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
fg plan move <project-id> <item-path> <target-folder-path> [--password-stdin | --password-env <name>]
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
fg plan remove <project-id> <item-path> [--password-stdin | --password-env <name>]
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
