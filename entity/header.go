package entity

type Header struct {
	Key, Val string

	// NeedToRemove is used only from REPL header command
	NeedToRemove bool
}
