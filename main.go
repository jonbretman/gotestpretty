package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
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
	fail           = color.New(color.BgHiRed).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	pass           = color.New(color.BgHiGreen).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	run            = color.New(color.BgHiYellow).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	boldGreen      = color.New(color.FgHiGreen).Add(color.Bold).SprintFunc()
	boldRed        = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	lightGrey      = color.New(color.FgWhite).Add(color.Faint).SprintFunc()
	fileNameRegexp = regexp.MustCompile("[a-zA-Z_]+?\\.go:\\d+")
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
			fmt.Printf("%s %s %s", run(" RUN "), lightGrey(currPackageName), currTestName)
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

			fmt.Printf("\r%s %s %s\n", tag, lightGrey(currPackageName), currTestName)

			// List test run with t.Run inside this test
			for name, result := range currFixtures {
				s := boldGreen("✓")
				if !result {
					s = boldRed("✕")
				}
				fmt.Printf("\t%s %s\n", s, lightGrey(strings.TrimPrefix(name, currTestName+"/")))
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

			// Ignore the "Error Trace: name_of_file.go" output from failing tests
			if len(currTestOutput) > 0 {
				f := fileNameRegexp.FindAllString(currTestOutput[len(currTestOutput)-1], -1)
				if len(f) == 1 && strings.Contains(output, "Error Trace:") && strings.Contains(output, f[0]) {
					continue
				}
			}

			// Ignore the "Test: NameOfTest" output from failing tests
			if strings.Contains(output, "Test:") && strings.Contains(output, o.Test) {
				continue
			}

			// Ignore output saying that a package has no tests
			if strings.HasPrefix(output, "?") && strings.Contains(output, "[no test files]") && strings.Contains(output, o.Package) {
				continue
			}

			// Ignore the RUN/FAIL/PASS output
			if strings.HasPrefix(output, "=== RUN") || strings.HasPrefix(output, "--- FAIL") || strings.HasPrefix(output, "--- PASS") {
				continue
			}

			if currTestName != "" {
				currTestOutput = append(currTestOutput, o.Output)
				continue
			}

			// Unknown output
			lightGrey(o.Output)
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
