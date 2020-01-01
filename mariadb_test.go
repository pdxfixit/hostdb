package hostdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMariadbWhereClauses_Stringify(t *testing.T) {

	//
	// basic
	//
	where := MariadbWhereClauses{
		Groups: []MariadbWhereGrouping{
			{
				Clauses: []MariadbWhereClause{
					{
						Relativity: "",
						Key:        []string{"type"},
						Operator:   "=",
						Value:      []string{"test"},
					},
				},
			},
		},
	}

	whereSQL, values, err := where.Stringify()
	if err != nil {
		t.Errorf("%v", err)
	}

	assert.Equal(t, "WHERE type = ? ", whereSQL, "stringify a basic WHERE clause")
	assert.NotEmpty(t, values, "stringify values")
	assert.Equal(t, values[0], "test", "stringify scalar value")

	//
	// _search
	//
	where = MariadbWhereClauses{
		Groups: []MariadbWhereGrouping{
			{
				Clauses: []MariadbWhereClause{
					{
						Relativity: "AND",
						Key:        []string{fmt.Sprintf("json_search(data, 'one', '%%%v%%')", "m2.local")},
						Operator:   "IS NOT NULL",
						Value:      []string{},
					}, {
						Relativity: "OR",
						Key:        []string{fmt.Sprintf("json_search(context, 'one', '%%%v%%')", "m2.local")},
						Operator:   "IS NOT NULL",
						Value:      []string{},
					}, {
						Relativity: "OR",
						Key:        []string{"hostname"},
						Operator:   "LIKE",
						Value:      []string{fmt.Sprintf("%%%s%%", "m2.local")},
					}, {
						Relativity: "OR",
						Key:        []string{"ip"},
						Operator:   "LIKE",
						Value:      []string{fmt.Sprintf("%%%s%%", "m2.local")},
					}, {
						Relativity: "OR",
						Key:        []string{"type"},
						Operator:   "LIKE",
						Value:      []string{fmt.Sprintf("%%%s%%", "m2.local")},
					}, {
						Relativity: "OR",
						Key:        []string{"committer"},
						Operator:   "LIKE",
						Value:      []string{fmt.Sprintf("%%%s%%", "m2.local")},
					},
				},
			}, {
				Clauses: []MariadbWhereClause{
					{
						Relativity: "",
						Key:        []string{"type"},
						Operator:   "=",
						Value:      []string{"test"},
					},
				},
			},
		},
	}

	whereSQL, values, err = where.Stringify()
	if err != nil {
		t.Errorf("%v", err)
	}

	assert.Equal(t, "WHERE ( json_search(data, 'one', '%m2.local%') IS NOT NULL OR json_search(context, 'one', '%m2.local%') IS NOT NULL OR hostname LIKE ? OR ip LIKE ? OR type LIKE ? OR committer LIKE ? ) AND type = ? ", whereSQL, "WHERE clauses for _search")
	assert.NotEmpty(t, values, "stringify values")
	assert.Len(t, values, 5)
	assert.Equal(t, "%m2.local%", values[0], "stringify scalar value")
	assert.Equal(t, "test", values[4], "stringify scalar value")

}

func TestMariadbLimit_Stringify(t *testing.T) {

	// no limit
	none := MariadbLimit{}
	assert.Empty(t, none.Stringify())

	// simple limit
	limit := MariadbLimit{Limit: 3}
	assert.Equal(t, fmt.Sprintf("LIMIT %d", limit.Limit), limit.Stringify())

	// limit with an offset
	both := MariadbLimit{Limit: 3, Offset: 5}
	assert.Equal(t, fmt.Sprintf("LIMIT %d OFFSET %d", both.Limit, both.Offset), both.Stringify())

}
