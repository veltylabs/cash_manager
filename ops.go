package cashmanager

import (
	"webtyp.com/model"
	"webtyp.com/money"
	"webtyp.com/router"
	"webtyp.com/time"

	"webtyp.com/orm"

)

type TodayResponse struct {
	Shift *Shift
	Facts StoredFactList
}

func (r *TodayResponse) Schema() []model.Field { return nil }
func (r *TodayResponse) Pointers() []any { return nil }
func (r *TodayResponse) EncodeFields(w model.FieldWriter) {}
func (r *TodayResponse) DecodeFields(r2 model.FieldReader) {}
func (r *TodayResponse) IsNil() bool { return r == nil }

type ReorderResponse struct {
	Items []LowStockItem
}

func (r *ReorderResponse) Schema() []model.Field { return nil }
func (r *ReorderResponse) Pointers() []any { return nil }
func (r *ReorderResponse) EncodeFields(w model.FieldWriter) {}
func (r *ReorderResponse) DecodeFields(r2 model.FieldReader) {}
func (r *ReorderResponse) IsNil() bool { return r == nil }

type ThirtyDayResponse struct {
	Facts    StoredFactList
	Forecast []ForecastLine
}

func (r *ThirtyDayResponse) Schema() []model.Field { return nil }
func (r *ThirtyDayResponse) Pointers() []any { return nil }
func (r *ThirtyDayResponse) EncodeFields(w model.FieldWriter) {}
func (r *ThirtyDayResponse) DecodeFields(r2 model.FieldReader) {}
func (r *ThirtyDayResponse) IsNil() bool { return r == nil }

func (m *Module) openShift(ctx router.Context) {
	var args OpenShiftArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" || args.CashierId == "" || args.Currency == "" {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	qb := m.db.Query(&Shift{})
	qb.Where(Shift_.TenantId).Eq(args.TenantId)
	qb.Where(Shift_.CashierId).Eq(args.CashierId)
	qb.Where(Shift_.Status).Eq(ShiftOpen)

	existing, err := ReadAllShift(qb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}
	if len(existing) > 0 {
		ctx.WriteStatus(409); ctx.Write([]byte("conflict"))
		return
	}

	now := time.NowUnix()

	shift := &Shift{
		Id:          m.deps.IDs.NewID(),
		TenantId:    args.TenantId,
		CashierId:   args.CashierId,
		Status:      ShiftOpen,
		OpeningCash: args.OpeningCash,
		Currency:    args.Currency,
		OpenedAt:    now,
		UpdatedAt:   time.NowNano(),
	}

	if err := shift.Validate(model.ActionCreate); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if err := m.db.Create(shift); err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	_ = ctx.Encode(shift)
}

func (m *Module) closeShift(ctx router.Context) {
	var args CloseShiftArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" || args.Id == "" {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	qb := m.db.Query(&Shift{})
	qb.Where(Shift_.TenantId).Eq(args.TenantId)
	qb.Where(Shift_.Id).Eq(args.Id)

	shift := &Shift{}
	_, err := ReadOneShift(qb, shift)
	if err != nil {
		if orm.IsNotFound(err) {
			ctx.WriteStatus(404); ctx.Write([]byte("not found"))
			return
		}
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	if shift.Status == ShiftClosed {
		ctx.WriteStatus(409); ctx.Write([]byte("conflict"))
		return
	}

	now := time.NowUnix()

	factQb := m.db.Query(&StoredFact{})
	factQb.Where(StoredFact_.TenantId).Eq(args.TenantId)
	factQb.Where(StoredFact_.At).Gte(shift.OpenedAt)
	factQb.Where(StoredFact_.At).Lte(now)

	facts, err := ReadAllStoredFact(factQb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	expected := shift.OpeningCash
	for _, f := range facts {
		if f.Currency == shift.Currency {
			if f.Kind == string(money.Charged) {
				expected += f.Amount
			} else if f.Kind == string(money.Refunded) {
				expected -= f.Amount
			}
		}
	}

	shift.Status = ShiftClosed
	shift.ClosingCashCounted = args.ClosingCashCounted
	shift.ClosingCashExpected = expected
	shift.Difference = args.ClosingCashCounted - expected
	shift.ClosedAt = now
	shift.UpdatedAt = time.NowNano()

	// We validate only the updated fields via UpdateFields, so skip Validate entirely. UpdateFields will validate the specific fields.

	if err := m.db.UpdateFields(shift, []string{Shift_.Status, Shift_.ClosingCashCounted, Shift_.ClosingCashExpected, Shift_.Difference, Shift_.ClosedAt, Shift_.UpdatedAt}, orm.Eq(Shift_.Id, args.Id), orm.Eq(Shift_.TenantId, args.TenantId)); err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	_ = ctx.Encode(shift)
}

func (m *Module) getToday(ctx router.Context) {
	var args GetTodayArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" || args.CashierId == "" {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	qb := m.db.Query(&Shift{})
	qb.Where(Shift_.TenantId).Eq(args.TenantId)
	qb.Where(Shift_.CashierId).Eq(args.CashierId)
	qb.Where(Shift_.Status).Eq(ShiftOpen)

	shifts, err := ReadAllShift(qb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	if len(shifts) == 0 {
		_ = ctx.Encode(&TodayResponse{})
		return
	}

	shift := shifts[0]

	now := time.NowUnix()
	factQb := m.db.Query(&StoredFact{})
	factQb.Where(StoredFact_.TenantId).Eq(args.TenantId)
	factQb.Where(StoredFact_.At).Gte(shift.OpenedAt)
	factQb.Where(StoredFact_.At).Lte(now)

	facts, err := ReadAllStoredFact(factQb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	_ = ctx.Encode(&TodayResponse{
		Shift: shift,
		Facts: facts,
	})
}

func (m *Module) listStoredFacts(ctx router.Context) {
	var args ListStoredFactsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" || args.Scope == "" || args.From == 0 || args.To == 0 {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	qb := m.db.Query(&StoredFact{})
	qb.Where(StoredFact_.TenantId).Eq(args.TenantId)
	qb.Where(StoredFact_.Scope).Eq(args.Scope)
	qb.Where(StoredFact_.At).Gte(args.From)
	qb.Where(StoredFact_.At).Lte(args.To)

	facts, err := ReadAllStoredFact(qb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	_ = ctx.Encode(&facts)
}

func (m *Module) reorder(ctx router.Context) {
	var items []LowStockItem
	if m.deps.LowStock != nil {
		var err error
		items, err = m.deps.LowStock.LowStock(m.deps.TenantID)
		if err != nil {
			ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
			return
		}
	}
	if items == nil {
		items = []LowStockItem{} // Ensure it's never nil
	}

	_ = ctx.Encode(&ReorderResponse{Items: items})
}

func (m *Module) thirtyDay(ctx router.Context) {
	var args ListStoredFactsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" || args.From == 0 || args.To == 0 {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	qb := m.db.Query(&StoredFact{})
	qb.Where(StoredFact_.TenantId).Eq(args.TenantId)
	qb.Where(StoredFact_.DueAt).Gte(args.From)
	qb.Where(StoredFact_.DueAt).Lte(args.To)

	if args.Scope != "" {
		qb.Where(StoredFact_.Scope).Eq(args.Scope)
	}

	facts, err := ReadAllStoredFact(qb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	var forecast []ForecastLine
	if m.deps.Forecast != nil {
		forecast, err = m.deps.Forecast.PendingAccruals(args.TenantId, args.From, args.To)
		if err != nil {
			ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
			return
		}
	}
	if forecast == nil {
		forecast = []ForecastLine{}
	}

	_ = ctx.Encode(&ThirtyDayResponse{
		Facts:    facts,
		Forecast: forecast,
	})
}
