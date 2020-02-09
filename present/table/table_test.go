package table

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPresenter(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		const expected = `+----------+--------+------------+
|  FIELD1  | FIELD2 |   FIELD3   |
+----------+--------+------------+
| field1-1 |    100 | {field3-1} |
| field1-2 |    300 | {field3-2} |
+----------+--------+------------+
`
		type Struct struct {
			StructField string
		}
		type v struct {
			Field1 string
			Field2 int
			Field3 Struct
		}
		type wrapper struct {
			Fields []*v
		}

		vi := &wrapper{
			Fields: []*v{
				{
					Field1: "field1-1",
					Field2: 100,
					Field3: Struct{"field3-1"},
				},
				{
					Field1: "field1-2",
					Field2: 300,
					Field3: Struct{"field3-2"},
				},
			},
		}
		p := NewPresenter()
		actual, err := p.Format(&vi)
		fmt.Println(actual)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("(-want, +got)\n%s", diff)
		}
	})
}
