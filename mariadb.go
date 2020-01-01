package hostdb

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

// MariadbConfig contains configuration parameters for connecting to the database
type MariadbConfig struct {
	Host   string   `mapstructure:"host"`
	Port   int      `mapstructure:"port"`
	DB     string   `mapstructure:"db"`
	User   string   `mapstructure:"user"`
	Pass   string   `mapstructure:"pass"`
	Params []string `mapstructure:"params"`
}

// MariadbLimit is used to define LIMIT values
type MariadbLimit struct {
	Limit  int
	Offset int
}

// MariadbVersion is used to store the current database version
type MariadbVersion struct {
	Version string `json:"version"`
}

// MariadbWhereClause represents a single argument in an SQL WHERE statement
type MariadbWhereClause struct {
	Relativity string
	Key        []string
	Operator   string
	Value      []string
}

// MariadbWhereGrouping is a slice of Clauses, grouped by parentheses
type MariadbWhereGrouping struct {
	Clauses []MariadbWhereClause
}

// MariadbWhereClauses contains groups of clauses
type MariadbWhereClauses struct {
	Relativity string // defaults to AND
	Groups     []MariadbWhereGrouping
}

// Stringify will convert the MariadbWhereClauses into a string
func (c MariadbWhereClauses) Stringify() (whereSQL string, values []interface{}, err error) {

	if len(c.Groups) < 1 {
		return
	}

	var whereBuilder strings.Builder

	if _, err := fmt.Fprintf(&whereBuilder, "WHERE "); err != nil {
		log.Println(err.Error())
	}

	for groupCounter, group := range c.Groups {

		if len(c.Groups) > 1 && groupCounter >= 1 {
			if c.Relativity == "" {
				c.Relativity = "AND"
			}

			if _, err := fmt.Fprintf(&whereBuilder, "%s ", c.Relativity); err != nil {
				log.Println(err.Error())
			}
		}

		if len(group.Clauses) > 1 {
			if _, err := fmt.Fprintf(&whereBuilder, "( "); err != nil {
				log.Println(err.Error())
			}
		}

		for clauseCounter, clause := range group.Clauses {
			// start by checking things

			// if there's no key, or if there's a missing value with an eval operator
			if len(clause.Key) < 1 || (len(clause.Value) < 1 && !strings.Contains(clause.Operator, "IS N")) {
				return "", nil, errors.New("incomplete WHERE argument")
			}

			if len(clause.Value) > 1 && strings.ToUpper(clause.Operator) != "IN" {
				return "", nil, errors.New("tried to equate more than one value")
			}

			if clause.Relativity == "" {
				clause.Relativity = "AND"
			}

			if clause.Operator == "" {
				clause.Operator = "="
			}

			// if there are existing arguments, use the relativity (AND/OR)
			if whereBuilder.Len() > 8 && clauseCounter >= 1 {
				if _, err := fmt.Fprintf(&whereBuilder, "%v ", clause.Relativity); err != nil {
					log.Println(err.Error())
				}
			}

			// if there are multiple keys, we'll generate multiple arguments, and
			// they should be encapsulated by parentheses, and use OR relativity
			// e.g. WHERE (a = 1 OR b = 1) AND (a = 2 OR b = 2)
			if len(clause.Key) > 1 {
				if _, err := fmt.Fprintf(&whereBuilder, "( "); err != nil {
					log.Println(err.Error())
				}
			}

			counter := 1
			for _, key := range clause.Key {

				// if this is not the first key, make sure we use OR
				if counter > 1 {
					if _, err := fmt.Fprintf(&whereBuilder, "%v ", "OR"); err != nil {
						log.Println(err.Error())
					}
				}

				if _, err := fmt.Fprintf(&whereBuilder, "%v %v ", key, clause.Operator); err != nil {
					log.Println(err.Error())
				}

				// if there are multiple values, wrap them in parentheses
				if len(clause.Value) > 1 {
					if _, err := fmt.Fprintf(&whereBuilder, "("); err != nil {
						log.Println(err.Error())
					}
					for i := 0; i < len(clause.Value); i++ {
						if _, err := fmt.Fprintf(&whereBuilder, "?"); err != nil {
							log.Println(err.Error())
						}
						values = append(values, clause.Value[i])
						if (i + 1) < len(clause.Value) {
							if _, err := fmt.Fprintf(&whereBuilder, ","); err != nil {
								log.Println(err.Error())
							}
						}
					}
					if _, err := fmt.Fprintf(&whereBuilder, ") "); err != nil {
						log.Println(err.Error())
					}
				} else if len(clause.Value) == 1 {
					if _, err := fmt.Fprintf(&whereBuilder, "? "); err != nil {
						log.Println(err.Error())
					}
					values = append(values, clause.Value[0])
				}

				counter++
			}

			if len(clause.Key) > 1 {
				if _, err := fmt.Fprintf(&whereBuilder, ")"); err != nil {
					log.Println(err.Error())
				}
			}
		}

		if len(group.Clauses) > 1 {
			if _, err := fmt.Fprintf(&whereBuilder, ") "); err != nil {
				log.Println(err.Error())
			}
		}
	}

	whereSQL = whereBuilder.String()

	return whereSQL, values, nil

}

// Stringify will convert MariadbLimit into a string
func (l MariadbLimit) Stringify() (sql string) {

	if l.Limit != 0 {
		sql = fmt.Sprintf("LIMIT %d", l.Limit)

		if l.Offset != 0 {
			sql = fmt.Sprintf("%s OFFSET %d", sql, l.Offset)
		}
	}

	return sql

}
