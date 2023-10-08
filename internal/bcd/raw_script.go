package bcd

import (
	"bytes"
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// RawScript -
type RawScript struct {
	Code      []byte `json:"-"`
	Parameter []byte `json:"-"`
	Storage   []byte `json:"-"`
	Views     []byte `json:"-"`
}

type prim struct {
	Prim string          `json:"prim"`
	Args json.RawMessage `json:"args"`
}

// UnmarshalJSON -
func (s *RawScript) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) < 3 {
		return errors.Errorf("length of script types must be 3 but got %d", len(raw))
	}

	var views bytes.Buffer
	if err := views.WriteByte('['); err != nil {
		return err
	}
	for i := range raw {
		var p prim
		if err := json.Unmarshal(raw[i], &p); err != nil {
			return err
		}
		switch p.Prim {
		case consts.PARAMETER:
			s.Parameter = p.Args
		case consts.STORAGE:
			s.Storage = p.Args
		case consts.CODE:
			s.Code = p.Args
		case consts.View:
			if views.Len() > 1 {
				if err := views.WriteByte(','); err != nil {
					return err
				}
			}
			if _, err := views.Write(raw[i]); err != nil {
				return err
			}
		default:
			return errors.Errorf("unknown script high level primitive: %s", p.Prim)
		}
	}

	if err := views.WriteByte(']'); err != nil {
		return err
	}

	if views.Len() > 2 {
		s.Views = views.Bytes()
	}

	return nil
}
