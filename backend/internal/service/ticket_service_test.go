package service

import (
	"testing"
)

// =============================================
// Helper functions (taruh juga di ticket_service.go kalau belum ada)
// =============================================

func isValidTransition(from, to, role string) bool {
	transitions := map[string]map[string][]string{
		"open": {
			"in_progress": {"agent", "admin"},
			"closed":      {"requester", "admin"},
		},
		"in_progress": {
			"resolved": {"agent", "admin"},
		},
		"resolved": {
			"closed": {"requester", "agent", "admin"},
			"open":   {"requester"},
		},
	}

	allowedRoles, ok := transitions[from][to]
	if !ok {
		return false
	}

	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

func canUserAccessTicket(userID, role, ticketOwnerID string) bool {
	if role == "agent" || role == "admin" {
		return true
	}
	return userID == ticketOwnerID
}

// =============================================
// TEST 1: Valid Status Transition
// =============================================

func TestStatusTransition_Valid(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		to     string
		role   string
		expect bool
	}{
		{"agent open to in_progress", "open", "in_progress", "agent", true},
		{"admin open to in_progress", "open", "in_progress", "admin", true},
		{"requester cannot open to in_progress", "open", "in_progress", "requester", false},
		{"agent in_progress to resolved", "in_progress", "resolved", "agent", true},
		{"requester resolved to closed", "resolved", "closed", "requester", true},
		{"requester resolved to open", "resolved", "open", "requester", true},
		{"agent cannot reopen", "resolved", "open", "agent", false},
		{"requester can cancel open ticket", "open", "closed", "requester", true},
		{"invalid transition closed to open", "closed", "open", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTransition(tt.from, tt.to, tt.role)
			if result != tt.expect {
				t.Errorf("got %v, want %v", result, tt.expect)
			}
		})
	}
}

// =============================================
// TEST 2: Requester Access Restriction
// =============================================

func TestRequesterCannotAccessOtherTicket(t *testing.T) {
	canAccess := canUserAccessTicket("user-a", "requester", "user-b")
	if canAccess {
		t.Error("requester should NOT access other user's ticket")
	}
}

// =============================================
// TEST 3: Agent Can Access Any Ticket
// =============================================

func TestAgentCanAccessAnyTicket(t *testing.T) {
	canAccess := canUserAccessTicket("agent-1", "agent", "user-b")
	if !canAccess {
		t.Error("agent SHOULD access any ticket")
	}
}

// =============================================
// TEST 4: Admin Can Access Any Ticket
// =============================================

func TestAdminCanAccessAnyTicket(t *testing.T) {
	canAccess := canUserAccessTicket("admin-1", "admin", "user-x")
	if !canAccess {
		t.Error("admin SHOULD access any ticket")
	}
}

// =============================================
// TEST 5: Owner Can Access Own Ticket
// =============================================

func TestOwnerCanAccessOwnTicket(t *testing.T) {
	canAccess := canUserAccessTicket("user-a", "requester", "user-a")
	if !canAccess {
		t.Error("requester SHOULD access their own ticket")
	}
}
