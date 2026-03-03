package vm

import (
	"errors"
	"fmt"
	"strings"

	"avenir/internal/ir"
	"avenir/internal/runtime"
	"avenir/internal/runtime/builtins"
	"avenir/internal/value"
)

// Frame represents a function call frame.
type Frame struct {
	Clo  *value.Closure // Current closure
	Fn   *ir.Function
	IP   int // Instruction pointer: index into Fn.Chunk.Code
	Base int // Stack index where local variables start
}

type exceptionHandler struct {
	FrameIndex int // index in vm.frames this handler belongs to
	TargetIP   int // IP to jump to when exception is handled
	StackSP    int // stack height to restore before jump
}

// VM is a stack-based virtual machine for Avenir.
type VM struct {
	mod    *ir.Module
	stack  []value.Value
	sp     int // Stack pointer: next free index
	frames []Frame

	env      *runtime.Env
	handlers []exceptionHandler
}

func (vm *VM) throwValue(exc value.Value) bool {
	if exc.Kind != value.KindError {
		exc = value.ErrorValue(fmt.Sprintf("thrown non-error: %s", exc.String()))
	}

	for len(vm.handlers) > 0 {
		h := vm.handlers[len(vm.handlers)-1]
		vm.handlers = vm.handlers[:len(vm.handlers)-1]

		if h.FrameIndex >= len(vm.frames) {
			continue
		}

		vm.frames = vm.frames[:h.FrameIndex+1]
		vm.sp = h.StackSP
		vm.push(exc)
		vm.frames[h.FrameIndex].IP = h.TargetIP
		return true
	}

	return false
}

func (vm *VM) raiseError(err error) bool {
	if err == nil {
		return false
	}
	return vm.throwValue(value.ErrorValue(err.Error()))
}

func errorMessage(val value.Value) string {
	if val.Kind != value.KindError {
		return val.String()
	}
	if val.Error != nil && val.Error.Message != "" {
		return val.Error.Message
	}
	return val.Str
}

// NewVM creates a VM for the given module.
func NewVM(m *ir.Module, env *runtime.Env) *VM {
	if env == nil {
		env = runtime.DefaultEnv()
	}
	if m != nil && len(m.StructTypes) > 0 {
		names := make([]string, len(m.StructTypes))
		for i, st := range m.StructTypes {
			names[i] = st.Name
		}
		env.SetStructTypeNames(names)
	}
	vm := &VM{
		mod:      m,
		stack:    make([]value.Value, 0, 1024),
		frames:   make([]Frame, 0, 16),
		env:      env,
		handlers: make([]exceptionHandler, 0, 16),
	}
	// Enable builtins to call closures by setting the closure caller
	// We need to wrap callClosure to match the ClosureCaller signature
	env.SetClosureCaller(func(clo *value.Closure, args []value.Value) (value.Value, error) {
		// Push arguments onto the stack
		for _, arg := range args {
			vm.push(arg)
		}
		// Call the closure with the number of arguments
		return vm.callClosure(clo, len(args))
	})
	return vm
}

// push/pop

func (vm *VM) push(v value.Value) {
	if vm.sp >= len(vm.stack) {
		vm.stack = append(vm.stack, v)
	} else {
		vm.stack[vm.sp] = v
	}
	vm.sp++
}

func (vm *VM) pop() (value.Value, error) {
	if vm.sp == 0 {
		return value.Value{}, errors.New("stack underflow")
	}
	vm.sp--
	return vm.stack[vm.sp], nil
}

func (vm *VM) peek(offset int) (value.Value, error) {
	idx := vm.sp - 1 - offset
	if idx < 0 || idx >= vm.sp {
		return value.Value{}, errors.New("stack underflow")
	}
	return vm.stack[idx], nil
}

// RunMain runs the main function of the module.
func (vm *VM) RunMain() (value.Value, error) {
	if vm.mod.MainIndex < 0 || vm.mod.MainIndex >= len(vm.mod.Functions) {
		return value.Value{}, fmt.Errorf("invalid main index %d", vm.mod.MainIndex)
	}
	fn := vm.mod.Functions[vm.mod.MainIndex]
	cloVal := value.NewClosure(fn, nil)
	return vm.callClosure(cloVal.Closure, 0)
}

// callClosure calls a closure with the given number of arguments.
//
// Stack layout for a function call:
//   - Arguments are already on the stack (pushed by caller)
//   - base = sp - numArgs (start of this function's stack frame)
//   - We allocate space for all locals (NumLocals >= NumParams)
//   - Parameters occupy the first NumParams slots
//   - Additional locals (declared variables) occupy slots [NumParams, NumLocals)
//
// Upvalues:
//   - Each closure carries its upvalues array
//   - Upvalues can be "open" (pointing to stack slots) or "closed" (holding copied values)
//   - When a function returns, open upvalues pointing into its frame are closed
func (vm *VM) callClosure(clo *value.Closure, numArgs int) (value.Value, error) {
	if clo == nil {
		return value.Value{}, fmt.Errorf("callClosure: nil closure")
	}
	fn := clo.Fn

	if numArgs != fn.NumParams {
		return value.Value{}, fmt.Errorf("function %s expects %d args, got %d",
			fn.Name, fn.NumParams, numArgs)
	}

	// Arguments are already on the stack
	// base = sp - numArgs (start of this function's stack frame)
	base := vm.sp - numArgs
	if base < 0 {
		return value.Value{}, errors.New("stack underflow before call")
	}

	// Allocate space for all local variables (NumLocals >= NumParams)
	// Parameters occupy the first NumParams slots, additional locals follow
	if fn.Chunk.NumLocals > fn.NumParams {
		additional := fn.Chunk.NumLocals - fn.NumParams
		for i := 0; i < additional; i++ {
			vm.push(value.Value{}) // zero / invalid
		}
	}

	frame := Frame{
		Clo:  clo,
		Fn:   fn,
		IP:   0,
		Base: base,
	}
	vm.frames = append(vm.frames, frame)
	callFrameIdx := len(vm.frames) - 1

	var lastRet value.Value
	var skipIncrement bool
	for {
		if len(vm.frames) == 0 {
			return lastRet, nil
		}
		frameIdx := len(vm.frames) - 1
		fr := &vm.frames[frameIdx]

		if fr.IP < 0 || fr.IP >= len(fr.Fn.Chunk.Code) {
			if vm.raiseError(fmt.Errorf("instruction pointer out of range in %s: %d", fr.Fn.Name, fr.IP)) {
				skipIncrement = true
				continue
			}
			return value.Value{}, fmt.Errorf("instruction pointer out of range in %s: %d", fr.Fn.Name, fr.IP)
		}

		inst := fr.Fn.Chunk.Code[fr.IP]
		shouldIncrementIP := !skipIncrement
		skipIncrement = false

		switch inst.Op {
		case ir.OpHalt:
			vm.frames = vm.frames[:0]
			return lastRet, nil

		case ir.OpConst:
			constIdx := inst.A
			if constIdx < 0 || constIdx >= len(fr.Fn.Chunk.Consts) {
				if vm.raiseError(fmt.Errorf("const index out of range: %d", constIdx)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("const index out of range: %d", constIdx)
			}
			c := fr.Fn.Chunk.Consts[constIdx]
			switch c.Kind {
			case ir.ConstInt:
				vm.push(value.Int(c.Int))
			case ir.ConstFloat:
				vm.push(value.Float(c.Float))
			case ir.ConstString:
				vm.push(value.Str(c.String))
			case ir.ConstBool:
				vm.push(value.Bool(c.Bool))
			case ir.ConstBytes:
				vm.push(value.Bytes(c.Bytes))
			case ir.ConstNone:
				vm.push(value.None())
			default:
				if vm.raiseError(fmt.Errorf("unsupported const kind %d", c.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("unsupported const kind %d", c.Kind)
			}

		case ir.OpLoadLocal:
			slot := fr.Base + inst.A
			if slot < 0 || slot >= vm.sp {
				if vm.raiseError(fmt.Errorf("OpLoadLocal: invalid slot %d", slot)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpLoadLocal: invalid slot %d", slot)
			}
			vm.push(vm.stack[slot])

		case ir.OpStoreLocal:
			slot := fr.Base + inst.A
			if slot < 0 || slot >= vm.sp {
				if vm.raiseError(fmt.Errorf("OpStoreLocal: invalid slot %d", slot)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpStoreLocal: invalid slot %d", slot)
			}
			v, err := vm.peek(0)
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			vm.stack[slot] = v

		case ir.OpPop:
			if _, err := vm.pop(); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}

		// Arithmetic operations
		case ir.OpAdd:
			if err := vm.binaryNumericOp(func(a, b float64) float64 { return a + b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpSub:
			if err := vm.binaryNumericOp(func(a, b float64) float64 { return a - b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpMul:
			if err := vm.binaryNumericOp(func(a, b float64) float64 { return a * b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpDiv:
			b, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			a, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if (b.Kind == value.KindInt && b.Int == 0) || (b.Kind == value.KindFloat && b.Float == 0) {
				if vm.raiseError(fmt.Errorf("division by zero")) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("division by zero")
			}
			vm.push(a)
			vm.push(b)
			if err := vm.binaryNumericOp(func(x, y float64) float64 { return x / y }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpMod:
			// Modulo only works on integers
			b, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			a, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if a.Kind != value.KindInt || b.Kind != value.KindInt {
				if vm.raiseError(fmt.Errorf("binary int op expects (int, int), got (%v, %v)", a.Kind, b.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("binary int op expects (int, int), got (%v, %v)", a.Kind, b.Kind)
			}
			if b.Int == 0 {
				if vm.raiseError(fmt.Errorf("modulo by zero")) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("modulo by zero")
			}
			vm.push(value.Int(a.Int % b.Int))
		case ir.OpNegate:
			v, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if v.Kind == value.KindInt {
				vm.push(value.Int(-v.Int))
			} else if v.Kind == value.KindFloat {
				vm.push(value.Float(-v.Float))
			} else {
				if vm.raiseError(fmt.Errorf("OpNegate: expected int or float, got %v", v.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpNegate: expected int or float, got %v", v.Kind)
			}

		// Comparisons / logic
		case ir.OpLt:
			if err := vm.binaryNumericCmp(func(a, b float64) bool { return a < b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpLte:
			if err := vm.binaryNumericCmp(func(a, b float64) bool { return a <= b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpGt:
			if err := vm.binaryNumericCmp(func(a, b float64) bool { return a > b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpGte:
			if err := vm.binaryNumericCmp(func(a, b float64) bool { return a >= b }); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpEq:
			if err := vm.binaryEq(); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
		case ir.OpNeq:
			if err := vm.binaryNeq(); err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}

		// Control flow
		case ir.OpJump:
			fr.IP = inst.A
			shouldIncrementIP = false

		case ir.OpJumpIfFalse:
			cond, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if cond.Kind != value.KindBool {
				if vm.raiseError(fmt.Errorf("OpJumpIfFalse: expected bool, got %v", cond.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpJumpIfFalse: expected bool, got %v", cond.Kind)
			}
			if !cond.Bool {
				fr.IP = inst.A
				shouldIncrementIP = false
			}

		// Function calls
		case ir.OpCall:
			// Direct call by function index - create closure with no upvalues
			fn := vm.mod.Functions[inst.A]
			cloVal := value.NewClosure(fn, nil)
			retVal, err := vm.callClosure(cloVal.Closure, inst.B)
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			lastRet = retVal

		case ir.OpCallValue:
			callee, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if callee.Kind != value.KindClosure {
				if vm.raiseError(fmt.Errorf("OpCallValue: expected closure, got %v", callee.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpCallValue: expected closure, got %v", callee.Kind)
			}
			retVal, err := vm.callClosure(callee.Closure, inst.A)
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			lastRet = retVal

		case ir.OpCallBuiltin:
			builtinID := builtins.ID(inst.A)
			n := inst.B
			if n < 0 || n > vm.sp {
				if vm.raiseError(fmt.Errorf("OpCallBuiltin: invalid arg count %d", n)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpCallBuiltin: invalid arg count %d", n)
			}
			args := make([]value.Value, n)
			for i := n - 1; i >= 0; i-- {
				v, err := vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						continue
					}
					return value.Value{}, err
				}
				args[i] = v
			}
			res, hasRes, err := runtime.CallBuiltin(vm.env, builtinID, args)
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if hasRes {
				vm.push(res)
			}

		case ir.OpClosure:
			// Create a closure: A = function index, B = number of upvalues
			//
			// Upvalue handling:
			//   - If IsLocal is true: the upvalue captures a local from the current frame.
			//     We create an "open" upvalue pointing directly to the stack slot.
			//   - If IsLocal is false: the upvalue captures a parent function's upvalue.
			//     The compiler has already pushed the value via OpLoadUpvalue.
			//     We try to share the parent's upvalue object (for nested closures).
			//
			// Open upvalues are closed when the function that owns the stack slot returns.
			fn := vm.mod.Functions[inst.A]
			numUpvalues := inst.B
			currentFrame := vm.frames[len(vm.frames)-1]

			// Build upvalues array
			upvalues := make([]*value.Upvalue, numUpvalues)
			for i := 0; i < numUpvalues; i++ {
				upvalueInfo := fn.Upvalues[i]
				if upvalueInfo.IsLocal {
					// Capture from current frame's locals
					// upvalueInfo.Index is the index in the parent function's locals array
					// Stack slot = base + index
					slotIndex := currentFrame.Base + upvalueInfo.Index
					if slotIndex < 0 || slotIndex >= vm.sp {
						if vm.raiseError(fmt.Errorf("OpClosure: invalid local slot %d (base=%d, index=%d, sp=%d)",
							slotIndex, currentFrame.Base, upvalueInfo.Index, vm.sp)) {
							skipIncrement = true
							shouldIncrementIP = false
							goto nextInstruction
						}
						return value.Value{}, fmt.Errorf("OpClosure: invalid local slot %d (base=%d, index=%d, sp=%d)",
							slotIndex, currentFrame.Base, upvalueInfo.Index, vm.sp)
					}

					// Create open upvalue pointing to stack slot
					// Multiple closures can point to the same slot (they share the variable)
					upvalues[i] = &value.Upvalue{
						IsClosed: false,
						Index:    slotIndex,
					}
				} else {
					// This references a parent function's upvalue
					// The compiler has already pushed the value via OpLoadUpvalue
					val, err := vm.pop()
					if err != nil {
						if vm.raiseError(err) {
							skipIncrement = true
							shouldIncrementIP = false
							goto nextInstruction
						}
						return value.Value{}, err
					}

					// Try to find and share the parent's upvalue
					// This is important for nested closures to share the same upvalue object
					parentClo := currentFrame.Clo
					if parentClo != nil && upvalueInfo.Index < len(parentClo.Upvalues) {
						// Share the parent's upvalue (allows multiple closures to share state)
						upvalues[i] = parentClo.Upvalues[upvalueInfo.Index]
					} else {
						// Fallback: create closed upvalue with the current value
						upvalues[i] = &value.Upvalue{
							IsClosed: true,
							Closed:   val,
						}
					}
				}
			}

			clo := value.NewClosure(fn, upvalues)
			vm.push(clo)

		case ir.OpLoadUpvalue:
			// Load upvalue: A = upvalue index
			idx := inst.A
			if fr.Clo == nil || idx >= len(fr.Clo.Upvalues) {
				if vm.raiseError(fmt.Errorf("OpLoadUpvalue: invalid upvalue index %d", idx)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpLoadUpvalue: invalid upvalue index %d", idx)
			}
			upv := fr.Clo.Upvalues[idx]
			if upv.IsClosed {
				vm.push(upv.Closed)
			} else {
				// Open upvalue: read from stack
				if upv.Index < 0 || upv.Index >= vm.sp {
					if vm.raiseError(fmt.Errorf("OpLoadUpvalue: invalid stack index %d", upv.Index)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpLoadUpvalue: invalid stack index %d", upv.Index)
				}
				vm.push(vm.stack[upv.Index])
			}

		case ir.OpStoreUpvalue:
			// Store to upvalue: A = upvalue index
			idx := inst.A
			v, err := vm.peek(0)
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if fr.Clo == nil || idx >= len(fr.Clo.Upvalues) {
				if vm.raiseError(fmt.Errorf("OpStoreUpvalue: invalid upvalue index %d", idx)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpStoreUpvalue: invalid upvalue index %d", idx)
			}
			upv := fr.Clo.Upvalues[idx]
			if upv.IsClosed {
				upv.Closed = v
			} else {
				// Open upvalue: write to stack
				if upv.Index < 0 || upv.Index >= vm.sp {
					if vm.raiseError(fmt.Errorf("OpStoreUpvalue: invalid stack index %d", upv.Index)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpStoreUpvalue: invalid stack index %d", upv.Index)
				}
				vm.stack[upv.Index] = v
			}

		case ir.OpReturn:
			currentFrameIdx := len(vm.frames) - 1
			f := vm.frames[currentFrameIdx]

			// Close all open upvalues that point into this frame's stack
			// Do this BEFORE we pop the return value, so we can read the values
			vm.closeUpvalues(f.Base)

			var ret value.Value
			var err error
			if inst.B == 1 {
				ret, err = vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						continue
					}
					return value.Value{}, err
				}
			} else {
				// For void, return "unit"/invalid so that call ALWAYS has one result
				ret = value.Value{}
			}

			vm.frames = vm.frames[:len(vm.frames)-1]

			if f.Base < 0 || f.Base > vm.sp {
				if vm.raiseError(fmt.Errorf("OpReturn: invalid base %d", f.Base)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpReturn: invalid base %d", f.Base)
			}
			vm.sp = f.Base

			// Always push one result onto the stack (even for void)
			vm.push(ret)
			lastRet = ret

			if len(vm.frames) == 0 {
				return lastRet, nil
			}
			// If we returned from the frame we're calling, return from callClosure
			if currentFrameIdx == callFrameIdx {
				// The frame we were calling has been popped, return the value
				return ret, nil
			}
			// Continue with next frame - restart loop to get new fr pointer
			continue

		case ir.OpMakeList:
			n := inst.A
			if n < 0 || n > vm.sp {
				if vm.raiseError(fmt.Errorf("OpMakeList: invalid count %d", n)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpMakeList: invalid count %d", n)
			}
			list := make([]value.Value, n)
			for i := n - 1; i >= 0; i-- {
				v, err := vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						shouldIncrementIP = false
						goto nextInstruction
					}
					return value.Value{}, err
				}
				list[i] = v
			}
			vm.push(value.List(list))

		case ir.OpMakeDict:
			n := inst.A
			if n < 0 || n*2 > vm.sp {
				if vm.raiseError(fmt.Errorf("OpMakeDict: invalid count %d", n)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpMakeDict: invalid count %d", n)
			}
			entries := make([]struct {
				key   string
				value value.Value
			}, n)
			for i := n - 1; i >= 0; i-- {
				val, err := vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						shouldIncrementIP = false
						goto nextInstruction
					}
					return value.Value{}, err
				}
				keyVal, err := vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						shouldIncrementIP = false
						goto nextInstruction
					}
					return value.Value{}, err
				}
				if keyVal.Kind != value.KindString {
					if vm.raiseError(fmt.Errorf("OpMakeDict: expected string key, got %v", keyVal.Kind)) {
						skipIncrement = true
						shouldIncrementIP = false
						goto nextInstruction
					}
					return value.Value{}, fmt.Errorf("OpMakeDict: expected string key, got %v", keyVal.Kind)
				}
				entries[i] = struct {
					key   string
					value value.Value
				}{key: keyVal.Str, value: val}
			}
			dict := make(map[string]value.Value, n)
			for _, entry := range entries {
				dict[entry.key] = entry.value
			}
			vm.push(value.Dict(dict))

		case ir.OpIndex:
			idxVal, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			listVal, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			switch listVal.Kind {
			case value.KindList:
				if idxVal.Kind != value.KindInt {
					if vm.raiseError(fmt.Errorf("OpIndex: expected int index, got %v", idxVal.Kind)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: expected int index, got %v", idxVal.Kind)
				}
				idx := idxVal.Int
				if idx < 0 || int(idx) >= len(listVal.List) {
					if vm.raiseError(fmt.Errorf("OpIndex: index out of range %d (len=%d)", idx, len(listVal.List))) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: index out of range %d (len=%d)", idx, len(listVal.List))
				}
				vm.push(listVal.List[idx])
			case value.KindBytes:
				if idxVal.Kind != value.KindInt {
					if vm.raiseError(fmt.Errorf("OpIndex: expected int index, got %v", idxVal.Kind)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: expected int index, got %v", idxVal.Kind)
				}
				idx := idxVal.Int
				if idx < 0 || int(idx) >= len(listVal.Bytes) {
					if vm.raiseError(fmt.Errorf("OpIndex: index out of range %d (len=%d)", idx, len(listVal.Bytes))) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: index out of range %d (len=%d)", idx, len(listVal.Bytes))
				}
				vm.push(value.Int(int64(listVal.Bytes[idx])))
			case value.KindDict:
				if idxVal.Kind != value.KindString {
					if vm.raiseError(fmt.Errorf("OpIndex: expected string key, got %v", idxVal.Kind)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: expected string key, got %v", idxVal.Kind)
				}
				val, ok := listVal.Dict[idxVal.Str]
				if !ok {
					if vm.raiseError(fmt.Errorf("OpIndex: key %q not found", idxVal.Str)) {
						skipIncrement = true
						continue
					}
					return value.Value{}, fmt.Errorf("OpIndex: key %q not found", idxVal.Str)
				}
				vm.push(val)
			default:
				if vm.raiseError(fmt.Errorf("OpIndex: expected list, bytes, or dict, got %v", listVal.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpIndex: expected list, bytes, or dict, got %v", listVal.Kind)
			}

		case ir.OpMakeSome:
			v, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			vm.push(value.Some(v))

		case ir.OpMakeStruct:
			// A = struct type index, B = number of fields
			structTypeIdx := inst.A
			fieldCount := inst.B

			// Pop field values from stack (in reverse order)
			fields := make([]value.Value, fieldCount)
			for i := fieldCount - 1; i >= 0; i-- {
				v, err := vm.pop()
				if err != nil {
					if vm.raiseError(err) {
						skipIncrement = true
						shouldIncrementIP = false
						goto nextInstruction
					}
					return value.Value{}, err
				}
				fields[i] = v
			}

			vm.push(value.Struct(structTypeIdx, fields))

		case ir.OpLoadField:
			// A = field index
			fieldIdx := inst.A

			// Pop struct from stack
			structVal, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}

			if structVal.Kind != value.KindStruct {
				if vm.raiseError(fmt.Errorf("OpLoadField: expected struct, got %v", structVal.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpLoadField: expected struct, got %v", structVal.Kind)
			}

			if structVal.Struct == nil {
				if vm.raiseError(fmt.Errorf("OpLoadField: nil struct")) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpLoadField: nil struct")
			}

			if fieldIdx < 0 || fieldIdx >= len(structVal.Struct.Fields) {
				if vm.raiseError(fmt.Errorf("OpLoadField: field index %d out of range (struct has %d fields)", fieldIdx, len(structVal.Struct.Fields))) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpLoadField: field index %d out of range (struct has %d fields)", fieldIdx, len(structVal.Struct.Fields))
			}

			vm.push(structVal.Struct.Fields[fieldIdx])

		case ir.OpStoreField:
			// A = field index
			// In-place mutation: mutate the struct field directly (no copying)
			fieldIdx := inst.A

			// Pop value to store
			fieldVal, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}

			// Pop struct from stack
			structVal, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}

			if structVal.Kind != value.KindStruct {
				if vm.raiseError(fmt.Errorf("OpStoreField: expected struct, got %v", structVal.Kind)) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpStoreField: expected struct, got %v", structVal.Kind)
			}

			if structVal.Struct == nil {
				if vm.raiseError(fmt.Errorf("OpStoreField: nil struct")) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpStoreField: nil struct")
			}

			if fieldIdx < 0 || fieldIdx >= len(structVal.Struct.Fields) {
				if vm.raiseError(fmt.Errorf("OpStoreField: field index %d out of range (struct has %d fields)", fieldIdx, len(structVal.Struct.Fields))) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpStoreField: field index %d out of range (struct has %d fields)", fieldIdx, len(structVal.Struct.Fields))
			}

			// Mutate the struct in-place (no copying)
			structVal.Struct.Fields[fieldIdx] = fieldVal

			// Push the mutated struct back (same struct value, mutated in-place)
			vm.push(structVal)

		case ir.OpStringify:
			val, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			vm.push(value.Str(val.String()))

		case ir.OpConcatString:
			right, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			left, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if left.Kind != value.KindString || right.Kind != value.KindString {
				if vm.raiseError(fmt.Errorf("OpConcatString expects strings")) {
					skipIncrement = true
					continue
				}
				return value.Value{}, fmt.Errorf("OpConcatString expects strings")
			}
			var builder strings.Builder
			builder.Grow(len(left.Str) + len(right.Str))
			builder.WriteString(left.Str)
			builder.WriteString(right.Str)
			vm.push(value.Str(builder.String()))

		case ir.OpBeginTry:
			vm.handlers = append(vm.handlers, exceptionHandler{
				FrameIndex: len(vm.frames) - 1,
				TargetIP:   inst.A,
				StackSP:    vm.sp,
			})

		case ir.OpEndTry:
			// Pop handlers belonging to the current frame
			for len(vm.handlers) > 0 {
				h := vm.handlers[len(vm.handlers)-1]
				if h.FrameIndex != len(vm.frames)-1 {
					break
				}
				vm.handlers = vm.handlers[:len(vm.handlers)-1]
			}

		case ir.OpThrow:
			exc, err := vm.pop()
			if err != nil {
				if vm.raiseError(err) {
					skipIncrement = true
					continue
				}
				return value.Value{}, err
			}
			if vm.throwValue(exc) {
				skipIncrement = true
				continue
			}
			return value.Value{}, fmt.Errorf("unhandled error: %s", errorMessage(exc))

		default:
			if vm.raiseError(fmt.Errorf("unknown opcode %d", inst.Op)) {
				skipIncrement = true
				continue
			}
			return value.Value{}, fmt.Errorf("unknown opcode %d", inst.Op)
		}

	nextInstruction:
		if shouldIncrementIP {
			fr.IP++
		}
	}
}

// callFunction calls a function by index in the module (legacy wrapper).
func (vm *VM) callFunction(fnIndex int, numArgs int) (value.Value, error) {
	fn := vm.mod.Functions[fnIndex]
	cloVal := value.NewClosure(fn, nil)
	return vm.callClosure(cloVal.Closure, numArgs)
}

// ---- Helpers for binary operations ----

func (vm *VM) binaryIntOp(op func(a, b int64) int64) error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}
	if a.Kind != value.KindInt || b.Kind != value.KindInt {
		return fmt.Errorf("binary int op expects (int, int), got (%v, %v)", a.Kind, b.Kind)
	}
	vm.push(value.Int(op(a.Int, b.Int)))
	return nil
}

// binaryNumericOp handles arithmetic operations for int and float, promoting to float if needed
func (vm *VM) binaryNumericOp(op func(a, b float64) float64) error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}

	// Convert to float64
	var aFloat, bFloat float64
	if a.Kind == value.KindInt {
		aFloat = float64(a.Int)
	} else if a.Kind == value.KindFloat {
		aFloat = a.Float
	} else {
		return fmt.Errorf("binary numeric op expects numeric types, got (%v, %v)", a.Kind, b.Kind)
	}

	if b.Kind == value.KindInt {
		bFloat = float64(b.Int)
	} else if b.Kind == value.KindFloat {
		bFloat = b.Float
	} else {
		return fmt.Errorf("binary numeric op expects numeric types, got (%v, %v)", a.Kind, b.Kind)
	}

	result := op(aFloat, bFloat)
	// If both operands were int and result is whole number, return int; otherwise float
	if a.Kind == value.KindInt && b.Kind == value.KindInt {
		if result == float64(int64(result)) {
			vm.push(value.Int(int64(result)))
		} else {
			vm.push(value.Float(result))
		}
	} else {
		vm.push(value.Float(result))
	}
	return nil
}

func (vm *VM) binaryIntCmp(op func(a, b int64) bool) error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}
	if a.Kind != value.KindInt || b.Kind != value.KindInt {
		return fmt.Errorf("binary int cmp expects (int, int), got (%v, %v)", a.Kind, b.Kind)
	}
	vm.push(value.Bool(op(a.Int, b.Int)))
	return nil
}

// binaryNumericCmp handles comparison operations for int and float
func (vm *VM) binaryNumericCmp(op func(a, b float64) bool) error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}

	// Convert to float64
	var aFloat, bFloat float64
	if a.Kind == value.KindInt {
		aFloat = float64(a.Int)
	} else if a.Kind == value.KindFloat {
		aFloat = a.Float
	} else {
		return fmt.Errorf("binary numeric cmp expects numeric types, got (%v, %v)", a.Kind, b.Kind)
	}

	if b.Kind == value.KindInt {
		bFloat = float64(b.Int)
	} else if b.Kind == value.KindFloat {
		bFloat = b.Float
	} else {
		return fmt.Errorf("binary numeric cmp expects numeric types, got (%v, %v)", a.Kind, b.Kind)
	}

	vm.push(value.Bool(op(aFloat, bFloat)))
	return nil
}

// Deep comparison of values (including lists and functions)
func equalValues(a, b value.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case value.KindInt:
		return a.Int == b.Int
	case value.KindFloat:
		return a.Float == b.Float
	case value.KindString:
		return a.Str == b.Str
	case value.KindBool:
		return a.Bool == b.Bool
	case value.KindError:
		return a.Str == b.Str
	case value.KindBytes:
		if len(a.Bytes) != len(b.Bytes) {
			return false
		}
		for i := range a.Bytes {
			if a.Bytes[i] != b.Bytes[i] {
				return false
			}
		}
		return true
	case value.KindList:
		if len(a.List) != len(b.List) {
			return false
		}
		for i := range a.List {
			if !equalValues(a.List[i], b.List[i]) {
				return false
			}
		}
		return true
	case value.KindDict:
		if len(a.Dict) != len(b.Dict) {
			return false
		}
		for k, av := range a.Dict {
			bv, ok := b.Dict[k]
			if !ok {
				return false
			}
			if !equalValues(av, bv) {
				return false
			}
		}
		return true
	case value.KindClosure:
		// Closures are equal if they reference the same function and have same upvalues
		// For simplicity, we compare function pointers
		if a.Closure == nil || b.Closure == nil {
			return a.Closure == b.Closure
		}
		return a.Closure.Fn == b.Closure.Fn
	default:
		return false
	}
}

func (vm *VM) binaryEq() error {
	b, err := vm.pop()
	if err != nil {
		return err
	}
	a, err := vm.pop()
	if err != nil {
		return err
	}
	vm.push(value.Bool(equalValues(a, b)))
	return nil
}

func (vm *VM) binaryNeq() error {
	if err := vm.binaryEq(); err != nil {
		return err
	}
	v, err := vm.pop()
	if err != nil {
		return err
	}
	if v.Kind != value.KindBool {
		return fmt.Errorf("binaryNeq: internal error, expected bool")
	}
	vm.push(value.Bool(!v.Bool))
	return nil
}

// closeUpvalues closes all open upvalues that point to stack slots >= base.
//
// When a function returns, we need to "close" any open upvalues that point
// into its stack frame. This copies the value from the stack into the upvalue
// object, so closures can still access the variable after the function returns.
//
// We scan:
//  1. All frames and their closures (for closures created during execution)
//  2. All closures on the stack (for closures passed as arguments or stored)
func (vm *VM) closeUpvalues(base int) {
	// Scan all frames and their closures for open upvalues pointing to this frame
	for i := len(vm.frames) - 1; i >= 0; i-- {
		frame := vm.frames[i]
		if frame.Clo != nil {
			for _, upv := range frame.Clo.Upvalues {
				if !upv.IsClosed && upv.Index >= base {
					// Close this upvalue: copy value from stack
					if upv.Index < vm.sp {
						upv.Closed = vm.stack[upv.Index]
						upv.IsClosed = true
					}
				}
			}
		}
	}

	// Also check any closures on the stack (they might have been passed as arguments)
	for i := 0; i < vm.sp; i++ {
		if vm.stack[i].Kind == value.KindClosure && vm.stack[i].Closure != nil {
			for _, upv := range vm.stack[i].Closure.Upvalues {
				if !upv.IsClosed && upv.Index >= base {
					if upv.Index < vm.sp {
						upv.Closed = vm.stack[upv.Index]
						upv.IsClosed = true
					}
				}
			}
		}
	}
}
