package libovsdb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIListSimple(t *testing.T) {
	cache := apiTestCache(t)
	lscacheList := []Model{
		&testLogicalSwitch{
			UUID:        aUUID0,
			Name:        "ls0",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitch{
			UUID:        aUUID1,
			Name:        "ls1",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID2,
			Name:        "ls2",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID3,
			Name:        "ls4",
			ExternalIds: map[string]string{"foo": "baz"},
			Ports:       []string{"port0", "port1"},
		},
	}
	lscache := map[string]Model{}
	for i := range lscacheList {
		lscache[lscacheList[i].(*testLogicalSwitch).UUID] = lscacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lscache}
	cache.cache["Logical_Switch_Port"] = newRowCache() // empty

	test := []struct {
		name       string
		initialCap int
		resultCap  int
		resultLen  int
		content    []Model
		err        bool
	}{
		{
			name:       "full",
			initialCap: 0,
			resultCap:  len(lscache),
			resultLen:  len(lscacheList),
			content:    lscacheList,
			err:        false,
		},
		{
			name:       "single",
			initialCap: 1,
			resultCap:  1,
			resultLen:  1,
			content:    lscacheList[0:0],
			err:        false,
		},
		{
			name:       "multiple",
			initialCap: 2,
			resultCap:  2,
			resultLen:  2,
			content:    lscacheList[0:2],
			err:        false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiList: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitch
			if tt.initialCap != 0 {
				result = make([]testLogicalSwitch, tt.initialCap)
			}
			api := newAPI(cache)
			err := api.List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Lenf(t, result, tt.resultLen, "Length should match expected")
				assert.Equal(t, cap(result), tt.resultCap, "Capability should match expected")
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}

	t.Run("ApiList: Error wrong type", func(t *testing.T) {
		var result []string
		api := newAPI(cache)
		err := api.List(&result)
		assert.NotNil(t, err)
	})

	t.Run("ApiList: Type Selection", func(t *testing.T) {
		var result []testLogicalSwitchPort
		api := newAPI(cache)
		err := api.List(&result)
		assert.Nil(t, err)
		assert.Len(t, result, 0, "Should be empty since cache is empty")
	})
}

func TestAPIListPredicate(t *testing.T) {
	cache := apiTestCache(t)
	lscacheList := []Model{
		&testLogicalSwitch{
			UUID:        aUUID0,
			Name:        "ls0",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitch{
			UUID:        aUUID1,
			Name:        "magicLs1",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID2,
			Name:        "ls2",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID3,
			Name:        "magicLs2",
			ExternalIds: map[string]string{"foo": "baz"},
			Ports:       []string{"port0", "port1"},
		},
	}
	lscache := map[string]Model{}
	for i := range lscacheList {
		lscache[lscacheList[i].(*testLogicalSwitch).UUID] = lscacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lscache}

	test := []struct {
		name      string
		predicate interface{}
		content   []Model
		err       bool
	}{
		{
			name: "none",
			predicate: func(t *testLogicalSwitch) bool {
				return false
			},
			content: []Model{},
			err:     false,
		},
		{
			name: "all",
			predicate: func(t *testLogicalSwitch) bool {
				return true
			},
			content: lscacheList,
			err:     false,
		},
		{
			name: "nil function must fail",
			err:  true,
		},
		{
			name: "arbitrary condition",
			predicate: func(t *testLogicalSwitch) bool {
				return strings.HasPrefix(t.Name, "magic")
			},
			content: []Model{lscacheList[1], lscacheList[3]},
			err:     false,
		},
		{
			name: "error wrong type",
			predicate: func(t testLogicalSwitch) string {
				return "foo"
			},
			err: true,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiListPredicate: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitch
			api := newAPI(cache)
			cond := api.Where(api.ConditionFromFunc(tt.predicate))
			err := cond.List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				if !assert.Nil(t, err) {
					t.Log(err)
				}
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}
}

func TestAPIListFields(t *testing.T) {
	cache := apiTestCache(t)
	lspcacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "magiclsp1",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			ExternalIds: map[string]string{"unique": "id"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "magiclsp2",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{true},
		},
	}
	lspcache := map[string]Model{}
	for i := range lspcacheList {
		lspcache[lspcacheList[i].(*testLogicalSwitchPort).UUID] = lspcacheList[i]
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspcache}

	testObj := testLogicalSwitchPort{}

	test := []struct {
		name    string
		fields  []interface{}
		prepare func(*testLogicalSwitchPort)
		content []Model
		err     bool
	}{
		{
			name:    "empty object must match everything",
			content: lspcacheList,
			err:     false,
		},
		{
			name: "List unique by UUID",
			prepare: func(t *testLogicalSwitchPort) {
				t.UUID = aUUID0
			},
			content: []Model{lspcache[aUUID0]},
			err:     false,
		},
		{
			name: "List unique by Index",
			prepare: func(t *testLogicalSwitchPort) {
				t.Name = "lsp2"
			},
			content: []Model{lspcache[aUUID2]},
			err:     false,
		},
		{
			name: "List unique by extra field",
			prepare: func(t *testLogicalSwitchPort) {
				t.ExternalIds = map[string]string{"unique": "id"}
			},
			content: []Model{lspcache[aUUID2]},
			fields:  []interface{}{&testObj.ExternalIds},
			err:     false,
		},
		{
			name: "List by extra field",
			prepare: func(t *testLogicalSwitchPort) {
				t.Enabled = []bool{true}
			},
			content: []Model{lspcache[aUUID0], lspcache[aUUID3]},
			fields:  []interface{}{&testObj.Enabled},
			err:     false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiListFields: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitchPort
			// Clean object
			testObj = testLogicalSwitchPort{}
			api := newAPI(cache)
			err := api.Where(api.ConditionFromModel(&testObj, tt.fields...)).List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}

	t.Run("ApiListFields: Wrong table", func(t *testing.T) {
		var result []testLogicalSwitchPort
		api := newAPI(cache)
		obj := testLogicalSwitch{
			UUID: aUUID0,
		}

		err := api.Where(api.ConditionFromModel(&obj)).List(&result)
		assert.NotNil(t, err)
	})

	t.Run("ApiListFields: Wrong object field", func(t *testing.T) {
		var result []testLogicalSwitchPort
		api := newAPI(cache)
		obj := testLogicalSwitch{}
		obj2 := testLogicalSwitch{
			UUID: aUUID0,
		}

		err := api.Where(api.ConditionFromModel(&obj, &obj2.UUID)).List(&result)
		assert.NotNil(t, err)
	})
}

func TestConditionFromFunc(t *testing.T) {
	test := []struct {
		name string
		arg  interface{}
		err  bool
	}{
		{
			name: "wrong function must fail",
			arg: func(s string) bool {
				return false
			},
			err: true,
		},
		{
			name: "wrong function must fail2 ",
			arg: func(t *testLogicalSwitch) string {
				return "foo"
			},
			err: true,
		},
		{
			name: "correct func should succeed",
			arg: func(t *testLogicalSwitch) bool {
				return true
			},
			err: false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ConditionFromFunc: %s", tt.name), func(t *testing.T) {
			cache := apiTestCache(t)
			api := newAPI(cache)
			condition := api.ConditionFromFunc(tt.arg)
			if tt.err {
				assert.IsType(t, &errorCondition{}, condition)
			} else {
				assert.IsType(t, &predicateCond{}, condition)
			}
		})
	}
}

func TestConditionFromModel(t *testing.T) {
	var testObj testLogicalSwitch
	test := []struct {
		name   string
		model  Model
		fields []interface{}
		err    bool
	}{
		{
			name:  "wrong model must fail",
			model: &struct{ a string }{},
			err:   true,
		},
		{
			name: "wrong fields must fail",
			model: &struct {
				a string `ovs:"_uuid"`
			}{},
			fields: []interface{}{"foo"},
			err:    true,
		},
		{
			name:  "correct model must succeed",
			model: &testLogicalSwitch{},
			err:   false,
		},
		{
			name:   "correct model with fields must succeed",
			model:  &testObj,
			fields: []interface{}{&testObj.Name},
			err:    false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ConditionFromModel: %s", tt.name), func(t *testing.T) {
			cache := apiTestCache(t)
			api := newAPI(cache)
			condition := api.ConditionFromModel(tt.model, tt.fields...)
			if tt.err {
				assert.IsType(t, &errorCondition{}, condition)
			} else {
				assert.IsType(t, &indexCond{}, condition)
			}
		})
	}
}
