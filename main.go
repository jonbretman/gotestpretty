package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type testOutput struct {
	Time    string
	Action  string
	Package string
	Test    string
	Output  string
}

var (
	fail        = color.New(color.BgHiRed).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	pass        = color.New(color.BgHiGreen).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	run         = color.New(color.BgHiYellow).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	boldGreen   = color.New(color.FgHiGreen).Add(color.Bold).SprintFunc()
	boldRed     = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	packageName = color.New(color.FgWhite).Add(color.Faint).SprintFunc()
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var line []byte

	currTestName := ""
	currPackageName := ""
	currTestOutput := []string{}
	currFixtures := map[string]bool{}

	passed := 0
	failed := 0

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		if input != '\n' {
			line = append(line, byte(input))
			continue
		}

		var o testOutput
		err = json.Unmarshal([]byte(line), &o)
		if err != nil {
			panic(err)
		}

		// reset line
		line = []byte{}

		switch o.Action {
		case "skip":
			// do nothing
			continue

		case "run":
			// A test run with t.Run inside another test
			if currTestName != "" && strings.HasPrefix(o.Test, currTestName+"/") {
				currFixtures[o.Test] = false
				continue
			}

			currTestName = o.Test
			currPackageName = o.Package
			fmt.Printf("%s %s %s", run(" RUN "), packageName(currPackageName), currTestName)
			continue

		case "pass", "fail":
			// Not interested in pass/fail actions that are not for tests
			if o.Test == "" {
				continue
			}

			// A test run with t.Run inside another test
			if o.Test != currTestName && strings.HasPrefix(o.Test, currTestName+"/") {
				if o.Action == "pass" {
					currFixtures[o.Test] = true
				}
				continue
			}

			tag := pass(" PASS ")
			if o.Action == "fail" {
				tag = fail(" FAIL ")
			}

			fmt.Printf("\r%s %s %s\n", tag, packageName(currPackageName), currTestName)

			// List test run with t.Run inside this test
			for name, result := range currFixtures {
				s := boldGreen("✓")
				if !result {
					s = boldRed("✕")
				}
				fmt.Printf("\t%s %s\n", s, packageName(strings.TrimPrefix(name, currTestName+"/")))
			}

			// Print output from failing test
			if o.Action == "fail" {
				failed++
				for _, o := range currTestOutput {
					fmt.Print(o)
				}
			} else {
				passed++
			}

			// reset everything
			currTestName = ""
			currTestOutput = nil
			currFixtures = map[string]bool{}
			continue
		case "output":
			output := strings.TrimLeft(o.Output, " ")

			// Noise
			if strings.HasPrefix(output, "?") && strings.Contains(output, "[no test files]") && strings.Contains(output, o.Package) {
				continue
			}

			// Noise
			if strings.HasPrefix(output, "=== RUN") || strings.HasPrefix(output, "--- FAIL") || strings.HasPrefix(output, "--- PASS") {
				continue
			}

			if currTestName != "" {
				currTestOutput = append(currTestOutput, o.Output)
				continue
			}

			// Unknown output
			fmt.Print(o.Output)
			continue
		default:
			panic(fmt.Sprintf("Unknown test action: %s", o.Action))
		}

	}

	// Print summary
	fmt.Printf(
		"\nTests:  %s, %s, %s\n",
		boldGreen(fmt.Sprintf("%d passed", passed)),
		boldRed(fmt.Sprintf("%d failed", failed)),
		fmt.Sprintf("%d total", passed+failed))
}
