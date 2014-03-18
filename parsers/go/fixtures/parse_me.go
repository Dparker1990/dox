// parseme package
// double line
package parseme

type Fuz interface {
	Kuz()
}

type Buz struct {
	mer string
	fer []byte
}

// recv comments
func (b *Buz) meth() {
	println(b.mer)
}

// testing some
// comments
func foo() {
	println("bar")
}

// calls foo
func bar() {
	foo()
}

// prints what is passed in.
func baz(str string) {
	println(str)
}
