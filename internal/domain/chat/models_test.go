package chat

import "testing"

func TestRoleValid(t *testing.T) {
	valid := []Role{RoleUser, RoleAssistant, RoleSystem}
	for _, r := range valid {
		if !r.Valid() {
			t.Errorf("expected %q to be valid", r)
		}
	}

	invalid := []Role{"admin", "moderator", ""}
	for _, r := range invalid {
		if r.Valid() {
			t.Errorf("expected %q to be invalid", r)
		}
	}
}
