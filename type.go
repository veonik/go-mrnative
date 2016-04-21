package mrnative

import "strings"

var validTypes = []string{
	"int",
	"int16",
	"int32",
	"string",
	"float32",
	"float64",
}

var typeMapHadoop = map[string]string{
	"int":     "LongWritable",
	"int16":   "ShortWritable",
	"int32":   "IntWritable",
	"string":  "Text",
	"float32": "FloatWritable",
	"float64": "DoubleWritable",
}

var typeMapJava = map[string]string{
	"int":     "long",
	"int16":   "short",
	"int32":   "int",
	"string":  "String",
	"float32": "float",
	"float64": "double",
}

// GoToJavaType converts the given go type into its
// corresponding Java type.
func GoToJavaType(gt string) string {
	if m, ok := typeMapJava[gt]; ok {
		return m
	}
	return gt
}

// GoToHadoopType converts the given go type into its
// corresponding MapReduce type.
func GoToHadoopType(gt string) string {
	if m, ok := typeMapHadoop[gt]; ok {
		return m
	}
	return ""
}

func GoToValueInHadoopType(gt string) string {
	res := GoToHadoopType(gt)
	if strings.HasPrefix(gt, "[]") {
		return "Iterable<" + res + ">"
	}
	return res
}
