package testdata

type docView struct {
	readme string
}

func DocView(readme string) *docView {
	return &docView{readme: readme}
}

t1 (c *docView) Render() {
	<div class="flex-col overflow-y-scroll">
		<markdown source={c.readme} width={80} />
	</div>
}
