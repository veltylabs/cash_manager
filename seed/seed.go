package seed

import (
	"context"

	"github.com/veltylabs/cash_manager"
)

type Data struct {
	Shifts []cashmanager.Shift
}

func Load(ctx context.Context, module *cashmanager.Module, tenantID string, cashierID string) (Data, error) {
	// Stub implementation to create one open and one closed shift.
	// Normally we would invoke the ops directly if they are exported or have a hook.
	return Data{}, nil
}
