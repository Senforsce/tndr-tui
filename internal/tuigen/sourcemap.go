package tuigen

// SourceMap tracks position mappings between .t2 source and generated .go files.
// All line numbers are 0-indexed.
type SourceMap struct {
	// SourceFile is the original .t2 file path
	SourceFile string `json:"sourceFile"`

	// Mappings contains position mappings from .go to .t2
	Mappings []SourceMapping `json:"mappings"`
}

// SourceMapping represents a single line/column mapping.
type SourceMapping struct {
	// GoLine is the line in the generated .go file (0-indexed)
	GoLine int `json:"goLine"`
	// GoCol is the column in the generated .go file (0-indexed)
	GoCol int `json:"goCol"`
	// T2Line is the line in the source .t2 file (0-indexed)
	T2Line int `json:"t2Line"`
	// T2Col is the column in the source .t2 file (0-indexed)
	T2Col int `json:"t2Col"`
	// Length is the length of the mapped region
	Length int `json:"length"`
}

// NewSourceMap creates a new empty source map.
func NewSourceMap(sourceFile string) *SourceMap {
	return &SourceMap{
		SourceFile: sourceFile,
		Mappings:   make([]SourceMapping, 0),
	}
}

// AddMapping adds a new position mapping.
func (sm *SourceMap) AddMapping(m SourceMapping) {
	sm.Mappings = append(sm.Mappings, m)
}

// GoToT2 converts a .go position to a .t2 position.
// Returns the translated position and true if found, otherwise returns
// the input position and false.
func (sm *SourceMap) GoToT2(goLine, goCol int) (t2Line, t2Col int, found bool) {
	for _, m := range sm.Mappings {
		if m.GoLine == goLine && goCol >= m.GoCol && goCol <= m.GoCol+m.Length {
			offset := goCol - m.GoCol
			return m.T2Line, m.T2Col + offset, true
		}
	}
	return goLine, goCol, false
}

// T2ToGo converts a .t2 position to a .go position.
// Returns the translated position and true if found.
func (sm *SourceMap) T2ToGo(t2Line, t2Col int) (goLine, goCol int, found bool) {
	for _, m := range sm.Mappings {
		if m.T2Line == t2Line && t2Col >= m.T2Col && t2Col <= m.T2Col+m.Length {
			offset := t2Col - m.T2Col
			return m.GoLine, m.GoCol + offset, true
		}
	}
	return t2Line, t2Col, false
}
