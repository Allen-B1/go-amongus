// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const CODE_REF = `
type %sRef Ref

func (r %sRef) Read() %s {
	var i %s
	Ref(r).Read(&i)
	return i
}

func (r %sRef) Write(i %s) {
	Ref(r).Write(&i)
}
`

func main() {
	types := []string{"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64",
		"float32", "float64", "bool"}

	code := "package amongus"
	for _, type_ := range types {
		uppercase := strings.ToUpper(type_[:1]) + type_[1:]
		code += fmt.Sprintf(CODE_REF, uppercase, uppercase, type_, type_, uppercase, type_)
	}

	f, err := os.Create("ref_basic.go")
	if err != nil {
		panic(err)
	}
	io.WriteString(f, code)
	f.Close()
}
