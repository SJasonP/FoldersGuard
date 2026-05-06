# WebUI General Behavior

## Global Behavior

The WebUI operates on local filesystem paths selected or entered by the user.

The WebUI uses FG's data directory as the source of local active projects.

Passwords are required before entering protected workflows. When a workflow creates a new project or protected share, the password must be set and confirmed. When a workflow opens an existing project or protected share, the password must be verified before sensitive metadata or content operations are shown.

The WebUI must ask for explicit confirmation before applying project changes, deleting projects, exporting sensitive data, writing operation instructions, decrypting content, or deleting source files.

The WebUI must show clear progress for encryption, decryption, import, export, share generation, verification, operation-guide generation, and project modification apply steps.

When a setting is configured to ask every time, the WebUI asks during the affected workflow and records the selected choice only for that operation.

## Application Shell

The WebUI uses a persistent application shell.

Shell areas:

- Top navigation area.
- Main content area.
- Job status area.
- Modal area for password prompts, confirmations, and blocking errors.

Global navigation actions:

- Home.
- Settings.
- About.

Window behavior:

- If no long-running job is active, closing the window exits the WebUI.
- If a long-running job is active, closing the window asks for confirmation.
- If a project modification session has unapplied changes, leaving the session asks the user to apply, discard, or stay.

## First Launch

On first launch, FG ensures the data directory exists.

First launch behavior:

- If the data directory can be created or opened, the WebUI shows the start screen.
- If the data directory cannot be created or opened, the WebUI shows a blocking error with the data directory path and the underlying error.
- The WebUI does not require an account, network connection, or remote service.

## Start Screen

The start screen shows primary actions and local projects.

Primary actions:

- Create Project.
- Import Project.
- Load Share.
- Settings.
- About.

Local project list:

- Lists all active `.fg` projects in FG's data directory.
- Shows enough non-secret project identity for the user to choose a project.
- Selecting a local project opens the project action menu.

Project list fields:

- Project id.
- Local database file name.
- Local database modified time.
- Availability status.

Project list behavior:

- If there are no active projects, the list shows an empty state and keeps primary actions available.
- The list supports search by non-secret displayed project identity.
- The list supports sorting by project id, local database file name, local database modified time, and availability status.
- The list can be refreshed from FG's data directory.
- Local files that cannot be accessed at the filesystem level are shown as unavailable entries without exposing decrypted metadata.
- Database validation failures are shown only after an open attempt.
- Project names, root folder names, item counts, and decrypted metadata are shown only after password verification.

## Project Action Menu

After selecting a local project, the WebUI shows project actions.

Actions:

- Inspect Project.
- Modify Project.
- Decrypt Project.
- Verify Project Content.
- Create Share.
- Export Project.
- Delete Project.

Each action verifies the project password before continuing.

The menu displays available non-secret project identity before password verification. Sensitive project metadata is displayed only after password verification.

## Password Prompts

Password prompts are modal and block the protected workflow until completed or cancelled.

Rules:

- Password input is hidden.
- Passwords are not shown in validation messages.
- Project password creation requires confirmation.
- Password-protected share creation requires confirmation.
- Password confirmation mismatch blocks continuation.
- Empty project passwords are rejected.
- Empty password-protected share passwords are rejected.
- Existing project or share unlock prompts do not require confirmation.
- Unprotected shares open without a password prompt.
- Cancelling a password prompt returns to the previous safe screen.

Password failure behavior:

- Incorrect passwords show a generic authentication failure.
- Authentication failure must not reveal whether metadata, keys, or content authentication failed internally.
- Retry remains available until the user cancels the workflow.

## Path Selection

Path fields support direct text entry and native path selection.

Path selection rules:

- Directory inputs must select directories.
- File inputs must select files or target file paths as appropriate.
- Existing output must follow FG overwrite rules.
- Path validation errors are shown next to the field and block continuation.
- The WebUI never creates plaintext output outside the user-selected output directory.
- Restored paths must not escape the requested output directory.

Recently used paths:

- The WebUI may remember recently used paths according to Settings.
- Recently used paths are convenience data only.
- Recently used paths are never used silently for destructive operations.

## Confirmation Dialogs

Confirmation dialogs summarize the exact operation before it starts.

Required confirmations:

- Create Project.
- Apply project changes.
- Decrypt Project.
- Decrypt Share.
- Create unprotected share database.
- Export Project.
- Delete Project.
- Write operation guide.
- Delete source files after successful processing.
- Overwrite existing output.
- Cancel a running destructive job.

Confirmation content:

- Operation name.
- Project or share identity when available.
- Input paths.
- Output paths.
- Whether source cleanup may delete files.
- Whether existing output may be overwritten.
- Expected item counts when available.

## Job Status

Long-running operations run as jobs.

Job states:

- Pending.
- Running.
- Cancelling.
- Completed.
- Failed.
- Cancelled.

Job display:

- Operation name.
- Current phase.
- Processed file count when available.
- Processed folder count when available.
- Processed part count when available.
- Total count when known.
- Current path or item name when safe to display.
- Error summary when failed.

Job rules:

- Jobs that can stop safely expose Cancel.
- Cancellation is best effort.
- Files that fail authentication or fail to decrypt are not deleted.
- Files that fail to encrypt are not deleted from the cleartext source.
- Completed jobs show a result summary.
- Failed jobs show recoverable details and keep sensitive values hidden.
