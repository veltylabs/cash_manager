package tests

import (
	"strings"
	"testing"
	"github.com/veltylabs/cash_manager"
	"webtyp.com/model"
)

func contains(err error, sub string) bool {
	if err == nil { return false }
	return strings.Contains(err.Error(), sub)
}

func TestErrorsStatuses(t *testing.T) {
	m, _, caller := setup()
	_ = m

	errCh := make(chan error, 1)

	caller.Call("cashmanager.OpenShift", &cashmanager.OpenShiftArgs{}, nil, func(err error) {
		errCh <- err
	})

	err := <-errCh
	if model.KindOf(err) != model.ErrorInvalid && !contains(err, "invalid") {
		t.Errorf("Expected ErrorInvalid for bad request, got %v (err: %v)", model.KindOf(err), err)
	}

	caller.Call("cashmanager.CloseShift", &cashmanager.CloseShiftArgs{TenantId: "t1", Id: "nonexistent"}, nil, func(err error) {
		errCh <- err
	})

	err = <-errCh
	if model.KindOf(err) != model.ErrorNotFound && !contains(err, "not found") {
		t.Errorf("Expected ErrorNotFound, got %v (err: %v)", model.KindOf(err), err)
	}
}
