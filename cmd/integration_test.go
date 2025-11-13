package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

// TestRealWorldExample tests the conversion with actual Go code from the example directory
func TestRealWorldExample(t *testing.T) {
	// Get the workspace root - go up one directory from cmd/
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	workspaceRoot := filepath.Dir(wd)
	examplePath := filepath.Join(workspaceRoot, "example")

	// Check if example directory exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("Example directory not found, skipping real-world test")
		return
	}

	// Generate PlantUML from the example Go code using the same approach as main.go
	result, err := goplantuml.NewClassDiagramWithMaxDepth([]string{examplePath}, []string{}, true, 4)
	if err != nil {
		t.Fatalf("Failed to parse example Go code: %v", err)
	}

	// Set default rendering options for consistency
	renderingOptions := map[goplantuml.RenderingOption]any{
		goplantuml.RenderFields:  true,
		goplantuml.RenderMethods: true,
	}
	_ = result.SetRenderingOptions(renderingOptions)

	plantUML := result.Render()
	t.Logf("Generated PlantUML:\n%s", plantUML)

	// Convert to Mermaid
	mermaidOutput, err := ConvertToMermaid(plantUML)
	if err != nil {
		t.Fatalf("ConvertToMermaid() error = %v", err)
	}

	t.Logf("Generated Mermaid:\n%s", mermaidOutput)

	// Verify the output contains expected elements
	expectedElements := []string{
		"classDiagram",
		"class User {",
		"class UserService {",
		"class DatabaseUserService {",
		"<<struct>>",
		"<<interface>>",
		"--|>", // inheritance relationship
	}

	for _, element := range expectedElements {
		if !strings.Contains(mermaidOutput, element) {
			t.Errorf("Expected Mermaid output to contain '%s', but it didn't", element)
		}
	}

	// Verify no PlantUML-specific artifacts remain
	unwantedElements := []string{
		"@startuml",
		"@enduml",
		"namespace",
		"<<(S,Aquamarine)>>",
		"<<(I,",
		"<font color=",
	}

	for _, element := range unwantedElements {
		if strings.Contains(mermaidOutput, element) {
			t.Errorf("Mermaid output should not contain PlantUML artifact '%s', but it did", element)
		}
	}

	// Verify the structure is valid Mermaid syntax
	lines := strings.Split(mermaidOutput, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "classDiagram" {
		t.Error("Mermaid output should start with 'classDiagram'")
	}

	// Count classes and relationships
	classCount := 0
	relationshipCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "class ") && strings.HasSuffix(line, " {") {
			classCount++
		}
		if strings.Contains(line, "--|>") || strings.Contains(line, "*--") ||
			strings.Contains(line, "<--") || strings.Contains(line, " -- ") {
			relationshipCount++
		}
	}

	if classCount < 2 {
		t.Errorf("Expected at least 2 classes in the output, got %d", classCount)
	}

	if relationshipCount < 1 {
		t.Errorf("Expected at least 1 relationship in the output, got %d", relationshipCount)
	}
}
