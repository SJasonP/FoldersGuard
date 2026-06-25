# Maintenance CLI Commands

These commands follow the same global rules, password input, project reference, and output conventions as the rest of
the CLI.

## `fg passwd`

Changes the password of a project database, or a password-protected share database, by re-keying it. Content is not
re-encrypted.

Usage:

```text
fg passwd <project-ref> [--password-stdin | --password-env <name>] [--new-password-stdin | --new-password-env <name>]
fg passwd --share <share.fgs> [--password-stdin | --password-env <name>] [--new-password-stdin | --new-password-env <name>]
```

Arguments:

- `<project-ref>`: the project id of an active project database.
- `--share <share.fgs>`: path to a share database instead of an active project.

Behavior:

- Verifies the old password by opening the database.
- Snapshots the database to a backup before re-keying.
- Re-keys a copy, confirms the copy opens under the new password, then atomically replaces the live database.
- Does not require encrypted content to be present.
- Does not change any encrypted content object.
- Internal per-file and per-folder keys are unchanged.

Validation:

- The old password must be correct.
- The new password must be confirmed; interactive prompts require confirmation.
- Empty new passwords are rejected for `.fg` project databases and password-protected `.fgs` share databases.
- A single command must not combine `--password-stdin` and `--new-password-stdin`, because stdin can provide only one
  password value.

Output:

```text
project_id=<uuid>
rekeyed=true
content_operations=0
```

## `fg backups list`

Lists the database backups retained for a project.

Usage:

```text
fg backups list <project-ref>
```

Behavior:

- Lists retained backups for the project, newest first.
- Does not require a password.

Output:

```text
project_id=<uuid>
backup_id=<id> reason=<reason> created=<timestamp> size=<bytes>
backup_id=<id> reason=<reason> created=<timestamp> size=<bytes>
```

## `fg backups restore`

Restores a project database from one of its backups.

Usage:

```text
fg backups restore <project-ref> <backup-id> [--force]
```

Arguments:

- `<backup-id>`: a backup id from `fg backups list`.

Behavior:

- Snapshots the current database to a new backup before restoring.
- Replaces the active project database with the selected backup atomically.
- Does not change any encrypted content object.

Validation:

- `<backup-id>` must identify an existing retained backup for the project.
- Replacing the active database requires `--force`.

Output:

```text
project_id=<uuid>
restored_from=<backup-id>
```
