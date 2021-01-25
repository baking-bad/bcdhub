package forge

// Object -
type Object struct {
	Node
	argsCount int
	hasAnnots bool
}

// NewObject -
func NewObject(argsCount int, hasAnnots bool) *Object {
	return &Object{
		Node:      Node{},
		argsCount: argsCount,
		hasAnnots: hasAnnots,
	}
}

// Unforge -
func (obj *Object) Unforge(data []byte) (int, error) {
	var length int

	primUnforger := new(prim)
	n, err := primUnforger.Unforge(data)
	if err != nil {
		return n, err
	}
	length += n
	data = data[n:]
	obj.Prim = primUnforger.Value

	if obj.argsCount > 0 {
		argsUnforger := newArgs(obj.argsCount)
		n, err = argsUnforger.Unforge(data)
		if err != nil {
			return n, err
		}
		length += n
		data = data[n:]
		obj.Args = argsUnforger.Args
	} else if obj.argsCount == -1 {
		argsUnforger := NewArray()
		n, err = argsUnforger.Unforge(data)
		if err != nil {
			return n, err
		}
		length += n
		data = data[n:]
		obj.Args = argsUnforger.Args
	}

	if obj.hasAnnots {
		a := newAnnots()
		n, err := a.Unforge(data)
		if err != nil {
			return n, err
		}
		length += n
		data = data[n:]
		obj.Annots = a.Value
	}
	return length, nil
}

// Forge -
func (obj *Object) Forge() ([]byte, error) {
	data := []byte{obj.getFirstByte()}

	primForger := newPrim(obj.Prim)
	primBody, err := primForger.Forge()
	if err != nil {
		return nil, err
	}
	data = append(data, primBody...)

	if len(obj.Args) > 0 {
		var argsForger Forger
		if len(obj.Args) < 3 {
			argsForger = newArgsFromNodes(obj.Args)
		} else {
			argsForger = newArrayFromNodes(obj.Args)
		}
		argsBody, err := argsForger.Forge()
		if err != nil {
			return nil, err
		}
		if len(obj.Args) > 2 {
			argsBody = argsBody[1:]
		}
		data = append(data, argsBody...)
	}

	if len(obj.Annots) > 0 {
		a := newAnnots()
		a.Value = obj.Annots
		annotsBody, err := a.Forge()
		if err != nil {
			return nil, err
		}
		data = append(data, annotsBody...)
	}
	return data, nil
}

func (obj *Object) getFirstByte() byte {
	switch {
	case len(obj.Args) == 0 && len(obj.Annots) == 0:
		return BytePrim
	case len(obj.Args) == 0 && len(obj.Annots) > 0:
		return BytePrimAnnots
	case len(obj.Args) == 1 && len(obj.Annots) == 0:
		return BytePrimArg
	case len(obj.Args) == 1 && len(obj.Annots) > 0:
		return BytePrimArgAnnots
	case len(obj.Args) == 2 && len(obj.Annots) == 0:
		return BytePrimArgs
	case len(obj.Args) == 2 && len(obj.Annots) > 0:
		return BytePrimArgsAnnots
	default:
		return ByteGeneralPrim
	}
}
