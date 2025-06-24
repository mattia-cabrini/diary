// SPDX-License-Identifier: MIT

package diary

import (
	"database/sql"

	_ "embed"
)

//go:embed res/license_info.txt
var LICENSE_INFO string

func cmdLicense(_ *sql.DB) (_ error) {
	print(LICENSE_INFO)
	return
}
