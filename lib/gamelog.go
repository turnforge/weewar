package weewar

// =============================================================================
// GameLog Core Data Structures
// =============================================================================

// =============================================================================
// GameLog Implementation
// =============================================================================

// Save persists the current session using the SaveHandler
/*
func (gl *GameLog) Save() error {
	if gl.session == nil {
		return fmt.Errorf("no active session to save")
	}

	if gl.saveHandler == nil {
		return fmt.Errorf("no save handler configured")
	}

	sessionData, err := json.Marshal(gl.session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := gl.saveHandler.Save(sessionData); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Clear pending entries after successful save
	gl.entries = make([]GameLogEntry, 0)

	return nil
}

// SetStatus updates the session status
func (gl *GameLog) SetStatus(status string) error {
	if gl.session == nil {
		return fmt.Errorf("no active session")
	}

	gl.session.Status = status
	gl.session.LastUpdated = time.Now()

	// Auto-save status change
	if gl.autoSave && gl.saveHandler != nil {
		return gl.Save()
	}

	return nil
}
*/
