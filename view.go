package cashmanager

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleCashManager = "Cash management"

// NewView builds the Presenter that lists the stored cash facts.
//
// Example: presenter := cashmanager.NewView(caller, view.WithDescription(descCash))
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	lister := view.NewCallerLister(caller,
		view.Ops{Module: ModelName, Page: view.Page{First: 50, Size: 50, More: true}},
		func() model.TableSlice { return &StoredFactList{} })
	allOpts := append([]view.Option{view.WithTitle(titleCashManager)}, opts...)
	return view.New(lister, &StoredFact{}, allOpts...)
}
