package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/translator"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/tidwall/gjson"
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
	return marshalJSON(consts.LAMBDA, l.annots, l.InputType, l.ReturnType)
}

// ParseType -
func (l *Lambda) ParseType(node *base.Node, id *int) error {
	if err := l.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], l.depth, id)
	if err != nil {
		return err
	}
	l.InputType = typ

	retTyp, err := typingNode(node.Args[1], l.depth, id)
	if err != nil {
		return err
	}
	l.ReturnType = retTyp
	return nil
}

// ParseValue -
func (l *Lambda) ParseValue(node *base.Node) error {
	return l.Default.ParseValue(node)
}

// ToJSONSchema -
func (l *Lambda) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(l.Default), nil
}

// ToParameters -
func (l *Lambda) ToParameters() ([]byte, error) {
	sLambda := fmt.Sprintf("%s", l.Value)
	t, err := translator.NewConverter(
		translator.WithDefaultGrammar(),
	)
	if err != nil {
		return nil, err
	}
	jsonLambda, err := t.FromString(sLambda)
	if err != nil {
		return nil, err
	}
	return []byte(jsonLambda.String()), nil
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

	iBytes, err := json.Marshal(l.InputType)
	if err != nil {
		return nil, "", err
	}
	parameter, err := formatter.MichelineToMichelson(gjson.ParseBytes(iBytes), true, formatter.DefLineSize)
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: "input", Value: parameter})

	rBytes, err := json.Marshal(l.ReturnType)
	if err != nil {
		return nil, "", err
	}
	returnValue, err := formatter.MichelineToMichelson(gjson.ParseBytes(rBytes), true, formatter.DefLineSize)
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: "return", Value: returnValue})

	typedef.Type = consts.LAMBDA
	return []Typedef{typedef}, makeVarDocString(name), nil
}
