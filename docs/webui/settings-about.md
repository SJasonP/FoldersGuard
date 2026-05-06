# WebUI Settings And Support

## Settings

Settings controls WebUI preferences.

Supported settings:

- Operation guide format: `txt` or `md`.
- Default maximum part size.
- Source cleanup mode: ask every time, keep source files, or delete each file after successful processing.
- Remember recently used paths.
- Clear recently used paths.
- Window state persistence.
- Theme: system, light, or dark.

Default settings:

- Source cleanup mode defaults to ask every time.
- Operation guide format defaults to `txt`.
- Default maximum part size defaults to no limit.
- Theme defaults to system.
- Recently used paths are remembered.

Settings behavior:

- Settings changes require confirmation before saving.
- Settings that affect running jobs apply only to future operations.
- Clearing recently used paths does not affect projects, encrypted content, or FG databases.

## About

About shows product and format information.

About information:

- Product name.
- App id.
- Native format version.
- Schema version.
- Data directory path.
- WebUI implementation stack.
- CLI executable name.
- Short CLI alias.

## Error Handling

The WebUI shows errors without exposing secrets.

Error display rules:

- User-actionable errors are shown near the relevant field or operation.
- Blocking errors use a modal dialog.
- Background job errors are shown in the job result.
- Technical details can be expanded when useful for debugging.
- Passwords, internal keys, database keys, and decrypted key material are never shown.

Common error categories:

- Password authentication failure.
- Database open failure.
- Database validation failure.
- Path not found.
- Path permission failure.
- Output conflict.
- Encrypted content missing.
- Encrypted content authentication failure.
- Job cancellation.

Unsupported filesystem entries are ignored silently during normal scanning and are not reported as errors.

## Accessibility And Keyboard Behavior

The WebUI supports basic keyboard navigation.

Keyboard behavior:

- Tab moves through interactive controls in visual order.
- Enter activates the primary action in forms and dialogs.
- Escape cancels dismissible dialogs or returns focus to the previous safe state.
- Destructive actions require explicit confirmation and are not triggered by a single accidental keypress.

Accessibility behavior:

- Form fields have labels.
- Validation errors are associated with the relevant field.
- Progress indicators include textual status.
- Color is not the only indicator for errors, warnings, or pending changes.
