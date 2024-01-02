package models

import "github.com/scylladb/gocqlx/table"

var personMetadata = table.Metadata{
	Name:    "person",
	Columns: []string{"first_name", "last_name", "email"},
	PartKey: []string{"first_name"},
	SortKey: []string{"last_name"},
}

var personTable = table.New(personMetadata)

type Person struct {
	FirstName string
	LastName  string
	Email     []string
	HairColor string `db:"-"` // exported and skipped
	eyeColor  string // unexported also skipped

}
