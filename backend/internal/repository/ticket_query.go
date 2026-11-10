package repository

import (
	"fmt"
	"strings"
)

var ticketSortColumns = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"priority":   "priority",
	"status":     "status",
	"title":      "title",
}

func buildTicketListQuery(filter TicketListFilter) (listSQL string, listArgs []any, countSQL string, countArgs []any, err error) {
	sortCol, ok := ticketSortColumns[filter.SortBy]
	if !ok {
		return "", nil, "", nil, fmt.Errorf("invalid sort column")
	}
	order := strings.ToUpper(filter.Order)
	if order != "ASC" && order != "DESC" {
		return "", nil, "", nil, fmt.Errorf("invalid sort order")
	}
	if filter.Limit < 1 {
		return "", nil, "", nil, fmt.Errorf("invalid limit")
	}
	if filter.Offset < 0 {
		return "", nil, "", nil, fmt.Errorf("invalid offset")
	}

	var where []string
	var args []any
	add := func(expr string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(expr, len(args)))
	}

	if filter.Status != "" {
		add("status = $%d", filter.Status)
	}
	if filter.Priority != "" {
		add("priority = $%d", filter.Priority)
	}
	if filter.Unassigned {
		where = append(where, "assignee_id IS NULL")
	} else if filter.AssigneeID != nil {
		add("assignee_id = $%d", *filter.AssigneeID)
	}
	if filter.RequesterID != nil {
		add("requester_id = $%d", *filter.RequesterID)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		add("title ILIKE $%d ESCAPE '\\'", "%"+escapeLike(search)+"%")
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	countSQL = "SELECT COUNT(*) FROM tickets" + whereSQL
	countArgs = append([]any{}, args...)

	args = append(args, filter.Limit, filter.Offset)
	listSQL = fmt.Sprintf(
		"SELECT %s FROM tickets%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		ticketColumns,
		whereSQL,
		sortCol,
		order,
		len(args)-1,
		len(args),
	)
	return listSQL, args, countSQL, countArgs, nil
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
