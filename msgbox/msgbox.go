package msgbox

import "github.com/andlabs/ui"

func New(titel, msg string) <-chan struct{} {
	done := make(chan struct{})
	msgLabel := ui.NewLabel(msg)
	btn := ui.NewButton("Ok")
	btn.OnClicked(func() {
		close(done)
	})
	stack := ui.NewVerticalStack(
		msgLabel,
		btn,
	)
	stack.SetStretchy(2)
	ui.NewWindow(titel, 500, 200, stack)
	return done
}
