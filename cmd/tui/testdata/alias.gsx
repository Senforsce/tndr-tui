package testdata

import (
	"fmt"

	gui "github.com/senforsce/tndr-tui"
)

type aliasHelper struct{}

func (foo aliasHelper) State() string {
	return "not tui state"
}

t1 Alias(counter *gui.State[int]) {
	label := gui.NewState("alias state")
	foo := aliasHelper{}

	<div border={gui.BorderSingle} padding={1}>
		<span>{label.Get()}</span>
		<span>{fmt.Sprintf("%d", counter.Get())}</span>
		<span>{foo.State()}</span>
	</div>
}
