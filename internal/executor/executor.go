package executor

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
	// TODO: Implement script execution
	return nil
}
