// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"fmt"
	"os"
)

func cmdInfo(db *sql.DB) (err error) {
	var tmp []int64
	var stat os.FileInfo

	tmp, err = querySingleInt64Array(db, "select count(*) from entries;")
	if te := tmp[0]; err == nil {
		fmt.Printf("Total entries:     %d\n", te)

		tmp, err = querySingleInt64Array(db, "select count(*) from attachments;")
		if ta := tmp[0]; err == nil {
			fmt.Printf("Total attachments: %d (avg. %.2f p.e.)\n", ta, float64(ta)/float64(te))

			tmp, err = querySingleInt64Array(db, "select sum(length(content)) from attachments;")
			if err == nil {
				fmt.Printf("Blob total size:   %s\n", sizeNorm(tmp[0]))

				stat, err = os.Stat(args.Path)
				if err == nil {
					fmt.Printf("DB size:           %s\n", sizeNorm(stat.Size()))

				}
			}
		}
	}

	return
}
