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
- Operation status area.
- Modal area for password prompts, confirmations, and blocking errors.

Global navigation actions:

- Home.
- Settings.
- About.

Window behavior:

- If no long-running operation is active, closing the window exits the WebUI.
- If a long-running operation is active, normal window close is blocked.
- If the user forcibly terminates the application or the host system terminates it, any resulting incomplete work is outside FG's control.
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
- Shows local project names for the user to choose a project.
- Local project names are stored outside `.fg` databases in FG's local data directory.
- Selecting a local project opens the project action menu.

Project list fields:

- Project id.
- Project name.
- Local database modified time.
- Availability status.

Project list behavior:

- If there are no active projects, the list shows an empty state and keeps primary actions available.
- The list supports search by non-secret displayed project identity.
- The list supports sorting by project id, project name, local database modified time, and availability status.
- The list can be refreshed from FG's data directory.
- Local files that cannot be accessed at the filesystem level are shown as unavailable entries without exposing decrypted metadata.
- Database validation failures are shown only after an open attempt.
- Root folder names, item counts, and decrypted metadata are shown only after password verification.

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

The menu includes an editable local project name field. Changing this field updates only FG's local project name record and does not modify the `.fg` database.

## Password Prompts

Password prompts are modal and block the protected workflow until completed or dismissed.

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
- Dismissing a password prompt returns to the previous safe screen.

Password failure behavior:

- Incorrect passwords show a generic authentication failure.
- Authentication failure must not reveal whether metadata, keys, or content authentication failed internally.
- Retry remains available until the user dismisses the workflow.

## Path Selection

Path fields support direct text entry and native path selection.

Path selection rules:

- Directory inputs must select directories.
- File inputs must select files or target file paths as appropriate.
- Existing output must follow FG overwrite rules.
- Path validation errors are shown next to the field and block continuation.
- The WebUI never creates plaintext output outside the user-selected output directory.
- Restored paths must not escape the requested output directory.

## Confirmation Dialogs

Confirmation dialogs prevent accidental actions by summarizing the operation before it starts.

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

Confirmation content:

- Operation name.
- Project or share identity when available.
- Input paths.
- Output paths.
- Whether source cleanup may delete files.
- Whether existing output may be overwritten.
- Expected item counts when available.

## Operation Status

Long-running operations show status while work is active.

Operation states:

- Pending.
- Running.
- Completed.
- Failed.

Status display:

- Operation name.
- Current phase.
- Processed file count when available.
- Processed folder count when available.
- Processed part count when available.
- Total count when known.
- Current path or item name when safe to display.
- Error summary when failed.

Status rules:

- Files that fail authentication or fail to decrypt are not deleted.
- Files that fail to encrypt are not deleted from the cleartext source.
- Completed operations show a result summary.
- Failed operations show recoverable details and keep sensitive values hidden.
