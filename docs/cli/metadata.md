# Metadata CLI Commands

## `fg rename`

Renames a file or folder in FG metadata.

Usage:

```text
fg rename <project-ref> <item-path> <new-name> [--password-stdin | --password-env <name>]
```

Arguments:

- `<item-path>`: real-name path inside the FG project, starting with the root folder name.
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
fg remove <project-ref> <item-path> --force [--content <encrypted-content-folder>] [--password-stdin | --password-env <name>]
```

Behavior:

- Marks or removes the item in FG metadata according to the storage format rules.
- Produces delete instructions for encrypted content.
- If `--content` is provided, FG deletes encrypted content directly.
- If `--content` is omitted, FG updates metadata and prints the delete operations for manual execution.
- Does not expose sibling or parent keys.

Validation:

- The root folder cannot be removed.
- `--force` is required because the command changes FG metadata and may delete encrypted content.

Output:

```text
project_id=<uuid>
operation_plan_id=<uuid>
operations=<count>
operation=delete target=<path>
```
