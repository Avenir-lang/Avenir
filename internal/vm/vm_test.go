package vm

import (
	"testing"

	"avenir/internal/ir"
	"avenir/internal/runtime"
	"avenir/internal/value"
)

// Simple test: 1 + 2 = 3 using "manual" bytecode.
func TestVM_SimpleAdd(t *testing.T) {
	var chunk ir.Chunk
	i1 := chunk.AddConstInt(1)
	i2 := chunk.AddConstInt(2)

	chunk.Emit(ir.OpConst, i1, 0) // push 1
	chunk.Emit(ir.OpConst, i2, 0) // push 2
	chunk.Emit(ir.OpAdd, 0, 0)    // 1 + 2
	chunk.Emit(ir.OpReturn, 0, 1) // return top

	fn := &ir.Function{
		Name:      "main",
		NumParams: 0,
		Chunk:     chunk,
	}
	mod := &ir.Module{
		Functions: []*ir.Function{fn},
		MainIndex: 0,
	}

	m := NewVM(mod, runtime.DefaultEnv())
	v, err := m.RunMain()
	if err != nil {
		t.Fatalf("RunMain error: %v", err)
	}

	if v.Kind != value.KindInt || v.Int != 3 {
		t.Fatalf("expected 3, got %v (%s)", v.Int, v.String())
	}
}
