package table

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPresenter(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		const expected = `+----------+--------+----------+
|  FIELD1  | FIELD2 |  FIELD3  |
+----------+--------+----------+
| field1-1 |    100 | {field3} |
| field1-2 |    200 | {field3} |
| field1-3 |        | {field3} |
+----------+--------+----------+
`
		type Struct struct {
			StructField string
		}
		type v struct {
			Field1 []string
			Field2 []int
			Field3 Struct
		}
		vi := &v{Field1: []string{"field1-1", "field1-2", "field1-3"}, Field2: []int{100, 200}, Field3: Struct{"field3"}}
		p := NewPresenter()
		actual, err := p.Format(&vi, "")
		fmt.Println(actual)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("(-want, +got)\n%s", diff)
		}
	})
}
