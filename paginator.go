package paginator

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"reflect"
	"time"
)

type CursorPaginator struct {
	Mode            string `json:"-"`
	baseQuery       *gorm.DB
	Order           OrderDirection     `json:"-"` // asc = 1, desc = 2
	Direction       PaginatorDirection `json:"-"`
	Limit           int                `json:"-"`
	DBFieldName     string             `json:"-"`
	CursorFieldName string             `json:"-"`
	Cursor          interface{}        `json:"-"`
	NextCursor      interface{}        `json:"next_cursor"`
	PreviousCursor  interface{}        `json:"previous_cursor"`
}

func getOperator(order OrderDirection, direction PaginatorDirection) string {
	if direction == PaginatorDirections.Next {
		if order == OrderDirections.Desc {
			return "<"
		}
		return ">"

	}
	if order == OrderDirections.Desc {
		return ">"
	}
	return "<"
}

func getOrder(order OrderDirection, direction PaginatorDirection) string {
	orderDirection := order.String()
	if direction == PaginatorDirections.Back {
		if order == OrderDirections.Asc {
			orderDirection = OrderDirections.Desc.String()
		} else {
			orderDirection = OrderDirections.Asc.String()
		}
	}
	return orderDirection
}

func (p *CursorPaginator) reverse() {
	temp := p.NextCursor
	p.NextCursor = p.PreviousCursor
	p.PreviousCursor = temp
}

func (p *CursorPaginator) hasLimit() bool {
	return p.Limit > 0
}

func (p *CursorPaginator) findCursor(output interface{}) error {
	if reflect.ValueOf(output).Elem().Type().Kind() == reflect.Slice && reflect.ValueOf(output).Elem().Len() > 0 {
		elems := reflect.ValueOf(output).Elem()
		hasMore := elems.Len() > p.Limit
		if hasMore {
			elems.Set(elems.Slice(0, elems.Len()-1))
			lastElement := elems.Index(elems.Len() - 1)
			p.NextCursor = lastElement.FieldByName(p.CursorFieldName).Interface()
			if p.Mode == TimeModeCursor {
				p.NextCursor = (p.NextCursor).(time.Time).UnixNano()
			}
		}

		// check if can go back
		if elems.Len() > 0 {
			firstElement := elems.Index(0)
			backCursor := firstElement.FieldByName(p.CursorFieldName).Interface()

			backQuery := p.baseQuery.Limit(1)
			operator := getOperator(p.Order, PaginatorDirections.Back)
			if p.Direction == PaginatorDirections.Back {
				operator = getOperator(p.Order, PaginatorDirections.Next)
			}

			backQuery = backQuery.Where(fmt.Sprintf("%s %s ?", p.DBFieldName, operator), backCursor)
			tableName := p.baseQuery.NewScope(firstElement.Interface()).TableName()

			var previousCount int
			err := backQuery.Table(tableName).Count(&previousCount).Error
			if err != nil {
				return err
			}
			if previousCount > 0 {
				p.PreviousCursor = backCursor
				if p.Mode == TimeModeCursor {
					p.PreviousCursor = (p.PreviousCursor).(time.Time).UnixNano()
				}
			}
		}
		if p.Direction == PaginatorDirections.Back {
			elems.Set(reverse(elems))
			p.reverse()
		}
	}
	return nil
}

func (p *CursorPaginator) Paginate(output interface{}) error {
	query := p.baseQuery
	if p.hasLimit() {
		query = query.Limit(p.Limit + 1)
	}
	if !p.Order.IsEmpty() {
		orderDirection := getOrder(p.Order, p.Direction)
		orderStmt := fmt.Sprintf("%s %s", p.DBFieldName, orderDirection)
		query = query.Order(orderStmt)
	}
	if !isNil(p.Cursor) {
		operator := getOperator(p.Order, p.Direction)
		query = query.Where(fmt.Sprintf("%s %s ?", p.DBFieldName, operator), p.Cursor)
	}
	err := query.Find(output).Error
	if err != nil {
		return err
	}

	if p.hasLimit() {
		if err = p.findCursor(output); err != nil {
			return err
		}

	}

	return nil
}

func isNil(c interface{}) bool {
	return c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil())
}

func reverse(v reflect.Value) reflect.Value {
	result := reflect.MakeSlice(v.Type(), 0, v.Cap())
	for i := v.Len() - 1; i >= 0; i-- {
		result = reflect.Append(result, v.Index(i))
	}
	return result
}

func NewCursorPaginator(baseQuery *gorm.DB,
	limit int,
	order OrderDirection,
	direction PaginatorDirection,
	dbFieldName string,
	cursorFieldName string,
	cursor *int64,
	mode string,
) *CursorPaginator {
	var paginationCursor interface{}
	if cursor != nil {
		if mode == TimeModeCursor {
			paginationCursor = time.Unix(*cursor/1e9, *cursor%1e9).UTC()
		} else {
			paginationCursor = cursor
		}
	}
	return &CursorPaginator{
		baseQuery:       baseQuery,
		Limit:           limit,
		Order:           order,
		DBFieldName:     dbFieldName,
		CursorFieldName: cursorFieldName,
		Cursor:          paginationCursor,
		Direction:       direction,
		Mode:            mode,
	}
}
