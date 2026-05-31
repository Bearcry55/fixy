package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	printBanner()

	if len(os.Args) < 2 {
		fmt.Printf("%s%s ERROR %s Missing parameters.\n", bold, bgRed, reset)
		fmt.Printf("%s Usage:%s  fixy \"your error message\"\n", yellow, reset)
		fmt.Printf("         fixy main.go \"your error message\"\n")
		fmt.Printf("%s Example:%s fixy main.go \"panic: index out of range\"\n\n", cyan, reset)
		os.Exit(1)
	}

	var fileName, errorMessage, langTag string
	var fileMatches []FileMatch

	if len(os.Args) >= 3 && isFilePath(os.Args[1]) {
		fileName     = os.Args[1]
		errorMessage = os.Args[2]
		langTag      = detectLang(fileName)

		fmt.Printf("%s📄 File:%s   %s\n", cyan, reset, fileName)
		if langTag != "" {
			fmt.Printf("%s🔤 Lang:%s   %s\n", cyan, reset, langTag)
		}

		fileMatches = scanFile(fileName, errorMessage, langTag)
		printFileMatches(fileName, fileMatches)
	} else {
		errorMessage = os.Args[1]
	}

	fmt.Printf("%s⚠  Error:%s  %s\n\n", yellow, reset, errorMessage)

	resultsChan := make(chan Result, 20)
	var wg sync.WaitGroup

	wg.Add(3)
	go fetchStackOverflow(&wg, errorMessage, langTag, resultsChan)
	go fetchGitHubIssues(&wg, errorMessage, langTag, resultsChan)
	go fetchReddit(&wg, errorMessage, langTag, resultsChan)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	printResults(resultsChan)

	_ = fileMatches // used via printFileMatches
}

func detectLang(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".ts", ".jsx", ".tsx":
		return "javascript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	default:
		return ""
	}
}