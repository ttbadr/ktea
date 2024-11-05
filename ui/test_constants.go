package ui

import "ktea/kontext"

var TestKontext = &kontext.ProgramKtx{
	Config:          nil,
	WindowWidth:     100,
	WindowHeight:    100,
	AvailableHeight: 100,
}

var TestRenderer = NewRenderer(TestKontext)
