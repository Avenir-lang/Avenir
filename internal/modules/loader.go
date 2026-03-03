package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"avenir/internal/ast"
	"avenir/internal/lexer"
	"avenir/internal/parser"
)

// ModuleAST represents a parsed module.
type ModuleAST struct {
	Name     string       // Fully-qualified module name, e.g. "std.io"
	FilePath string       // Path to the .av file
	Prog     *ast.Program // Parsed AST
}

// World represents all loaded modules.
type World struct {
	Modules map[string]*ModuleAST // by FQN
	Entry   string                // Entry module name
}

// LoadWorld loads the entry file and all its dependencies recursively.
func LoadWorld(entryFile string) (*World, []error) {
	w := &World{
		Modules: make(map[string]*ModuleAST),
	}

	// Determine project root (directory of entry file)
	entryAbs, err := filepath.Abs(entryFile)
	if err != nil {
		return nil, []error{fmt.Errorf("cannot resolve entry file path: %v", err)}
	}
	projectRoot := filepath.Dir(entryAbs)

	// Track visited modules for cycle detection
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var errors []error

	// Load entry module
	entryMod, errs := loadModule(entryFile, projectRoot, visited, visiting, w)
	if len(errs) > 0 {
		errors = append(errors, errs...)
	}
	if entryMod == nil {
		return nil, errors
	}

	w.Entry = entryMod.Name
	return w, errors
}

// loadModule loads a single module and recursively loads its dependencies.
func loadModule(filePath string, projectRoot string, visited map[string]bool, visiting map[string]bool, world *World) (*ModuleAST, []error) {
	moduleFiles, err := moduleFilesForEntry(filePath)
	if err != nil {
		return nil, []error{err}
	}

	var parseErrors []error
	var programs []*ast.Program
	var moduleName string

	for _, path := range moduleFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("cannot read file %s: %v", path, err))
			continue
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		prog := p.ParseProgram()

		if errs := p.Errors(); len(errs) > 0 {
			for _, e := range errs {
				parseErrors = append(parseErrors, fmt.Errorf("%s: %s", path, e))
			}
			continue
		}

		if prog.Package == nil {
			parseErrors = append(parseErrors, fmt.Errorf("%s: missing package declaration", path))
			continue
		}

		if moduleName == "" {
			moduleName = prog.Package.Name
		} else if prog.Package.Name != moduleName {
			parseErrors = append(parseErrors, fmt.Errorf("%s: package %q does not match module %q", path, prog.Package.Name, moduleName))
			continue
		}

		if err := validateFileStructMapping(path, prog); err != nil {
			parseErrors = append(parseErrors, err)
			continue
		}

		programs = append(programs, prog)
	}

	if len(parseErrors) > 0 {
		return nil, parseErrors
	}

	if moduleName == "" {
		return nil, []error{fmt.Errorf("cannot determine module name for %s", filePath)}
	}

	// Check for cycles
	if visiting[moduleName] {
		return nil, []error{fmt.Errorf("import cycle detected involving module %q", moduleName)}
	}
	if visited[moduleName] {
		// Already loaded, return existing
		return world.Modules[moduleName], nil
	}

	visiting[moduleName] = true
	defer func() {
		visiting[moduleName] = false
		visited[moduleName] = true
	}()

	merged := mergePrograms(programs)

	mod := &ModuleAST{
		Name:     moduleName,
		FilePath: filePath,
		Prog:     merged,
	}
	world.Modules[moduleName] = mod

	// Load all imports recursively
	var allErrors []error
	for _, imp := range merged.Imports {
		// Build FQN from import path
		importFQN := strings.Join(imp.Path, ".")

		// Find the file for this import
		importFile, err := findModuleFile(importFQN, projectRoot)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("%s:%d:%d: %v", filePath, imp.ImportPos.Line, imp.ImportPos.Column, err))
			continue
		}

		// Recursively load the imported module
		_, errs := loadModule(importFile, projectRoot, visited, visiting, world)
		if len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}

	if len(allErrors) > 0 {
		return mod, allErrors
	}

	return mod, nil
}

// findModuleFile locates the .av file for a given module FQN.
// Supports both flat files (module.av) and folder-based imports (module/module.av).
func findModuleFile(moduleFQN string, projectRoot string) (string, error) {
	// Check if it's a std module
	if strings.HasPrefix(moduleFQN, "std.") {
		// std.io -> try std/io/io.av (folder) then std/io.av (flat)
		path := strings.TrimPrefix(moduleFQN, "std.")
		parts := strings.Split(path, ".")
		lastPart := parts[len(parts)-1]

		// Try folder-based: std/io/io.av
		folderPath := filepath.Join(projectRoot, "std", filepath.Join(parts...))
		folderFile := filepath.Join(folderPath, lastPart+".av")
		if _, err := os.Stat(folderFile); err == nil {
			return folderFile, nil
		}

		// Try flat file: std/io.av
		flatFile := filepath.Join(projectRoot, "std", path+".av")
		if _, err := os.Stat(flatFile); err == nil {
			return flatFile, nil
		}

		// Fallback: try repo root std/
		repoRoot := findRepoRoot(projectRoot)
		if repoRoot != "" {
			// Try folder-based in repo root
			folderFile = filepath.Join(repoRoot, "std", filepath.Join(parts...), lastPart+".av")
			if _, err := os.Stat(folderFile); err == nil {
				return folderFile, nil
			}
			// Try flat file in repo root
			flatFile = filepath.Join(repoRoot, "std", path+".av")
			if _, err := os.Stat(flatFile); err == nil {
				return flatFile, nil
			}
		}
		return "", fmt.Errorf("cannot find module %q (looked for folder %s/%s.av and file %s)", moduleFQN, folderPath, lastPart, flatFile)
	}

	// Non-std module: app.utils -> try app/utils/utils.av (folder) then app/utils.av (flat)
	parts := strings.Split(moduleFQN, ".")
	lastPart := parts[len(parts)-1]

	// Try folder-based: app/utils/utils.av
	folderPath := filepath.Join(projectRoot, filepath.Join(parts...))
	folderFile := filepath.Join(folderPath, lastPart+".av")
	if _, err := os.Stat(folderFile); err == nil {
		return folderFile, nil
	}

	// Try flat file: app/utils.av
	flatFile := filepath.Join(projectRoot, filepath.Join(parts...)+".av")
	if _, err := os.Stat(flatFile); err == nil {
		return flatFile, nil
	}

	// Check if folder exists but file is missing
	if info, err := os.Stat(folderPath); err == nil && info.IsDir() {
		return "", fmt.Errorf("folder %q exists but does not contain required file %q. Expected: %s", folderPath, lastPart+".av", folderFile)
	}

	return "", fmt.Errorf("cannot find module %q (looked for folder %s/%s.av and file %s)", moduleFQN, folderPath, lastPart, flatFile)
}

// findRepoRoot tries to find the repository root by looking for go.mod or .git
func findRepoRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func moduleFilesForEntry(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ".av")
	if filepath.Base(dir) != base {
		return []string{filePath}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read module directory %s: %v", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".av") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func mergePrograms(programs []*ast.Program) *ast.Program {
	if len(programs) == 0 {
		return &ast.Program{}
	}
	merged := &ast.Program{
		Package: programs[0].Package,
	}
	for _, prog := range programs {
		merged.Imports = append(merged.Imports, prog.Imports...)
		merged.Funcs = append(merged.Funcs, prog.Funcs...)
		merged.Structs = append(merged.Structs, prog.Structs...)
		merged.Interfaces = append(merged.Interfaces, prog.Interfaces...)
	}
	return merged
}

func validateFileStructMapping(filePath string, prog *ast.Program) error {
	fileName := filepath.Base(filePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, ".av")

	var foundMatchingStruct bool
	var structNames []string
	for _, st := range prog.Structs {
		structNames = append(structNames, st.Name)
		if st.Name == fileNameWithoutExt {
			foundMatchingStruct = true
		}
	}

	if len(prog.Structs) > 0 && !foundMatchingStruct {
		return fmt.Errorf("%s: file %q does not contain struct %q (found structs: %s). A file can only be imported if it contains a struct with the same name as the file",
			filePath, fileName, fileNameWithoutExt, strings.Join(structNames, ", "))
	}
	return nil
}
