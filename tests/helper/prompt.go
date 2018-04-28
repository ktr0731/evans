package helper

import (
	prompt "github.com/c-bata/go-prompt"
)

type MockPrompt struct {
	iq []string
	sq []string
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

type MockRepeatedPrompt struct {
	*MockPrompt

	iq [][]string
}

func NewMockRepeatedPrompt(iq [][]string, sq []string) *MockRepeatedPrompt {
	return &MockRepeatedPrompt{
		MockPrompt: &MockPrompt{
			sq: sq,
		},
		iq: iq,
	}
}

func (p *MockRepeatedPrompt) Input() string {
	line := p.iq[0]
	if len(line) == 0 {
		p.iq = p.iq[1:]
	}
	in := p.iq[0][0]
	if len(p.iq[0]) > 1 {
		p.iq[0] = p.iq[0][1:]
	}
	return in
}
