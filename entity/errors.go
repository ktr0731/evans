package entity

type warn interface {
	isWarn() bool
}

func IsWarn(err error) bool {
	_, ok := err.(warn)
	return ok
}
