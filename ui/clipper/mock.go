package clipper

type WriteFunc func(text string) error

type MockClipper struct {
	WriteFunc WriteFunc
}

func (d *MockClipper) Write(text string) error {
	if d.WriteFunc != nil {
		return d.WriteFunc(text)
	}
	return nil
}

func NewMock() *MockClipper {
	return &MockClipper{}
}
