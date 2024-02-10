package models

import (
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/table"
	"github.com/scylladb/gocqlx/v2"
)

var messageMetaData = table.Metadata{
	Name:    "messages",
	Columns: []string{"id", "user_id", "body"},
}

var messageTable = table.New(messageMetaData)

type Message struct {
	Id     string
	UserId string
	Body   string
}

func InsertMessage(db gocqlx.Session, message *Message) error {
	message.Id = uuid.NewString()
	q := db.Query(messageTable.Insert()).BindStruct(message)
	if err := q.ExecRelease(); err != nil {
		return err
	}
	return nil
}
