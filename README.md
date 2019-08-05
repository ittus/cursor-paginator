## Cursor paginator
Two-way cursor pagination with Golang and GORM (bidirections cursor)

[![CircleCI](https://circleci.com/gh/ittus/cursor-paginator.svg?style=shield)](https://circleci.com/gh/ittus/cursor-paginator)

## Install

```go
go get -u github.com/ittus/cursor-paginator
```

## How to use 
Support we have gorm model
```go
type item struct {
	ID        int       `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}
```

then we can use
```go
   import paginator "https://github.com/ittus/cursor-paginator"    


    var itemResults []item
    var cursor *int64
    p := paginator.NewCursorPaginator(
    	baseQuery, 
    	perPage, 
    	paginator.OrderDirections.Desc, 
    	paginator.PaginatorDirections.Next, 
    	"id", 
    	"ID", 
    	cursor, 
    	paginator.IDModeCursor)
    p.Paginate(&itemResults)
    
    // p.NextCursor
    // p.PreviousCursor
```
	
## Test

```go
go test -v -covermode=count -coverprofile=c.out
```

## License
Released under the [MIT License](/LICENSE)