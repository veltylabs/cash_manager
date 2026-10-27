package cashmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	titleCashManager  = "Cash management"
	OpListStoredFacts = "ListStoredFacts"
)

// NewView construye el Presenter de consulta de hechos de caja.
// Ejemplo: pres := cashmanager.NewView(caller, view.WithDescription(descCash))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	l := view.NewCallerLister(caller,
		view.Ops{Module: (&StoredFact{}).ModelName(), List: OpListStoredFacts},
		func() model.ModelSlice { return &StoredFactList{} })
	allOpts := append([]view.Option{view.WithTitle(titleCashManager)}, opts...)
	return view.New(l, &StoredFact{}, allOpts...)
}
