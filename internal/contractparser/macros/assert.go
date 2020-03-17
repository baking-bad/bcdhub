package macros

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// =======================
// ======= ASSERT ========
// =======================

type assertMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newAssertMacros() *assertMacros {
	return &assertMacros{
		defaultMacros: &defaultMacros{
			Name: assert,
		},
		NewValue: map[string]interface{}{
			"prim": assert,
		},
	}
}

func (m *assertMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != ifp {
		return false
	}
	args := data.Get("args").Array()
	if len(args) != 2 {
		return false
	}

	return len(args[0].Array()) == 0 && getPrim(args[1].Get("0")) == fail
}

func (m *assertMacros) Replace(json, path string) (res string, err error) {
	return sjson.Set(json, path, m.NewValue)
}

// =======================
// ===== ASSERT_NONE =====
// =======================

type assertNoneMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newAssertNoneMacros() *assertNoneMacros {
	return &assertNoneMacros{
		defaultMacros: &defaultMacros{
			Name: assertNone,
		},
		NewValue: map[string]interface{}{
			"prim": assertNone,
		},
	}
}

func (m *assertNoneMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != ifNone {
		return false
	}
	args := data.Get("args").Array()
	if len(args) != 2 {
		return false
	}

	return len(args[0].Array()) == 0 && getPrim(args[1].Get("0")) == fail
}

func (m *assertNoneMacros) Replace(json, path string) (res string, err error) {
	return sjson.Set(json, path, m.NewValue)
}

// =======================
// ====== ASSERT_EQ ======
// =======================

type assertEqMacros struct {
	*defaultMacros

	IfType   string
	NewValue map[string]interface{}
}

func newAssertEqMacros() *assertEqMacros {
	return &assertEqMacros{
		defaultMacros: &defaultMacros{
			Name: "ASSERT_EQ",
		},
	}
}

func (m *assertEqMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	arr := data.Array()
	if len(arr) != 2 {
		return false
	}

	prim := getPrim(arr[0])
	if !helpers.StringInArray(prim, []string{
		eq, neq, gt, lt, le, ge,
	}) {
		return false
	}

	assertPrim := getPrim(arr[1])
	if assertPrim != assert {
		return false
	}

	m.IfType = prim
	return true
}

func (m *assertEqMacros) Collapse(data gjson.Result) {
	m.NewValue = map[string]interface{}{
		"prim": fmt.Sprintf("ASSERT_%s", m.IfType),
	}
}

func (m *assertEqMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

// =======================
// ===== ASSERT_CMP ======
// =======================

type assertCmpMacros struct {
	*defaultMacros

	CmpType  string
	NewValue map[string]interface{}
}

func newAssertCmpMacros() *assertCmpMacros {
	return &assertCmpMacros{
		defaultMacros: &defaultMacros{
			Name: "ASSERT_CMP",
		},
	}
}

func (m *assertCmpMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	arr := data.Array()
	if len(arr) != 2 {
		return false
	}

	prim := getPrim(arr[0].Get("0"))
	if len(prim) <= 3 || !strings.HasPrefix(prim, cmp) {
		return false
	}
	assertPrim := getPrim(arr[1])
	if assertPrim != assert {
		return false
	}

	m.CmpType = prim
	return true
}

func (m *assertCmpMacros) Collapse(data gjson.Result) {
	m.NewValue = map[string]interface{}{
		"prim": fmt.Sprintf("ASSERT_%s", m.CmpType),
	}
}

func (m *assertCmpMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

// =======================
// ===== ASSERT_SOME =====
// =======================

type assertSomeMacros struct {
	*defaultMacros
	NewValue map[string]interface{}
}

func newAssertSomeMacros() *assertSomeMacros {
	return &assertSomeMacros{
		defaultMacros: &defaultMacros{
			Name: assertSome,
		},
	}
}

func (m *assertSomeMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != ifNone {
		return false
	}
	args := data.Get("args").Array()
	if len(args) != 2 {
		return false
	}

	return getPrim(args[0].Get("0")) == fail && getPrim(args[1].Get("0")) == rename
}

func (m *assertSomeMacros) Collapse(data gjson.Result) {
	res := map[string]interface{}{
		"prim": m.Name,
	}

	annots := data.Get("args.1.annots")
	if annots.Exists() {
		res["annots"] = annots.Value()
	}

	m.NewValue = res
}

func (m *assertSomeMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

// =======================
// ===== ASSERT_LEFT =====
// =======================

type assertLeftMacros struct {
	*defaultMacros
	NewValue map[string]interface{}
}

func newAssertLeftMacros() *assertLeftMacros {
	return &assertLeftMacros{
		defaultMacros: &defaultMacros{
			Name: assertLeft,
		},
	}
}

func (m *assertLeftMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != ifLeft {
		return false
	}
	args := data.Get("args").Array()
	if len(args) != 2 {
		return false
	}

	return getPrim(args[1].Get("0")) == fail && getPrim(args[0].Get("0")) == rename
}

func (m *assertLeftMacros) Collapse(data gjson.Result) {
	res := map[string]interface{}{
		"prim": m.Name,
	}

	annots := data.Get("args.0.annots")
	if annots.Exists() {
		res["annots"] = annots.Value()
	}

	m.NewValue = res
}

func (m *assertLeftMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

// =======================
// ==== ASSERT_RIGHT =====
// =======================

type assertRightMacros struct {
	*defaultMacros
	NewValue map[string]interface{}
}

func newAssertRightMacros() *assertRightMacros {
	return &assertRightMacros{
		defaultMacros: &defaultMacros{
			Name: assertRight,
		},
	}
}

func (m *assertRightMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != ifLeft {
		return false
	}
	args := data.Get("args").Array()
	if len(args) != 2 {
		return false
	}

	return getPrim(args[0].Get("0")) == fail && getPrim(args[1].Get("0")) == rename
}

func (m *assertRightMacros) Collapse(data gjson.Result) {
	res := map[string]interface{}{
		"prim": m.Name,
	}

	annots := data.Get("args.1.annots")
	if annots.Exists() {
		res["annots"] = annots.Value()
	}

	m.NewValue = res
}

func (m *assertRightMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}
