package mrnative

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tyler-sommer/stick"
)

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	env     *stick.Env
	pkgs    []*Package
	targets []*Target // Targets we are generating code for.
}

// NewGenerator creates a new Generator, ready for use.
func NewGenerator(packages []string) *Generator {
	g := &Generator{newEnv(), []*Package{}, []*Target{}}
	g.parsePackages(packages)
	g.locateTargets()
	return g
}

func newEnv() *stick.Env {
	env := stick.NewEnv(newTemplateLoader())
	env.Filters["hadoop_type"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		return GoToHadoopType(stick.CoerceString(val))
	}
	env.Filters["valuein_type"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		return GoToValueInHadoopType(stick.CoerceString(val))
	}
	env.Filters["java_type"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		return GoToJavaType(stick.CoerceString(val))
	}
	env.Filters["transform"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		baseTyp := stick.CoerceString(val)
		passedVar := stick.CoerceString(args[0])
		transformedVar := stick.CoerceString(args[1])
		getMethod := ".get();"
		if baseTyp == "string" {
			getMethod = ".toString();"
		}
		if strings.HasPrefix(baseTyp, "[]") {
			// Go actually wants a slice. We gotta make magic baby!
			return ""
		}

		return fmt.Sprintf("%s %s = %s%s", GoToJavaType(baseTyp), transformedVar, passedVar, getMethod)
	}
	return env
}

func (g *Generator) parsePackages(packages []string) {
	for _, path := range packages {
		absPath := path
		if !build.IsLocalImport(path) {
			absPath = prefixPath(path)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			log.Fatalln("parsing package:", err)
		}
		var p *build.Package
		var err error
		if !isDirectory(absPath) {
			path = filepath.Dir(path)
			absPath = filepath.Dir(path)
		}
		p, err = build.ImportDir(absPath, 0)
		if err != nil {
			log.Fatalln("importing dir:", err)
		}
		g.pkgs = append(g.pkgs, NewPackage(absPath, p))
	}
}

func (g *Generator) Generate() {
	if len(g.targets) == 0 {
		log.Fatalln("no targets found.")
	}
	for _, target := range g.targets {
		g.genJava(target)
	}
}

func tplParams(t *Target) map[string]stick.Value {
	gobindClassRoot := strings.ToTitle(string(t.pkg.name[0])) + t.pkg.name[1:]
	var mapredMethodName string
	var mapredClassName string
	if t.IsMapper() {
		mapredMethodName = "map"
		mapredClassName = "Mapper"
	} else {
		mapredMethodName = "reduce"
		mapredClassName = "Reducer"
	}
	return map[string]stick.Value{
		"target": t,

		"goStructName":   t.decl.name,
		"goCtxInterface": t.ctx.name,

		"javaPackage":   "go." + t.pkg.name,
		"javaClassName": gobindClassRoot + t.decl.name,

		"gobindClassRoot":   gobindClassRoot,
		"gobindCtxClass":    gobindClassRoot + "." + t.ctx.name,
		"gobindClass":       gobindClassRoot + "." + t.decl.name,
		"gobindConstructor": gobindClassRoot + "." + t.ctor.name,
		"gobindMethodName":  t.method.name,

		"mapredMethodName": mapredMethodName,
		"mapredClassName":  mapredClassName,

		"keyIn":    t.keyIn.typ,
		"valueIn":  t.valueIn.typ,
		"keyOut":   t.keyOut.typ,
		"valueOut": t.valueOut.typ,
	}
}

func (g *Generator) genJava(target *Target) {
	params := tplParams(target)
	dir, _ := os.Readlink(target.pkg.dir)
	dir = filepath.Join(dir, "build/java/go")
	if err := os.MkdirAll(dir, os.ModeDir|os.ModePerm); err != nil {
		log.Fatalf("rendering: %s", err.Error())
	}
	f, err := os.Create(filepath.Join(dir, params["javaClassName"].(string)+".java"))
	if err != nil {
		log.Fatalf("rendering: %s", err.Error())
	}
	err = g.env.Execute("tpl/class_template.java.twig", f, params)
	if err != nil {
		log.Fatalf("rendering: %s", err.Error())
	}
}

func (g *Generator) locateTargets() {
	for _, pkg := range g.pkgs {
		for _, s := range pkg.structs {
			if strings.Contains(s.comment, "@mapper") {
				g.targets = append(g.targets, NewTarget(pkg, s, targetMapper))
			}
			if strings.Contains(s.comment, "@reducer") {
				g.targets = append(g.targets, NewTarget(pkg, s, targetReducer))
			}
		}
	}
}

// isDirectory returns true if name is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatalln("checking is directory:", err)
	}
	return info.IsDir()
}

// prefixPath prefixes the given path with GOROOT and GOPATH
// in an attempt to locate the actual full path.
func prefixPath(path string) string {
	res := filepath.Join(build.Default.GOROOT, "src", path)
	if _, err := os.Stat(res); os.IsNotExist(err) {
		res = filepath.Join(build.Default.GOPATH, "src", path)
		if _, err := os.Stat(res); os.IsNotExist(err) {
			log.Fatalln("prefixing path: %s", err)
		}
	}
	return res
}
