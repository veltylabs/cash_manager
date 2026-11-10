//go:build wasm

package tests

import (
	"github.com/veltylabs/cash_manager/seed"
	"testing"
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
