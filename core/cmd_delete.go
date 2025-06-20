// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"errors"

	_ "embed"
)

func cmdDelete(db *sql.DB) (err error) {
	var aff int64

	if args.Id < 0 {
		err = errors.New("invalid id")
		return
	}

	res, err := db.Exec("UPDATE entries set deleted = 1 where id = ?", args.Id)
	if err == nil {
		aff, err = res.RowsAffected()

		if err == nil {
			logger.info.Printf("%d row(s) deleted\n", aff)
		}
	}

	return
}
