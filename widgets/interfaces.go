package widgets

import "github.com/odvcencio/fluffyui/forms"

// Searchable represents widgets that expose a searchable query.
type Searchable interface {
	SetQuery(query string)
	Query() string
}

// Validatable represents widgets that can be validated with form validators.
type Validatable interface {
	SetValidators(validators ...forms.Validator)
	Validate() []forms.ValidationError
	Errors() []string
	Valid() bool
}

// LazyLoadable represents widgets that can request more data as users scroll.
type LazyLoadable interface {
	SetLazyLoad(fn func(start, end, total int))
	SetLazyLoadThreshold(threshold int)
}

// TabularDataSource provides virtualized tabular data.
type TabularDataSource interface {
	RowCount() int
	Cell(row, col int) string
}

// TabularRowProvider optionally provides full row slices.
type TabularRowProvider interface {
	TabularDataSource
	Row(row int) []string
}

// TabularEditable allows editing tabular data.
type TabularEditable interface {
	TabularDataSource
	SetCell(row, col int, value string)
}
