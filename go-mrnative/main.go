// Gomrnative is a tool to automate the boilerplate generation of Java code
// necessary to natively interact with Hadoop MapReduce.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tyler-sommer/go-mrnative"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-mrnative: ")
	flag.Parse()
	if flag.NArg() == 0 {
		name := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", name)
		fmt.Fprintf(os.Stderr, "\t%s [flags] package [ package [ ... ] ]\n", name)
		os.Exit(2)
	}
	g := mrnative.NewGenerator(flag.Args())
	g.Generate()
}
