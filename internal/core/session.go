package core

type SessionManager interface {
	TryLock(sessionID string) bool
	Unlock(sessionID string)
}
