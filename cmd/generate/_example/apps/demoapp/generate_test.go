package demoapp

import (
	"testing"

	"github.com/morehao/gocli/cmd/generate"
)

func TestGenerateModelCode(t *testing.T) {
	_, err := generate.ExecuteCommand(generate.Cmd, "--mode", "model")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
}

func TestGenerateModuleCode(t *testing.T) {
	_, err := generate.ExecuteCommand(generate.Cmd, "--mode", "module")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
}

func TestGenerateApiCode(t *testing.T) {
	_, err := generate.ExecuteCommand(generate.Cmd, "--mode", "api")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
}
