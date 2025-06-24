// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

func cmdResume(db *sql.DB) (err error) {
	date, _ := time.ParseInLocation(time.DateOnly, args.DateInit.Format(time.DateOnly), time.Now().Location())

	rows, err := db.Query(QUERY_ENTRY_ALL+" where init >= ? and init < ? and deleted = 0 order by init", date.Unix(), date.Add(24*time.Hour).Unix())
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() && err == nil {
		var entry Entry

		entry, err = CreateEntryByScan(rows)
		if err != nil {
			break
		}

		entry.FPrintResume(db, os.Stdout)
		fmt.Fprintln(os.Stdout)
	}

	return
}
