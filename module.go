package main

import (
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

type Module struct {
	id     uint64
	path   string
	input  []byte
	output []byte
	hash   []byte
	files  []string
	ast    *ast.Program

	deps map[string]string
	uses map[string]bool
	refs map[string]bool
}

func newModule(id uint64, path string, hash []byte, input []byte) *Module {
	return &Module{
		id:     id,
		path:   path,
		hash:   hash,
		input:  input,
		output: input,
		deps:   make(map[string]string),
		uses:   make(map[string]bool),
		refs:   make(map[string]bool),
	}
}

func (f *Module) getAST() (*ast.Program, error) {
	if f.ast != nil {
		return f.ast, nil
	}

	p, err := parser.ParseFile(nil, f.path, string(f.input), parser.IgnoreRegExpErrors)
	if err != nil {
		return nil, err
	}

	f.ast = p

	return f.ast, nil
}

func (f *Module) ID() uint64 { return f.id }

func (f *Module) Path() string { return f.path }
func (f *Module) Size() int    { return len(f.input) }

func (f *Module) Deps() map[string]string { return f.deps }

func (f *Module) Dependencies() []string {
	p, err := f.getAST()
	if err != nil {
		panic(err)
	}

	var c dependencyCollector

	ast.Walk(&c, p)

	return c.paths
}

func (f *Module) Output() string {
	return string(f.output)
}

type dependencyCollector struct {
	paths []string
}

func (c *dependencyCollector) Enter(n ast.Node) ast.Visitor {
	ce, ok := n.(*ast.CallExpression)
	if !ok {
		return c
	}

	id, ok := ce.Callee.(*ast.Identifier)
	if !ok || id.Name != "require" {
		return c
	}

	if len(ce.ArgumentList) != 1 {
		return c
	}

	arg, ok := ce.ArgumentList[0].(*ast.StringLiteral)
	if !ok {
		return c
	}

	c.paths = append(c.paths, arg.Value)

	return c
}

func (c *dependencyCollector) Exit(n ast.Node) {}
