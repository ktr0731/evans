package table

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPresenter(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		const expected = `+--------+--------+----------+
| FIELD1 | FIELD2 |  FIELD3  |
+--------+--------+----------+
| field1 |    100 | {field3} |
+--------+--------+----------+
`
		type Struct struct {
			StructField string
		}
		type v struct {
			Field1 string
			Field2 int
			Field3 Struct
		}
		vi := &v{Field1: "field1", Field2: 100, Field3: Struct{"field3"}}
		p := NewPresenter()
		actual, err := p.Format(&vi, "")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("(-want, +got)\n%s", diff)
		}
	})

	t.Run("slice", func(t *testing.T) {
		const expected = `+--------+--------+
| FIELD1 | FIELD2 |
+--------+--------+
| field1 |    100 |
| field2 |    300 |
+--------+--------+
`
		type v struct {
			Field1 string
			Field2 int
		}
		vi := []*v{
			{"field1", 100},
			{"field2", 300},
		}
		p := NewPresenter()
		actual, err := p.Format(&vi, "")
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("(-want, +got)\n%s", diff)
		}
	})
}
