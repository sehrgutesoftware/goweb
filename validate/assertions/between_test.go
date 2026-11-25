package assertions_test

import (
	"reflect"
	"testing"

	"github.com/sehrgutesoftware/goweb/validate/assertions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type title string

func TestBetweenLen(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		actual uint64
		ok     bool
	}{
		{name: "it accepts a string within the range", value: "abc", ok: true},
		{name: "it counts the characters of a string, not its bytes", value: "äöü", ok: true},
		{name: "it counts emoji as one character each", value: "👍👍👍", ok: true},
		{name: "it counts the characters of a named string type", value: title("ßßß"), ok: true},
		{name: "it refuses a string that is too long", value: "äöüß", actual: 4},
		{name: "it refuses a string that is too short", value: "", actual: 0},
		{name: "it counts the elements of a slice", value: []string{"äöü", "b"}, ok: true},
		{name: "it refuses a slice with too many elements", value: []int{1, 2, 3, 4}, actual: 4},
		{name: "it counts the entries of a map", value: map[string]int{"ä": 1}, ok: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := assertions.Resolve("between", reflect.TypeOf(tt.value), "1:3")
			require.NoError(t, err)

			err = a.Validate(tt.value)

			if tt.ok {
				assert.NoError(t, err)
				return
			}
			var res interface {
				Code() string
				Values() map[string]any
			}
			require.ErrorAs(t, err, &res)
			assert.Equal(t, "between", res.Code())
			assert.Equal(t, map[string]any{"min": uint64(1), "max": uint64(3), "actual": tt.actual}, res.Values())
		})
	}
}
