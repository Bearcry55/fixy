package main

import (
	"fmt"
	"strings"
)

func printBanner() {
	fmt.Printf("\n%s%s", bold, cyan)
	fmt.Println(`  ███████╗██╗██╗  ██╗██╗   ██╗`)
	fmt.Println(`  ██╔════╝██║╚██╗██╔╝╚██╗ ██╔╝`)
	fmt.Println(`  █████╗  ██║ ╚███╔╝  ╚████╔╝ `)
	fmt.Println(`  ██╔══╝  ██║ ██╔██╗   ╚██╔╝  `)
	fmt.Println(`  ██║     ██║██╔╝ ██╗   ██║   `)
	fmt.Println(`  ╚═╝     ╚═╝╚═╝  ╚═╝   ╚═╝   `)
	fmt.Printf("%s", reset)
	fmt.Printf("  %sCLI Error Hunter — No logins. No BS.%s\n\n", dim, reset)
}

func printFileMatches(fileName string, matches []FileMatch) {
	divider := strings.Repeat("─", 70)
	fmt.Printf("%s%s%s\n", dim, divider, reset)
	fmt.Printf("  %s%s🔍 FOUND IN FILE%s\n", bold, green, reset)
	fmt.Printf("%s%s%s\n\n", dim, divider, reset)

	if len(matches) == 0 {
		fmt.Printf("  %sNo matching lines found in %s%s\n\n", dim, fileName, reset)
		return
	}

	for _, m := range matches {
		if m.FuncName != "" {
			fmt.Printf("  %s%s⚙ Function:%s %s%s%s\n", bold, cyan, reset, bold, m.FuncName, reset)
		}
		fmt.Printf("  %s%s⚠ Risk:%s    %s%s%s\n", bold, red, reset, yellow, m.Reason, reset)
		fmt.Printf("  %s%sLine %d:%s   %s%s%s\n", bold, yellow, m.LineNumber, reset, white, m.Line, reset)

		if len(m.Context) > 0 {
			fmt.Printf("  %s  Context:\n", dim)
			for _, c := range m.Context {
				fmt.Printf("    %s%s\n", c, reset)
			}
		}
		fmt.Println()
	}
}

func printResults(resultsChan <-chan Result) {
	divider := strings.Repeat("─", 70)
	fmt.Printf("%s%s%s\n", dim, divider, reset)
	fmt.Printf("  %s%s⚡ ONLINE RESULTS%s\n", bold, yellow, reset)
	fmt.Printf("%s%s%s\n\n", dim, divider, reset)

	count := 0
	for res := range resultsChan {
		count++
		label, labelColor := "", ""
		switch res.Source {
		case "StackOverflow":
			label, labelColor = " StackOverflow ", blue
		case "GitHub Issues":
			label, labelColor = " GitHub Issues ", magenta
		case "Reddit":
			label, labelColor = " Reddit ", yellow
		}

		fmt.Printf("  %s%s%s%s\n", bold+labelColor, label, reset, dim)
		fmt.Printf("  %s%s%s\n", reset+bold+white, res.Title, reset)
		fmt.Printf("  %s%s%s\n", dim, res.Info, reset)
		fmt.Printf("  %s🔗 %s%s\n", cyan, res.Link, reset)
		fmt.Printf("%s%s%s\n\n", dim, divider, reset)
	}

	if count == 0 {
		fmt.Printf("  %s%s No results found.%s Try a shorter error phrase.\n\n", bold, red, reset)
	} else {
		fmt.Printf("  %s%s✓ %d result(s) — links are clickable in most terminals.%s\n\n", bold, green, count, reset)
	}
}