// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"
	"errors"
	"fmt"
)

func cmdAddAttach(db *sql.DB) (err error) {
	if args.Id <= 0 {
		err = errors.New("you must specify an id")
	}

	if err == nil {
		_, err = RetrieveEntryByID(db, args.Id)

		if err == NOT_FOUND {
			err = fmt.Errorf("entry #%d not found", args.Id)
		}

		if err == nil {
			askForAttachments(db, args.Id)
		}
	}

	return
}
