# WebUI General Behavior

## Global Behavior

The WebUI operates on local filesystem paths selected or entered by the user.

The WebUI uses FG's data directory as the source of local active projects.

Passwords are required before entering protected workflows. When a workflow creates a new project or protected share,
the password must be set and confirmed. When a workflow opens an existing project or protected share, the password must
be verified before sensitive metadata or content operations are shown.

The WebUI must ask for explicit confirmation before applying project changes, deleting projects, exporting sensitive
data, showing manual processing instructions, decrypting content, or deleting source files.

The WebUI must show clear progress for encryption, decryption, import, export, share generation, verification,
manual processing guide generation, and project modification apply steps.

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
- If a long-running operation is active, closing the window and quitting the app are blocked, and the user is warned that
  the operation cannot be cancelled and must finish first.
- If the user forcibly terminates the application or the host system terminates it, any resulting incomplete or damaged
  work is entirely the user's own responsibility and is not the responsibility of FoldersGuard or its developers.
- If a project modification session has unapplied changes, leaving the session asks the user to apply, discard, or stay.

## First Launch

On first launch, FG ensures the data directory exists.

First launch behavior:

- If the data directory can be created or opened, the WebUI shows the start screen.
- If the data directory cannot be created or opened, the WebUI shows a blocking error with the data directory path.
- Startup errors must not expose low-level technical details in the dialog.
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
- Local files that cannot be accessed at the filesystem level are shown as unavailable entries without exposing
  decrypted metadata.
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

The menu displays available non-secret project identity before password verification. Sensitive project metadata is
displayed only after password verification.

The menu includes an editable local project name field. Changing this field updates only FG's local project name record
and does not modify the `.fg` database.

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

- Incorrect passwords show a clear password failure message.
- Authentication failure must not reveal whether metadata, keys, or content authentication failed internally.
- Retry remains available until the user dismisses the workflow.

## Path Selection

Path fields support direct text entry and native path selection.

Path selection rules:

- Directory inputs must select directories.
- File inputs must select files or target file paths as appropriate.
- Existing output must follow FG overwrite rules.
- Path validation errors are shown next to the field and block continuation.
- Non-empty output directories must be reported clearly. When noise file handling is ignore everywhere, recognized noise
  files alone do not make an output directory non-empty. In other modes, the message must tell the user to choose an
  empty directory, remove existing files including hidden files such as `.DS_Store`, or explicitly enable force
  overwrite.
- Backend path safety rejections must be translated into user-facing messages. The WebUI must distinguish at least:
  output inside the source folder, output containing the source folder, identical source and target paths, and non-empty
  output folders.
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
- Delete source files after successful processing.
- Overwrite existing output.

Confirmation content:

- Operation name.
- Project or share identity when available.
- Input paths.
- Output paths.
- Whether source file handling may delete files.
- Whether existing output may be overwritten.
- Expected item counts when available.

## Operation Status

Long-running operations show status while work is active. Because a project may hold hundreds of gigabytes, the status
shows real progress reported by the Go core, not a generic busy indicator.

Operation states:

- Pending.
- Running.
- Completed.
- Failed.

Status display:

- Operation name.
- Current phase, with phase position and phase count.
- A determinate progress percentage when totals are known.
- An indeterminate running indicator when totals are not yet known.
- Processed bytes and total bytes, shown in localized human-readable sizes.
- Processed file count when available.
- Processed folder count when available.
- Processed part count when available.
- Total count when known.
- Throughput and estimated time remaining when they can be derived.
- Current path or item name when safe to display.
- Error summary when failed.

Progress display rules:

- Progress is byte-weighted as the primary measure and advances within a large file, not only when whole files finish.
- The reported percentage never moves backward and reaches one hundred percent only on successful completion.
- The frontend renders only progress events that match the operation it is currently waiting on.

No cancellation and locking:

- Operations cannot be cancelled. No operation shows a cancel control.
- The form or dialog that started the operation closes as soon as the operation begins, leaving only the progress
  status visible.
- While an operation runs, the project list is locked: refreshing, searching, selecting a project, opening project
  actions, and starting another operation are all disabled until the operation finishes.
- Closing the window and quitting the app are blocked while an operation runs. If the user forces the app to quit
  anyway, any resulting incomplete or damaged data is entirely the user's own responsibility and not the responsibility
  of FoldersGuard or its developers.

Status rules:

- Files that fail authentication or fail to decrypt are not deleted.
- Files that fail to encrypt are not deleted from the cleartext source.
- Completed operations show a result summary.
- Failed operations show recoverable details and keep sensitive values hidden.

## Operation Options

Planned; not yet implemented. These options apply to the long-running content operations and default to the
v1 behavior.

Resume:

- When an encryption or decryption output already exists from an interrupted run, the WebUI offers to resume or to start
  fresh.
- Resuming skips objects that are already complete and processes only what remains; the progress display counts
  already-completed work as processed.
- Starting fresh re-processes every object and is the default.

Failure handling:

- An operation can abort on the first error or continue past item-level failures, following the default failure handling
  setting and an optional per-operation override.
- When continuing, the result summary lists the failed item count and per-item reasons, with sensitive values hidden.

Concurrency:

- Encryption can process several files at once, following the encryption concurrency setting.
- Concurrency does not change the progress display, which remains byte-weighted.
