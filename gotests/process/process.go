// Package process is a thin wrapper around the gotests library. It is intended
// to be called from a binary and handle its arguments, flags, and output when
// generating tests.
package process

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ppltools/gotests"
)

const newFilePerm os.FileMode = 0644
const msgFmt string = "\033[%sm%s\033[m%s\n"
const msgInfo string = "0;32"
const msgError string = "0;31"

// Set of options to use when generating tests.
type Options struct {
	OnlyFuncs     string // Regexp string for filter matches.
	ExclFuncs     string // Regexp string for excluding matches.
	ExportedFuncs bool   // Only include exported functions.
	AllFuncs      bool   // Include all non-tested functions.
	PrintInputs   bool   // Print function parameters as part of error messages.
	Subtests      bool   // Print tests using Go 1.7 subtests
	WriteOutput   bool   // Write output to test file(s).
	AllowError    bool   // allow error during test, otherwise exit when error occurs
}

// Generates tests for the Go files defined in args with the given options.
// Logs information and errors to out. By default outputs generated tests to
// out unless specified by opt.
func Run(out io.Writer, args []string, opts *Options) {
	if opts == nil {
		opts = &Options{}
	}
	opt := parseOptions(out, opts)
	if opt == nil {
		return
	}
	if len(args) == 0 {
		fmt.Fprintf(out, msgFmt, msgError, "[ERROR]\t", "-> please specify a file or directory containing the source")
		return
	}
	for _, path := range args {
		generateTests(out, path, opts.WriteOutput, opt)
	}
}

func parseOptions(out io.Writer, opt *Options) *gotests.Options {
	if opt.OnlyFuncs == "" && opt.ExclFuncs == "" && !opt.ExportedFuncs && !opt.AllFuncs {
		fmt.Fprintf(out, msgFmt, msgError, "[ERROR]\t", "-> please specify either the -only, -excl, -export, or -all flag")
		return nil
	}
	onlyRE, err := parseRegexp(opt.OnlyFuncs)
	if err != nil {
		fmt.Fprintf(out, msgFmt, msgError, "[ERROR]\t", "-> invalid -only regex:" + err.Error())
		return nil
	}
	exclRE, err := parseRegexp(opt.ExclFuncs)
	if err != nil {
		fmt.Fprintf(out, msgFmt, msgError, "[ERROR]\t", "-> invalid -excl regex:" + err.Error())
		return nil
	}
	return &gotests.Options{
		Only:        onlyRE,
		Exclude:     exclRE,
		Exported:    opt.ExportedFuncs,
		PrintInputs: opt.PrintInputs,
		Subtests:    opt.Subtests,
		AllowError:  opt.AllowError,
	}
}

func parseRegexp(s string) (*regexp.Regexp, error) {
	if s == "" {
		return nil, nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	return re, nil
}

func generateTests(out io.Writer, path string, writeOutput bool, opt *gotests.Options) {
	gts, err := gotests.GenerateTests(path, opt)
	if err != nil {
		fmt.Fprintf(out, err.Error())
		return
	}
	if len(gts) == 0 {
		fmt.Fprintf(out, msgFmt, msgInfo, "[INFO]\t", "-> no tests generated for: " + path)
		return
	}
	for _, t := range gts {
		outputTest(out, t, writeOutput)
	}
}

func outputTest(out io.Writer, t *gotests.GeneratedTest, writeOutput bool) {
	if writeOutput {
		if err := ioutil.WriteFile(t.Path, t.Output, newFilePerm); err != nil {
			fmt.Fprintf(out, msgFmt, msgError, err)
			return
		}
	}
	for _, t := range t.Functions {
		fmt.Fprintf(out, msgFmt, msgInfo, "[INFO]\t", "-> generated: " + t.TestName())
	}
	if !writeOutput {
		out.Write(t.Output)
	}
}
