package ui

import (

	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const ID = "cash_manager"
const Label = "Caja"

func Browser(caller router.Caller, ids model.IDGenerator, tenantID string) (view.Presenter, error) {
	// 3 screens logic, as per guidelines for view.Presenter. Stub for compilation
	return nil, nil
}
