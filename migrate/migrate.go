package migrate

import (
	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/ddl"
)

func Migrate(conn ddl.Execer, compiler ddl.Compiler) error {
	return ddl.New(conn, compiler).Sync(&cashmanager.Shift{}, &cashmanager.StoredFact{})
}
