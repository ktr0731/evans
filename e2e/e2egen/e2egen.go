// Package main provides a generator for E2E test cases.
//
// e2egen generates a new testcase stub from execution result.
// e2egen accepts following args.
//
//   e2egen <file> <Evans flags>
//
// file is the source code of REPL E2E test. Updated code is written to it.
// Evans flags are options of Evans.
//
// First, e2egen tries to read source code from file.
// Second, e2egen launches Evans with Evans flags. In this time, e2egen records
// all input from the prompt internally.
// Third, e2egen generates updated source code that appended the new testcase.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	goprompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/prompt"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	if len(os.Args) < 3 || strings.HasPrefix(os.Args[1], "-") {
		fmt.Println("Usage: e2egen <file> <Evans flags>")
		os.Exit(1)
	}

	fileName, args := os.Args[1], os.Args[2:]

	src, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read the file: %s", err)
		os.Exit(1)
	}

	f, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open the file: %s", err)
		os.Exit(1)
	}
	defer f.Close()

	p := &recorderPrompt{Prompt: prompt.New()}
	prompt.New = func() prompt.Prompt {
		return p
	}
	code := app.New(nil).Run(args)

	if code != 0 {
		os.Exit(code)
	}

	testCaseName := goprompt.Input("testcase name: ", func(goprompt.Document) []goprompt.Suggest { return nil })
	if testCaseName == "" {
		fmt.Println("abort")
		f.Write(src)
		os.Exit(1)
	}

	generateFile(f, string(src), testCaseName, p.inputHistory, args)

	os.Exit(code)
}

func generateFile(w io.Writer, src, testCaseName string, input, args []string) error {
	args = filterArgs(args)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return errors.Wrap(err, "failed to parse source test file")
	}

	inputExprs := make([]ast.Expr, len(input))
	for i, s := range input {
		inputExprs[i] = &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(s),
		}
	}

	astutil.Apply(f, func(cr *astutil.Cursor) bool {
		if cr.Name() != "Rhs" {
			return true
		}

		compLit, ok := cr.Node().(*ast.CompositeLit)
		if !ok {
			return true
		}

		as, ok := cr.Parent().(*ast.AssignStmt)
		if !ok {
			return true
		}

		ident, ok := as.Lhs[0].(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name != "cases" {
			return true
		}

		// Add a large value.
		file := fset.File(1)
		lastElt := compLit.Elts[len(compLit.Elts)-1].(*ast.KeyValueExpr).Value.(*ast.CompositeLit)
		nextLine := file.LineStart(fset.Position(lastElt.Rbrace).Line + 1)

		l1 := file.LineStart(fset.Position(nextLine).Line + 1)
		l2 := file.LineStart(fset.Position(nextLine).Line + 2)
		l3 := file.LineStart(fset.Position(nextLine).Line + 3)
		l4 := file.LineStart(fset.Position(nextLine).Line + 4)
		l5 := file.LineStart(fset.Position(nextLine).Line + 5)

		elts := append(compLit.Elts, &ast.KeyValueExpr{
			Key: &ast.BasicLit{
				ValuePos: l1,
				Kind:     token.STRING,
				Value:    strconv.Quote(testCaseName),
			},
			Colon: l1,
			Value: &ast.CompositeLit{
				Lbrace: l1,
				Elts: []ast.Expr{
					&ast.KeyValueExpr{
						Key: &ast.Ident{
							NamePos: l2,
							Name:    "args",
						},
						Value: &ast.BasicLit{
							ValuePos: l2,
							Kind:     token.STRING,
							Value:    strconv.Quote(strings.Join(args, " ")),
						},
					},
					&ast.KeyValueExpr{
						Key: &ast.Ident{
							NamePos: l3,
							Name:    "input",
						},
						Value: &ast.CompositeLit{
							Lbrace: l3,
							Type: &ast.ArrayType{
								Lbrack: l3,
								Elt: &ast.InterfaceType{
									Interface: l3,
									Methods: &ast.FieldList{
										Opening: l3,
										Closing: l3,
									},
								},
							},
							Elts:   inputExprs,
							Rbrace: l3,
						},
					},
				},
				Rbrace: l4,
			},
		})
		compLit.Rbrace = l5
		compLit.Elts = elts
		return false
	}, nil)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	b, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to format modified source code")
	}
	w.Write(b)
	return nil
}

type recorderPrompt struct {
	prompt.Prompt

	inputHistory []string
}

func (p *recorderPrompt) Input() (string, error) {
	s, err := p.Prompt.Input()
	if err == nil {
		p.inputHistory = append(p.inputHistory, s)
	}
	return s, err
}

func (p *recorderPrompt) Select(message string, options []string) (string, error) {
	s, err := p.Prompt.Select(message, options)
	if err == nil {
		p.inputHistory = append(p.inputHistory, s)
	}
	return s, err
}

func filterArgs(args []string) []string {
	newArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--port" {
			// Skip arg of --port and itself.
			i++
			continue
		}
		newArgs = append(newArgs, args[i])
	}
	return newArgs
}
