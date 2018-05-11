package e2e

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/tests/e2e/repl"
	"github.com/ktr0731/evans/tests/helper"
)

// replHelper has gateway.REPL and special fields for REPL-mode e2e testing.
// replHelper is used from TestREPL.
//
// r, w and ew are re-created at run()
// users can replace each reader/writer by any implementation for testing.
type replHelper struct {
	r     io.Reader
	w, ew io.Writer

	commonArgs []string

	iq []string

	reseted bool
}

// newREPLHelper initializes new replHelper.
// this func must be call once at each test.
func newREPLHelper(commonArgs []string) *replHelper {
	h := &replHelper{
		commonArgs: commonArgs,
		iq:         []string{},
		reseted:    true,
	}
	return h
}

func (h *replHelper) reset() {
	h.r = nil
	h.w = nil
	h.ew = nil
	h.iq = []string{}
	h.reseted = true
}

func (h *replHelper) registerInput(in ...repl.CmdAndArgs) {
	for i := range in {
		h.iq = append(h.iq, in[i]()...)
	}
}

func (h *replHelper) run(args []string) int {
	if !h.reseted {
		panic("must be call reset() before each run()")
	}
	old := controller.DefaultREPLUI
	defer func() {
		controller.DefaultREPLUI = old
	}()

	if h.r == nil {
		h.r = os.Stdin
	}
	if h.w == nil {
		h.w = ioutil.Discard
	}
	if h.ew == nil {
		// h.ew = ioutil.Discard
		h.ew = os.Stderr
	}

	controller.DefaultREPLUI = &controller.REPLUI{
		UI: controller.NewUI(h.r, h.w, h.ew),
	}

	h.iq = append(h.iq, repl.Exit()...)
	p := helper.NewMockPrompt(h.iq, []string{})
	cleanup := SetPrompt(p)
	defer cleanup()

	h.reseted = false

	return newCLI(controller.NewUI(os.Stdin, ioutil.Discard, ioutil.Discard)).
		Run(append(h.commonArgs, args...))
}
