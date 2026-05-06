# FoldersGuard WebUI Implementation

This document defines the v1 implementation model for the local desktop WebUI.

## Stack

The v1 WebUI uses:

- Wails as the local desktop application shell.
- React for the frontend component model.
- TypeScript for frontend type checking.
- Vite for frontend development and bundling.
- Go application services as the only bridge to FG core behavior.

The WebUI frontend is bundled into the desktop application. FG does not run a general remote web service for the WebUI by default.

## Implementation Rules

- The WebUI must be implemented as a local desktop application.
- The WebUI must use Go application services for all FG operations.
- The WebUI must keep cryptography, database access, filesystem validation, and encrypted content operations in Go.
- The WebUI must bundle frontend assets into the desktop application release.
- The WebUI must not require a remote server for normal use.

## Service Boundary

The WebUI calls Go application services. Those services expose user-level operations rather than low-level database or cryptographic primitives.

Service operations include:

- List active projects.
- Create project.
- Inspect project.
- Import project.
- Load share.
- Inspect share.
- Open project.
- Modify project.
- Apply project changes.
- Decrypt project.
- Decrypt share.
- Verify project content.
- Verify share content.
- Create share.
- Export project.
- Delete project.
- Read and write settings.
- Report application and format information.

The frontend must not:

- Open SQLCipher databases directly.
- Read or write FG encrypted content directly.
- Implement encryption or decryption.
- Hold internal file keys, folder keys, database keys, or decrypted key material as UI state.

## Password Boundary

Passwords are collected by the frontend only as user input for the current operation.

Rules:

- Password fields use hidden input.
- Password values are sent only to the Go service method that needs them.
- Password values are not logged.
- Password values are not stored in frontend persistent state.
- The frontend does not display password-derived information.
- Go services must clear password-derived temporary state as soon as the operation allows.

## Job Model

Long-running operations run as Go-owned jobs.

Job rules:

- Each job has an id.
- Each job reports progress events.
- Each job reports completion, failure, or cancellation.
- Cancellable jobs expose a cancel action.
- The frontend renders progress from job events and does not infer completion by scanning files directly.

Long-running operations include:

- Project creation.
- Project decryption.
- Share loading and decryption.
- Project content verification.
- Share content verification.
- Project import validation.
- Project export.
- Share creation.
- Project modification apply.

## Frontend State Model

The frontend stores UI state and pending user choices.

Project modification state is represented as a pending change set owned by Go services and displayed by the frontend.

The frontend may cache display models for responsiveness, but Go services remain authoritative for:

- Project identity.
- Item identity.
- Path validation.
- Conflict validation.
- Pending change applicability.
- Storage operation plans.

## Localization Implementation

The frontend uses structured localization resources for all user-visible strings.

Rules:

- English (United States) and Simplified Chinese resources must both be present.
- English (United States) is the fallback locale.
- UI components reference translation keys instead of hard-coded display strings.
- Adding a new language must be limited to adding locale resources and registering the locale.
- Locale-aware formatting is used for dates, times, numbers, and file sizes.
- Localization resources do not contain passwords, internal keys, database keys, or decrypted key material.

## Theme Implementation

The frontend uses theme tokens rather than hard-coded colors in components.

Rules:

- Light and dark token sets must cover all UI states.
- The default theme follows the host system appearance.
- System theme changes are observed while the WebUI is running.
- A user-selected light or dark theme overrides system matching.
- Theme state is stored as a user preference, not as project data.
- Theme tokens must preserve readable contrast for normal text, disabled controls, warnings, errors, selected items, and destructive actions.

## Packaging

The distributed desktop application is built from the Go backend and bundled frontend assets.

The primary command-line executable remains `foldersguard`, with `fg` provided as a filesystem link for CLI use.
