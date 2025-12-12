package page

type Page struct {
	number int
	size   int
}

type Document struct {
	Page       int
	PageSize   int
	TotalCount int
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

func Parse(page int, pageSize int) Page {
	if page <= 0 {
		page = 1
	}
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}
	return Page{number: page, size: pageSize}
}

func (p Page) Number() int { return p.number }
func (p Page) Size() int   { return p.size }

func (p Page) Offset() int {
	return (p.number - 1) * p.size
}

func NewDocument(p Page, totalCount int) Document {
	totalPages := (totalCount + p.size - 1) / p.size

	return Document{
		Page:       p.number,
		PageSize:   p.size,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    p.number < totalPages,
		HasPrev:    p.number > 1,
	}
}
