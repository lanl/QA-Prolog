// Execute an external command

package main

import (
	"io/ioutil"
	"os"
)

// CreateWorkDir creates a directory to hold byproducts of Prolog compilation.
// If the parameter list names a directory, that one is used.  Otherwise, a
// temporary directory is created and used.
func CreateWorkDir(p *Parameters) {
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
