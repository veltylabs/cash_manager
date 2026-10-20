package migrate

import (
	"github.com/veltylabs/cash_manager"
	"webtyp.com/ddl"
)

func Migrate(conn ddl.Execer, c ddl.Compiler) error {
	m := ddl.New(conn, c)
	return m.Sync(
		&cashmanager.Shift{},
		&cashmanager.StoredFact{},
	)
}
