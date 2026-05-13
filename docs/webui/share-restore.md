# WebUI Sharing And Restore

## Decrypt Project

Decrypt Project decrypts encrypted content selected by the user.

The user enters the project password.

After password verification, the WebUI asks for:

- The encrypted content path.
- The output directory path.

Encrypted content path rules:

- The path is required.
- The path must exist.
- The path must be a directory.
- The path may contain the full encrypted content tree.
- The path may contain only part of the encrypted content tree.
- The path is interpreted by UUID names, not by restored real-name structure.

Output path rules:

- The path is required.
- The path must not be inside the encrypted content path.
- Existing output must follow FG overwrite rules.
- Restored paths must not escape the requested output directory.

Before decrypting, the WebUI shows a summary and asks for confirmation.

When decryption starts, FG validates that the selected encrypted content path can be matched to encrypted content known
by the project metadata.

Decrypt behavior:

- FG decrypts every recognized encrypted file or part under the selected directory.
- FG restores the original directory tree structure for the selected content.
- Restored outputs keep supported restored file and directory metadata.
- Directory objects do not contain encrypted file data and do not need to be decrypted.
- Source file handling follows Settings.
- When source file handling is set to delete, each encrypted file is deleted immediately after that file is successfully
  decrypted.
- Encrypted files that fail authentication or cannot be decrypted are not deleted.

Output conflict behavior:

- Existing output paths require overwrite confirmation.
- Without overwrite confirmation, conflicting output paths block decryption.
- If decryption fails after writing some outputs, completed restored files are kept and reported.

After completion, the WebUI reports:

- Decrypted file count.
- Restored folder count.
- Deleted encrypted file count.
- Skipped directory count.
- Failed encrypted file count.
- Output directory path.

## Verify Project Content

Verify Project Content checks an active project database against encrypted content without writing plaintext output.

The user enters the project password.

After password verification, the WebUI asks for the encrypted content path.

Path rules:

- The encrypted content path is required.
- The encrypted content path must exist.
- The encrypted content path must be a directory.
- The path may contain the full encrypted content tree or a recognized subset.

Verify behavior:

- FG checks that required encrypted content paths exist.
- FG authenticates encrypted file objects and split parts.
- FG reports missing, extra, and tampered content.
- Extra content is informational and does not make verification fail.
- FG does not write plaintext output.
- FG does not delete encrypted content.
- Source file handling settings do not apply to verification.

After completion, the WebUI reports:

- Checked object count.
- Missing object count.
- Tampered object count.
- Extra object count, for information only.
- Missing, tampered, and extra object paths when present.
- Verification status.

## Create Share

Create Share creates an `.fgs` share database for selected project content.

The user enters the project password.

After password verification, the WebUI opens the project browser in share-selection mode.

Create Share asks for:

- One or more selected files or folders.
- The share database output path.
- Whether the share database is password-protected.
- The share password and confirmation when password-protected.

Selection rules:

- A share may contain a single file, multiple files, a single folder, multiple folders, or a mixed top-level set.
- Selecting the same item more than once has no effect.
- If both a folder and one of its descendants are selected, the descendant is included through the folder and is not
  duplicated.
- The root folder can be selected.

Share behavior:

- Sharing always creates `.fgs`.
- The share database contains only the metadata and internal keys required for selected content.
- Create Share creates only the `.fgs` share database.
- Create Share does not copy, stage, upload, move, or delete encrypted content.
- Password-protected shares require password confirmation.
- Unprotected shares require explicit confirmation because the share database is a bearer secret.

After completion, the WebUI reports:

- Share database path.
- Selected top-level item count.
- Included file count.
- Included folder count.
- Whether a password is required to open the share.
- Encrypted content locations the user must provide to the recipient.

## Load Share

Load Share opens an `.fgs` share database from the start screen.

The WebUI asks for:

- The `.fgs` path.
- The share password when required.

Path rules:

- The `.fgs` path is required.
- The input must be an `.fgs` database.

After the share database is opened, the WebUI displays share information:

- Share id.
- Database type.
- Format version.
- Top-level item count.
- File count.
- Folder count.
- Part count.
- Storage object count.
- Whether the share database is password-protected.

After loading, the WebUI opens the share action menu.

## Share Action Menu

The share action menu is shown after Load Share opens an `.fgs` database.

Actions:

- Inspect Share.
- Decrypt Share.
- Verify Share Content.

Share actions use the loaded `.fgs` database. If the user leaves the share action menu, the loaded share session is
closed.

## Decrypt Share

Decrypt Share decrypts shared encrypted content.

The WebUI asks for:

- The encrypted content path.
- The output directory path.

Path rules:

- The encrypted content path is required.
- The encrypted content path must be a directory.
- The encrypted content path may contain all shared encrypted content or a recognized subset.
- The output directory path is required.
- The output directory must not be inside the encrypted content path.
- Restored paths must not escape the requested output directory.

Before decrypting, the WebUI shows a summary and asks for confirmation.

Share decryption behavior:

- FG decrypts every recognized encrypted file or part under the selected directory.
- FG restores the shared content's directory structure relative to the output directory.
- Restored outputs keep supported restored file and directory metadata.
- Source file handling follows Settings.
- When source file handling is set to delete, each encrypted file is deleted immediately after that file is successfully
  decrypted.
- Encrypted files that fail authentication or cannot be decrypted are not deleted.

After completion, the WebUI reports:

- Decrypted file count.
- Restored folder count.
- Deleted encrypted file count.
- Failed encrypted file count.
- Output directory path.

## Inspect Share

Inspect Share displays unlocked share metadata without requiring encrypted content to be present.

The WebUI asks for:

- The `.fgs` path.
- The share password when required.

Displayed information:

- Share id.
- Database type.
- Format version.
- Top-level item count.
- File count.
- Folder count.
- Part count.
- Storage object count.
- Whether the share database is password-protected.

Inspect behavior:

- Inspect Share opens the `.fgs` database directly.
- Unprotected share databases can be opened without a password.
- Inspect Share does not require encrypted content to be present.
- Inspect Share does not decrypt file content.
- Inspect Share does not modify FG data or encrypted content.

## Verify Share Content

Verify Share Content checks a share database against encrypted content without writing plaintext output.

The WebUI asks for:

- The `.fgs` path.
- The encrypted content path.
- The share password when required.

Path rules:

- The `.fgs` path is required.
- The encrypted content path is required.
- The encrypted content path must exist.
- The encrypted content path must be a directory.
- The encrypted content path may contain all shared encrypted content or a recognized subset.

Verify behavior:

- FG checks that required encrypted content paths exist.
- FG authenticates encrypted file objects and split parts.
- FG reports missing, extra, and tampered content.
- Extra content is informational and does not make verification fail.
- FG does not write plaintext output.
- FG does not delete encrypted content.
- Source file handling settings do not apply to verification.

After completion, the WebUI reports:

- Checked object count.
- Missing object count.
- Tampered object count.
- Extra object count.
- Verification status.
