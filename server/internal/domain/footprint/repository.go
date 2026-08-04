package footprint

import "context"

type ListOptions struct {
	Page                int
	PageSize            int
	Published           *bool
	PublishedAlbumsOnly bool
	Search              string
}

type Repository interface {
	Create(ctx context.Context, journey *Journey, place Place, albumIDs []int64) error
	GetByID(ctx context.Context, id int64, publishedAlbumsOnly bool) (*Journey, error)
	Update(ctx context.Context, journey *Journey, place Place, albumIDs []int64) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, opts ListOptions) ([]*Journey, int64, error)
	ListPlaces(ctx context.Context) ([]Place, error)
}
