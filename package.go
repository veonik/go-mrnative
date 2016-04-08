package mrnative

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"strings"
)

type Struct struct {
	name    string
	comment string
	methods []*Method
}

func (s *Struct) String() string {
	return fmt.Sprintf("%s{%s}", s.name, s.comment)
}

func NewStruct(tdef structDecl, methods []*Method) *Struct {
	var cmt []string
	if tdef.decl.Doc != nil {
		declComments := tdef.decl.Doc.List
		for i := 0; i < len(declComments); i++ {
			cmt = append(cmt, declComments[i].Text)
		}
	}
	if tdef.spec.Doc != nil {
		specComments := tdef.spec.Doc.List
		for i := 0; i < len(specComments); i++ {
			cmt = append(cmt, specComments[i].Text)
		}
	}
	return &Struct{tdef.spec.Name.String(), strings.Join(cmt, "\n"), methods}
}

type Interface struct {
	name    string
	methods []*Method
}

func NewInterface(spec *ast.TypeSpec) *Interface {
	name := spec.Name.String()
	typ := spec.Type.(*ast.InterfaceType)
	var m []*Method
	methods := typ.Methods.List
	for i := 0; i < len(methods); i++ {
		m = append(m, NewInterfaceMethod(name, methods[i]))
	}
	return &Interface{name, m}
}

type Method struct {
	*Func
	recv string
}

func NewInterfaceMethod(recv string, f *ast.Field) *Method {
	var p []*Param
	var r []*Param
	if par := f.Type.(*ast.FuncType).Params; par != nil {
		params := par.List
		for i := 0; i < len(params); i++ {
			p = append(p, NewParam(params[i]))
		}
	}
	if ret := f.Type.(*ast.FuncType).Results; ret != nil {
		returns := ret.List
		for i := 0; i < len(returns); i++ {
			r = append(r, NewParam(returns[i]))
		}
	}
	return &Method{&Func{f.Names[0].Name, p, r}, recv}
}

func NewMethod(fn *ast.FuncDecl) *Method {
	return &Method{NewFunc(fn), fmt.Sprintf("%+v", fn.Recv.List[0].Type.(*ast.StarExpr).X)}
}

type Func struct {
	name    string
	params  []*Param
	returns []*Param
}

func NewFunc(fn *ast.FuncDecl) *Func {
	var p []*Param
	var r []*Param
	if par := fn.Type.Params; par != nil {
		params := par.List
		for i := 0; i < len(params); i++ {
			p = append(p, NewParam(params[i]))
		}
	}
	if ret := fn.Type.Results; ret != nil {
		returns := ret.List
		for i := 0; i < len(returns); i++ {
			r = append(r, NewParam(returns[i]))
		}
	}
	return &Func{fn.Name.String(), p, r}
}

type Param struct {
	name string
	typ  string
}

func NewParam(param *ast.Field) *Param {
	var typName string
	if typ, ok := param.Type.(*ast.ArrayType); ok {
		typName = "[]" + fmt.Sprintf("%+v", typ.Elt)
	} else {
		typName = fmt.Sprintf("%+v", param.Type)
	}
	var name string
	if param.Names == nil {
		name = ""
	} else {
		name = param.Names[0].Name
	}
	return &Param{name, typName}
}

// A Package contains information about a given Package.
type Package struct {
	name       string
	dir        string
	structs    []*Struct
	interfaces []*Interface
	functions  []*Func

	gend bool // True if this package already contains generated Go code.
}

type structDecl struct {
	decl *ast.GenDecl
	spec *ast.TypeSpec
}

// NewPackage creates a new package, ready for use.
func NewPackage(absPath string, p *build.Package) *Package {
	pkg := &Package{p.Name, absPath, []*Struct{}, []*Interface{}, []*Func{}, false}

	fs := token.NewFileSet()
	var astFiles []*ast.File
	for _, filename := range p.GoFiles {
		filename = filepath.Join(absPath, filename)
		parsedFile, err := parser.ParseFile(fs, filename, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("parsing package: %s: %s", filename, err)
		}
		astFiles = append(astFiles, parsedFile)
	}
	pkg.check(fs, astFiles)
	pkg.walkFiles(astFiles)
	return pkg
}

func (pkg *Package) walkFiles(files []*ast.File) {
	var structs []structDecl
	var methods []*ast.FuncDecl
	for h := 0; h < len(files); h++ {
		file := files[h]
		for i := 0; i < len(file.Decls); i++ {
			switch decl := file.Decls[i].(type) {
			case *ast.GenDecl:
				for j := 0; j < len(decl.Specs); j++ {
					spec, ok := decl.Specs[j].(*ast.TypeSpec)
					if !ok {
						continue
					}
					switch spec.Type.(type) {
					case *ast.InterfaceType:
						pkg.interfaces = append(pkg.interfaces, NewInterface(spec))
						log.Printf("%s: type %s interface\n", file.Name, spec.Name)
					case *ast.StructType:
						structs = append(structs, structDecl{decl, spec})
						log.Printf("%s: type %s struct\n", file.Name, spec.Name)
					}
				}
			case *ast.FuncDecl:
				if decl.Recv == nil {
					pkg.functions = append(pkg.functions, NewFunc(decl))
					log.Printf("%s: func %s(...)\n", file.Name, decl.Name)
				} else {
					methods = append(methods, decl)
					log.Printf("%s: func (%s) %s(...)\n", file.Name, decl.Recv.List[0].Type, decl.Name)
				}

			default:
				continue
			}
		}
	}
	// After all Package members are known, we can create our Structs.
	for i := 0; i < len(structs); i++ {
		var sm []*Method
		name := structs[i].spec.Name.String()
		for j := 0; j < len(methods); j++ {
			fn := methods[j]
			recv := fmt.Sprintf("%+v", fn.Recv.List[0].Type.(*ast.StarExpr).X)
			if recv == name {
				sm = append(sm, NewMethod(fn))
			}
		}
		pkg.structs = append(pkg.structs, NewStruct(structs[i], sm))
	}
	fmt.Println(pkg.structs)
}

// check ensures the package has no parse errors.
func (pkg *Package) check(fs *token.FileSet, astFiles []*ast.File) {
	config := types.Config{Importer: importer.Default(), FakeImportC: true}
	_, err := config.Check(pkg.dir, fs, astFiles, nil)
	if err != nil {
		log.Fatalf("checking package: %s", err)
	}
}
