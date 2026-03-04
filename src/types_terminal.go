package plugin

type TerminalHistoryEntry struct {
	Command   string
	Output    string
	ExitCode  string
	Timestamp string
}

type TerminalViewModel struct {
	CommandInput string
	LastOutput   string
	History      []TerminalHistoryEntry
}
