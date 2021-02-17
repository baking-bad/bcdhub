package ast

import "github.com/baking-bad/bcdhub/internal/bcd/types"

// Schema types
const (
	JSONSchemaTypeInt    = "integer"
	JSONSchemaTypeString = "string"
	JSONSchemaTypeBool   = "boolean"
	JSONSchemaTypeArray  = "array"
	JSONSchemaTypeObject = "object"
)

// JSONSchema -
type JSONSchema struct {
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Prim       string                 `json:"prim,omitempty"`
	Tag        string                 `json:"tag,omitempty"`
	Format     string                 `json:"format,omitempty"`
	Default    interface{}            `json:"default,omitempty"`
	MinLength  int                    `json:"minLength,omitempty"`
	MaxLength  int                    `json:"maxLength,omitempty"`
	Properties map[string]*JSONSchema `json:"properties,omitempty"`
	OneOf      []*JSONSchema          `json:"oneOf,omitempty"`
	Required   []string               `json:"required,omitempty"`
	XItemTitle string                 `json:"x-itemTitle,omitempty"`
	Const      string                 `json:"const,omitempty"`
	SchemaKey  *SchemaKey             `json:"schemaKey,omitempty"`
	Items      *SchemaKey             `json:"items,omitempty"`
}

// SchemaKey -
type SchemaKey JSONSchema

func getStringJSONSchema(d Default) *JSONSchema {
	return wrapObject(&JSONSchema{
		Prim:    d.Prim,
		Type:    JSONSchemaTypeString,
		Default: "",
		Title:   d.GetName(),
	})
}

func getIntJSONSchema(d Default) *JSONSchema {
	return wrapObject(&JSONSchema{
		Prim:    d.Prim,
		Type:    JSONSchemaTypeInt,
		Default: 0,
		Title:   d.GetName(),
	})
}

func getAddressJSONSchema(d Default) *JSONSchema {
	return wrapObject(&JSONSchema{
		Prim:      d.Prim,
		Type:      JSONSchemaTypeString,
		MinLength: 36,
		MaxLength: 36,
		Default:   "",
		Title:     d.GetName(),
	})
}

func setIntJSONSchema(d *Default, data map[string]interface{}) {
	for key := range data {
		if key == d.GetName() {
			f := data[key].(float64)
			d.Value = types.NewBigInt(int64(f))
			d.ValueKind = valueKindInt
			break
		}
	}
}

func setBytesJSONSchema(d *Default, data map[string]interface{}) {
	for key := range data {
		if key == d.GetName() {
			d.Value = data[key]
			d.ValueKind = valueKindBytes
			break
		}
	}
}

func setOptimizedJSONSchema(d *Default, data map[string]interface{}, optimizer func(string) (string, error)) {
	for key, value := range data {
		if key == d.GetName() {
			val, err := optimizer(value.(string))
			if err != nil {
				val = value.(string)
			}
			d.ValueKind = valueKindString
			d.Value = val
			break
		}
	}
}

type mergeFields struct {
	reqs   []string
	xTitle string
}

func mergePropertiesMap(src, dest map[string]*JSONSchema, required, needXTitle bool) *mergeFields {
	fields := mergeFields{}
	if required {
		fields.reqs = make([]string, 0)
	}
	for key, value := range src {
		dest[key] = value

		if required {
			fields.reqs = append(fields.reqs, key)
		}
		if needXTitle {
			fields.xTitle = key
		}
	}
	return &fields
}

func setChildSchema(child Node, required bool, parent *JSONSchema) error {
	s, err := child.ToJSONSchema()
	if err != nil {
		return err
	}

	if len(s.Properties) > 0 {
		if parent.Items == nil {
			fields := mergePropertiesMap(s.Properties, parent.Properties, required, false)
			parent.Required = append(parent.Required, fields.reqs...)
		} else {
			if parent.Items.Properties == nil {
				parent.Items.Properties = make(map[string]*JSONSchema)
			}
			if parent.Items.Required == nil {
				parent.Items.Required = make([]string, 0)
			}
			fields := mergePropertiesMap(s.Properties, parent.Items.Properties, required, false)
			parent.Items.Required = append(parent.Items.Required, fields.reqs...)
		}
	} else {
		parent.Properties[child.GetName()] = s
	}
	return nil
}

func setChildSchemaForMap(child Node, needXTitle bool, parent *JSONSchema) error {
	s, err := child.ToJSONSchema()
	if err != nil {
		return err
	}

	if len(s.Properties) > 0 {
		if parent.Items.Properties == nil {
			parent.Items.Properties = make(map[string]*JSONSchema)
		}
		if parent.Items.Required == nil {
			parent.Items.Required = make([]string, 0)
		}
		fields := mergePropertiesMap(s.Properties, parent.Items.Properties, true, needXTitle)
		parent.Items.Required = append(parent.Items.Required, fields.reqs...)
		if fields.xTitle != "" {
			parent.XItemTitle = fields.xTitle
		}
	} else {
		parent.Items.Properties[child.GetName()] = s
	}
	return nil
}

func wrapObject(schema *JSONSchema) *JSONSchema {
	return &JSONSchema{
		Type: JSONSchemaTypeObject,
		Properties: map[string]*JSONSchema{
			schema.Title: schema,
		},
	}
}
