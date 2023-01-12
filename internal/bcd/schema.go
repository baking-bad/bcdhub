package bcd

var (
	SchemaApprove = []byte(`{"type":"object","properties":{"allowances":{"type":"array","prim":"list","default":[],"items":{"type":"object","required":["token_type"],"properties":{"token_type":{"oneOf":[{"title":"FA1.2","properties":{"schema_key":{"type":"integer","const":1},"token_contract":{"type":"string","title":"Token contract","prim":"address","default":"","minLength":36,"maxLength":36},"allowance":{"type":"integer","title":"Allowance","prim":"nat","default":0}}},{"title":"FA2","properties":{"schema_key":{"type":"integer","const":2},"token_contract":{"type":"string","title":"Token contract","prim":"address","default":"","minLength":36,"maxLength":36},"owner":{"type":"string","title":"Owner","prim":"address","default":"","minLength":36,"maxLength":36},"token_id":{"type":"integer","title":"Token id","prim":"nat","default":0}}}],"title":"Token type","prim":"or","type":"object"}},"x-options":{"sectionsClass":"pl-0"}}}}}`)
)
