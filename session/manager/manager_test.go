package manager

import (
	"net/url"
	"testing"

	"github.com/coreos/dex/db"
	"github.com/coreos/dex/session"
	"github.com/coreos/go-oidc/oidc"
)

func staticGenerateCodeFunc(code string) GenerateCodeFunc {
	return func() (string, error) {
		return code, nil
	}
}

func newManager() *SessionManager {
	dbMap := db.NewMemDB()
	return NewSessionManager(db.NewSessionRepo(dbMap), db.NewSessionKeyRepo(dbMap))
}

func TestSessionManagerNewSession(t *testing.T) {
	sm := newManager()
	sm.GenerateCode = staticGenerateCodeFunc("boo")
	got, err := sm.NewSession("bogus_idpc", "XXX", "bogus", url.URL{}, "", false, []string{"openid"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != "boo" {
		t.Fatalf("Incorrect Session ID: want=%s got=%s", "boo", got)
	}
}

func TestSessionAttachRemoteIdentityTwice(t *testing.T) {
	sm := newManager()
	sessionID, err := sm.NewSession("bogus_idpc", "XXX", "bogus", url.URL{}, "", false, []string{"openid"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ident := oidc.Identity{ID: "YYY", Name: "elroy", Email: "elroy@example.com"}
	if _, err := sm.AttachRemoteIdentity(sessionID, ident); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := sm.AttachRemoteIdentity(sessionID, ident); err == nil {
		t.Fatalf("Expected non-nil error")
	}
}

func TestSessionManagerExchangeKey(t *testing.T) {
	sm := newManager()
	sessionID, err := sm.NewSession("connector_id", "XXX", "bogus", url.URL{}, "", false, []string{"openid"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	key, err := sm.NewSessionKey(sessionID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	got, err := sm.ExchangeKey(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != sessionID {
		t.Fatalf("Incorrect Session ID: want=%s got=%s", sessionID, got)
	}

	if _, err := sm.ExchangeKey(key); err == nil {
		t.Fatalf("Received nil response from attempt with spent Session key")
	}
}

func TestSessionManagerGetSessionInStateNoExist(t *testing.T) {
	sm := newManager()
	ses, err := sm.getSessionInState("123", session.SessionStateNew)
	if err == nil {
		t.Errorf("Expected non-nil error")
	}
	if ses != nil {
		t.Errorf("Expected nil Session")
	}
}

func TestSessionManagerGetSessionInStateWrongState(t *testing.T) {
	sm := newManager()
	sessionID, err := sm.NewSession("connector_id", "XXX", "bogus", url.URL{}, "", false, []string{"openid"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	ses, err := sm.getSessionInState(sessionID, session.SessionStateDead)
	if err == nil {
		t.Errorf("Expected non-nil error")
	}
	if ses != nil {
		t.Errorf("Expected nil Session")
	}
}

func TestSessionManagerKill(t *testing.T) {
	sm := newManager()
	sessionID, err := sm.NewSession("connector_id", "XXX", "bogus", url.URL{}, "", false, []string{"openid"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ident := oidc.Identity{ID: "YYY", Name: "elroy", Email: "elroy@example.com"}
	if _, err := sm.AttachRemoteIdentity(sessionID, ident); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ses, err := sm.Kill(sessionID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ses == nil {
		t.Fatalf("Expected non-nil Session")
	}

	if ses.ClientState != "bogus" {
		t.Errorf("Unexpected Session: %#v", ses)
	}
}
