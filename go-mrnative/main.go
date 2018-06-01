// Gomrnative is a tool to automate the boilerplate generation of Java code
// necessary to natively interact with Hadoop MapReduce.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/veonik/go-mrnative"
)

const ver = "0.1.0"

var usage = func() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`Usage of %s:
	%s [-version] <command> [<args>]

Arguments:
  -version
  	Displays the current version and exits.

Commands:
  help <command>
	Shows additional information about the specified command.

  init
	Initializes a MapReduce project in the current directory. This command
	is interactive and takes no arguments. It is safe to use this command
	on existing projects.

  build [<package> [, <package> , ... ] ]
	Generates Java source, compiles and jars it. This command accepts zero
	or more package names, which, if passed, will be included in the final
	jar. If no packages are passed, %s will use the current directory.
`, name, name, name)
}

var version = func() {
	name := filepath.Base(os.Args[0])
	fmt.Printf("%s version %s\n", name, ver)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-mrnative: ")
	var v = flag.Bool("version", false, "Displays the current version and exits.")
	flag.Usage = usage
	flag.Parse()
	if *v {
		version()
		os.Exit(0)
	}
	if flag.NArg() < 1 {
		fmt.Println("command not specified")
		fmt.Println("")
		flag.Usage()
		os.Exit(2)
	}
	args := flag.Args()
	switch cmd := args[0]; cmd {
	case "help":
		helpCmd()
	case "init":
		initCmd()
	case "build":
		generateCmd(args[1:])
	default:
		fmt.Printf("unknown command: %s", cmd)
		fmt.Println("")
		flag.Usage()
		os.Exit(2)
	}
}

func helpCmd() {
	flag.Usage()
	os.Exit(0)
}

func initCmd() {
	dir, _ := os.Getwd()
	i := mrnative.NewInitializer(dir)
	i.CollectConfig(os.Stdout, os.Stdin)
	i.Initialize()
}

func generateCmd(pkgs []string) {
	g := mrnative.NewGenerator(pkgs)
	g.Generate()
}
