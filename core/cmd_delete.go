// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"errors"
)

func cmdDelete(db *sql.DB) (err error) {
	if args.Id < 0 {
		err = errors.New("invalid id")
		return
	}

	aff, err := DeleteAttachment(db, args.Id)
	if err == nil {
		logger.info.Printf("%d row(s) deleted\n", aff)
	}

	return
}
