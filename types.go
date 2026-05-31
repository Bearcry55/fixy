package main

// Result holds a single online search result from any source
type Result struct {
	Source string
	Title  string
	Link   string
	Info   string
}

// FileMatch holds a flagged line from the scanned file
type FileMatch struct {
	LineNumber int
	Line       string
	Context    []string
	FuncName   string
	Reason     string
}

// RiskPattern describes a dangerous code pattern the scanner detects
type RiskPattern struct {
	Name  string
	Check func(line, lower string) bool
}