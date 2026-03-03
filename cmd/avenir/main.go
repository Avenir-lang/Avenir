package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"avenir/internal/ir"
	"avenir/internal/modules"
	"avenir/internal/runtime"
	"avenir/internal/types"
	"avenir/internal/vm"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "run":
		if err := cmdRun(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "build":
		if err := cmdBuild(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	case "version", "-v", "--version":
		fmt.Println("avenir", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Avenir language CLI

Usage:
  avenir run <file.av|file.avc>
  avenir build <file.av> [-o out.avc] [-target=bytecode|native]

Commands:
  version  Avenir Language version
  run      Compile+run .av source or run .avc bytecode
  build    Compile .av source into .avc file

Flags (build):
  -o       Output file name (default: <input>.avc)
  -target  Build target: "bytecode" (default) or "native" (native not implemented yet)`)
}

// -------------- RUN --------------

func cmdRun(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("run: missing input file")
	}
	input := args[0]
	ext := filepath.Ext(input)

	switch ext {
	case ".av":
		// компиляция в памяти → VM
		mod, err := compileSourceFile(input)
		if err != nil {
			return err
		}
		env := runtime.DefaultEnv()
		absInput, err := filepath.Abs(input)
		if err == nil {
			env.SetExecRoot(filepath.Dir(absInput))
		}
		m := vm.NewVM(mod, env)
		_, err = m.RunMain()
		return err
	case ".avc":
		// запуск байткода
		mod, err := ir.ReadModuleFromFile(input)
		if err != nil {
			return fmt.Errorf("failed to read bytecode: %w", err)
		}
		env := runtime.DefaultEnv()
		absInput, err := filepath.Abs(input)
		if err == nil {
			env.SetExecRoot(filepath.Dir(absInput))
		}
		m := vm.NewVM(mod, env)
		_, err = m.RunMain()
		return err
	default:
		return fmt.Errorf("run: unsupported file extension %q (use .av or .avc)", ext)
	}
}

// -------------- BUILD --------------

func cmdBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var out string
	var target string

	fs.StringVar(&out, "o", "", "output file (default: <input>.avc)")
	fs.StringVar(&target, "target", "bytecode", "build target: bytecode|native")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("build: missing input file")
	}
	input := fs.Arg(0)

	if target == "native" {
		return fmt.Errorf("native target not implemented")
	}
	if target != "bytecode" {
		return fmt.Errorf("unknown target %q (supported: bytecode, native)", target)
	}

	if filepath.Ext(input) != ".av" {
		return fmt.Errorf("build: input must be .av source file")
	}

	if out == "" {
		base := input[:len(input)-len(filepath.Ext(input))]
		out = base + ".avc"
	}

	mod, err := compileSourceFile(input)
	if err != nil {
		return err
	}

	if err := ir.WriteModuleToFile(out, mod); err != nil {
		return fmt.Errorf("failed to write bytecode: %w", err)
	}

	return nil
}

// -------------- Unified compilation pipeline: .av -> *ir.Module --------------

// compileSourceFile compiles a source file using the unified module-based pipeline.
// Single-file programs are handled as a trivial world with one module.
func compileSourceFile(path string) (*ir.Module, error) {
	// Load world (handles both single-file and multi-module cases)
	world, errs := modules.LoadWorld(path)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s\n", e)
		}
		return nil, fmt.Errorf("module loading failed with %d errors", len(errs))
	}

	// Find entry module
	entryModName := world.Entry
	if _, ok := world.Modules[entryModName]; !ok {
		return nil, fmt.Errorf("entry module %q not found in world", entryModName)
	}

	// Build types.World from modules.World
	typeWorld := &types.World{
		Modules: make(map[string]*types.ModuleInfo),
		Entry:   entryModName,
	}
	for modName, modAST := range world.Modules {
		typeWorld.Modules[modName] = &types.ModuleInfo{
			Name:  modName,
			Prog:  modAST.Prog,
			Scope: nil, // Will be set by CheckWorldWithBindings
		}
	}

	// Type-check with bindings
	bindings, typeErrs := types.CheckWorldWithBindings(typeWorld)
	if len(typeErrs) > 0 {
		for _, e := range typeErrs {
			fmt.Fprintf(os.Stderr, "%s\n", e)
		}
		return nil, fmt.Errorf("type checking failed with %d errors", len(typeErrs))
	}

	// Compile using unified pipeline
	entryModInfo := typeWorld.Modules[entryModName]
	mod, errs := ir.CompileWorld(typeWorld, entryModInfo, bindings)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s\n", e)
		}
		return nil, fmt.Errorf("compilation failed with %d errors", len(errs))
	}

	return mod, nil
}
