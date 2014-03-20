package goparser

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
)

type ParsedSource struct {
	PackageName   string          `packageName`
	PackageDocs   string          `packageDocs`
	Types         map[string]Type `types`
	TopLevelFuncs map[string]Func `topLevelFuncs`
}

type Func struct {
	Doc  string `doc`
	Body string `body`
}

type Type struct {
	Name    string          `name`
	Body    string          `body`
	Docs    string          `docs`
	Methods map[string]Func `methods`
}

func NewParsedSource() *ParsedSource {
	return &ParsedSource{
		TopLevelFuncs: make(map[string]Func),
		Types:         make(map[string]Type),
	}
}

func (ps *ParsedSource) DumpJSON() {
	out, err := os.Create(ps.PackageName + ".json")
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(out)

	err = encoder.Encode(ps)
	if err != nil {
		log.Fatal(err)
	}

	err = out.Close()
	if err != nil {
		log.Fatal(err)
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
			parseType(parsedsource, x, fset, fi)
		case *ast.FuncDecl:
			parseFunc(parsedsource, x, fset, fi)
		}

		return true
	})

	return parsedsource
}

func parseFunc(parsedsource *ParsedSource, fdecl *ast.FuncDecl, fset *token.FileSet, fi *os.File) {
	var (
		fname      = fdecl.Name.String()
		start, end int64
		recv       string
	)

	if fdecl.Recv == nil {
		start = int64(fset.Position(fdecl.Type.Func).Offset)
		end = int64(fset.Position(fdecl.Body.End()).Offset)
		// In this case we are not a method, but a
		// top level function.
		parsedsource.TopLevelFuncs[fname] = Func{
			Doc:  fdecl.Doc.Text(),
			Body: readCodeBlock(start, end, fi),
		}
	} else {
		for _, y := range fdecl.Recv.List {
			start = int64(fset.Position(y.Type.Pos()).Offset)
			end = int64(fset.Position(y.Type.End()).Offset)
			recv = readCodeBlock(start, end, fi)
			recv = strings.Replace(recv, "*", "", -1)
		}

		start = int64(fset.Position(fdecl.Type.Func).Offset)
		end = int64(fset.Position(fdecl.Body.End()).Offset)
		parsedsource.Types[recv].Methods[fname] = Func{
			Doc:  fdecl.Doc.Text(),
			Body: readCodeBlock(start, end, fi),
		}
	}
}

func parseType(parsedsource *ParsedSource, tspec *ast.TypeSpec, fset *token.FileSet, fi *os.File) {
	start := int64(fset.Position(tspec.Pos()).Offset)
	end := int64(fset.Position(tspec.Type.End()).Offset)
	name := tspec.Name.String()
	body := "type " + readCodeBlock(start, end, fi)

	parsedsource.Types[name] = Type{
		Name:    name,
		Docs:    tspec.Doc.Text(),
		Body:    body,
		Methods: make(map[string]Func),
	}
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
