package pagination

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/internal/platform/page"
)

func ToMetaData(pg page.Page, total int) api.PageMetadata {
	doc := page.NewDocument(pg, total)

	return api.PageMetadata{
		Number:          doc.Page,
		Size:            doc.PageSize,
		TotalElements:   doc.TotalCount,
		TotalPages:      doc.TotalPages,
		HasNextPage:     doc.HasNext,
		HasPreviousPage: doc.HasPrev,
	}
}

func ToLinks(baseURL string, pg page.Page, total int) *api.Links {
	doc := page.NewDocument(pg, total)
	links := make(api.Links)

	makeHref := func(pageNum int) string {
		return fmt.Sprintf("%s?page=%d&pageSize=%d", baseURL, pageNum, doc.PageSize)
	}

	links["self"] = api.Link{Href: makeHref(doc.Page)}
	links["first"] = api.Link{Href: makeHref(1)}
	links["last"] = api.Link{Href: makeHref(doc.TotalPages)}

	if doc.HasNext {
		links["next"] = api.Link{Href: makeHref(doc.Page + 1)}
	}
	if doc.HasPrev {
		links["prev"] = api.Link{Href: makeHref(doc.Page - 1)}
	}

	return &links
}
