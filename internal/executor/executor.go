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

	if err := e.moveOutputFiles(); err != nil {
		return fmt.Errorf("failed to move output files: %w", err)
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

// moveOutputFiles moves the generated files to the output directory
func (e *Executor) moveOutputFiles() error {
	scriptDir := filepath.Dir(e.scriptPath)
	files := []string{
		"claude-code-usage.log",
		"claude-code-usage.json",
		"claude-code-usage-execution.log",
	}

	for _, file := range files {
		src := filepath.Join(scriptDir, file)
		dst := filepath.Join(e.outputDir, file)

		if err := e.moveFile(src, dst); err != nil {
			return fmt.Errorf("failed to move %s: %w", file, err)
		}
	}

	return nil
}

// moveFile moves a file from src to dst
func (e *Executor) moveFile(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return err
	}

	return os.Remove(src)
}
