// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"time"

	_ "embed"
)

const QUERY_ATTACHMENT_ALL = "select id, name, inserted, entry_id, content from attachments"

type Attachment struct {
	Id       int64
	Name     string
	Inserted time.Time
	EntryId  int64
	Content  []byte
}

func CreateAttachmentByScan(rows *sql.Rows) (a Attachment, err error) {
	var insertedIn int64

	err = rows.Scan(&a.Id, &a.Name, &insertedIn, &a.EntryId, &a.Content)
	if err != nil {
		return
	}

	a.Inserted = time.Unix(insertedIn, 0)

	return
}

func (a *Attachment) Insert(db *sql.DB) (err error) {
	a.Inserted = time.Now()
	_, err = db.Exec("insert into attachments (name, inserted, content, entry_id) values (?, ?, ?, ?)", a.Name, a.Inserted.Unix(), a.Content, a.EntryId)
	return
}

func DeleteAttachment(db *sql.DB, id int64) (aff int64, err error) {
	res, err := db.Exec("UPDATE entries set deleted = 1 where id = ?", id)
	if err == nil {
		aff, err = res.RowsAffected()
	}

	return
}
