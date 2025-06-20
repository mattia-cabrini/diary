// SPDX-License-Identifier: MIT

package diary

import (
	"bufio"
	"database/sql"
	"os"
	"time"

	_ "embed"
)

func cmdAdd(db *sql.DB) (err error) {
	var reg = time.Now()
	var note = args.Note

	if note == "" {
		note, err = editor()
		if err != nil {
			return
		}
	}

	res, err := db.Exec("insert into entries (init, fin, inserted, note, deleted) values (?, ?, ?, ?, 0)", args.DateInit.Unix(), args.DateEnd.Unix(), reg.Unix(), note)
	if err != nil {
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.info.Println("Inserted, could not retrieve id")
		return
	} else {
		logger.info.Printf("Inserted, with id #%d", id)
	}

	if !args.NoAttach {
		askForAttachments(db, id)
	}

	return
}

func askForAttachments(db *sql.DB, id int64) {
	var k = bufio.NewScanner(os.Stdin)
	var buf = make([]byte, 0)

	for {
		var errF error
		var fp *os.File

		print("Attachment: ")
		k.Scan()

		if k.Text() == "" {
			break
		}

		var attachmentPath = k.Text()

		stat, errStat := os.Stat(attachmentPath)
		if errStat != nil {
			logger.warn.Printf("file does not exist: %s\n", attachmentPath)
			continue
		}

		fp, errF = os.OpenFile(attachmentPath, os.O_RDONLY, 0444)
		if errF != nil {
			logger.err.Printf("could not open file: %v\n", errF)
			continue
		}

		if stat.Size() > getMaxBlobSize() {
			logger.err.Printf("file too big: max %s\n", sizeNorm(getMaxBlobSize()))
			continue
		}

		clear(buf)
		buf = make([]byte, stat.Size())

		_, errF = fp.Read(buf)
		if errF != nil {
			logger.err.Printf("could not open file: %v\n", errF)
			continue
		}

		_, errF = db.Exec("insert into attachments (name, inserted, content, entry_id) values (?, ?, ?, ?)", stat.Name(), time.Now().Unix(), buf, id)
		if errF != nil {
			logger.err.Printf("could not store file: %v\n", errF)
			continue
		}
		logger.info.Printf("Attached %s (%s)\n", stat.Name(), sizeNorm(len(buf)))
	}
}
