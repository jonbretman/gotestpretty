package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

type testOutput struct {
	Action  string
	Package string
	Test    string
	Output  string
}

type testResult struct {
	packageName string
	name        string
	fixtures    map[string]*testResult
	isFixture   bool
	skipped     bool
	pass        bool
	output      []string
}

var (
	failTag        = color.New(color.BgHiRed).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	passTag        = color.New(color.BgHiGreen).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	skipTag        = color.New(color.BgYellow).Add(color.FgHiBlack).Add(color.Bold).SprintFunc()
	boldGreen      = color.New(color.FgHiGreen).Add(color.Bold).SprintFunc()
	red            = color.New(color.FgHiRed).SprintFunc()
	boldRed        = color.New(color.FgHiRed).Add(color.Bold).SprintFunc()
	lightGrey      = color.New(color.FgWhite).Add(color.Faint).SprintFunc()
	orange         = color.New(color.FgYellow).SprintFunc()
	fileNameRegexp = regexp.MustCompile("[a-zA-Z_]+?\\.go:\\d+")
	panicPrefix    = "panic: runtime error: "
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	tests := map[string]*testResult{}

	// A list of tests in the order they ran in so that the summary of
	// failed tests can be shown in the same order
	testNames := []string{}

	for scanner.Scan() {
		currLine := scanner.Text()

		var o testOutput
		err := json.Unmarshal([]byte(currLine), &o)
		if err != nil {
			// Must not be JSON - just output it
			fmt.Println(string(currLine))
			continue
		}

		switch o.Action {
		case "skip":
			t, ok := tests[o.Test]
			if !ok {
				continue
			}

			t.skipped = true
			fmt.Printf("\r%s %s %s\n", skipTag(" SKIP "), lightGrey(t.packageName), t.name)
			continue

		case "run":
			testNames = append(testNames, o.Test)

			// A test run with t.Run() inside another test
			if strings.Contains(o.Test, "/") {
				parts := strings.Split(o.Test, "/")
				t, ok := tests[parts[0]]
				if ok {
					t.fixtures[o.Test] = &testResult{
						name:        o.Test,
						packageName: o.Package,
						isFixture:   true,
					}
					tests[o.Test] = t.fixtures[o.Test]
					continue
				}
			}

			t := &testResult{
				name:        o.Test,
				packageName: o.Package,
				fixtures:    map[string]*testResult{},
			}
			tests[t.name] = t
			fmt.Printf("%s %s %s", skipTag(" RUNS "), lightGrey(t.packageName), t.name)
			continue

		case "pass", "fail":
			t, ok := tests[o.Test]
			if !ok {
				continue
			}

			t.pass = o.Action == "pass"

			// A test run with t.Run inside another test
			if t.isFixture {
				continue
			}

			tag := passTag(" PASS ")
			if o.Action == "fail" {
				tag = failTag(" FAIL ")
			}

			fmt.Printf("\r%s %s %s\n", tag, lightGrey(t.packageName), t.name)

			// List test run with t.Run inside this test
			for _, fixture := range t.fixtures {
				s := boldGreen("✓")
				if !fixture.pass {
					s = boldRed("✕")
				}
				fmt.Printf("\t%s %s\n", s, lightGrey(strings.TrimPrefix(fixture.name, t.name+"/")))
			}

			// If the tests pass don't hide any ouput since they are
			// likely to be debug print statements
			if t.pass {
				for _, o := range t.output {
					fmt.Print(o)
				}
			}

			continue
		case "output":
			t, ok := tests[o.Test]
			if !ok {
				// Unknown output
				lightGrey(o.Output)
				continue
			}

			output := strings.TrimLeft(o.Output, " ")

			// Ignore the "Error Trace: name_of_file.go" output from failing tests
			if len(t.output) > 0 {
				f := fileNameRegexp.FindAllString(t.output[len(t.output)-1], -1)
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

			t.output = append(t.output, o.Output)
			continue
		default:
			panic(fmt.Sprintf("Unknown test action: %s", o.Action))
		}
	}

	passed := 0
	failed := 0
	skipped := 0

	for _, testName := range testNames {
		t := tests[testName]

		if t.skipped {
			skipped++
			continue
		}

		if t.pass {
			passed++
			continue
		}

		failed++

		// If a failed tests has no output then it will be the "parent" test of one
		// run with t.Run()
		if len(t.output) == 0 {
			continue
		}

		// Package name and tests name
		fmt.Printf("\n%s %s %s\n", failTag(" FAIL "), lightGrey(t.packageName), red(t.name))

		isPanic := false
		for _, o := range t.output {

			// Check for a panic
			if strings.HasPrefix(o, panicPrefix) {

				// Try to find the relevant bit of code that panic'd
				for _, o := range t.output {
					if !strings.Contains(o, t.packageName) {
						continue
					}
					m := fileNameRegexp.FindAllString(o, 1)
					if len(m) == 1 {
						code := getCode(t.packageName, m[0])
						fmt.Printf("%s\n\n", code)
						break
					}
				}

				fmt.Printf("    %s%s", boldRed(panicPrefix), strings.TrimPrefix(o, panicPrefix))
				isPanic = true
				continue
			}

			if isPanic {
				if strings.Contains(o, t.packageName) && !strings.HasPrefix(o, "FAIL\t") {
					fmt.Printf("    %s", o)
				} else {
					fmt.Printf("    %s", lightGrey(o))
				}
				continue
			}

			// Try to find the file containing the test and print the relevant lines
			m := fileNameRegexp.FindAllString(o, -1)
			if len(m) == 1 && strings.TrimSpace(o) == m[0]+":" {
				code := getCode(t.packageName, m[0])
				fmt.Printf("%s\n\n", code)
				continue
			}

			// highlight some key lines
			o = strings.Replace(o, "expected:", boldGreen("expected:"), 1)
			o = strings.Replace(o, "actual  :", boldRed("actual  :"), 1)
			fmt.Print(o)
		}
	}

	// Output a summary
	summary := []string{}
	if passed > 0 {
		summary = append(summary, boldGreen(fmt.Sprintf("%d passed", passed)))
	}

	if failed > 0 {
		summary = append(summary, boldRed(fmt.Sprintf("%d failed", failed)))
	}

	if skipped > 0 {
		summary = append(summary, orange(fmt.Sprintf("%d skipped", skipped)))
	}

	summary = append(summary, fmt.Sprintf("%d total", len(tests)))
	fmt.Printf("\nSummary:  %s\n", strings.Join(summary, ", "))
}

func getCode(packageName string, filenameLineNumber string) string {
	parts := strings.Split(filenameLineNumber, ":")
	filename := parts[0]
	lineNumber, _ := strconv.Atoi(parts[1])

	dirs := strings.Split(packageName, "/")

	// try to find file in subdirectory
	for i := len(dirs) - 1; i >= 0; i-- {
		possibleFilePath := path.Join(path.Join(dirs[i:]...), filename)
		b, err := ioutil.ReadFile(possibleFilePath)
		if err != nil {
			continue
		}

		return formatCode(possibleFilePath, string(b), lineNumber)
	}

	// try in the current directory
	b, err := ioutil.ReadFile(filename)
	if err == nil {
		return formatCode(filename, string(b), lineNumber)
	}

	// fall back to just showing the filename
	return filename
}

func formatCode(filename string, code string, lineNumber int) string {
	lineNumbers := []int{
		lineNumber - 2,
		lineNumber - 1,
		lineNumber,
		lineNumber + 1,
		lineNumber + 2,
	}
	lines := strings.Split(code, "\n")[lineNumber-3 : lineNumber+2]
	result := []string{}

	// Get the number of chars in the highest line number so that we can
	// correctly pad all line numbers so they take up the same number of
	// chars. This is important when your test failure is on line 10 for
	// example so that we can format it like: " 8"," 9", "10" etc.
	maxLineNumberWidth := len(strconv.Itoa(lineNumbers[len(lineNumbers)-1]))

	for i, line := range lines {
		lineNumberFormatted := fmt.Sprintf("%*d", maxLineNumberWidth, lineNumbers[i])

		if lineNumbers[i] == lineNumber {
			result = append(result, fmt.Sprintf(" %s %s |%s", boldRed(">"), lineNumberFormatted, line))
			continue
		}

		result = append(result, fmt.Sprintf("   %s |%s", lightGrey(fmt.Sprintf("%s", lineNumberFormatted)), lightGrey(line)))
	}

	return fmt.Sprintf("%s:%d\n%s", filename, lineNumber, strings.Join(result, "\n"))
}
