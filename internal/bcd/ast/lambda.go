package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/translator"
)

// Lambda -
type Lambda struct {
	Default
	InputType  Node
	ReturnType Node
}

// NewLambda -
func NewLambda(depth int) *Lambda {
	return &Lambda{
		Default: NewDefault(consts.LAMBDA, -1, depth),
	}
}

// MarshalJSON -
func (l *Lambda) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.LAMBDA, l.Annots, l.InputType, l.ReturnType)
}

// ParseType -
func (l *Lambda) ParseType(node *base.Node, id *int) error {
	if err := l.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], l.Depth, id)
	if err != nil {
		return err
	}
	l.InputType = typ

	retTyp, err := typingNode(node.Args[1], l.Depth, id)
	if err != nil {
		return err
	}
	l.ReturnType = retTyp
	return nil
}

// ParseValue -
func (l *Lambda) ParseValue(node *base.Node) error {
	str, err := json.MarshalToString(node.Args)
	if err != nil {
		return err
	}
	l.Value = str
	return nil
}

// ToBaseNode -
func (l *Lambda) ToBaseNode(optimized bool) (*base.Node, error) {
	var lambda base.Node
	if err := json.UnmarshalFromString(l.Value.(string), &lambda); err != nil {
		return nil, err
	}
	return &lambda, nil
}

// ToMiguel -
func (l *Lambda) ToMiguel() (*MiguelNode, error) {
	var formatted string
	if s, ok := l.Value.(string); ok {
		val, err := formatter.MichelineStringToMichelson(s, false, formatter.DefLineSize)
		if err != nil {
			return nil, err
		}
		formatted = val
	}
	name := l.GetTypeName()
	return &MiguelNode{
		Value: formatted,
		Type:  l.Prim,
		Prim:  l.Prim,
		Name:  &name,
	}, nil
}

// FromJSONSchema -
func (l *Lambda) FromJSONSchema(data map[string]interface{}) error {
	for key := range data {
		if l.GetTypeName() == key {
			t, err := translator.NewConverter()
			if err != nil {
				return err
			}
			jsonLambda, err := t.FromString(data[key].(string))
			if err != nil {
				return err
			}
			l.Value = jsonLambda
			l.ValueKind = valueKindString
		}
	}
	return nil
}

// ToJSONSchema -
func (l *Lambda) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(l.Default), nil
}

// ToParameters -
func (l *Lambda) ToParameters() ([]byte, error) {
	return []byte(l.Value.(string)), nil
}

// Docs -
func (l *Lambda) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(l, inferredName)
	typedef := Typedef{
		Name: name,
		Type: fmt.Sprintf("lambda(%s, %s)", l.InputType.GetPrim(), l.ReturnType.GetPrim()),
		Args: make([]TypedefArg, 0),
	}
	if isSimpleDocType(l.InputType.GetPrim()) && isSimpleDocType(l.ReturnType.GetPrim()) {
		return []Typedef{typedef}, typedef.Type, nil
	}

	iStr, err := json.MarshalToString(l.InputType)
	if err != nil {
		return nil, "", err
	}
	parameter, err := formatter.MichelineStringToMichelson(iStr, true, formatter.DefLineSize)
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: "input", Value: parameter})

	rStr, err := json.MarshalToString(l.ReturnType)
	if err != nil {
		return nil, "", err
	}
	returnValue, err := formatter.MichelineStringToMichelson(rStr, true, formatter.DefLineSize)
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: "return", Value: returnValue})

	typedef.Type = consts.LAMBDA
	return []Typedef{typedef}, makeVarDocString(name), nil
}

// Distinguish -
func (l *Lambda) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Lambda)
	if !ok {
		return nil, nil
	}
	curr, err := l.ToMiguel()
	if err != nil {
		return nil, err
	}
	prev, err := second.ToMiguel()
	if err != nil {
		return nil, err
	}
	if prev.Value != curr.Value {
		curr.From = prev.Value

		switch {
		case curr.Value == "":
			curr.DiffType = MiguelKindDelete
		case prev.Value == "":
			curr.DiffType = MiguelKindCreate
		default:
			curr.DiffType = MiguelKindUpdate
		}
	}

	return curr, nil

}

// EqualType -
func (l *Lambda) EqualType(node Node) bool {
	if !l.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Lambda)
	if !ok {
		return false
	}
	if !l.InputType.EqualType(second.InputType) {
		return false
	}

	return l.ReturnType.EqualType(second.ReturnType)
}
