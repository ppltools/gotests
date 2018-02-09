// Package process is a thin wrapper around the gotests library. It is intended
// to be called from a binary and handle its arguments, flags, and output when
// generating tests.
package process

import (
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ppltools/gotests"

	"io"

	"github.com/ppltools/cmsg"
)

const newFilePerm os.FileMode = 0644

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
	cmsg.Default.Stderr = out

	if opts == nil {
		opts = &Options{}
	}
	opt := parseOptions(opts)
	if opt == nil {
		return
	}
	if len(args) == 0 {
		cmsg.Die("-> please specify a file or directory containing the source")
	}
	for _, path := range args {
		generateTests(path, opts.WriteOutput, opt)
	}
}

func parseOptions(opt *Options) *gotests.Options {
	if opt.OnlyFuncs == "" && opt.ExclFuncs == "" && !opt.ExportedFuncs && !opt.AllFuncs {
		cmsg.Die("-> please specify either the -only, -excl, -export, or -all flag")
	}
	onlyRE, err := parseRegexp(opt.OnlyFuncs)
	if err != nil {
		cmsg.Die("-> invalid -only regex: %s", err)
	}
	exclRE, err := parseRegexp(opt.ExclFuncs)
	if err != nil {
		cmsg.Die("-> invalid -excl regex: %s", err)
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

func generateTests(path string, writeOutput bool, opt *gotests.Options) {
	gts, err := gotests.GenerateTests(path, opt)
	if err != nil {
		cmsg.Die("-> generate test failed: %s", err)
	}
	if len(gts) == 0 {
		cmsg.Warn("-> no tests generated for: %s", path)
	}
	for _, t := range gts {
		outputTest(t, writeOutput)
	}
}

func outputTest(t *gotests.GeneratedTest, writeOutput bool) {
	if writeOutput {
		if err := ioutil.WriteFile(t.Path, t.Output, newFilePerm); err != nil {
			cmsg.Die("-> write file %s failed: %s", t.Path, err)
		}
	}
	for _, t := range t.Functions {
		cmsg.Info("-> generated: %s", t.TestName())
	}
}
