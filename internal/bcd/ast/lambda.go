package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// Lambda -
type Lambda struct {
	Default
	Type Node
}

// NewLambda -
func NewLambda(depth int) *Lambda {
	return &Lambda{
		Default: NewDefault(consts.LAMBDA, -1, depth),
	}
}

// MarshalJSON -
func (l *Lambda) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.LAMBDA, l.annots, l.Type)
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
	l.Type = typ
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
