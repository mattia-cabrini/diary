// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "embed"
)

func cmdResume(db *sql.DB) (err error) {
	date, _ := time.ParseInLocation(time.DateOnly, args.DateInit.Format(time.DateOnly), time.Now().Location())

	rows, err := db.Query("select id, init, fin, inserted, note from entries where init >= ? and init < ? and deleted = 0 order by init", date.Unix(), date.Add(24*time.Hour).Unix())
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() && err == nil {
		var idIn int64
		var initIn int64
		var endIn int64
		var insertedIn int64
		var noteIn string
		var attachmentCount int

		err = rows.Scan(&idIn, &initIn, &endIn, &insertedIn, &noteIn)

		n, _ := fmt.Printf("[%d] %s --> %s\n", idIn, time.Unix(initIn, 0).Format(time.DateTime), time.Unix(endIn, 0).Format(time.DateTime))
		printLine(n, '-', os.Stdout)

		fmt.Printf("%s\n", noteIn)

		rowsAttachments, err2 := db.Query("select id, name, length(content) from attachments where entry_id = ? order by inserted", idIn)
		if err2 != nil {
			err = err2
			return
		}

		for attachmentCount = 0; rowsAttachments.Next() && err == nil; attachmentCount++ {
			var attIdIn int64
			var nameIn string
			var lengthIn int64

			err = rowsAttachments.Scan(&attIdIn, &nameIn, &lengthIn)

			if attachmentCount == 0 {
				printLine(n, '-', os.Stdout)
				fmt.Println("Attachments:")
			}

			fmt.Printf("[%d] %s (%s)\n", attIdIn, nameIn, sizeNorm(lengthIn))
		}

		rowsAttachments.Close()
		if attachmentCount > 0 {
			printLine(n, '-', os.Stdout)
		}

		fmt.Fprintln(os.Stdout)
	}

	return
}
