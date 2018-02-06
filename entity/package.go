package entity

type Package struct {
	Name     string
	Services []Service
	Messages []Message
}

func NewPackage(name string, msgs []Message, svcs []Service) *Package {
	return &Package{
		Name:     name,
		Services: svcs,
		Messages: msgs,
	}
}

// GetMessage is only used to get a root message
func (p *Package) GetMessage(name string) (Message, error) {
	// nested message は辿る必要がない
	for _, msg := range p.Messages {
		if msg.Name() == name {
			return msg, nil
		}
	}
	return nil, ErrInvalidMessageName
}

type Packages []*Package
