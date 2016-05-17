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
	Field []StructField
}

// StructField describes field itself
type StructField struct {
	Name string
	Type string
	Tags map[string]string
}

func getTypeName(t ast.Expr) string {
	switch e := t.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.ArrayType:
		return "[]" + getTypeName(e.Elt)
	case *ast.StarExpr:
		return "*" + getTypeName(e.X)
	}
	return "unknown"
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
							newStruct := StructDesc{Name: ts.Name.String(), Field: make([]StructField, 0, len(tt.Fields.List))}
							for _, field := range tt.Fields.List {
								newField := StructField{}
								if len(field.Names) < 1 {
									continue
								}
								newField.Name = field.Names[0].Name
								newField.Type = getTypeName(field.Type)
								if nil != field.Tag {
									tags := strings.Split(reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get(tag), ",")
									newField.Tags = make(map[string]string, len(tags))
									for _, tag := range tags {
										ts := strings.SplitN(tag, "=", 2)
										if len(ts) == 1 {
											newField.Tags[ts[0]] = ""
										} else {
											newField.Tags[ts[0]] = ts[1]
										}
									}
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
