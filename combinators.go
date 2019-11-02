package pars

import (
	"fmt"
	"reflect"
	"strings"
)

func typeRep(q ParserLike) string {
	return reflect.TypeOf(q).String()
}

func typeReps(qs []ParserLike) []string {
	r := make([]string, len(qs))
	for i, q := range qs {
		r[i] = typeRep(q)
	}
	return r
}

func Seq(qs ...ParserLike) Parser {
	ps := AsParsers(qs...)
	name := fmt.Sprintf("Seq(%s)", strings.Join(typeReps(qs), ", "))

	return func(state *State, result *Result) error {
		state.Push()
		v := make([]Result, len(ps))
		for i, p := range ps {
			if err := p(state, &v[i]); err != nil {
				state.Pop()
				return NewTraceError(name, err)
			}
		}
		state.Drop()
		result.SetChildren(v)
		return nil
	}
}

func Any(qs ...ParserLike) Parser {
	ps := AsParsers(qs...)
	name := fmt.Sprintf("Any(%s)", strings.Join(typeReps(qs), ", "))

	return func(state *State, result *Result) error {
		state.Push()
		for _, p := range ps {
			if p(state, result) == nil {
				state.Drop()
				return nil
			}
		}
		state.Pop()
		return NewParserError(name, state.Position())
	}
}

func Maybe(q ParserLike) Parser {
	p := AsParser(q)

	return func(state *State, result *Result) error {
		state.Push()
		if p(state, result) != nil {
			state.Pop()
			return nil
		}
		state.Drop()
		return nil
	}
}

func Many(q ParserLike) Parser {
	p := AsParser(q)
	name := fmt.Sprintf("Many(%s)", typeRep(q))
	fmt.Println(name)

	return func(state *State, result *Result) error {
		v := make([]Result, 1)

		state.Push()
		if err := p(state, &v[0]); err != nil {
			state.Pop()
			return NewTraceError(name, err)
		}
		state.Drop()

		tmp := Result{}

		state.Push()
		for p(state, &tmp) == nil {
			state.Drop()
			state.Push()
			v = append(v, tmp)
			tmp = Result{}
		}
		state.Pop()

		result.SetChildren(v)
		return nil
	}
}
