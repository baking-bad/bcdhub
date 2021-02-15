package ast

import (
	"bytes"
)

func buildMapItemParameters(key, value Node) ([]byte, error) {
	var builder bytes.Buffer
	if _, err := builder.WriteString(`{"prim":"Elt","args":[`); err != nil {
		return nil, err
	}

	bKey, err := key.ToParameters()
	if err != nil {
		return nil, err
	}
	if _, err := builder.Write(bKey); err != nil {
		return nil, err
	}

	if err := builder.WriteByte(','); err != nil {
		return nil, err
	}

	bValue, err := value.ToParameters()
	if err != nil {
		return nil, err
	}
	if _, err := builder.Write(bValue); err != nil {
		return nil, err
	}
	if _, err := builder.WriteString(`]}`); err != nil {
		return nil, err
	}
	return builder.Bytes(), nil
}

func buildMapParameters(data *OrderedMap) ([]byte, error) {
	if data == nil {
		return nil, nil
	}
	var builder bytes.Buffer
	if err := builder.WriteByte('['); err != nil {
		return nil, err
	}

	err := data.Range(func(key, value Node) (bool, error) {
		if builder.Len() != 1 {
			if err := builder.WriteByte(','); err != nil {
				return true, err
			}
		}
		b, err := buildMapItemParameters(key, value)
		if err != nil {
			return true, err
		}
		if _, err := builder.Write(b); err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	if err := builder.WriteByte(']'); err != nil {
		return nil, err
	}
	return builder.Bytes(), nil
}

func buildListParameters(data []Node) ([]byte, error) {
	var builder bytes.Buffer
	if err := builder.WriteByte('['); err != nil {
		return nil, err
	}
	for i := range data {
		if i != 0 {
			if err := builder.WriteByte(','); err != nil {
				return nil, err
			}
		}
		b, err := data[i].ToParameters()
		if err != nil {
			return nil, err
		}
		if _, err := builder.Write(b); err != nil {
			return nil, err
		}
	}
	if err := builder.WriteByte(']'); err != nil {
		return nil, err
	}
	return builder.Bytes(), nil
}
