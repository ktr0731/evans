package helper

import prompt "github.com/c-bata/go-prompt"

type MockPrompt struct {
	iq []string
	sq []string

	repeatedCnt int
}

func NewMockPrompt(iq, sq []string) *MockPrompt {
	return &MockPrompt{
		iq: iq,
		sq: sq,
	}
}

func (p *MockPrompt) Input() string {
	in := p.iq[0]
	if len(p.iq) > 1 {
		p.iq = p.iq[1:]
	}
	return in
}

func (p *MockPrompt) Select(_ string, _ []string) (string, error) {
	in := p.sq[0]
	if len(p.sq) > 1 {
		p.sq = p.sq[1:]
	}
	return in, nil
}

func (p *MockPrompt) SetPrefix(_ string) {}

func (p *MockPrompt) SetPrefixColor(_ prompt.Color) error {
	return nil
}
