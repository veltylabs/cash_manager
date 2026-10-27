//go:build browser

package tests

import (
	"testing"

	cashmanager "github.com/veltylabs/cash_manager"
	"github.com/veltylabs/cash_manager/ui"
	"webtyp.com/model"
	"webtyp.com/router/mock"
	"webtyp.com/view"
)

func TestUIScreens(t *testing.T) {
	caller := &mock.Caller{}
	wrappedCaller := view.NewRecords().Caller(caller)
	screens := ui.Screens(wrappedCaller)
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

func TestRowOf(t *testing.T) {
	fact := cashmanager.StoredFact{
		Kind:       "charge",
		Amount:     1000,
		Currency:   "EUR",
		Scope:      "shift-123",
		SourceType: "order",
		SourceId:   "ord-456",
	}

	row := model.RowOf(&fact)
	if row.Title != "charge" {
		t.Fatalf("expected Title 'charge', got %q", row.Title)
	}

	expectedDetail := "EUR · 1000 · shift-123 · order · ord-456"
	if row.Detail != expectedDetail {
		t.Fatalf("expected Detail %q, got %q", expectedDetail, row.Detail)
	}
}
