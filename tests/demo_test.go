//go:build wasm

package tests

import (
	"testing"
	"github.com/veltylabs/cash_manager/seed"
)

func TestDemo_Valid(t *testing.T) {
	demo, err := seed.Demo()
	if err != nil {
		t.Fatalf("Demo() err = %v", err)
	}
	if demo.Label == "" {
		t.Errorf("Demo() returned empty label")
	}
}
