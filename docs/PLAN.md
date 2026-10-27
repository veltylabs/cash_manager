---
PLAN: "feat: export ui.Screens, introduce NewView and resolve langc reuse warnings"
TAG: v0.2.7
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.

## Tools

```
go install webtyp.com/devflow/cmd/gotest@latest
go install webtyp.com/lang/cmd/langc@latest
go install webtyp.com/ormc/cmd/ormc@latest
```

## Why

Phase M of webtyp app composition (`webtyp/docs/APP_COMPOSITION_MASTER_PLAN.md`): modules declare their screens via `Screens(caller router.Caller) []view.Screen` instead of legacy `Browser(...) (view.Presenter, error)`. Each leaf screen declares required `ID`, `Label`, `Icon`, `Permission`, and a non-nil `Presenter` with `view.WithDescription(...)`.
Additionally, `langc sync .` detected a reuse warning on `closing_cash_counted` and `closing_cash_expected` sharing `closing cash`; distinct labels must be assigned in `model.go`, `view.Presenter` introduced via `view.go`, and Spanish translations completed in `lang.json`.

## APIs used

| symbol | exact signature | tag | file |
|---|---|---|---|
| `view.Screen` | `type Screen struct { ID string; Label lang.Text; Icon string; Permission model.Resource; Presenter Presenter; Lookups []Lookup; Help lang.Text; Children []Screen }` | `webtyp.com/view@v0.6.45` | `view/screen.go` |
| `view.WithDescription` | `func WithDescription(d lang.Text) Option` | `webtyp.com/view@v0.6.45` | `view/options.go` |
| `view.WithTitle` | `func WithTitle(t string) Option` | `webtyp.com/view@v0.6.45` | `view/options.go` |
| `view.New` | `func New(source Lister, target model.Model, opts ...Option) Presenter` | `webtyp.com/view@v0.6.45` | `view/view.go` |
| `view.Item` | `type Item struct { ID, Label, Description, Sublabel, LeadTop, LeadMain, LeadBottom string }` | `webtyp.com/view@v0.6.45` | `view/item.go` |
| `view.Itemizer` | `type Itemizer interface { Item() Item }` | `webtyp.com/view@v0.6.45` | `view/item.go` |
| `router.Caller` | `type Caller interface { Call(op string, args, out any, done func(error)) }` | `webtyp.com/router@v0.4.6` | `router/caller.go` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Stage 1: Explicit labels in model.go and ormc generation

In `model.go`:
1. Define constant `ResourceCashManager`:
```go
const ResourceCashManager = "cashmanager"
```
2. In `ShiftModel`, add explicit distinct labels to avoid the two-word share warning:
```go
		{Name: "closing_cash_counted", Type: model.Int(), Label: "Counted cash"},
		{Name: "closing_cash_expected", Type: model.Int(), Label: "Expected cash"},
```
3. Run `ormc` to update `model_orm.go`.

Files: `model.go`, `model_orm.go`

## Stage 2: View presenter and default list in ops.go

1. Create `view.go` at module root:
```go
package cashmanager

import (
	"webtyp.com/fmt"
	"webtyp.com/lang"
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
	l.caller.Call(ModelName+"."+OpListStoredFacts, nil, out, func(err error) {
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
```

2. In `ops.go`, update `listStoredFacts` to support listing when `From`/`To` are zero (default to past 30 days) and when `Scope` is empty (query all facts for the tenant):
```go
func (m *Module) listStoredFacts(ctx router.Context) {
	var args ListStoredFactsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	if args.TenantId == "" {
		ctx.WriteStatus(400); ctx.Write([]byte("bad request"))
		return
	}

	now := time.NowUnix()
	if args.To == 0 {
		args.To = now
	}
	if args.From == 0 {
		args.From = now - 30*time.Day
	}

	qb := m.db.Query(&StoredFact{})
	qb.Where(StoredFact_.TenantId).Eq(args.TenantId)
	if args.Scope != "" {
		qb.Where(StoredFact_.Scope).Eq(args.Scope)
	}
	qb.Where(StoredFact_.At).Gte(int64(args.From))
	qb.Where(StoredFact_.At).Lte(int64(args.To))

	facts, err := ReadAllStoredFact(qb)
	if err != nil {
		ctx.WriteStatus(500); ctx.Write([]byte("internal error"))
		return
	}

	_ = ctx.Encode(&facts)
}
```

Files: `view.go`, `ops.go`

## Stage 3: Declarative Screens in ui/browser.go

Replace `ui/browser.go` with:
```go
//go:build browser

package ui

import (
	cashmanager "github.com/veltylabs/cash_manager"
	"webtyp.com/lang"
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
			Permission: cashmanager.ResourceCashManager,
			Presenter:  cashmanager.NewView(caller, view.WithDescription(descCash)),
		},
	}
}
```

Files: `ui/browser.go`

## Stage 4: Test and Translations

1. Add `tests/ui_view_browser_test.go`:
```go
package tests

import (
	"testing"

	"github.com/veltylabs/cash_manager/ui"
	"webtyp.com/router/mock"
)

func TestUIScreens(t *testing.T) {
	caller := mock.NewCaller()
	screens := ui.Screens(caller)
	if len(screens) == 0 {
		t.Fatal("expected at least 1 screen")
	}
	for _, s := range screens {
		if err := s.Validate(); err != nil {
			t.Fatalf("screen %s invalid: %v", s.ID, err)
		}
		if s.Presenter == nil {
			t.Fatalf("screen %s expected Presenter", s.ID)
		}
		if s.Presenter.Description() == "" {
			t.Fatalf("screen %s expected non-empty description", s.ID)
		}
	}
}
```

2. Run `langc sync .`.
3. Complete 100% of Spanish translations `(es)` in `lang.json`.
4. Ensure `langc sync .` reports:
   - 0 untranslated `(es)`
   - 0 reuse warnings
   - 0 identical values

Files: `tests/ui_view_browser_test.go`, `lang.json`

## Delivery check

Run:
```
gotest
langc sync .
```
Both commands must exit 0 with 0 errors, 0 untranslated `(es)`, and 0 reuse warnings.
Commit after every stage. When all checks pass, push and open the PR.

## Stages

| # | Stage | Files |
|---|---|---|
| 1 | Model labels and ormc | `model.go`, `model_orm.go` |
| 2 | View presenter and list defaults | `view.go`, `ops.go` |
| 3 | Declarative screens | `ui/browser.go` |
| 4 | Test and translations | `tests/ui_view_browser_test.go`, `lang.json` |
