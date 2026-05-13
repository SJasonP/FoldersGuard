# WebUI Settings And Support

## Settings

Settings controls WebUI preferences.

Supported settings:

- Default maximum part size in MB.
- Source file handling: keep source files, or delete source files after successful processing.
- Theme: system, light, or dark.
- Language: system, English (United States), or Simplified Chinese.

Default settings:

- Source file handling defaults to delete source files.
- Default maximum part size defaults to disabled file splitting.
- Theme defaults to system.
- Language defaults to system.

Settings behavior:

- Settings are saved when the user clicks Save Settings.
- Settings that affect running operations apply only to future operations.

## Localization

The WebUI supports localization.

Supported languages:

- English (United States).
- Simplified Chinese.

Localization behavior:

- All user-visible WebUI text must come from localization resources.
- English (United States) is the fallback language.
- Missing translations fall back to English (United States).
- The language setting supports system matching.
- System language changes are applied automatically when the language setting is system.
- Adding a new language must not require changing UI component logic.
- Dates, times, numbers, and file sizes are formatted through localization-aware formatters.
- Paths, project ids, UUID names, command names, file extensions, and cryptographic algorithm identifiers are not
  translated.

## Theme

The WebUI supports complete light and dark themes.

Theme behavior:

- Every WebUI screen, modal, form, table, tree, progress indicator, error state, warning state, empty state, and
  disabled state must support both light and dark themes.
- Theme defaults to system.
- When theme is system, the WebUI automatically matches the host operating system light or dark appearance.
- When the host system appearance changes while the WebUI is running, the WebUI updates without restart.
- User-selected light or dark theme overrides system matching.
- Theme changes apply immediately.
- Color is not the only indicator for errors, warnings, pending changes, selected items, or destructive actions.

## About

About shows product and format information.

About information:

- Product name.
- App id.
- Product version.
- Format version.
- Data directory path.
- Copyright notice.
- Project link.

## Error Handling

The WebUI shows errors without exposing secrets.

Error display rules:

- User-actionable errors are shown near the relevant field or operation.
- Blocking errors use a modal dialog.
- Background operation errors are shown in the operation result.
- Error dialogs show user-facing messages only.
- Error dialogs must not include expandable technical details.
- Passwords, internal keys, database keys, and decrypted key material are never shown.
- Unknown backend errors fall back to a generic operation failure message instead of displaying the raw backend error.
- Known backend errors should be identified by stable error codes or sentinel errors, not by exposing raw low-level
  messages to users.

Common error categories:

- Password authentication failure.
- Database open failure.
- Database validation failure.
- Path not found.
- Path permission failure.
- Output conflict.
- Output folder is not empty, including hidden-file cases such as `.DS_Store`.
- Output path is inside the source folder.
- Output path contains the source folder.
- Source and target paths are identical.
- Encrypted content missing.
- Encrypted content authentication failure.

Unsupported filesystem entries are ignored silently during normal scanning and are not reported as errors.

## Accessibility And Keyboard Behavior

The WebUI supports basic keyboard navigation.

Keyboard behavior:

- Tab moves through interactive controls in visual order.
- Enter activates the primary action in forms and dialogs.
- Escape dismisses dismissible dialogs or returns focus to the previous safe state.
- Destructive actions require explicit confirmation and are not triggered by a single accidental keypress.

Accessibility behavior:

- Form fields have labels.
- Validation errors are associated with the relevant field.
- Progress indicators include textual status.
- Color is not the only indicator for errors, warnings, or pending changes.
