// Execute an external command

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateWorkDir creates a directory to hold byproducts of Prolog compilation.
// If the parameter list names a directory, that one is used.  Otherwise, a
// temporary directory is created and used.
func CreateWorkDir(p *Parameters) {
	// Before we return, output the name we chose.
	if p.Verbose {
		defer func() {
			dir, err := filepath.Abs(p.WorkDir)
			if err != nil {
				notify.Fatal(err)
			}
			VerbosePrintf(p, "Storing intermediate files in %s", dir)
		}()
	}

	// If the user specified a directory, create it if necessary and return.
	if p.WorkDir != "" {
		err := os.MkdirAll(p.WorkDir, 0777)
		if err != nil {
			notify.Fatal(err)
		}
		p.DeleteWorkDir = false
		return
	}

	// If the user did not specify a directory, create a random one.
	nm, err := ioutil.TempDir("", "qap-")
	if err != nil {
		notify.Fatal(err)
	}
	p.WorkDir = nm
	p.DeleteWorkDir = true
}

// CreateYosysScript creates a synthesis script for Yosys.
func CreateYosysScript(p *Parameters) {
	// Create a .ys file.
	yName := p.OutFileBase + ".ys"
	VerbosePrintf(p, "Writing a Yosys synthesis script to %s", yName)
	ys, err := os.Create(filepath.Join(p.WorkDir, yName))
	if err != nil {
		notify.Fatal(err)
	}

	// Write some boilerplate text to it.
	fmt.Fprintln(ys, "### Design synthesis")
	fmt.Fprintf(ys, "### Usage: yosys %s.v %s.ys -b edif -o %s.edif\n",
		p.OutFileBase, p.OutFileBase, p.OutFileBase)
	fmt.Fprint(ys, `
# Check design hierarchy.
hierarchy -auto-top

# Translate processes.
proc; opt

# Detect and optimize FSM encodings.
fsm; opt

# Convert to gate logic.
techmap; opt

# Recast in terms of more gate types.
abc -g AND,NAND,OR,NOR,XOR,XNOR,MUX,AOI3,OAI3,AOI4,OAI4; opt

# Clean up.
clean
`)
}

// RunCommand executes a given command, aborting on error.
func RunCommand(p *Parameters, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	VerbosePrintf(p, "Executing %s %s", name, strings.Join(arg, " "))
	err := cmd.Run()
	if err != nil {
		notify.Fatal(err)
	}
}
