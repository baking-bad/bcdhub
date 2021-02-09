package ast

import "bytes"

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

func buildMapParameters(data map[Node]Node) ([]byte, error) {
	var builder bytes.Buffer
	if err := builder.WriteByte('['); err != nil {
		return nil, err
	}
	for key, value := range data {
		if builder.Len() != 1 {
			if err := builder.WriteByte(','); err != nil {
				return nil, err
			}
		}
		b, err := buildMapItemParameters(key, value)
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
