package provider

import "sync"

type CommitPromptOptions struct {
	Generate int
	Language string
	Emoji    bool
}

var (
	runtimeOptionsMu sync.RWMutex
	runtimeOptions   CommitPromptOptions
)

func SetRuntimeCommitPromptOptions(opts CommitPromptOptions) {
	runtimeOptionsMu.Lock()
	defer runtimeOptionsMu.Unlock()
	runtimeOptions = opts
}

func GetRuntimeCommitPromptOptions() CommitPromptOptions {
	runtimeOptionsMu.RLock()
	defer runtimeOptionsMu.RUnlock()
	return runtimeOptions
}

func ResetRuntimeCommitPromptOptions() {
	SetRuntimeCommitPromptOptions(CommitPromptOptions{})
}
