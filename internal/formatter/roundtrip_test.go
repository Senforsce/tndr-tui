package formatter

import (
	"testing"
)

func TestFormat_Idempotent(t *testing.T) {
	type tc struct {
		input string
	}

	tests := map[string]tc{
		"simple component": {
			input: `package test

t1 Hello() {
	<div class="flex-col">
		<span>Hello</span>
	</div>
}
`,
		},
		"component with imports": {
			input: `package test

import "fmt"

t1 Hello(name string) {
	<span>{fmt.Sprintf("Hello %s", name)}</span>
}
`,
		},
		"if else": {
			input: `package test

t1 Cond(show bool) {
	<div>
		if show {
			<span>Yes</span>
		} else {
			<span>No</span>
		}
	</div>
}
`,
		},
		"for loop": {
			input: `package test

t1 List(items []string) {
	<div class="flex-col">
		for _, item := range items {
			<span>{item}</span>
		}
	</div>
}
`,
		},
		"self-closing elements": {
			input: `package test

t1 Divider() {
	<div>
		<hr />
		<br />
	</div>
}
`,
		},
		"let binding": {
			input: `package test

import "fmt"

t1 Counter(count int) {
	countText := <span>{fmt.Sprintf("Count: %d", count)}</span>
	<div>{countText}</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := New()

			// First format
			result1, err := f.Format("test.t2", tt.input)
			if err != nil {
				t.Fatalf("first format failed: %v", err)
			}

			// Second format (should be identical)
			result2, err := f.Format("test.t2", result1)
			if err != nil {
				t.Fatalf("second format failed: %v", err)
			}

			if result1 != result2 {
				t.Errorf("format is not idempotent\nfirst:\n%s\nsecond:\n%s", result1, result2)
			}
		})
	}
}
