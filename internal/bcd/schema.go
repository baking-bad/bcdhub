package bcd

var (
	SchemaApproveFa1 = []byte(`{"type":"object","properties":{"allowances":{"type":"array","prim":"list","default":[],"items":{"type":"object","required":["token_contract"],"properties":{"token_contract":{"type":"string","title":"token_contract","prim":"address","default":"","minLength":36,"maxLength":36},"allowance":{"type":"integer","title":"allowance","prim":"nat","default":0}},"x-options":{"sectionsClass":"pl-0"}}}}}`)
	SchemaApproveFa2 = []byte(`{"type":"object","properties":{"allowances":{"type":"array","prim":"list","default":[],"items":{"type":"object","required":["token_contract","token_id","owner"],"properties":{"token_contract":{"type":"string","title":"token_contract","prim":"address","default":"","minLength":36,"maxLength":36},"owner":{"type":"string","title":"owner","prim":"address","default":"","minLength":36,"maxLength":36},"token_id":{"type":"integer","title":"token_id","prim":"nat","default":0}},"x-options":{"sectionsClass":"pl-0"}}}}}`)
)
