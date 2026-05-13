# WebUI Project Lifecycle

## Create Project

Create Project creates one active FG project from one cleartext directory.

The WebUI asks for:

- The source directory path.
- The encrypted content output path.
- The project password and confirmation.

Source path rules:

- The path is required.
- The path must exist.
- The path must be a directory.
- The selected directory becomes the project's top-level folder.

Encrypted content output path rules:

- The path is required.
- The path must not be inside the source directory.
- Existing output must follow FG overwrite rules.

Maximum part size rules:

- The WebUI always uses the default maximum part size from Settings.
- The value is configured in MB in Settings.
- Values greater than 4 MB enable balanced file splitting.
- Values less than or equal to 4 MB disable file splitting.

Create behavior:

- FG creates the active project database in FG's data directory.
- FG encrypts the selected directory into encrypted content.
- FG preserves supported directory structure, names, metadata, and file content according to the native format.
- Unsupported filesystem entries are ignored as if they do not exist.
- Hard links are treated as normal files.
- Source file handling follows Settings.
- When source file handling is set to delete, each cleartext file is deleted immediately after that file is successfully
  encrypted.
- Cleartext files that fail to encrypt are not deleted.
- When source file handling is set to delete, directories are removed after their child entries have been processed and
  only if they are empty.

Before creation starts, the WebUI shows a confirmation summary.

After completion, the WebUI reports:

- Project id.
- Project name.
- Encrypted content path.
- Encrypted file count.
- Encrypted folder count.
- Encrypted part count.
- Deleted cleartext file count.
- Failed file count.

## Inspect Project

Inspect Project displays unlocked project metadata without requiring encrypted content to be present.

The user enters the project password before project details are shown.

Displayed information:

- Project id.
- Database type.
- Project name.
- Root folder id.
- Root folder name.
- Format version.
- Created time.
- Updated time.
- Item count.
- File count.
- Folder count.
- Part count.
- Storage object count.

Inspect behavior:

- Inspect Project opens the active project database from FG's data directory.
- Inspect Project does not require encrypted content to be present.
- Inspect Project does not decrypt file content.
- Inspect Project does not modify FG data or encrypted content.

## Import Project

Import Project imports an exported `.fg` project database into FG's data directory.

The WebUI asks for the exported `.fg` path.

Import rules:

- The input path is required.
- The input must be a `.fg` database.
- The imported database becomes an active local project only after validation.
- If an active project with the same project id already exists, FG either overwrites it when the user explicitly allows
  overwrite or stops the import without changing it.
- Import does not require encrypted content to be present.
- Import does not decrypt file content.

After import, the WebUI refreshes the local project list and shows the imported project.

## Export Project

Export Project writes an active local project database to a user-selected `.fg` output path.

The user enters the project password before export.

Export rules:

- Exported `.fg` files are backups or transfer files.
- Exporting does not change the active project.
- Export must follow FG overwrite rules.
- Export does not require encrypted content to be present.
- Export does not decrypt file content.

After completion, the WebUI reports the exported `.fg` path.

## Delete Project

Delete Project removes an active local project database from FG's data directory.

The user enters the project password before deletion.

Delete rules:

- Deleting a project removes local FG data only.
- Deleting a project does not delete encrypted content unless another explicit content operation is selected.
- The WebUI must ask for confirmation before deletion.
- The confirmation shows the project id, local database file name, and data directory path affected.

After deletion, the WebUI refreshes the local project list.
