package mrnative

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"strconv"

	"bufio"

	"github.com/tyler-sommer/stick"
)

type Initializer struct {
	env    *stick.Env
	pieces map[string]stick.Value
}

func NewInitializer(dir string) *Initializer {
	return &Initializer{newInitializerEnv(), make(map[string]stick.Value)}
}

func newInitializerEnv() *stick.Env {
	env := stick.New(newTemplateLoader())
	return env
}

func (i *Initializer) CollectConfig(out io.Writer, in io.Reader) {
	c := &collector{bufio.NewReader(in), out}
	d, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	i.pieces["name"] = c.AskDefault("Enter package name", filepath.Base(d))
	types := make(map[string]stick.Value)
	mapper := false
	reducer := false
	for {
		dat := make(map[string]stick.Value)
		var def string
		if mapper == false {
			def = "mapper"
		} else if reducer == false {
			def = "reducer"
		}
		t := c.AskDefaultOptions("Define a", def, "mapper", "reducer", "combiner")
		dat["type"] = t
		if t == "mapper" {
			mapper = true
		} else if t == "reducer" {
			reducer = true
		}
		dat["type_name"] = c.AskDefault("Enter type name", strings.ToTitle(string(t[0]))+t[1:])
		dat["keyIn"] = c.AskDefaultOptions("Enter key in type", "int", validTypes...)
		dat["valueIn"] = c.AskDefaultOptions("Enter value in type", "string", validTypes...)
		dat["keyOut"] = c.AskDefaultOptions("Enter key out type", dat["keyIn"].(string), validTypes...)
		dat["valueOut"] = c.AskDefaultOptions("Enter value out type", dat["valueIn"].(string), validTypes...)
		types[dat["type_name"].(string)] = dat

		if !c.AskDefaultBool("Define another?", !(mapper && reducer)) {
			break
		}
	}
	i.pieces["types"] = types

	target := filepath.Join(c.AskDefault("Target for generated files", d), i.pieces["name"].(string)+".go")
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		a := c.AskDefaultOptions("Target exists", "Overwrite", "Overwrite", "Keep both")
		if a == "Keep both" {
			for i := 0; true; i++ {
				try := strings.Replace(target, ".go", "_"+strconv.FormatInt(int64(i), 10)+".go", 1)
				if _, err := os.Stat(try); os.IsNotExist(err) {
					target = try
					break
				}
			}
		}
	}
	i.pieces["target"] = target
}

func (i *Initializer) Initialize() {
	tgt := i.pieces["target"].(string)
	if err := os.MkdirAll(filepath.Dir(tgt), os.ModeDir|os.ModePerm); err != nil {
		panic(err)
	}
	f, err := os.Create(tgt)
	if err != nil {
		panic(err)
	}
	err = i.env.Execute("tpl/init_template.go.twig", f, i.pieces)
	if err != nil {
		panic(err)
	}
}

type collector struct {
	in  *bufio.Reader
	out io.Writer
}

func (c *collector) printf(format string, args ...interface{}) {
	fmt.Fprintf(c.out, format, args...)
}

func (c *collector) readline() string {
	var res string
	res, err := c.in.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimRight(res, "\n")
}

func (c *collector) Ask(question string) string {
	c.printf(question + ": ")
	return c.readline()
}

func (c *collector) AskDefault(question, def string) string {
	if def != "" {
		question = question + " (default: " + def + ")"
	}
	res := c.Ask(question)
	if res == "" {
		return def
	}
	return res
}

func (c *collector) AskBool(question string) bool {
	res := c.AskOptions(question, "Yes", "No")
	return res == "Yes"
}

func (c *collector) AskDefaultBool(question string, def bool) bool {
	var sdef string
	if def {
		sdef = "Yes"
	} else {
		sdef = "No"
	}
	res := c.AskDefaultOptions(question, sdef, "Yes", "No")
	return res == "Yes"
}

func (c *collector) AskOptions(question string, options ...string) string {
	return c.AskDefaultOptions(question, "", options...)
}

func (c *collector) AskDefaultOptions(question, def string, options ...string) string {
	if def != "" {
		question = question + " (default: " + def + ")"
	}
	opts := make(map[string]string)
	for i := 0; i < len(options); i++ {
		opts[strings.ToLower(options[i])] = options[i]
	}
	for i := 0; i < 3; i++ { // three tries. TODO: return error
		res := c.Ask(question + " [" + strings.Join(options, ", ") + "]")
		if res == "" {
			return def
		}
		lres := strings.ToLower(res)
		if v, ok := opts[lres]; ok {
			return v
		}
	}
	return def
}
