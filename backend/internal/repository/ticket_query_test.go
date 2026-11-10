package repository

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildTicketListQueryBindsFilters(t *testing.T) {
	requester := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	assignee := uuid.MustParse("22222222-2222-4222-8222-222222222222")

	listSQL, listArgs, countSQL, countArgs, err := buildTicketListQuery(TicketListFilter{
		Status:      "open",
		Priority:    "high",
		AssigneeID:  &assignee,
		Search:      `100%_off\`,
		RequesterID: &requester,
		SortBy:      "title",
		Order:       "ASC",
		Limit:       5,
		Offset:      10,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, fragment := range []string{
		"status = $1",
		"priority = $2",
		"assignee_id = $3",
		"requester_id = $4",
		"title ILIKE $5 ESCAPE '\\'",
		"ORDER BY title ASC",
		"LIMIT $6 OFFSET $7",
	} {
		if !strings.Contains(listSQL, fragment) {
			t.Errorf("list SQL missing %q\n%s", fragment, listSQL)
		}
	}
	if strings.Contains(listSQL, "100%") {
		t.Fatal("search value was interpolated into SQL")
	}
	if len(listArgs) != 7 || len(countArgs) != 5 {
		t.Fatalf("args list=%d count=%d", len(listArgs), len(countArgs))
	}
	if countArgs[4] != `%100\%\_off\\%` {
		t.Fatalf("escaped search = %#v", countArgs[4])
	}
	if !strings.Contains(countSQL, "requester_id = $4") || strings.Contains(countSQL, "LIMIT") {
		t.Fatalf("unexpected count SQL: %s", countSQL)
	}
	if listArgs[5] != 5 || listArgs[6] != 10 {
		t.Fatalf("limit/offset = %#v", listArgs[5:])
	}
}

func TestBuildTicketListQueryUnassignedAndRejectedSort(t *testing.T) {
	listSQL, _, _, _, err := buildTicketListQuery(TicketListFilter{
		Unassigned: true,
		SortBy:     "created_at",
		Order:      "DESC",
		Limit:      20,
		Offset:     0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(listSQL, "assignee_id IS NULL") {
		t.Fatalf("missing unassigned filter: %s", listSQL)
	}

	_, _, _, _, err = buildTicketListQuery(TicketListFilter{
		SortBy: "created_at; drop table tickets",
		Order:  "DESC",
		Limit:  20,
	})
	if err == nil {
		t.Fatal("expected sort column to be rejected")
	}
}
