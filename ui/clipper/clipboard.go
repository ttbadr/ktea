package clipper

import "github.com/atotto/clipboard"

type Writer interface {
	Write(text string) error
}

type DefaultClipper struct {
}

func (d *DefaultClipper) Write(text string) error {
	err := clipboard.WriteAll(text)
	if err != nil {
		return err
	}
	return nil
}

func New() Writer {
	return &DefaultClipper{}
}
