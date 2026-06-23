# FoldersGuard WebUI Implementation

This document defines the v1 implementation model for the local desktop WebUI.

## Stack

The v1 WebUI uses:

- Wails as the local desktop application shell.
- React for the frontend component model.
- TypeScript for frontend type checking.
- Vite for frontend development and bundling.
- Ant Design for the primary React component library.
- Go application services as the only bridge to FG core behavior.

The WebUI frontend is bundled into the desktop application. FG does not run a general remote web service for the WebUI
by default.

## UI Component Model

The frontend uses Ant Design components for standard application UI.

Ant Design is used for:

- Application layout.
- Menus and navigation.
- Forms and validation display.
- Tables, lists, and empty states.
- Trees and directory-style item selection.
- Modals, drawers, confirmations, and progress feedback.
- Tabs, segmented controls, switches, checkboxes, radio groups, selects, and file-operation controls.
- Messages and notifications.

The frontend should prefer Ant Design components over custom controls when Ant Design provides the needed behavior.

Custom components may wrap Ant Design primitives to express FG-specific workflows, but they must keep Ant Design
accessibility, keyboard behavior, validation behavior, and theme integration intact.

Icons use Ant Design icons by default. Additional icon packages may be introduced only when Ant Design does not provide
a suitable icon for a required command or state.

## Implementation Rules

- The WebUI must be implemented as a local desktop application.
- The WebUI must use Go application services for all FG operations.
- The WebUI must keep cryptography, database access, filesystem validation, and encrypted content operations in Go.
- The WebUI must bundle frontend assets into the desktop application release.
- The WebUI must not require a remote server for normal use.

## Service Boundary

The WebUI calls Go application services. Those services expose user-level operations rather than low-level database or
cryptographic primitives.

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

## Error Boundary

Errors crossing from Go services to the WebUI must be treated as a product boundary, not as raw debug output.

Rules:

- Go services expose stable, user-actionable error categories for common failures.
- The WebUI maps known error categories to localized messages.
- Unknown operation errors use a generic localized operation failure message.
- Error dialogs do not show raw backend errors, stack traces, SQLCipher messages, driver messages, or expandable
  technical-detail sections.
- Wrong database passwords are shown as password failures, but must not disclose whether the database header, metadata,
  keys, or authenticated content check failed.
- Non-empty output folders are shown as output conflicts. When noise file handling is ignore everywhere, recognized
  noise files alone do not make an output folder non-empty. In other modes, the localized message must mention hidden
  files such as `.DS_Store`.
- Output path safety failures are distinct from password failures and must remain clear to the user.

## Operation Progress Model

Long-running operations run through Go application services. A project may contain hundreds of gigabytes of content, so
progress reporting must be reliable, byte-accurate, and informative rather than a generic busy indicator.

### Authority And Transport

- The Go core is the single source of truth for progress. It computes totals and reports advancement.
- The Go core reports progress to the frontend through application events. The frontend subscribes to these events and
  renders them.
- The frontend does not infer progress or completion by scanning files, counting outputs, or timing operations itself.
- Every long-running operation call carries an operation id. The frontend renders only events that match the operation
  id it is currently waiting on, and ignores events from earlier or superseded operations.

### Granularity

- Progress is byte-weighted as the primary measure, because file and folder counts are unreliable at scale. One large
  file or many small files must both produce a meaningful percentage.
- Processed bytes advance within a file as content is streamed, not only when a whole file finishes. A single large file
  must not leave the progress bar stalled.
- Item counts (files, folders, parts) are reported as secondary information alongside byte progress.
- When a precise byte total cannot be established, the operation reports indeterminate progress and still reports
  processed counts and the current phase.

### Totals

- Totals are established before the main work begins, during a measuring phase or from existing project metadata.
- Project creation derives the byte total from the planned source content size.
- Decryption, restore, and verification derive the byte total from stored object and part sizes.
- Export and import derive the byte total from the database payload size.
- If actual sizes differ from the established total during work, reported progress is clamped so it never exceeds the
  total and never moves backward.

### Phases

- Each operation is modeled as an ordered set of phases, for example: measuring, processing content, writing metadata,
  and cleanup.
- Each phase reports its own determinate or indeterminate progress and contributes to an overall operation progress.
- The current phase, the phase position, and the total phase count are reported so the frontend can show both phase-level
  and overall progress.

### Reliability Rules

- Reported progress is monotonic per phase and per operation; it never decreases.
- Progress events are throttled by elapsed time and by processed bytes so that very large operations do not emit an
  excessive number of events. Updates are coalesced between throttle points.
- A progress event is always emitted on phase change, on failure, and on completion.
- Every operation reaches a terminal state: completed, failed, or cancelled. A completed operation reports its final
  byte and item totals as fully processed. The cancelled state is reached only when the application shuts down while an
  operation is running; operations cannot be cancelled by the user.
- Progress reporting must not block, slow, or alter the correctness of the underlying work.

### Reported Fields

Each progress event reports:

- Operation id and operation kind.
- Operation state: pending, running, completed, failed, or cancelled.
- Current phase, phase position, and phase count.
- Whether the current progress is determinate.
- Processed bytes and total bytes when known.
- Processed and total item counts when known.
- The current item name when it is safe to display.
- Throughput and an estimated time remaining when they can be derived.
- An error summary when the operation failed, with sensitive values kept hidden.

### No Cancellation

- Long-running operations cannot be cancelled. Operations such as encryption, decryption, and applying changes would
  leave encrypted output or source files in a partial state if interrupted, so FoldersGuard does not expose a cancel
  control for any operation.
- The operation context is cancelled only internally, when the operation finishes or when the application shuts down, to
  release resources. It is never cancelled in response to a user action.

### Operation Locking

- While an operation runs, the WebUI prevents starting any other operation and prevents changing project selection.
- The form or dialog that triggers an operation closes as soon as the operation starts, so only the progress display
  remains visible.
- The project list is locked while an operation runs: refreshing, searching, selecting a project, and opening project
  actions are all disabled until the operation reaches a terminal state.
- The progress display is shown above other surfaces, including modals and drawers, and is non-interactive.
- Closing the window and quitting the app are blocked while an operation runs. If the user forces the app to quit anyway
  — for example, Force Quit or killing the process — encrypted output and source files may be left incomplete or
  damaged. Any resulting errors or data loss are entirely the user's own responsibility and are not the responsibility
  of FoldersGuard or its developers.

### Long-Running Operations

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

The frontend uses structured localization resources for all user-visible strings and connects the selected locale to Ant
Design's locale provider.

Rules:

- English (United States) and Simplified Chinese resources must both be present.
- English (United States) is the fallback locale.
- UI components reference translation keys instead of hard-coded display strings.
- Adding a new language must be limited to adding locale resources and registering the locale.
- Ant Design locale configuration must match the active FG locale.
- Locale-aware formatting is used for dates, times, numbers, and file sizes.
- Localization resources do not contain passwords, internal keys, database keys, or decrypted key material.

## Theme Implementation

The frontend uses Ant Design theme tokens rather than hard-coded colors in components.

Rules:

- Light and dark token sets must cover all UI states.
- The default theme follows the host system appearance.
- System theme changes are observed while the WebUI is running.
- A user-selected light or dark theme overrides system matching.
- Ant Design theme configuration must switch between light and dark algorithms according to FG theme state.
- Theme state is stored as a user preference, not as project data.
- Theme tokens must preserve readable contrast for normal text, disabled controls, warnings, errors, selected items, and
  destructive actions.

## Packaging

The distributed desktop application is built from the Go backend and bundled frontend assets.

The primary command-line executable remains `foldersguard`, with `fg` provided as a filesystem link for CLI use.
