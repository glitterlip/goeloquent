package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestString", "test_string"},
		{"AnotherTestString", "another_test_string"},
		{"testString", "test_string"},
		{"test_string", "test_string"},
		{"Test123String", "test123_string"},
		{"ID", "id"},
		{"Id", "id"},
	}

	for _, test := range tests {
		result := goeloquent.ToSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToSnakeCase(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		a        interface{}
		b        interface{}
		expected bool
		kind     reflect.Kind
	}{
		{int64(1), int64(1), true, reflect.Int64},
		{int64(1), int64(2), false, reflect.Int64},
		{1, 1, true, reflect.Int},
		{1, 2, false, reflect.Int},
		{uint64(1), uint64(1), true, reflect.Uint64},
		{uint64(1), uint64(2), false, reflect.Uint64},
		{float32(1.0), float32(1.0), true, reflect.Float32},
		{float32(1.0), float32(2.0), false, reflect.Float32},
		{1.0, 1.0, true, reflect.Float64},
		{1.0, 2.0, false, reflect.Float64},
		{"test", "test", true, reflect.String},
		{"test", "different", false, reflect.String},
		{complex64(1 + 2i), complex64(1 + 2i), true, reflect.Complex64},
		{complex64(1 + 2i), complex64(2 + 3i), false, reflect.Complex64},
		{1 + 2i, 1 + 2i, true, reflect.Complex128},
		{1 + 2i, 2 + 3i, false, reflect.Complex128},
		{[]int{1, 2, 3}, []int{1, 2, 3}, true, reflect.Slice},
		{[]int{1, 2, 3}, []int{4, 5, 6}, false, reflect.Slice},
		{[3]int{1, 2, 3}, [3]int{1, 2, 3}, true, reflect.Array},
		{[3]int{1, 2, 3}, [3]int{4, 5, 6}, false, reflect.Array},
		{map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 2}, true, reflect.Map},
		{map[string]int{"a": 1, "b": 2}, map[string]int{"a": 1, "b": 3}, false, reflect.Map},
		{map[string]interface{}{"a": 1, "b": "test"}, map[string]interface{}{"a": 1, "b": "test"}, true, reflect.Map},
		{map[string]interface{}{"a": 1, "b": "test"}, map[string]interface{}{"a": 1, "b": "different"}, false, reflect.Map},
		{struct{ A int }{A: 1}, struct{ A int }{A: 1}, true, reflect.Struct},
		{struct{ A int }{A: 1}, struct{ A int }{A: 2}, false, reflect.Struct},
		{a: true, b: true, expected: true, kind: reflect.Bool},
		{a: true, b: false, expected: false, kind: reflect.Bool},
	}

	for _, test := range tests {
		result := goeloquent.Equals(test.a, test.b, test.kind)
		if result != test.expected {
			t.Errorf("Equals(%v, %v) = %v; want %v", test.a, test.b, result, test.expected)
		}
	}
}

func TestGroupItems(t *testing.T) {
	type GroupUser struct {
		Id      int64  `json:"id" goelo:"column:id;primaryKey;autoIncrement:false"`
		Name    string `json:"name" goelo:"column:name;"`
		GroupId int64  `json:"group_id" goelo:"column:group_id"`
	}

	var users = []GroupUser{
		{Id: 1, Name: "Alice", GroupId: 2},
		{Id: 2, Name: "Bob", GroupId: 2},
		{Id: 3, Name: "Alice", GroupId: 3},
		{Id: 4, Name: "Charlie", GroupId: 3},
		{Id: 5, Name: "Bob", GroupId: 1},
		{Id: 6, Name: "Bob1", GroupId: 3},
	}
	config, _ := goeloquent.GetParsedModel(&users)
	grouped := goeloquent.GroupItemsByKey(users, "id", config, false)
	assert.Equal(t, len(users), len(grouped))
	for id, slice := range grouped {
		userGroup := slice.([]*GroupUser)
		assert.Equal(t, 1, len(userGroup), "Expected one user per group for id %s", id)
		for _, user := range userGroup {
			assert.Equal(t, id, fmt.Sprint(user.Id), "Expected user ID to match group key %s", id)
		}
	}

	groupByGroup := goeloquent.GroupItemsByKey(users, "group_id", config, false)
	assert.Equal(t, 3, len(groupByGroup), "Expected 3 groups based on group_id")
	for groupId, slice := range groupByGroup {
		groupSlice := slice.([]*GroupUser)
		assert.Equal(t, groupId, fmt.Sprint(len(groupSlice)))
		for _, groupUser := range groupSlice {
			assert.Equal(t, groupId, fmt.Sprint(groupUser.GroupId), "Expected user GroupId to match group key %s", groupId)
		}
	}
}
