package goparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
)

type ParsedSource struct {
	PackageName   string
	PackageDocs   string
	Types         map[string]Type
	TopLevelFuncs map[string]Func
}

type Func struct {
	Doc  string
	Body string
}

type Type struct {
	Name    string
	Body    string
	Docs    string
	Methods map[string]Func
}

func NewParsedSource() *ParsedSource {
	return &ParsedSource{
		TopLevelFuncs: make(map[string]Func),
		Types:         make(map[string]Type),
	}
}

func Parse(filename string) *ParsedSource {
	var (
		parsedsource = NewParsedSource()
		fset         = token.NewFileSet()
	)

	fi, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	parsedsource.PackageName = f.Name.String()
	parsedsource.PackageDocs = f.Doc.Text()

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			start := int64(fset.Position(x.Pos()).Offset)
			end := int64(fset.Position(x.Type.End()).Offset)
			name := x.Name.String()
			body := "type " + readCodeBlock(start, end, fi)

			parsedsource.Types[name] = Type{
				Name:    name,
				Docs:    x.Doc.Text(),
				Body:    body,
				Methods: make(map[string]Func),
			}
		case *ast.FuncDecl:
			var (
				fname      = x.Name.String()
				start, end int64
				recv       string
			)

			if x.Recv == nil {
				start = int64(fset.Position(x.Type.Func).Offset)
				end = int64(fset.Position(x.Body.End()).Offset)
				// In this case we are not a method, but a
				// top level function.
				parsedsource.TopLevelFuncs[fname] = Func{
					Doc:  x.Doc.Text(),
					Body: readCodeBlock(start, end, fi),
				}
			} else {
				for _, y := range x.Recv.List {
					start = int64(fset.Position(y.Type.Pos()).Offset)
					end = int64(fset.Position(y.Type.End()).Offset)
					recv = readCodeBlock(start, end, fi)
					recv = strings.Replace(recv, "*", "", -1)
				}

				start = int64(fset.Position(x.Type.Func).Offset)
				end = int64(fset.Position(x.Body.End()).Offset)
				parsedsource.Types[recv].Methods[fname] = Func{
					Doc:  x.Doc.Text(),
					Body: readCodeBlock(start, end, fi),
				}
			}
		}

		return true
	})

	return parsedsource
}

func readCodeBlock(start, end int64, f *os.File) string {
	var (
		delta = end - start
		body  = make([]byte, delta)
		srdr  = io.NewSectionReader(f, start, delta)
	)

	srdr.Read(body)
	return string(body)
}
