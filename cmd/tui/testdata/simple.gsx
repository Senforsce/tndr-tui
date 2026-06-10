package testdata

import tui "github.com/senforsce/tndr-tui"

t1 Header(title string) {
	<div border={tui.BorderSingle} padding={1}>
		<span>{title}</span>
	</div>
}

t1 Footer() {
	<div padding={1}>
		<span>Footer content</span>
	</div>
}
