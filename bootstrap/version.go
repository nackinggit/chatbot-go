package bootstrap

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type Version struct {
	version     string
	date        string
	commit      string
	branch      string
	showVersion *bool
}

func (v *Version) init() {
	vv := flag.Lookup("version")
	if vv == nil {
		v.showVersion = flag.Bool("version", false, "Print version of this binary (only valid if compiled with make)")
	}

	os.Setenv("server.version", v.version)
	os.Setenv("server.date", v.date)
	os.Setenv("server.commit", v.commit)
	os.Setenv("server.branch", v.branch)
}

func (v *Version) Init() {
	if !flag.Parsed() {
		flag.Parse()
	}
	vv := flag.Lookup("version")
	if vv != nil && vv.Value.String() == "true" {
		v.printVersion(os.Stdout)
		os.Exit(0)
	}
}

func (v *Version) printVersion(w io.Writer) {
	fmt.Fprintf(w, "Version: %s\n", v.version)
	fmt.Fprintf(w, "Branch: %s\n", v.branch)
	fmt.Fprintf(w, "CommitID: %s\n", v.commit)
	fmt.Fprintf(w, "Binary: %s\n", os.Args[0])
	fmt.Fprintf(w, "Compile date: %s\n", v.date)
	fmt.Fprintf(w, "(version and date only valid if compiled with make)\n")
}
