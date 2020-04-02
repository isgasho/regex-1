package syntax

import (
	"fmt"
	"strings"
)

type Regexp struct {
	Source string
	Expr   Expr
}

func (re *Regexp) ExprString(e Expr) string {
	return re.Source[e.Begin():e.End()]
}

type Expr struct {
	Pos  Position
	Op   Operation
	Args []Expr
}

// Begin returns expression leftmost offset.
func (e Expr) Begin() uint16 { return e.Pos.Begin }

// End returns expression rightmost offset.
func (e Expr) End() uint16 { return e.Pos.End }

// LastArg returns expression last argument.
//
// Should not be called on expressions that may have 0 arguments.
func (e Expr) LastArg() Expr {
	return e.Args[len(e.Args)-1]
}

type Operation byte

func FormatSyntax(re *Regexp) string {
	return formatExprSyntax(re, re.Expr)
}

func formatExprSyntax(re *Regexp, e Expr) string {
	switch e.Op {
	case OpChar, OpLiteral:
		s := re.ExprString(e)
		switch s {
		case "{":
			return "'{'"
		case "}":
			return "'}'"
		default:
			return s
		}
	case OpString, OpEscape, OpEscapeMeta, OpEscapeOctal, OpEscapeUni, OpEscapeUniFull, OpEscapeHex, OpEscapeHexFull, OpPosixClass:
		return re.ExprString(e)
	case OpRepeat:
		return fmt.Sprintf("(repeat %s %s)", formatExprSyntax(re, e.Args[0]), re.ExprString(e.Args[1]))
	case OpCaret:
		return "^"
	case OpDollar:
		return "$"
	case OpDot:
		return "."
	case OpQuote:
		return fmt.Sprintf("(q %s)", re.ExprString(e))
	case OpCharRange:
		return fmt.Sprintf("%s-%s", formatExprSyntax(re, e.Args[0]), formatExprSyntax(re, e.Args[1]))
	case OpCharClass:
		return fmt.Sprintf("[%s]", formatArgsSyntax(re, e.Args))
	case OpNegCharClass:
		return fmt.Sprintf("[^%s]", formatArgsSyntax(re, e.Args))
	case OpConcat:
		return fmt.Sprintf("{%s}", formatArgsSyntax(re, e.Args))
	case OpAlt:
		return fmt.Sprintf("(or %s)", formatArgsSyntax(re, e.Args))
	case OpCapture:
		return fmt.Sprintf("(capture %s)", formatExprSyntax(re, e.Args[0]))
	case OpNamedCapture:
		return fmt.Sprintf("(capture %s %s)", formatExprSyntax(re, e.Args[0]), re.ExprString(e.Args[1]))
	case OpGroup:
		return fmt.Sprintf("(group %s)", formatExprSyntax(re, e.Args[0]))
	case OpGroupWithFlags:
		return fmt.Sprintf("(group %s %s)", formatExprSyntax(re, e.Args[0]), re.ExprString(e.Args[1]))
	case OpFlagOnlyGroup:
		return fmt.Sprintf("(flags %s)", formatExprSyntax(re, e.Args[0]))
	case OpPlus:
		return fmt.Sprintf("(+ %s)", formatExprSyntax(re, e.Args[0]))
	case OpStar:
		return fmt.Sprintf("(* %s)", formatExprSyntax(re, e.Args[0]))
	case OpQuestion:
		return fmt.Sprintf("(? %s)", formatExprSyntax(re, e.Args[0]))
	case OpNonGreedy:
		return fmt.Sprintf("(non-greedy %s)", formatExprSyntax(re, e.Args[0]))
	default:
		return fmt.Sprintf("<op=%d>", e.Op)
	}
}

func formatArgsSyntax(re *Regexp, args []Expr) string {
	parts := make([]string, len(args))
	for i, e := range args {
		parts[i] = formatExprSyntax(re, e)
	}
	return strings.Join(parts, " ")
}

//go:generate stringer -type=Operation -trimprefix=Op
const (
	OpNone Operation = iota

	// OpConcat is a concatenation of ops.
	// Examples: `xy` `abc\d` ``
	// Args - concatenated ops
	//
	// As a special case, OpConcat with 0 Args is used for "empty"
	// set of operations.
	OpConcat // xy concatenation of ops

	// OpDot is a '.' wildcard.
	OpDot

	// OpAlt is x|y alternation of ops.
	// Examples: `a|bc` `x(.*?)|y(.*?)`
	// Args - union-connected regexp branches
	OpAlt

	// OpStar is a shorthand for {0,} repetition.
	// Examples: `x*`
	// Args[0] - repeated expression
	OpStar

	// OpPlus is a shorthand for {1,} repetition.
	// Examples: `x+`
	// Args[0] - repeated expression
	OpPlus

	// OpQuestion is a shorthand for {0,1} repetition.
	// Examples: `x?`
	// Args[0] - repeated expression
	OpQuestion // x? zero or one x

	// OpNonGreedy makes its operand quantifier non-greedy.
	// Examples: `x??` `x*?` `x+?`
	// Args[0] - quantified expression
	OpNonGreedy

	// OpCaret is ^ anchor.
	OpCaret

	// OpDollar is $ anchor.
	OpDollar

	// OpLiteral is a collection of consecutive chars.
	// Examples: `ab` `10x`
	// Args - enclosed characters (OpChar)
	OpLiteral

	// OpChar is a single literal pattern character.
	// Examples: `a` `6` `ф`
	OpChar

	// OpString is an artificial element that is used in other expressions.
	OpString

	// OpQuote is a \Q...\E enclosed literal.
	// Examples: `\Q.?\E` `\Q?q[]=1`
	//
	// Note that closing \E is not mandatory.
	OpQuote

	// OpEscape is a single characted escape.
	// Examples: `\d` `\a` `\n`
	OpEscape

	// OpEscapeMeta is an escaped meta char.
	// Examples: `\(` `\[` `\+`
	OpEscapeMeta

	// OpEscapeOctal is an octal char code escape (up to 3 digits).
	// Examples: `\123` `\12`
	OpEscapeOctal

	// OpEscapeHex is a hex char code escape (exactly 2 digits).
	// Examples: `\x7F` `\xF7`
	OpEscapeHex

	// OpEscapeHexFull is a hex char code escape.
	// Examples: `\x{10FFFF}` `\x{F}`
	OpEscapeHexFull

	// OpEscapeUni is a Unicode char class escape (one-letter name).
	// Examples: `\pS` `\pL` `\PL`
	OpEscapeUni

	// OpEscapeUniFull is a Unicode char class escape.
	// Example: `\p{Greek}` `\p{Symbol}` `\p{^L}`
	OpEscapeUniFull

	// OpCharClass is a char class enclosed in [].
	// Examples: `[abc]` `[a-z0-9\]]`
	// Args - char class elements (can include OpCharRange and OpPosixClass).
	OpCharClass

	// OpNegCharClass is a negated char class enclosed in [].
	// Examples: `[^abc]` `[^a-z0-9\]]`
	// Args - char class elements (can include OpCharRange and OpPosixClass).
	OpNegCharClass

	// OpCharRange is an inclusive char range inside a char class.
	// Examples: `0-9` `A-Z`
	// Args[0] - range lower bound (OpChar or OpEscape).
	// Args[1] - range upper bound (OpChar or OpEscape).
	OpCharRange

	// OpPosixClass is a named ASCII char set inside a char class.
	// Examples: `[:alpha:]` `[:blank:]`
	OpPosixClass

	// OpRepeat is a {min,max} repetition quantifier.
	// Examples: `x{5}` `x{min,max}` `x{min,}`
	// Args[0] - repeated expression
	// Args[1] - repeat count (OpString)
	OpRepeat

	// OpCapture is `(re)` capturing group.
	// Examples: `(abc)` `(x|y)`
	// Args[0] - enclosed expression
	OpCapture

	// OpNamedCapture is `(?P<name>re)` capturing group.
	// Examples: `(?P<foo>abc)` `(?P<name>x|y)`
	// Args[0] - enclosed expression (OpConcat with 0 args for empty group)
	// Args[1] - group name (OpString)
	OpNamedCapture

	// OpGroup is `(?:re)` non-capturing group.
	// Examples: `(?:abc)` `(?:x|y)`
	// Args[0] - enclosed expression (OpConcat with 0 args for empty group)
	OpGroup

	// OpGroupWithFlags is `(?flags:re)` non-capturing group.
	// Examples: `(?i:abc)` `(?i:x|y)`
	// Args[0] - enclosed expression (OpConcat with 0 args for empty group)
	// Args[1] - flags (OpString)
	OpGroupWithFlags

	// OpFlagOnlyGroup is `(?flags)` form that affects current group flags.
	// Examples: `(?i)` `(?i-m)` `(?-im)`
	// Args[0] - flags (OpString)
	OpFlagOnlyGroup

	// OpNone2 is a sentinel value that is never part of the AST.
	// OpNone and OpNone2 can be used to cover all ops in a range.
	OpNone2
)
