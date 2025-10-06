package assemblers

import (
	"fmt"
	"hosting-contracts/api"
)

func to[T any](v T) *T {
	return &v
}

func newPaginationLinks(basePath string, page, pageSize int, totalPages int, hasPrev, hasNext bool) api.Links {
	links := make(api.Links)

	links["self"] = api.Link{Href: to(fmt.Sprintf("%s?page=%d&pageSize=%d", basePath, page, pageSize))}
	links["first"] = api.Link{Href: to(fmt.Sprintf("%s?page=1&pageSize=%d", basePath, pageSize))}
	links["last"] = api.Link{Href: to(fmt.Sprintf("%s?page=%d&pageSize=%d", basePath, totalPages, pageSize))}

	if hasPrev {
		links["prev"] = api.Link{Href: to(fmt.Sprintf("%s?page=%d&pageSize=%d", basePath, page-1, pageSize))}
	}
	if hasNext {
		links["next"] = api.Link{Href: to(fmt.Sprintf("%s?page=%d&pageSize=%d", basePath, page+1, pageSize))}
	}

	return links
}
