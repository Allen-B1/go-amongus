// + build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

type Field [2]interface{}

func (f Field) Offset() uintptr {
	return uintptr(f[0].(int64))
}

func (f Field) Type() string {
	return f[1].(string)
}

func transformName(name string) string {
	if len(name) == 0 {
		return ""
	}
	name = strings.Trim(name, "_")
	return strings.ToUpper(name[:1]) + name[1:]
}

func refOf(type_ string) string {
	builtin := []string{"int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64",
		"float32", "float64", "bool", "ptr", "List", "string", "Color", "array"}
	name := transformName(type_)
	if strings.Contains(strings.Join(builtin, ","), type_) {
		name = "amongus." + name
	}
	if isPtrType(type_) {
		name += "Ptr"
	}

	return name + "Ref"
}

type Type interface {
	Generate(name string, static uintptr) string
}

var Types = make(map[string]Type)

func isPtrType(type_ string) bool {
	if c, ok := Types[type_].(*Class); ok {
		return !c.Struct
	}
	if type_ == "List" || type_ == "string" || type_ == "array" {
		return true
	}
	return false
}

type Class struct {
	Obfuscated string           `toml:"obfuscated"`
	Struct     bool             `toml:"struct"`
	Addr       uintptr          `toml:"addr"`
	Static     map[string]Field `toml:"static"`
	Fields     map[string]Field `toml:"fields"`
}

const CLASS_TEMPLATE = `
package aurefs

import (
	"github.com/allen-b1/go-amongus"
)

type %sPtrRef amongus.PtrRef

func (r %sPtrRef) Null() bool {
	return amongus.PtrRef(r).Null()
}

func (r %sPtrRef) Deref() %sRef {
	return %sRef(amongus.PtrRef(r).Deref())
}

// Type %sRef represents an instance of %s.
type %sRef amongus.Ref

// Obfuscated: %s
`

const CLASS_FIELD_TEMPLATE = `
func (r %sRef) %s() %s {
	ref := amongus.Ref(r).Ref(%v)
	return %s(ref)
}
`

const CLASS_STATIC_TEMPLATE = `
func %s%s(au *amongus.AmongUs) %s {
	ref := au.Ref(au.ModBaseAddr, %v)
	return %s(ref)
}
`

const CLASS_SIG_TEMPLATE = `
func Sig%s(au *amongus.AmongUs) uintptr {
	ref := au.Ref(au.ModBaseAddr, %v, 0)
	return ref.Addr
}
`

func (c *Class) Generate(name string, static uintptr) string {
	transName := transformName(name)

	code := ""
	code += fmt.Sprintf(CLASS_TEMPLATE,
		transName, transName, transName, transName, transName,
		transName, name, transName, c.Obfuscated)
	if c.Addr != 0 {
		code += fmt.Sprintf(CLASS_SIG_TEMPLATE, transName, c.Addr)
	}
	for name, field := range c.Fields {
		transFieldName := transformName(name)
		fieldRefType := refOf(field.Type())

		offsets := fmt.Sprint(field.Offset())

		code += fmt.Sprintf(CLASS_FIELD_TEMPLATE, transName, transFieldName, fieldRefType, offsets, fieldRefType)
	}
	for name, field := range c.Static {
		transFieldName := transformName(name)
		fieldRefType := refOf(field.Type())

		offsets := []uintptr{c.Addr}
		offsets = append(offsets, static)
		offsets = append(offsets, field.Offset())

		offsetsString := ""
		for _, offset := range offsets {
			offsetsString += "," + fmt.Sprint(offset)
		}
		offsetsString = offsetsString[1:]

		code += fmt.Sprintf(CLASS_STATIC_TEMPLATE, transName, transFieldName, fieldRefType, offsetsString, fieldRefType)
	}
	return code
}

const ENUM_TEMPLATE = `
package aurefs

import (
	"github.com/allen-b1/go-amongus"
)

// Type %s represents an instance of %s.
type %s %s

// Obfuscated: %s

const (
	%s
)

// Type %sRef represents a reference to an instance of %s.
type %sRef %s

func (r %sRef) Read() %s {
	arg := %s(r).Read()
	return %s(arg)
}

func (r %sRef) Write(arg %s) {
	%s(r).Write(%s(arg))
}
`

type Enum struct {
	Obfuscated string                 `toml:"obfuscated"`
	Type       string                 `toml:"type"`
	Items      map[string]interface{} `toml:"items"`
}

func (e *Enum) Generate(name string, static uintptr) string {
	transName := transformName(name)
	refType := refOf(e.Type)

	constants := ""
	for name, value := range e.Items {
		body, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		valueStr := string(body)
		constants += transName + transformName(name) + " " + transName + " = " + valueStr + "\n\t"
	}

	code := fmt.Sprintf(ENUM_TEMPLATE, transName, transName, transName, e.Type,
		e.Obfuscated,
		constants,
		transName, e.Type,
		transName, refType,
		transName, transName, refType, transName,
		transName, transName, refType, e.Type)
	return code
}

func unmarshal(t *toml.Tree) Type {
	if t.Has("items") {
		obj := new(Enum)
		err := t.Unmarshal(obj)
		if err != nil {
			panic(err)
		}
		return obj
	} else {
		obj := new(Class)
		err := t.Unmarshal(obj)
		if err != nil {
			panic(err)
		}
		return obj
	}
}

func main() {
	tree, err := toml.LoadFile("offsets/" + os.Args[1] + ".toml")
	if err != nil {
		panic(err)
	}

	static := uintptr(tree.Get("static").(int64))
	for _, key := range tree.Keys() {
		if key == "static" {
			continue
		}

		info := unmarshal(tree.Get(key).(*toml.Tree))
		Types[key] = info
	}
	for name, type_ := range Types {
		f, err := os.Create("refs/" + name + ".go")
		if err != nil {
			panic(err)
		}
		io.WriteString(f, type_.Generate(name, static))
		f.Close()
	}
}
