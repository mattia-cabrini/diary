// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "embed"
)

//go:embed res/head_dump_day.html
var headDumpDay string

func cmdDumpDay(db *sql.DB) (err error) {
	dateI, _ := time.ParseInLocation(time.DateOnly, args.DateInit.Format(time.DateOnly), time.Now().Location())
	dateE := dateI.Add(24 * time.Hour)
	var fp *os.File

	fp, err = os.OpenFile("index.html", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(args.OutputPerm))
	if err != nil {
		return
	}
	defer fp.Close()

	fmt.Fprintln(fp, "<html>")
	fmt.Fprintf(fp, headDumpDay, dateI.Format(time.DateOnly))

	rows, err := db.Query(QUERY_ENTRY_ALL+" where init >= ? and init < ? and deleted = 0 order by init", dateI.Unix(), dateE.Unix())
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() && err == nil {
		var entry Entry

		entry, err = CreateEntryByScan(rows)
		if err != nil {
			return
		}

		err = entry.FPrintDumpDay(fp, db)
		if err != nil {
			logger.err.Println(err)
			break
		}
	}

	fmt.Fprintln(fp, "</html>")

	return
}
