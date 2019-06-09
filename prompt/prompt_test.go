package prompt

import (
	"io"
	"testing"

	goprompt "github.com/ktr0731/go-prompt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

func TestPrompt_Input(t *testing.T) {
	cases := map[string]struct {
		InputFunc func(prefix string, completer goprompt.Completer, opts ...goprompt.Option) (string, error)

		expected    string
		expectedErr error
	}{
		"normal": {
			InputFunc: func(prefix string, completer goprompt.Completer, opts ...goprompt.Option) (string, error) {
				return "an input", nil
			},
			expected: "an input",
		},
		"returns io.EOF as it is if InputFunc returns io.EOF": {
			InputFunc: func(prefix string, completer goprompt.Completer, opts ...goprompt.Option) (string, error) {
				return "", io.EOF
			},
			expectedErr: io.EOF,
		},
		"returns ErrAbort as it is if InputFunc returns goprompt.ErrAbort": {
			InputFunc: func(prefix string, completer goprompt.Completer, opts ...goprompt.Option) (string, error) {
				return "", goprompt.ErrAbort
			},
			expectedErr: ErrAbort,
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			p := newPrompt()
			p.(*prompt).InputFunc = c.InputFunc
			actual, err := p.Input()
			if c.expectedErr == nil {
				if err != nil {
					t.Fatalf("Input must not return an error, but got nil")
				}
			} else {
				if err != c.expectedErr {
					t.Errorf("expected error '%s', but got '%s'", c.expectedErr, err)
				}
				return
			}

			if actual != c.expected {
				t.Errorf("expected '%s', but got '%s'", c.expected, actual)
			}
		})
	}
}

func TestPrompt_Select(t *testing.T) {
	// used for the second testcase.
	var counter int
	_ = counter

	cases := map[string]struct {
		SelectFunc func(message string, options []string) (string, error)

		expectedErr error
	}{
		"normal": {
			SelectFunc: func(message string, options []string) (string, error) {
				return "an selection", nil
			},
		},
		"retry if prompttui.ErrInterrupt is returned from SelectFunc": {
			SelectFunc: func(message string, options []string) (string, error) {
				// To escape from the loop, check whether this call is the second or not.
				if counter != 1 {
					counter++
					return "", promptui.ErrInterrupt
				}
				return "an selection", nil
			},
		},
		"returns io.EOF if promptui.ErrEOF is returned from SelectFunc": {
			SelectFunc: func(message string, options []string) (string, error) {
				return "", promptui.ErrEOF
			},
			expectedErr: io.EOF,
		},
		"returns an error if an error is returned from SelectFunc": {
			SelectFunc: func(message string, options []string) (string, error) {
				return "", io.ErrUnexpectedEOF
			},
			expectedErr: io.ErrUnexpectedEOF,
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			p := newPrompt()
			p.(*prompt).SelectFunc = c.SelectFunc
			_, err := p.Select("", []string{"foo", "bar"})
			if c.expectedErr == nil {
				if err != nil {
					t.Fatalf("Input must not return an error, but got nil")
				}
			} else {
				if errors.Cause(err) != c.expectedErr {
					t.Errorf("expected error '%s', but got '%s'", c.expectedErr, err)
				}
				return
			}
		})
	}
}

type dummyCompleter struct{}

func (c *dummyCompleter) Complete(d Document) []*Suggest {
	return []*Suggest{
		NewSuggestion("foo", ""),
		NewSuggestion("bar", ""),
	}
}

func Test_toGoPromptCompleter(t *testing.T) {
	complete := toGoPromptCompleter(&dummyCompleter{})
	doc := goprompt.NewDocument()
	suggestions := complete(*doc)
	if suggestions[0].Text != "foo" {
		t.Errorf("expected 'foo', but got '%s'", suggestions[0].Text)
	}
	if suggestions[1].Text != "bar" {
		t.Errorf("expected 'bar', but got '%s'", suggestions[1].Text)
	}
}
