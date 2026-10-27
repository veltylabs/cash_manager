//go:build browser

package tests

import (
	"testing"

	"github.com/veltylabs/cash_manager/ui"
	"webtyp.com/router/mock"
)

func TestUIScreens(t *testing.T) {
	caller := &mock.Caller{}
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
