package builder

import "fmt"

type UI interface {
	Write(p []byte) (n int, err error)
	Print(string)
	Printf(string, ...interface{})
}

type ConsolUI struct {
}

func (ui *ConsolUI) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

func (ui *ConsolUI) Print(s string) {
	fmt.Println(s)
}

func (ui *ConsolUI) Printf(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}
