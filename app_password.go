package main

// ChangeProjectPassword changes an active project database password by re-keying
// it. Encrypted content is not rewritten.
func (a *App) ChangeProjectPassword(request ChangeProjectPasswordRequest) error {
	if err := a.service.ChangeProjectPassword(a.ctx, request.ProjectID, request.OldPassword, request.NewPassword); err != nil {
		return frontendError(err)
	}
	return nil
}

// ChangeSharePassword changes a share database password by re-keying it. It
// protects only future copies of the share database.
func (a *App) ChangeSharePassword(request ChangeSharePasswordRequest) error {
	if err := a.service.ChangeSharePassword(a.ctx, request.DatabasePath, request.OldPassword, request.NewPassword); err != nil {
		return frontendError(err)
	}
	return nil
}
