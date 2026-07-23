package mediarecord

import "context"

type ListOptions struct {
	Page       int
	PageSize   int
	Status     string
	MediaType  string
	Published  *bool
	SearchTerm string
}

type Repository interface {
	Create(ctx context.Context, record *Record) error
	GetByID(ctx context.Context, id int64) (*Record, error)
	Update(ctx context.Context, record *Record) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, opts ListOptions) ([]*Record, int64, error)
}
