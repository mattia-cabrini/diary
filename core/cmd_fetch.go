// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"errors"
)

func cmdFetch(db *sql.DB) (err error) {
	var lengthIn int64
	var buf []byte

	if args.OutputFile == nil {
		err = errors.New("no file provided, use -output \"-\" to print on stdout")
		return
	}

	if args.Id < 0 {
		err = errors.New("invalid id")
		return
	}

	row := db.QueryRow("SELECT length(content) from attachments where id = ?", args.Id)
	err = row.Scan(&lengthIn)

	if err == nil {
		buf = make([]byte, lengthIn)

		row := db.QueryRow("SELECT content from attachments where id = ?", args.Id)
		err = row.Scan(&buf)

		if err == nil {
			args.OutputFile.Write(buf)
		}
	}

	return
}
