// SPDX-License-Identifier: MIT

package diary

import (
	"bufio"
	"database/sql"
	"os"
)

func cmdAdd(db *sql.DB) (err error) {
	var note = args.Note

	if note == "" {
		note, err = editor()
		if err != nil {
			return
		}
	}

	var entry = Entry{
		Init: args.DateInit,
		End:  args.DateEnd,
		Note: note,
	}

	err = entry.Insert(db)
	if err != nil {
		return
	}

	if entry.Id == -1 {
		logger.info.Println("Inserted, could not retrieve id")
		return
	}

	logger.info.Printf("Inserted, with id #%d", entry.Id)

	if !args.NoAttach {
		askForAttachments(db, entry.Id)
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

		var attachment = Attachment{
			Name:    stat.Name(),
			EntryId: id,
			Content: buf,
		}

		errF = attachment.Insert(db)
		if errF != nil {
			logger.err.Printf("could not store file: %v\n", errF)
			continue
		}
		logger.info.Printf("Attached %s (%s)\n", stat.Name(), sizeNorm(len(buf)))
	}
}
