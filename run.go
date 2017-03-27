// Execute an external command

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// CreateWorkDir creates a directory to hold byproducts of Prolog compilation.
// If the parameter list names a directory, that one is used.  Otherwise, a
// temporary directory is created and used.
func CreateWorkDir(p *Parameters) {
	// Before we return, output the name we chose.
	if p.Verbose {
		defer func() {
			dir, err := filepath.Abs(p.WorkDir)
			CheckError(err)
			VerbosePrintf(p, "Storing intermediate files in %s", dir)
		}()
	}

	// If the user specified a directory, create it if necessary and return.
	if p.WorkDir != "" {
		err := os.MkdirAll(p.WorkDir, 0777)
		CheckError(err)
		p.DeleteWorkDir = false
		return
	}

	// If the user did not specify a directory, create a random one.
	nm, err := ioutil.TempDir("", "qap-")
	CheckError(err)
	p.WorkDir = nm
	p.DeleteWorkDir = true
}

// CreateYosysScript creates a synthesis script for Yosys.
func CreateYosysScript(p *Parameters) {
	// Create a .ys file.
	yName := p.OutFileBase + ".ys"
	VerbosePrintf(p, "Writing a Yosys synthesis script to %s", yName)
	ys, err := os.Create(filepath.Join(p.WorkDir, yName))
	CheckError(err)

	// Write some boilerplate text to it.
	fmt.Fprintln(ys, "### Design synthesis")
	fmt.Fprintf(ys, "### Usage: yosys %s.v %s.ys -b edif -o %s.edif\n",
		p.OutFileBase, p.OutFileBase, p.OutFileBase)
	fmt.Fprint(ys, `
# Check design hierarchy.
hierarchy -top Query

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
	CheckError(err)
}

// parseQMASMOutput is a helper function for RunQMASM that parses all of the
// solutions and reports them in a user-friendly format.
func (a *ASTNode) parseQMASMOutput(p *Parameters, haveVar bool, tys TypeInfo) {
	// Open the QMASM output file.
	r, err := os.Open(p.OutFileBase + ".out")
	CheckError(err)
	rb := bufio.NewReader(r)

	// Discard lines until we find a solution.
	for {
		ln, err := rb.ReadString('\n')
		if err == io.EOF {
			notify.Fatal("No solutions were found")
		}
		CheckError(err)
		if len(ln) > 10 && ln[:10] == "Solution #" {
			break
		}
	}

	// Parse lines until we reach the end of the file.
	for {
		// Read a line.
		ln, err := rb.ReadString('\n')
		if err == io.EOF {
			break
		}
		CheckError(err)

		// Output a blank line between solutions.
		if len(ln) > 10 && ln[:10] == "Solution #" {
			fmt.Println("")
			continue
		}

		// Extract a query variable and decimal value if both are
		// present.
		fields := strings.Fields(ln)
		if len(fields) != 3 {
			continue
		}
		if len(fields[0]) < 7 || fields[0][:6] != "Query." {
			continue
		}
		nm := fields[0][6:]
		val, err := strconv.Atoi(fields[2])
		CheckError(err)

		// Output the variable and its value.
		switch {
		case nm == "Valid":
			switch {
			case haveVar:
			case val == 0:
				fmt.Println("false")
			case val == 1:
				fmt.Println("true")
			}

		case tys[nm] == InfNumeral:
			// Output numeric values.
			fmt.Printf("%s = %d\n", nm, val)

		case tys[nm] == InfAtom:
			// Output symbolic values.
			sym := "[invalid]"
			if val >= 0 && val < len(p.IntToSym) {
				sym = p.IntToSym[val]
			}
			fmt.Printf("%s = %s\n", nm, sym)

		default:
			notify.Printf("Internal error processing %q", strings.TrimSpace(ln))
		}
	}
	err = r.Close()
	CheckError(err)
}

// showTail is a helper function for RunQMASM that outputs the last non-blank
// line of a file.
func (a *ASTNode) showTail(fn string) error {
	r, err := os.Open(fn)
	if err != nil {
		return err
	}
	rb := bufio.NewReader(r)
	last := ""
	for {
		ln, err := rb.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if ln != "" {
			last = ln
		}
	}
	if last != "" {
		fmt.Fprintln(os.Stderr, last)
	}
	return nil
}

// RunQMASM runs qmasm, parses the results, and outputs them.
func (a *ASTNode) RunQMASM(p *Parameters, clVarTys map[*ASTNode]TypeInfo) {
	// Find the type of each query argument.
	cl := a.FindByType(QueryType)[0]
	tys := clVarTys[cl]

	// Write verbose output to a file in case the user wants to look at it
	// later.
	out, err := os.Create(p.OutFileBase + ".out")
	CheckError(err)

	// Construct a QMASM argument list.
	args := make([]string, 0, 7)
	args = append(args, "--run", "-O", "--verbose", "--values=ints", "--postproc=opt")
	haveVar := false
	for nm := range tys {
		if unicode.IsUpper(rune(nm[0])) {
			// If the query contains at least one variable, we're
			// trying to find valid values for all variables.
			// Otherwise, we're trying to determine if the
			// arguments represent a true statement.
			haveVar = true
			args = append(args, "--pin=Query.Valid := true")
			break
		}
	}
	args = append(args, p.OutFileBase+".qmasm")

	// Execute QMASM.
	cmd := exec.Command("qmasm", args...)
	cmd.Stdout = out
	cmd.Stderr = out
	VerbosePrintf(p, "Executing qmasm %s", strings.Join(args, " "))
	err = cmd.Run()
	out.Close()
	if err != nil {
		// Output the last line of the .out file before aborting.
		_ = a.showTail(p.OutFileBase + ".out")
		CheckError(err)
	}

	// Report QMASM's output in terms of the query variables.
	a.parseQMASMOutput(p, haveVar, tys)
}
