# WebUI Project Modification

## Modify Project

Modify Project lets the user edit the project through a virtual project browser controlled by FG.

Project identity is tracked by FG item ids in the project database.

## Open Project

The user enters the project password.

After password verification, the WebUI asks for the encrypted content storage path.

Path behavior:

- If the user enters a path, the WebUI connects the project to that encrypted content location.
- If the user leaves the path empty, the project is loaded without encrypted content access.
- If a path is entered, FG validates that recognized encrypted content can be matched to project metadata.
- Content connection status is shown in the project browser.

After loading, the WebUI displays project information:

- Project name.
- Project id.
- Root folder name.
- Created time.
- Updated time.
- File count.
- Folder count.
- Part count.
- Whether encrypted content is connected.

## Project Browser Layout

The project browser is the main editing surface.

Browser areas:

- Folder tree.
- Current folder item list.
- Breadcrumb path.
- Details panel.
- Search field.
- Pending changes panel.
- Action toolbar.

Displayed entry fields:

- Item name.
- Item type.
- Size for files.
- Child count for folders when available.
- Modification time when available.
- Metadata capture status.
- Encrypted content presence when a content path is connected.
- Pending change state.

Selection behavior:

- Single selection shows item details.
- Multiple selection is supported for share creation, remove, and move operations.
- Selecting a folder can show its children in the item list.
- Breadcrumb navigation changes the current folder.
- Search filters displayed items by real names inside the unlocked project.

## Project Browser Actions

Supported edit actions:

- Rename file.
- Rename folder.
- Move file.
- Move folder.
- Remove file.
- Remove folder.
- Add local file.
- Add local folder.
- Create empty folder.

Rules:

- Rename and move are recorded as pending changes.
- Remove marks the item for deletion from the project.
- Add imports selected local files or folders as new project items.
- Adding a local folder recursively imports supported regular files and directories under it.
- Unsupported filesystem entries are ignored as if they do not exist.
- Hard links are treated as normal files.
- Each added item receives new FG identity, metadata, UUID names, and internal key material.
- The edit session remains pending until the user applies or discards it.

Validation rules:

- The root folder cannot be renamed, moved, or removed.
- Names must be valid single filesystem names.
- Names must not be empty.
- Names must not contain path separators.
- Names must not be `.` or `..`.
- Sibling names must be unique after pending changes are applied.
- An item cannot be moved into itself.
- A folder cannot be moved into its own descendant.
- Added local content must exist at apply time.

## Pending Changes

The WebUI keeps a pending change set during modification.

The pending change set includes:

- Renamed files.
- Renamed folders.
- Moved files.
- Moved folders.
- Removed files.
- Removed folders.
- Added files.
- Added folders.
- New encrypted content that must be uploaded or written.
- Encrypted content that must be moved or deleted.

Pending changes panel:

- Groups changes by type.
- Shows item names and logical paths.
- Shows whether each change needs encrypted content access.
- Allows undoing a pending change when it does not invalidate later changes.
- Allows discarding all pending changes.

Conflict handling:

- Conflicts are shown before apply.
- Apply is disabled while blocking conflicts exist.
- Non-blocking warnings are shown in the apply summary.
- If a pending change becomes invalid because source local files changed or disappeared, apply is blocked until the user removes or updates that change.

## Apply Changes

When the user chooses to apply changes, the WebUI shows a summary and asks for confirmation.

After confirmation:

- FG validates that the pending change set is still applicable.
- FG updates project metadata.
- FG encrypts added local files.
- FG creates any required encrypted content objects for added files and folders.
- FG creates a storage operation plan for uploads, moves, and deletes.

If encrypted content is connected:

- FG applies content operations directly.
- FG reports the operations that were performed.

If encrypted content is not connected:

- FG writes encrypted output for new content into FG's data directory.
- FG writes a human-readable operation guide into FG's data directory.
- The operation guide describes exactly how the user should upload, move, or delete encrypted objects.
- The operation guide includes real file names where useful for user understanding.

The operation guide format is controlled by Settings.

Apply result:

- Completed apply clears the pending change set.
- Failed apply keeps a result summary.
- Failed file encryption does not delete the source file.
- Failed content authentication or missing content blocks destructive content deletion.
