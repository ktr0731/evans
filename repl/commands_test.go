package repl

import "testing"

func TestValidate(t *testing.T) {
	type testCase struct {
		args   []string
		hasErr bool
	}

	type cmdTestCase struct {
		cmd       commander
		testCases []testCase
	}

	cases := map[string]cmdTestCase{
		"desc": cmdTestCase{
			cmd: &descCommand{},
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"package": cmdTestCase{
			cmd: &packageCommand{},
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"service": cmdTestCase{
			cmd: &serviceCommand{},
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"show": cmdTestCase{
			cmd: &showCommand{},
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"call": cmdTestCase{
			cmd: &callCommand{},
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"header": cmdTestCase{
			cmd: newHeaderCommand(),
			testCases: []testCase{
				{args: []string{"kumiko"}},
				{args: []string{}, hasErr: true},
			},
		},
		"exit": cmdTestCase{
			cmd: &exitCommand{},
			testCases: []testCase{
				{args: []string{}},
			},
		},
	}
	for cmdName, cmdTestCase := range cases {
		cmdTestCase := cmdTestCase
		for _, c := range cmdTestCase.testCases {
			c := c
			t.Run(cmdName, func(t *testing.T) {
				err := cmdTestCase.cmd.Validate(c.args)
				if c.hasErr {
					if err == nil {
						t.Errorf("Validate must return an error, but got nil")
					}
				} else if err != nil {
					t.Errorf("Validate must not return an error, but got '%s'", err)
				}
			})
		}
	}
}
