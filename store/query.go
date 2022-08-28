package store

type ComparisonOperator int

const (
	Eq ComparisonOperator = 1 << iota
)

type Filter struct {
	AttrName   string
	Value      string
	ValueFloat *float64
	Operator   ComparisonOperator
}

type Sort struct {
	AttrName  string
	Ascending bool
}

type ListQuery struct {
	Skip    int
	Limit   int
	Filters []Filter
	Sort    *Sort
}
