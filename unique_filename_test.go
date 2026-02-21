package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateUniqueFilename_NoCollision(t *testing.T) {
	// Test with a non-existing file
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "output.md")

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}

	if result != baseName {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, baseName)
	}
}

func TestGenerateUniqueFilename_SingleCollision(t *testing.T) {
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "output.md")

	// Create the base file to cause collision
	if err := os.WriteFile(baseName, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}
	expected := filepath.Join(tmpDir, "output-1.md")

	if result != expected {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, expected)
	}
}

func TestGenerateUniqueFilename_MultipleCollisions(t *testing.T) {
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "output.md")

	// Create base file and first two numbered versions
	if err := os.WriteFile(baseName, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "output-1.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "output-2.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}
	expected := filepath.Join(tmpDir, "output-3.md")

	if result != expected {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, expected)
	}
}

func TestGenerateUniqueFilename_HTMLExtension(t *testing.T) {
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "page.html")

	// Create the base file
	if err := os.WriteFile(baseName, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}
	expected := filepath.Join(tmpDir, "page-1.html")

	if result != expected {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, expected)
	}
}

func TestGenerateUniqueFilename_NoExtension(t *testing.T) {
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "output")

	// Create the base file
	if err := os.WriteFile(baseName, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}
	expected := filepath.Join(tmpDir, "output-1")

	if result != expected {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, expected)
	}
}

func TestGenerateUniqueFilename_CleanedSuffix(t *testing.T) {
	tmpDir := t.TempDir()
	baseName := filepath.Join(tmpDir, "document-cleaned.md")

	// Create the base file
	if err := os.WriteFile(baseName, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := GenerateUniqueFilename(baseName)
	if err != nil {
		t.Fatalf("GenerateUniqueFilename() error = %v", err)
	}
	expected := filepath.Join(tmpDir, "document-cleaned-1.md")

	if result != expected {
		t.Errorf("GenerateUniqueFilename() = %v, want %v", result, expected)
	}
}
