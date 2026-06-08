// Package tuigen compiles .t2 template files into type-safe Go code.
//
// The pipeline consists of:
//   - [Lexer]: tokenizes .t2 source into a token stream
//   - [Parser]: builds an AST from the token stream
//   - [Analyzer]: performs semantic analysis (imports, refs, state bindings)
//   - [Generator]: emits Go source code from the analyzed AST
//
// Additionally, [ParseTailwindClasses] translates Tailwind-style class strings
// into tui element options.
package tuigen
