package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Executor handles the execution of the claude-code-usage.sh script
type Executor struct {
	scriptPath string
	outputDir  string
}

// New creates a new Executor instance
func New(scriptPath, outputDir string) *Executor {
	return &Executor{
		scriptPath: scriptPath,
		outputDir:  outputDir,
	}
}

// Execute runs the script and handles output
func (e *Executor) Execute() error {
	if err := e.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := e.runScript(); err != nil {
		return fmt.Errorf("failed to run script: %w", err)
	}

	return nil
}

// ensureOutputDir creates the output directory if it doesn't exist
func (e *Executor) ensureOutputDir() error {
	return os.MkdirAll(e.outputDir, 0755)
}

// runScript executes the claude-code-usage.sh script
func (e *Executor) runScript() error {
	cmd := exec.Command("/bin/bash", e.scriptPath)
	cmd.Dir = filepath.Dir(e.scriptPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
