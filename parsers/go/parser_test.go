package goparser

import (
	"os"
	"testing"
)

func TestParsePackage(t *testing.T) {
	p := Parse("fixtures/parse_me.go")
	expectedDoc := "parseme package\ndouble line\n"

	if p.PackageDocs != expectedDoc {
		t.Errorf("Did not parse package docs, got: %#v", p.PackageDocs)
	}

	if p.PackageName != "parseme" {
		t.Errorf("Did not parse package name")
	}
}

func TestParseTypes(t *testing.T) {
	p := Parse("fixtures/parse_me.go")
	if len(p.Types) != 2 {
		t.Error("No types parsed")
	}

	expected := []string{"Fuz", "Buz"}
	for _, v := range expected {
		if _, ok := p.Types[v]; !ok {
			t.Error("Did not parse type correctly")
		}
	}
}

func TestParseTopLevelFuncs(t *testing.T) {
	p := Parse("fixtures/parse_me.go")
	if len(p.TopLevelFuncs) != 3 {
		t.Error("Expected 2 top level funcs")
	}

	expectedTopLevelFuncs := []struct {
		name string
		docs string
		body string
	}{
		{"foo", "testing some\ncomments\n", "func foo() {\n\tprintln(\"bar\")\n}"},
		{"bar", "calls foo\n", "func bar() {\n\tfoo()\n}"},
		{"baz", "prints what is passed in.\n", "func baz(str string) {\n\tprintln(str)\n}"},
	}

	for _, f := range expectedTopLevelFuncs {
		sut := p.TopLevelFuncs[f.name]

		if sut.Doc != f.docs {
			t.Errorf("Body did not match, got: %#v", sut.Doc)
		}

		if sut.Body != f.body {
			t.Errorf("Body did not match, got: %#v", sut.Body)
		}
	}
}

func TestParsingRecvFuncs(t *testing.T) {
	p := Parse("fixtures/parse_me.go")
	expectedRecvFuncs := []struct {
		name string
		recv string
		docs string
		body string
	}{
		{"meth", "Buz", "recv comments\n", "func (b *Buz) meth() {\n\tprintln(b.mer)\n}"},
	}

	for _, f := range expectedRecvFuncs {
		sut := p.Types[f.recv].Methods[f.name]

		if sut.Doc != f.docs {
			t.Errorf("Body did not match, got: %#v", sut.Doc)
		}

		if sut.Body != f.body {
			t.Errorf("Body did not match, got: %#v", sut.Body)
		}
	}
}

func TestDumpJSON(t *testing.T) {
	p := Parse("fixtures/parse_me.go")
	p.DumpJSON()

	f, err := os.Open("parseme.json")
	if err != nil {
		t.Error(err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}

	if fi.Size() == 0 {
		t.Error("Json was not dumped")
	}

	err = f.Close()
	if err != nil {
		t.Error(err)
	}

	err = os.Remove("parseme.json")
	if err != nil {
		t.Error(err)
	}
}
