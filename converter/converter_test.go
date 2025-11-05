// converter/converter_test.go
package converter

import (
	"testing"
)

// TestConvertWorkflow tests the ConvertWorkflow function
func TestConvertWorkflow(t *testing.T) {
	// Simple GitHub workflow YAML for testing
	yamlData := []byte(`
name: Test Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        run: echo "Checkout code"
      - name: Build
        run: echo "Building application"
  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Test
        run: echo "Running tests"
  deploy:
    needs: [build, test]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    container:
      image: alpine:latest
    steps:
      - name: Deploy
        run: echo "Deploying application"
`)

	result, err := ConvertWorkflow(yamlData)
	if err != nil {
		t.Errorf("ConvertWorkflow() error = %v", err)
		return
	}

	if result == "" {
		t.Error("ConvertWorkflow() returned empty result")
		return
	}

	// Check that the result contains expected Argo Workflow elements
	if !contains(result, "apiVersion") {
		t.Error("ConvertWorkflow() result missing apiVersion")
	}

	if !contains(result, "kind") {
		t.Error("ConvertWorkflow() result missing kind")
	}

	if !contains(result, "Workflow") {
		t.Error("ConvertWorkflow() result missing Workflow kind")
	}

	// Check that templates are created
	if !contains(result, "templates") {
		t.Error("ConvertWorkflow() result missing templates")
	}

	// Check that DAG tasks are created
	if !contains(result, "dag") {
		t.Error("ConvertWorkflow() result missing dag")
	}

	t.Logf("Conversion successful. Result length: %d characters", len(result))
}

// TestConvertWorkflowWithEmptyJobs tests conversion with empty jobs
func TestConvertWorkflowWithEmptyJobs(t *testing.T) {
	yamlData := []byte(`
name: Empty Workflow
on: [push]
jobs: {}
`)

	result, err := ConvertWorkflow(yamlData)
	if err != nil {
		t.Errorf("ConvertWorkflow() error = %v", err)
		return
	}

	if result == "" {
		t.Error("ConvertWorkflow() returned empty result")
		return
	}

	// Should still have basic structure
	if !contains(result, "apiVersion") {
		t.Error("ConvertWorkflow() result missing apiVersion")
	}

	if !contains(result, "kind") {
		t.Error("ConvertWorkflow() result missing kind")
	}

	t.Logf("Conversion successful. Result length: %d characters", len(result))
}

// TestConvertWorkflowWithSingleJob tests conversion with a single job
func TestConvertWorkflowWithSingleJob(t *testing.T) {
	yamlData := []byte(`
name: Single Job Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        run: echo "Checkout code"
      - name: Build
        run: echo "Building application"
`)

	result, err := ConvertWorkflow(yamlData)
	if err != nil {
		t.Errorf("ConvertWorkflow() error = %v", err)
		return
	}

	if result == "" {
		t.Error("ConvertWorkflow() returned empty result")
		return
	}

	// Check that the result contains expected Argo Workflow elements
	if !contains(result, "apiVersion") {
		t.Error("ConvertWorkflow() result missing apiVersion")
	}

	if !contains(result, "kind") {
		t.Error("ConvertWorkflow() result missing kind")
	}

	t.Logf("Conversion successful. Result length: %d characters", len(result))
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(s) == len(substr) && s == substr ||
			len(s) > len(substr) && (s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsHelper(s, substr)))
}

// Helper function for substring search
func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
