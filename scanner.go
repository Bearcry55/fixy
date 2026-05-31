package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// riskPatterns detect dangerous code patterns directly — no keyword guessing
var riskPatterns = []RiskPattern{
	{
		Name: "index access (potential out of range)",
		Check: func(line, lower string) bool {
			if !strings.Contains(lower, "[") {
				return false
			}
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") ||
				strings.HasPrefix(trimmed, "import") ||
				strings.HasPrefix(trimmed, "var ") ||
				strings.HasPrefix(trimmed, "type ") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 {
						last := prev[len(prev)-1]
						if (last >= 'a' && last <= 'z') ||
							(last >= 'A' && last <= 'Z') ||
							(last >= '0' && last <= '9') ||
							last == ')' {
							return true
						}
					}
				}
			}
			return false
		},
	},
	{
	Name: "nil dereference risk",
Check: func(line, lower string) bool {
    trimmed := strings.TrimSpace(lower)
    if strings.HasPrefix(trimmed, "//") ||
        strings.HasPrefix(trimmed, "func ") {  
        return false
    }
    return (strings.Contains(line, ".") && strings.Contains(line, "(")) ||
        strings.Contains(line, "*")
},
	},
	{
		Name: "unchecked error (blank identifier)",
		Check: func(line, lower string) bool {
			return strings.Contains(line, ", _") || strings.Contains(line, ",_")
		},
	},
	{
		Name: "type assertion without ok check",
		Check: func(line, lower string) bool {
			return strings.Contains(line, ".(") && !strings.Contains(line, ", ok")
		},
	},
	{
		Name: "channel send/receive (possible deadlock)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			return strings.Contains(line, "<-")
		},
	},
	{
		Name: "division (possible divide by zero)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			return strings.Contains(line, "/") &&
				!strings.Contains(lower, "//") &&
				!strings.Contains(lower, "http")
		},
	},
}

func scanFile(fileName, errorMsg string) []FileMatch {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("%s  ⚠ Could not open file: %s%s\n", red, err.Error(), reset)
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	activePatterns := relevantPatterns(errorMsg)
	var matches []FileMatch
	currentFunc := ""

	for i, line := range lines {
		if fn := extractFuncName(line); fn != "" {
			currentFunc = fn
		}
		lower := strings.ToLower(line)

		for _, pattern := range activePatterns {
			if pattern.Check(line, lower) {
				ctx := []string{}
				for j := clampMin(0, i-2); j <= clampMax(len(lines)-1, i+2); j++ {
					if j != i {
						ctx = append(ctx, fmt.Sprintf("%d: %s", j+1, lines[j]))
					}
				}
				matches = append(matches, FileMatch{
					LineNumber: i + 1,
					Line:       strings.TrimSpace(line),
					Context:    ctx,
					FuncName:   currentFunc,
					Reason:     pattern.Name,
				})
				break
			}
		}
	}

	return matches
}

func relevantPatterns(errorMsg string) []RiskPattern {
	lower := strings.ToLower(errorMsg)
	switch {
	case strings.Contains(lower, "index") || strings.Contains(lower, "range") || strings.Contains(lower, "slice"):
		return patternsNamed("index access (potential out of range)")
	case strings.Contains(lower, "nil") || strings.Contains(lower, "dereference") || strings.Contains(lower, "null"):
		return patternsNamed("nil dereference risk")
	case strings.Contains(lower, "type") || strings.Contains(lower, "assertion") || strings.Contains(lower, "interface"):
		return patternsNamed("type assertion without ok check")
	case strings.Contains(lower, "deadlock") || strings.Contains(lower, "channel") || strings.Contains(lower, "goroutine"):
		return patternsNamed("channel send/receive (possible deadlock)")
	case strings.Contains(lower, "divide") || strings.Contains(lower, "division") || strings.Contains(lower, "zero"):
		return patternsNamed("division (possible divide by zero)")
	case strings.Contains(lower, "error") || strings.Contains(lower, "err"):
		return patternsNamed("unchecked error (blank identifier)")
	default:
		return riskPatterns
	}
}

func patternsNamed(names ...string) []RiskPattern {
	nameSet := map[string]bool{}
	for _, n := range names {
		nameSet[n] = true
	}
	var out []RiskPattern
	for _, p := range riskPatterns {
		if nameSet[p.Name] {
			out = append(out, p)
		}
	}
	return out
}

func extractFuncName(line string) string {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range []string{"func ", "def ", "function "} {
		if strings.HasPrefix(trimmed, prefix) {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := parts[1]
				if idx := strings.Index(name, "("); idx != -1 {
					name = name[:idx]
				}
				return name
			}
		}
	}
	return ""
}

func isFilePath(s string) bool {
	// has a known code extension = treat as file
	switch strings.ToLower(s[max(0, len(s)-5):]) {
	case ".go", ".py", ".js", ".ts", ".rs", ".rb", ".cs", "java", ".cpp", ".php":
		return true
	}
	return strings.Contains(s, ".")
}

func clampMin(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampMax(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}