// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "embed"
)

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

	rows, err := db.Query("select id, init, fin, inserted, note from entries where init >= ? and init < ? and deleted = 0 order by init", dateI.Unix(), dateE.Unix())
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
		if err != nil {
			return
		}

		fmt.Fprintf(fp, `<p>
			<span class="record-id">#%d</span>
			From %s to %s<br>
		`, idIn, time.Unix(initIn, 0).Format(time.DateTime), time.Unix(endIn, 0).Format(time.DateTime))

		fmt.Fprintf(fp, "<div>%s</div>", noteIn)

		var buf []byte
		rowsAttachments, err2 := db.Query("select id, name, length(content), content from attachments where entry_id = ? order by inserted", idIn)
		if err2 != nil {
			err = err2
			return
		}

		for attachmentCount = 0; rowsAttachments.Next() && err == nil; attachmentCount++ {
			var attIdIn int64
			var nameIn string
			var lengthIn int64

			err = rowsAttachments.Scan(&attIdIn, &nameIn, &lengthIn, &buf)

			if attachmentCount == 0 {
				fmt.Fprintln(fp, "<table><tr><th>#</th><th>Size</th><th>Name</th></tr>")
			}

			var afp *os.File
			afp, err = os.OpenFile(nameIn, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.FileMode(args.OutputPerm))
			if err != nil {
				return
			}

			afp.Write(buf)
			clear(buf)
			afp.Close()

			fmt.Fprintf(fp, "<tr><td>%d</td><td>%s</td><td><a href=\"%s\" target=\"_blank\">%s</a></td></tr>", attIdIn, sizeNorm(lengthIn), nameIn, nameIn)
		}

		rowsAttachments.Close()
		if attachmentCount > 0 {
			fmt.Fprintln(fp, "</table>")
		}
		fmt.Fprintln(fp, "</p>")
	}

	fmt.Fprintln(fp, "</html>")

	return
}
