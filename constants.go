package paginator

const (
	TimeModeCursor = "time"
	IDModeCursor   = "id"
)

type OrderDirection string

type orderDirections struct {
	Asc  OrderDirection
	Desc OrderDirection
}

var OrderDirections = &orderDirections{
	Asc:  "asc",
	Desc: "desc",
}

func (t *OrderDirection) IsEmpty() bool {
	return t == nil || len(*t) == 0
}

func (t *OrderDirection) IsValid() bool {
	return *t == OrderDirections.Asc || *t == OrderDirections.Desc
}

func (t *OrderDirection) String() string {
	switch *t {
	case OrderDirections.Asc:
		return "ASC"
	case OrderDirections.Desc:
		return "DESC"
	default:
		return "Unknown"
	}
}

type PaginatorDirection string

type paginatorDirections struct {
	Next PaginatorDirection
	Back PaginatorDirection
}

var PaginatorDirections = &paginatorDirections{
	Next: "next",
	Back: "back",
}
