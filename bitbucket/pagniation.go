package bitbucket

import (
	"iter"
	"net/http"
)

type PaginatedResponse[T any] struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	PageLen  int    `json:"pagelen"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Values   []T    `json:"values"`

	client  *Client
	iterErr error
}

func (r *PaginatedResponse[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		page := r
		for {
			for _, v := range page.Values {
				if !yield(v) {
					return
				}
			}
			if page.Next == "" {
				return
			}

			if page.Size < page.PageLen*page.Page {
				return
			}
			page = &PaginatedResponse[T]{}
			err := r.client.rawRequest(http.MethodGet, page.Next, http.NoBody, page)
			if err != nil {
				r.iterErr = err
				return
			}
		}

	}
}
func (r *PaginatedResponse[T]) AllError() error {
	return r.iterErr
}
