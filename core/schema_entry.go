// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

const QUERY_ENTRY_ALL = "select id, init, fin, inserted, note, deleted from entries"

type Entry struct {
	Id int64

	Init     time.Time
	End      time.Time
	Inserted time.Time

	Note    string
	Deleted bool
}

func CreateEntryByScan(rows *sql.Rows) (e Entry, err error) {
	var initIn int64
	var endIn int64
	var insertedIn int64
	var deleted int64

	err = rows.Scan(&e.Id, &initIn, &endIn, &insertedIn, &e.Note, &deleted)
	if err != nil {
		return
	}

	e.Init = time.Unix(initIn, 0)
	e.End = time.Unix(endIn, 0)
	e.Inserted = time.Unix(insertedIn, 0)
	e.Deleted = deleted != 0

	return
}

func RetrieveEntryByID(db *sql.DB, id int64) (e Entry, err error) {
	rows, err := db.Query(QUERY_ENTRY_ALL+" where id = ?", id)

	if err == nil {
		if rows.Next() {
			e, err = CreateEntryByScan(rows)
		} else {
			err = NOT_FOUND
		}
	}

	return
}

func (e *Entry) Insert(db *sql.DB) (err error) {
	e.Inserted = time.Now()

	res, err := db.Exec("insert into entries (init, fin, inserted, note, deleted) values (?, ?, ?, ?, 0)", e.Init.Unix(), e.End.Unix(), e.Inserted.Unix(), e.Note)
	if err != nil {
		return
	}

	e.Deleted = false
	e.Id, err = res.LastInsertId()

	if err != nil {
		e.Id = -1
	}

	return
}

func (e *Entry) FPrintDumpDay(fp *os.File, db *sql.DB) (err error) {
	var attachmentCount int

	logger.info.Printf("Entry #%d\n", e.Id)

	fmt.Fprintf(fp, `<p>
			<span class="record-id">#%d</span>
			<span class="time">From %s to %s</span><br>
		`, e.Id, e.Init.Format(time.DateTime), e.End.Format(time.DateTime))

	noteHtml := strings.Replace(e.Note, "\n", "<br>", -1)
	fmt.Fprintf(fp, "%s", noteHtml)

	rows, err := db.Query(QUERY_ATTACHMENT_ALL+" where entry_id = ? order by inserted", e.Id)
	if err != nil {
		return
	}
	defer rows.Close()

	for attachmentCount = 0; rows.Next(); attachmentCount++ {
		var attachment Attachment
		attachment, err = CreateAttachmentByScan(rows)
		if err != nil {
			return
		}

		logger.info.Printf("Attachment #%d\n", attachment.Id)

		if attachmentCount == 0 {
			fmt.Fprintln(fp, "<table><tr><th>#</th><th>Size</th><th>Name</th></tr>")
		}

		var afp *os.File
		afp, err = os.OpenFile(attachment.Name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.FileMode(args.OutputPerm))
		if err != nil {
			return
		}

		afp.Write(attachment.Content)
		afp.Close()

		fmt.Fprintf(fp, "<tr><td>%d</td><td>%s</td><td><a href=\"%s\" target=\"_blank\">%s</a></td></tr>", attachment.Id, sizeNorm(len(attachment.Content)), attachment.Name, attachment.Name)
	}

	if attachmentCount > 0 {
		fmt.Fprintln(fp, "</table>")
	}

	fmt.Fprintln(fp, "</p><hr>")

	return
}

func (e *Entry) FPrintResume(db *sql.DB, fp *os.File) (n int, err error) {
	var attachmentCount int

	n, _ = fmt.Fprintf(fp, "[%d] %s --> %s\n", e.Id, e.Init.Format(time.DateTime), e.End.Format(time.DateTime))
	printLine(n, '-', fp)
	fmt.Fprintf(fp, "%s\n", e.Note)

	if db == nil {
		return
	}

	rows, err := db.Query("select id, name, length(content) from attachments where entry_id = ? order by inserted", e.Id)
	if err != nil {
		return
	}

	for attachmentCount = 0; rows.Next() && err == nil; attachmentCount++ {
		var attIdIn int64
		var nameIn string
		var lengthIn int64

		err = rows.Scan(&attIdIn, &nameIn, &lengthIn)

		if attachmentCount == 0 {
			printLine(n, '-', fp)
			fmt.Fprintln(fp, "Attachments:")
		}

		fmt.Fprintf(fp, "[%d] %s (%s)\n", attIdIn, nameIn, sizeNorm(lengthIn))
	}

	rows.Close()
	if attachmentCount > 0 {
		printLine(n, '-', fp)
	}

	return
}
