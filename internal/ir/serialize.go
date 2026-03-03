package ir

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var magicV1 = [4]byte{'A', 'V', 'C', '1'}
var magicV2 = [4]byte{'A', 'V', 'C', '2'}

func WriteModuleToFile(filename string, m *Module) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteModule(f, m)
}

func ReadModuleFromFile(filename string) (*Module, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadModule(f)
}

func WriteModule(w io.Writer, m *Module) error {
	// magic
	if _, err := w.Write(magicV2[:]); err != nil {
		return err
	}

	// num functions
	if err := binary.Write(w, binary.LittleEndian, uint32(len(m.Functions))); err != nil {
		return err
	}

	for _, fn := range m.Functions {
		nameBytes := []byte(fn.Name)
		if len(nameBytes) > 0xFFFF {
			return fmt.Errorf("function name too long: %s", fn.Name)
		}
		// name len (uint16) + name bytes
		if err := binary.Write(w, binary.LittleEndian, uint16(len(nameBytes))); err != nil {
			return err
		}
		if _, err := w.Write(nameBytes); err != nil {
			return err
		}

		// num params, num locals
		if err := binary.Write(w, binary.LittleEndian, uint32(fn.NumParams)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(fn.Chunk.NumLocals)); err != nil {
			return err
		}

		// consts
		if err := binary.Write(w, binary.LittleEndian, uint32(len(fn.Chunk.Consts))); err != nil {
			return err
		}
		for _, c := range fn.Chunk.Consts {
			if err := binary.Write(w, binary.LittleEndian, uint8(c.Kind)); err != nil {
				return err
			}
			switch c.Kind {
			case ConstInt:
				if err := binary.Write(w, binary.LittleEndian, c.Int); err != nil {
					return err
				}
			case ConstBool:
				var b byte
				if c.Bool {
					b = 1
				}
				if err := binary.Write(w, binary.LittleEndian, b); err != nil {
					return err
				}
			case ConstString:
				bs := []byte(c.String)
				if err := binary.Write(w, binary.LittleEndian, uint32(len(bs))); err != nil {
					return err
				}
				if _, err := w.Write(bs); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown const kind %d", c.Kind)
			}
		}

		// code
		if err := binary.Write(w, binary.LittleEndian, uint32(len(fn.Chunk.Code))); err != nil {
			return err
		}
		for _, inst := range fn.Chunk.Code {
			if err := binary.Write(w, binary.LittleEndian, uint8(inst.Op)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, int32(inst.A)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, int32(inst.B)); err != nil {
				return err
			}
		}
	}

	// struct types
	if err := binary.Write(w, binary.LittleEndian, uint32(len(m.StructTypes))); err != nil {
		return err
	}
	for _, st := range m.StructTypes {
		nameBytes := []byte(st.Name)
		if len(nameBytes) > 0xFFFF {
			return fmt.Errorf("struct name too long: %s", st.Name)
		}
		if err := binary.Write(w, binary.LittleEndian, uint16(len(nameBytes))); err != nil {
			return err
		}
		if _, err := w.Write(nameBytes); err != nil {
			return err
		}
	}

	// main index
	if err := binary.Write(w, binary.LittleEndian, int32(m.MainIndex)); err != nil {
		return err
	}

	return nil
}

func ReadModule(r io.Reader) (*Module, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	if hdr != magicV1 && hdr != magicV2 {
		return nil, fmt.Errorf("invalid magic header: %q", string(hdr[:]))
	}

	var numFuncs uint32
	if err := binary.Read(r, binary.LittleEndian, &numFuncs); err != nil {
		return nil, err
	}

	mod := &Module{
		Functions: make([]*Function, 0, numFuncs),
		MainIndex: -1,
	}

	for i := uint32(0); i < numFuncs; i++ {
		var nameLen uint16
		if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
			return nil, err
		}

		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(r, nameBytes); err != nil {
			return nil, err
		}
		name := string(nameBytes)

		var numParamsU32 uint32
		if err := binary.Read(r, binary.LittleEndian, &numParamsU32); err != nil {
			return nil, err
		}
		var numLocalsU32 uint32
		if err := binary.Read(r, binary.LittleEndian, &numLocalsU32); err != nil {
			return nil, err
		}

		fn := &Function{
			Name:      name,
			NumParams: int(numParamsU32),
			Chunk: Chunk{
				NumLocals: int(numLocalsU32),
			},
		}

		// consts
		var numConsts uint32
		if err := binary.Read(r, binary.LittleEndian, &numConsts); err != nil {
			return nil, err
		}
		fn.Chunk.Consts = make([]Constant, numConsts)
		for ci := uint32(0); ci < numConsts; ci++ {
			var kindU8 uint8
			if err := binary.Read(r, binary.LittleEndian, &kindU8); err != nil {
				return nil, err
			}
			c := Constant{Kind: ConstKind(kindU8)}
			switch c.Kind {
			case ConstInt:
				if err := binary.Read(r, binary.LittleEndian, &c.Int); err != nil {
					return nil, err
				}
			case ConstBool:
				var b byte
				if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
					return nil, err
				}
				c.Bool = b != 0
			case ConstString:
				var slen uint32
				if err := binary.Read(r, binary.LittleEndian, &slen); err != nil {
					return nil, err
				}
				sb := make([]byte, slen)
				if _, err := io.ReadFull(r, sb); err != nil {
					return nil, err
				}
				c.String = string(sb)
			default:
				return nil, fmt.Errorf("unknown const kind %d", c.Kind)
			}
			fn.Chunk.Consts[ci] = c
		}

		// code
		var numInstr uint32
		if err := binary.Read(r, binary.LittleEndian, &numInstr); err != nil {
			return nil, err
		}
		fn.Chunk.Code = make([]Instruction, numInstr)
		for pi := uint32(0); pi < numInstr; pi++ {
			var opU8 uint8
			if err := binary.Read(r, binary.LittleEndian, &opU8); err != nil {
				return nil, err
			}
			var a, b int32
			if err := binary.Read(r, binary.LittleEndian, &a); err != nil {
				return nil, err
			}
			if err := binary.Read(r, binary.LittleEndian, &b); err != nil {
				return nil, err
			}
			fn.Chunk.Code[pi] = Instruction{
				Op: OpCode(opU8),
				A:  int(a),
				B:  int(b),
			}
		}

		mod.Functions = append(mod.Functions, fn)
	}

	if hdr == magicV2 {
		var numStructs uint32
		if err := binary.Read(r, binary.LittleEndian, &numStructs); err != nil {
			return nil, err
		}
		mod.StructTypes = make([]StructTypeInfo, numStructs)
		for i := uint32(0); i < numStructs; i++ {
			var nameLen uint16
			if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
				return nil, err
			}
			nameBytes := make([]byte, nameLen)
			if _, err := io.ReadFull(r, nameBytes); err != nil {
				return nil, err
			}
			mod.StructTypes[i] = StructTypeInfo{Name: string(nameBytes)}
		}
	}

	var mainIdx int32
	if err := binary.Read(r, binary.LittleEndian, &mainIdx); err != nil {
		return nil, err
	}
	mod.MainIndex = int(mainIdx)

	return mod, nil
}
