package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	pass        bool
	output      []string
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

	var currLine []byte
	var currTest *testResult

	tests := map[string]*testResult{}

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		if input != '\n' {
			currLine = append(currLine, byte(input))
			continue
		}

		var o testOutput
		err = json.Unmarshal(currLine, &o)
		if err != nil {
			panic(err)
		}

		// reset line
		currLine = []byte{}

		switch o.Action {
		case "skip":
			// do nothing
			continue

		case "run":
			// A test run with t.Run() inside another test
			if currTest != nil && strings.HasPrefix(o.Test, currTest.name+"/") {
				currTest.fixtures[o.Test] = &testResult{
					name:        o.Test,
					packageName: o.Package,
					isFixture:   true,
				}
				tests[o.Test] = currTest.fixtures[o.Test]
				continue
			}

			t := &testResult{
				name:        o.Test,
				packageName: o.Package,
				fixtures:    map[string]*testResult{},
			}
			tests[t.name] = t
			currTest = t
			fmt.Printf("%s %s %s", run(" RUN "), lightGrey(t.packageName), t.name)
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

			tag := pass(" PASS ")
			if o.Action == "fail" {
				tag = fail(" FAIL ")
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

			// reset everything
			currTest = nil
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
	for _, t := range tests {
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
		fmt.Printf("\n%s - %s:\n", lightGrey(t.packageName), boldRed(t.name))

		// Output of failed test
		for _, o := range t.output {

			// Try to find the file containing the test and print the relevant lines
			m := fileNameRegexp.FindAllString(o, -1)
			if len(m) == 1 && strings.TrimSpace(o) == m[0]+":" {
				p := strings.Split(strings.TrimSpace(o), ":")
				filename := p[0]
				lineNo, _ := strconv.Atoi(p[1])
				code := getCode(t.packageName, filename, lineNo)
				fmt.Printf("\n%s\n\n", code)
				continue
			}

			// highlight some key lines
			o = strings.Replace(o, "expected:", boldGreen("expected:"), 1)
			o = strings.Replace(o, "actual  :", boldRed("actual  :"), 1)
			fmt.Print(o)
		}
	}

	// Print summary
	fmt.Printf(
		"\nSummary:  %s, %s, %s\n",
		boldGreen(fmt.Sprintf("%d passed", passed)),
		boldRed(fmt.Sprintf("%d failed", failed)),
		fmt.Sprintf("%d total", passed+failed))
}

func getCode(packageName string, filename string, lineNumber int) string {
	dirs := strings.Split(packageName, "/")

	// try to find file in subdirectory
	for i := len(dirs) - 1; i >= 0; i-- {
		possibleFilePath := path.Join(path.Join(dirs[i:]...), filename)
		b, err := ioutil.ReadFile(possibleFilePath)
		if err != nil {
			continue
		}

		return formatCode(filename, string(b), lineNumber)
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
	// chars. This is important when your test failure is on line 11 for
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

	return filename + ":\n" + strings.Join(result, "\n")
}
