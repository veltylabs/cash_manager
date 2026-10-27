package cashmanager

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const (
	titleCashManager  = "Cash management"
	OpListStoredFacts = "ListStoredFacts"
)

// Item implementa view.Itemizer para StoredFact.
func (f *StoredFact) Item() view.Item {
	return view.Item{
		ID:          f.Id,
		Label:       f.Kind + " · " + f.Currency + " " + fmt.Sprint(f.Amount),
		Description: f.Scope + " · " + f.SourceType + " " + f.SourceId,
	}
}

// NewView construye el Presenter de consulta de hechos de caja.
func NewView(caller router.Caller, opts ...view.Option) view.Presenter {
	allOpts := append([]view.Option{view.WithTitle(titleCashManager)}, opts...)
	return view.New(&factLister{caller: caller}, &StoredFact{}, allOpts...)
}

type factLister struct {
	caller router.Caller
}

func (l *factLister) List(done func([]model.Model, error)) {
	out := &StoredFactList{}
	l.caller.Call("cashmanager."+OpListStoredFacts, nil, out, func(err error) {
		if err != nil {
			done(nil, err)
			return
		}
		rows := make([]model.Model, len(*out))
		for i, f := range *out {
			rows[i] = f
		}
		done(rows, nil)
	})
}
