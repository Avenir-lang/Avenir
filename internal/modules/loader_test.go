package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadWorld_FileToStructMapping_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create geometry/Point.av with struct Point (matching file name)
	geometryDir := filepath.Join(tmpDir, "geometry")
	if err := os.MkdirAll(geometryDir, 0755); err != nil {
		t.Fatalf("failed to create geometry dir: %v", err)
	}
	
	pointFile := filepath.Join(geometryDir, "Point.av")
	pointContent := `pckg geometry.Point;

pub struct Point {
    x | int
    y | int
}
`
	if err := os.WriteFile(pointFile, []byte(pointContent), 0644); err != nil {
		t.Fatalf("failed to write Point.av: %v", err)
	}

	// Create main.av that imports Point using full path
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import geometry.Point;

fun main() | void {
    var p | Point = Point{x = 1, y = 2};
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load should succeed - file Point.av contains struct Point (matches)
	world, errs := LoadWorld(mainFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	if world == nil {
		t.Fatalf("expected world to be loaded")
	}
}

func TestLoadWorld_StructNameMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Point.av with NON-matching struct name
	pointFile := filepath.Join(tmpDir, "Point.av")
	pointContent := `pckg geometry;

pub struct Rectangle {
    x | int
    y | int
}
`
	if err := os.WriteFile(pointFile, []byte(pointContent), 0644); err != nil {
		t.Fatalf("failed to write Point.av: %v", err)
	}

	// Try to load - should fail with struct name mismatch
	_, errs := LoadWorld(pointFile)
	if len(errs) == 0 {
		t.Fatalf("expected error for struct name mismatch, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "does not contain struct") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about struct name mismatch, got: %v", errs)
	}
}

func TestLoadWorld_FolderImport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create geometry/geometry.av (folder A contains A.av)
	geometryDir := filepath.Join(tmpDir, "geometry")
	if err := os.MkdirAll(geometryDir, 0755); err != nil {
		t.Fatalf("failed to create geometry dir: %v", err)
	}

	geometryFile := filepath.Join(geometryDir, "geometry.av")
	geometryContent := `pckg geometry;

pub struct geometry {
    x | int
    y | int
}
`
	if err := os.WriteFile(geometryFile, []byte(geometryContent), 0644); err != nil {
		t.Fatalf("failed to write geometry.av: %v", err)
	}

	// Create main.av that imports using simplified syntax: import geometry
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import geometry;

fun main() | void {
    var g | geometry = geometry{x = 1, y = 2};
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load should succeed - import geometry resolves to geometry/geometry.av
	world, errs := LoadWorld(mainFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	if world == nil {
		t.Fatalf("expected world to be loaded")
	}

	// Verify geometry module was loaded
	if _, ok := world.Modules["geometry"]; !ok {
		t.Fatalf("expected geometry module to be loaded")
	}
}

func TestLoadWorld_MultiFileModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fs/fs.av and fs/file.av (same package)
	fsDir := filepath.Join(tmpDir, "fs")
	if err := os.MkdirAll(fsDir, 0755); err != nil {
		t.Fatalf("failed to create fs dir: %v", err)
	}

	fsFile := filepath.Join(fsDir, "fs.av")
	fsContent := `pckg fs;

struct fs {}
`
	if err := os.WriteFile(fsFile, []byte(fsContent), 0644); err != nil {
		t.Fatalf("failed to write fs.av: %v", err)
	}

	fileFile := filepath.Join(fsDir, "file.av")
	fileContent := `pckg fs;

struct file {}

pub struct File {
    handle | any
}
`
	if err := os.WriteFile(fileFile, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to write file.av: %v", err)
	}

	// Create main.av that imports fs
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import fs;

fun main() | void {
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	world, errs := LoadWorld(mainFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	mod, ok := world.Modules["fs"]
	if !ok {
		t.Fatalf("expected fs module to be loaded")
	}
	found := false
	for _, st := range mod.Prog.Structs {
		if st.Name == "File" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected File struct to be loaded from file.av")
	}
}

func TestLoadWorld_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main.av that imports non-existent module
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import nonexistent;

fun main() | void {
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load should fail with clear error
	_, errs := LoadWorld(mainFile)
	if len(errs) == 0 {
		t.Fatalf("expected error for missing module, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "cannot find module") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about missing module, got: %v", errs)
	}
}

func TestLoadWorld_FolderExistsButFileMissing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create geometry folder but no geometry.av file
	geometryDir := filepath.Join(tmpDir, "geometry")
	if err := os.MkdirAll(geometryDir, 0755); err != nil {
		t.Fatalf("failed to create geometry dir: %v", err)
	}

	// Create main.av that imports geometry
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import geometry;

fun main() | void {
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load should fail with clear error about missing file
	_, errs := LoadWorld(mainFile)
	if len(errs) == 0 {
		t.Fatalf("expected error for missing file in folder, got none")
	}

	hasError := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "does not contain required file") || strings.Contains(e.Error(), "cannot find module") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatalf("expected error about missing file, got: %v", errs)
	}
}

func TestLoadWorld_FileWithoutStructs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create utils.av with only functions (no structs)
	utilsFile := filepath.Join(tmpDir, "utils.av")
	utilsContent := `pckg utils;

pub fun helper() | void {
    print("help");
}
`
	if err := os.WriteFile(utilsFile, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("failed to write utils.av: %v", err)
	}

	// Files without structs should still be loadable (for function-only modules)
	// But according to the requirement, they cannot be imported
	// However, they can be the entry point
	world, errs := LoadWorld(utilsFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		// Entry file doesn't need structs, so this should work
		t.Fatalf("expected no errors for entry file without structs, got %d", len(errs))
	}

	if world == nil {
		t.Fatalf("expected world to be loaded")
	}
}

func TestLoadWorld_MultipleStructsOneMatches(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Point.av with multiple structs, one matches file name
	pointFile := filepath.Join(tmpDir, "Point.av")
	pointContent := `pckg geometry;

pub struct Point {
    x | int
    y | int
}

pub struct Rectangle {
    w | int
    h | int
}
`
	if err := os.WriteFile(pointFile, []byte(pointContent), 0644); err != nil {
		t.Fatalf("failed to write Point.av: %v", err)
	}

	// Load should succeed - Point.av contains struct Point (matches)
	world, errs := LoadWorld(pointFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	if world == nil {
		t.Fatalf("expected world to be loaded")
	}
}

func TestLoadWorld_NestedFolderImport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create app/utils/utils.av (nested folder)
	appDir := filepath.Join(tmpDir, "app")
	utilsDir := filepath.Join(appDir, "utils")
	if err := os.MkdirAll(utilsDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	utilsFile := filepath.Join(utilsDir, "utils.av")
	utilsContent := `pckg app.utils;

pub struct utils {
    data | string = ""
}
`
	if err := os.WriteFile(utilsFile, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("failed to write utils.av: %v", err)
	}

	// Create main.av that imports app.utils
	mainFile := filepath.Join(tmpDir, "main.av")
	mainContent := `pckg main;

import app.utils;

fun main() | void {
    var u | utils = utils{data = "test"};
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.av: %v", err)
	}

	// Load should succeed - import app.utils resolves to app/utils/utils.av
	world, errs := LoadWorld(mainFile)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s", e)
		}
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	if world == nil {
		t.Fatalf("expected world to be loaded")
	}

	// Verify app.utils module was loaded
	if _, ok := world.Modules["app.utils"]; !ok {
		t.Fatalf("expected app.utils module to be loaded")
	}
}
