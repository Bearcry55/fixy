package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ─────────────────────────────────────────────
//  LANGUAGE PATTERN REGISTRY
// ─────────────────────────────────────────────

// langPatterns maps language tag → its full set of risk patterns
var langPatterns = map[string][]RiskPattern{
	"go":         goPatterns,
	"python":     pythonPatterns,
	"javascript": jsPatterns,
	"rust":       rustPatterns,
	"java":       javaPatterns,
	"ruby":       rubyPatterns,
	"cpp":        cppPatterns,
	"csharp":     csharpPatterns,
	"php":        phpPatterns,
}

// ─────────────────────────────────────────────
//  GO PATTERNS
// ─────────────────────────────────────────────

var goPatterns = []RiskPattern{
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
						if isAlphanumeric(last) || last == ')' {
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
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "func ") {
				return false
			}
			// method call on variable that could be nil
			hasDotCall := strings.Contains(line, ".") && strings.Contains(line, "(")
			// pointer dereference but not in type position
			hasStar := strings.Contains(line, "*") &&
				!strings.HasPrefix(trimmed, "type ") &&
				!strings.HasPrefix(trimmed, "var ")
			return hasDotCall || hasStar
		},
	},
	{
		Name: "unchecked error (blank identifier)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
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
		Name: "channel op (possible deadlock)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(line)
			return !strings.HasPrefix(trimmed, "//") && strings.Contains(line, "<-")
		},
	},
	{
		Name: "division (possible divide by zero)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") &&
				strings.Contains(line, "/") &&
				!strings.Contains(lower, "http") &&
				!strings.Contains(lower, "//")
		},
	},
	{
		Name: "goroutine without sync (possible race)",
		Check: func(line, lower string) bool {
			return strings.Contains(lower, "go ") &&
				strings.Contains(lower, "func") &&
				!strings.HasPrefix(strings.TrimSpace(lower), "//")
		},
	},
}

// ─────────────────────────────────────────────
//  PYTHON PATTERNS
// ─────────────────────────────────────────────

var pythonPatterns = []RiskPattern{
	{
		Name: "index access (potential out of range)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "#") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "bare except (swallows all errors)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			// except: or except Exception: pass
			return trimmed == "except:" || strings.HasPrefix(trimmed, "except: ")
		},
	},
	{
		Name: "eval/exec usage (dangerous)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "#") {
				return false
			}
			return strings.Contains(lower, "eval(") || strings.Contains(lower, "exec(")
		},
	},
	{
		Name: "mutable default argument",
		Check: func(line, lower string) bool {
			// def func(arg=[]) or def func(arg={})
			return strings.Contains(lower, "def ") &&
				(strings.Contains(lower, "=[]") || strings.Contains(lower, "={}") ||
					strings.Contains(lower, "= []") || strings.Contains(lower, "= {}"))
		},
	},
	{
		Name: "division (possible divide by zero)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "#") &&
				strings.Contains(line, "/") &&
				!strings.Contains(lower, "http")
		},
	},
	{
		Name: "None dereference risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "def ") {
				return false
			}
			// method call that could fail if object is None
			return strings.Contains(line, ".") && strings.Contains(line, "(")
		},
	},
}

// ─────────────────────────────────────────────
//  JAVASCRIPT / TYPESCRIPT PATTERNS
// ─────────────────────────────────────────────

var jsPatterns = []RiskPattern{
	{
		Name: "undefined/null dereference risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			// method call without optional chaining ?.
			hasDotCall := strings.Contains(line, ".") &&
				strings.Contains(line, "(") &&
				!strings.Contains(line, "?.")
			return hasDotCall
		},
	},
	{
		Name: "== instead of === (loose equality)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			return strings.Contains(line, "==") &&
				!strings.Contains(line, "===") &&
				!strings.Contains(line, "!==") &&
				!strings.Contains(line, "=>") // not arrow function
		},
	},
	{
		Name: "JSON.parse without try/catch (parse error risk)",
		Check: func(line, lower string) bool {
			return strings.Contains(lower, "json.parse(")
		},
	},
	{
		Name: "unhandled promise (missing .catch)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			return strings.Contains(lower, ".then(") && !strings.Contains(lower, ".catch(")
		},
	},
	{
		Name: "index access (potential undefined)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "var usage (prefer let/const)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return strings.HasPrefix(trimmed, "var ") && !strings.HasPrefix(trimmed, "//")
		},
	},
}

// ─────────────────────────────────────────────
//  RUST PATTERNS
// ─────────────────────────────────────────────

var rustPatterns = []RiskPattern{
	{
		Name: "unwrap() (panics if None/Err)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") && strings.Contains(lower, ".unwrap()")
		},
	},
	{
		Name: "expect() (panics with message)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") && strings.Contains(lower, ".expect(")
		},
	},
	{
		Name: "index access (potential out of range)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "panic! macro",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") && strings.Contains(lower, "panic!(")
		},
	},
	{
		Name: "division (possible divide by zero)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") &&
				strings.Contains(line, "/") &&
				!strings.Contains(lower, "//") &&
				!strings.Contains(lower, "http")
		},
	},
}

// ─────────────────────────────────────────────
//  JAVA PATTERNS
// ─────────────────────────────────────────────

var javaPatterns = []RiskPattern{
	{
		Name: "null dereference risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "public ") ||
				strings.HasPrefix(trimmed, "private ") || strings.HasPrefix(trimmed, "protected ") {
				return false
			}
			return strings.Contains(line, ".") && strings.Contains(line, "(")
		},
	},
	{
		Name: "index access (potential ArrayIndexOutOfBounds)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "empty catch block (swallows exception)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return trimmed == "} catch (exception e) {" ||
				(strings.Contains(lower, "catch") && strings.Contains(lower, "{}"))
		},
	},
	{
		Name: "== for string comparison (use .equals())",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") &&
				strings.Contains(line, "==") &&
				strings.Contains(lower, "string")
		},
	},
}

// ─────────────────────────────────────────────
//  RUBY PATTERNS
// ─────────────────────────────────────────────

var rubyPatterns = []RiskPattern{
	{
		Name: "nil method call risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "def ") {
				return false
			}
			return strings.Contains(line, ".") && strings.Contains(line, "(")
		},
	},
	{
		Name: "eval usage (dangerous)",
		Check: func(line, lower string) bool {
			return strings.Contains(lower, "eval(") || strings.Contains(lower, "eval \"")
		},
	},
	{
		Name: "index access (potential nil)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "#") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
}

// ─────────────────────────────────────────────
//  C++ PATTERNS
// ─────────────────────────────────────────────

var cppPatterns = []RiskPattern{
	{
		Name: "raw pointer dereference",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			return strings.Contains(line, "->") || strings.Contains(line, "*(")
		},
	},
	{
		Name: "index access (potential out of bounds)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "memory allocation without null check",
		Check: func(line, lower string) bool {
			return strings.Contains(lower, "malloc(") ||
				strings.Contains(lower, "new ") && strings.Contains(lower, "*")
		},
	},
	{
		Name: "division (possible divide by zero)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") &&
				strings.Contains(line, "/") &&
				!strings.Contains(lower, "//") &&
				!strings.Contains(lower, "http")
		},
	},
}

// ─────────────────────────────────────────────
//  C# PATTERNS
// ─────────────────────────────────────────────

var csharpPatterns = []RiskPattern{
	{
		Name: "null dereference risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "public ") ||
				strings.HasPrefix(trimmed, "private ") {
				return false
			}
			return strings.Contains(line, ".") && strings.Contains(line, "(")
		},
	},
	{
		Name: "index access (potential out of range)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && isAlphanumeric(prev[len(prev)-1]) {
						return true
					}
				}
			}
			return false
		},
	},
	{
		Name: "empty catch block",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return strings.HasPrefix(trimmed, "catch") && strings.Contains(lower, "{}")
		},
	},
}

// ─────────────────────────────────────────────
//  PHP PATTERNS
// ─────────────────────────────────────────────

var phpPatterns = []RiskPattern{
	{
		Name: "eval usage (dangerous)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			return !strings.HasPrefix(trimmed, "//") && strings.Contains(lower, "eval(")
		},
	},
	{
		Name: "null dereference risk",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "function ") {
				return false
			}
			return strings.Contains(line, "->") || strings.Contains(line, "::")
		},
	},
	{
		Name: "index access (potential undefined)",
		Check: func(line, lower string) bool {
			trimmed := strings.TrimSpace(lower)
			if strings.HasPrefix(trimmed, "//") {
				return false
			}
			for i, ch := range line {
				if ch == '[' && i > 0 {
					prev := strings.TrimSpace(line[:i])
					if len(prev) > 0 && (isAlphanumeric(prev[len(prev)-1]) || prev[len(prev)-1] == ']') {
						return true
					}
				}
			}
			return false
		},
	},
}

// ─────────────────────────────────────────────
//  SCANNER CORE
// ─────────────────────────────────────────────

func scanFile(fileName, errorMsg, lang string) []FileMatch {
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

	// Pick pattern set for this language
	patterns, ok := langPatterns[lang]
	if !ok {
		// Unknown language — use Go patterns as fallback
		patterns = goPatterns
	}

	// Filter to only relevant patterns for this error
	activePatterns := relevantPatterns(errorMsg, patterns)

	var matches []FileMatch
	currentFunc := ""

	for i, line := range lines {
		if fn := extractFuncName(line, lang); fn != "" {
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

// relevantPatterns filters the language pattern set based on the error message
func relevantPatterns(errorMsg string, patterns []RiskPattern) []RiskPattern {
	lower := strings.ToLower(errorMsg)

	// map error keywords → pattern name substrings to activate
	keywords := map[string]string{
		"index":       "index",
		"range":       "index",
		"slice":       "index",
		"bounds":      "index",
		"nil":         "nil",
		"null":        "nil",
		"none":        "nil",
		"dereference": "nil",
		"unwrap":      "unwrap",
		"expect":      "expect",
		"type":        "type assertion",
		"assertion":   "type assertion",
		"deadlock":    "deadlock",
		"channel":     "channel",
		"divide":      "division",
		"zero":        "division",
		"panic":       "panic",
		"error":       "unchecked",
		"err":         "unchecked",
		"exception":   "catch",
		"parse":       "parse",
		"promise":     "promise",
		"memory":      "memory",
		"pointer":     "pointer",
		"eval":        "eval",
	}

	// Find which pattern name substring to activate
	target := ""
	for keyword, patternHint := range keywords {
		if strings.Contains(lower, keyword) {
			target = patternHint
			break
		}
	}

	// Unknown error — run all patterns
	if target == "" {
		return patterns
	}

	// Filter patterns whose name contains the target hint
	var out []RiskPattern
	for _, p := range patterns {
		if strings.Contains(strings.ToLower(p.Name), target) {
			out = append(out, p)
		}
	}

	// If nothing matched the hint, run all (safety fallback)
	if len(out) == 0 {
		return patterns
	}
	return out
}

// ─────────────────────────────────────────────
//  HELPERS
// ─────────────────────────────────────────────

func extractFuncName(line, lang string) string {
	trimmed := strings.TrimSpace(line)
	prefixes := []string{"func "} // Go default

	switch lang {
	case "python":
		prefixes = []string{"def "}
	case "javascript":
		prefixes = []string{"function ", "const ", "let ", "var "}
	case "java", "csharp":
		// Java/C# — look for method signatures (public/private/void etc.)
		prefixes = []string{"public ", "private ", "protected ", "void "}
	case "rust":
		prefixes = []string{"fn "}
	case "ruby":
		prefixes = []string{"def "}
	case "php":
		prefixes = []string{"function "}
	case "cpp":
		prefixes = []string{"void ", "int ", "auto ", "bool "}
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := parts[1]
				if idx := strings.IndexAny(name, "({"); idx != -1 {
					name = name[:idx]
				}
				if name != "" {
					return name
				}
			}
		}
	}
	return ""
}

func isFilePath(s string) bool {
	switch strings.ToLower(s[max(0, len(s)-5):]) {
	case ".go", ".py", ".js", ".ts", ".rs", ".rb", ".cs", "java", ".cpp", ".php":
		return true
	}
	return strings.Contains(s, ".")
}

func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
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