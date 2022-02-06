package tests

import (
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ShouldEqual(t *testing.T, expected interface{}, b *goeloquent.Builder) {
	assert.Equal(t, expected, b.ToSql())
}
func ElementsShouldMatch(t *testing.T, a interface{}, b interface{}) {
	assert.ElementsMatch(t, a, b)
}
