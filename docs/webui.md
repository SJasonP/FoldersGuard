# FoldersGuard WebUI

This document defines the v1 local desktop WebUI for FoldersGuard.

The WebUI is the primary interactive interface for normal users. It guides users through project creation, project
import, share loading, project modification, decryption, sharing, export, deletion, settings, and progress review
without requiring command-line knowledge.

The CLI remains the stable automation interface.

## Interface Model

FG provides a local desktop WebUI backed by the same Go core used by the CLI.

Rules:

- The WebUI runs as a local application.
- The WebUI must not expose a general remote HTTP API by default.
- The Go core is responsible for filesystem access, encryption, decryption, database access, validation, and storage
  operations.
- The frontend is responsible for display, navigation, local user interaction, and collecting explicit user choices.
- The frontend must not implement cryptography, parse FG databases directly, or manipulate encrypted content directly.
- Long-running work is owned by the Go core.
- The WebUI shows a running status and progress feedback while long-running work is active.
- Passwords, internal file keys, folder keys, database keys, and decrypted key material must never be shown in the
  WebUI.

## Document Index

General behavior:

- [Application shell, first launch, start screen, project menu, passwords, paths, confirmations, and operation status](webui/general.md).

Project lifecycle:

- [Create Project](webui/project-lifecycle.md#create-project).
- [Inspect Project](webui/project-lifecycle.md#inspect-project).
- [Import Project](webui/project-lifecycle.md#import-project).
- [Export Project](webui/project-lifecycle.md#export-project).
- [Delete Project](webui/project-lifecycle.md#delete-project).

Project modification:

- [Modify Project](webui/project-modify.md#modify-project).
- [Project Browser Layout](webui/project-modify.md#project-browser-layout).
- [Pending Changes](webui/project-modify.md#pending-changes).
- [Apply Changes](webui/project-modify.md#apply-changes).

Sharing and restore:

- [Decrypt Project](webui/share-restore.md#decrypt-project).
- [Verify Project Content](webui/share-restore.md#verify-project-content).
- [Create Share](webui/share-restore.md#create-share).
- [Load Share](webui/share-restore.md#load-share).
- [Share Action Menu](webui/share-restore.md#share-action-menu).
- [Decrypt Share](webui/share-restore.md#decrypt-share).
- [Inspect Share](webui/share-restore.md#inspect-share).
- [Verify Share Content](webui/share-restore.md#verify-share-content).

Settings and support behavior:

- [Settings](webui/settings-about.md#settings).
- [About](webui/settings-about.md#about).
- [Error Handling](webui/settings-about.md#error-handling).
- [Accessibility And Keyboard Behavior](webui/settings-about.md#accessibility-and-keyboard-behavior).
