package goparse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// StructDesc contains description of parsed struct
type StructDesc struct {
	Name  string
	Field []struct {
		Name string
		Type string
		Tags []string
	}
}

// GetFileStructs returns structs descriptions from parsed go file
func GetFileStructs(filename string, prefix string, tag string) ([]StructDesc, error) {
	result := make([]StructDesc, 0, 5)

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, 0)
	if nil != err {
		return result, err
	}

	for i := range f.Decls {
		if g, ok := f.Decls[i].(*ast.GenDecl); ok {
			for _, s := range g.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok {
					if "" == prefix || strings.HasPrefix(ts.Name.String(), prefix) {
						if tt, ok := ts.Type.(*ast.StructType); ok {
							newStruct := StructDesc{Name: ts.Name.String(), Field: make([]struct {
								Name string
								Type string
								Tags []string
							}, 0, len(tt.Fields.List))}
							for _, field := range tt.Fields.List {
								newField := struct {
									Name string
									Type string
									Tags []string
								}{}
								if len(field.Names) < 1 {
									continue
								}
								newField.Name = field.Names[0].Name
								if e, ok := field.Type.(*ast.Ident); ok {
									newField.Type = e.Name
								}
								if e, ok := field.Type.(*ast.ArrayType); ok {
									if e2, ok := e.Elt.(*ast.Ident); ok {
										newField.Type = "[]" + e2.Name
									}
								}
								if nil != field.Tag {
									newField.Tags = strings.Split(reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get(tag), ",")
								}
								newStruct.Field = append(newStruct.Field, newField)
							}
							result = append(result, newStruct)
						}
					}
				}
			}
		}
	}

	return result, nil
}
