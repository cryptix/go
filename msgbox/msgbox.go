package msgbox

import "github.com/andlabs/ui"

// New creates a new Window and hides the parent window until the "Ok" button is pressed
// once clicked, the done channel is closed so that the calling function can continue
//
// WARNING: New can't be called by the goroutine that created the parent window
func New(p ui.Window, titel, msg string) {
	done := make(chan struct{})
	go ui.Do(func() {
		p.Hide()
		msgField := ui.NewTextField()
		msgField.SetReadOnly(true)
		msgField.SetText(msg)
		btn := ui.NewButton("Ok")
		stack := ui.NewVerticalStack(
			msgField,
			btn,
		)
		stack.SetStretchy(0)
		w := ui.NewWindow(titel, 500, 200, stack)
		btn.OnClicked(func() {
			close(done)
			w.Close()
			p.Show()
		})
		w.Show()
	})
	<-done
}
