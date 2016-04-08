package mrnative

import "strings"

var typeMapHadoop = map[string]string{
	"int":    "LongWritable",
	"long":   "LongWritable",
	"string": "Text",
}

var typeMapJava = map[string]string{
	"int":    "long",
	"string": "String",
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
