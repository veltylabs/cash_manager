//go:build browser

package ui

import (
	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/lang"
	"webtyp.com/model"

	"webtyp.com/router"
	"webtyp.com/view"
)

const ID = "cash_manager"
const Label = "Caja"

const descCash lang.Text = "Cash shift management, cash register opening and closing, and financial fact audits."

// Screens declara las pantallas que este módulo ofrece al menú de la app.
func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{
		{
			ID:         ID,
			Label:      Label,
			Icon:       ID,
			Permission: model.ResourceOf(&cashmanager.StoredFactModel),
			Presenter:  cashmanager.NewView(caller, view.WithDescription(descCash)),
		},
	}
}
