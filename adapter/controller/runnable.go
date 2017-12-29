package controller

type Runnable interface {
	Run(args []string) int
}
