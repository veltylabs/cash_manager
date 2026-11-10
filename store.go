package cashmanager

import "webtyp.com/crud"

// StoreShift is the store of the shifts.
func (m *Module) StoreShift() crud.Store { return m.store }

// StoreStoredFact is the store of the stored facts.
func (m *Module) StoreStoredFact() crud.Store { return m.store }
