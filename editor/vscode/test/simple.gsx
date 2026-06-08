// Simple GSX component example
// Tests basic syntax highlighting

package example

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

t1 Header(title string) {
	<div class="border-single p-1">
		<span class="font-bold">{title}</span>
	</div>
}

t1 Footer() {
	<div class="p-1">
		<span>Footer content</span>
	</div>
}

t1 SimpleCard(title string, content string) {
	<div class="border-rounded">
		<span class="font-bold">{title}</span>
		<span>{content}</span>
	</div>
}

// Ref attribute example
t1 Layout(title string) {
	main := tui.NewRef()
	titleRef := tui.NewRef()
	<div ref={main} class="flex-col gap-1">
		<span ref={titleRef} class="font-bold">{title}</span>
		<span>Body content</span>
	</div>
}

// State variable example
t1 Counter() {
	count := tui.NewState(0)
	<div class="flex-col gap-1 p-1 border-single">
		<span>{fmt.Sprintf("Count: %d", count.Get())}</span>
	</div>
}
