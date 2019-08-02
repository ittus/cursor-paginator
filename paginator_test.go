package paginator

import (
	"math"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

func TestPaginator(t *testing.T) {
	suite.Run(t, &paginatorSuite{})
}

type paginatorSuite struct {
	suite.Suite
	db *gorm.DB
}

type item struct {
	ID        int       `gorm:"primary_key"`
	CreatedAt time.Time `gorm:"type:timestamp;not null"`
}

func (s *paginatorSuite) SetupSuite() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 dbname=test user=test password=test sslmode=disable")
	if err != nil {
		s.FailNow(err.Error())
	}
	s.db = db
	s.db.AutoMigrate(&item{})
}

func (s *paginatorSuite) TearDownTest() {
	s.db.Exec("TRUNCATE items;")
}

func (s *paginatorSuite) TearDownSuite() {
	s.db.DropTable(&item{})
	s.db.Close()
}

func (s *paginatorSuite) TestPaginateWithIDCursor() {
	var items = []item{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now().Add(3 * time.Hour)},
		{CreatedAt: time.Now().Add(4 * time.Hour)},
		{CreatedAt: time.Now().Add(5 * time.Hour)},
		{CreatedAt: time.Now().Add(6 * time.Hour)},
		{CreatedAt: time.Now().Add(7 * time.Hour)},
		{CreatedAt: time.Now().Add(8 * time.Hour)},
		{CreatedAt: time.Now().Add(9 * time.Hour)},
		{CreatedAt: time.Now().Add(10 * time.Hour)},
	}
	s.createItems(items)

	perPage := 3
	total := len(items)
	pageNum := (total / perPage) + 1
	baseQuery := s.db

	var cursor *int64
	for i := 0; i < pageNum; i++ {
		var itemResults []item
		paginator := NewCursorPaginator(baseQuery, perPage, OrderDirections.Desc, PaginatorDirections.Next, "id", "ID", cursor, IDModeCursor)
		paginator.Paginate(&itemResults)

		remainingItems := total - (i * perPage)
		expectedItemCount := math.Min(float64(remainingItems), float64(perPage))
		s.Equal(int(expectedItemCount), len(itemResults))

		if i != pageNum-1 {
			nextCursor := int64(paginator.NextCursor.(int))
			cursor = &nextCursor
		} else {
			cursor = nil
		}
		var prevItem item
		for index, item := range itemResults {
			if index != 0 {
				s.True(item.ID <= prevItem.ID)
			}
			prevItem = item
		}
	}

}

func (s *paginatorSuite) TestPaginateWithTimeCursor() {
	var items = []item{
		{CreatedAt: time.Now()},
		{CreatedAt: time.Now().Add(2 * time.Hour)},
		{CreatedAt: time.Now().Add(3 * time.Hour)},
		{CreatedAt: time.Now().Add(4 * time.Hour)},
		{CreatedAt: time.Now().Add(5 * time.Hour)},
		{CreatedAt: time.Now().Add(6 * time.Hour)},
		{CreatedAt: time.Now().Add(7 * time.Hour)},
		{CreatedAt: time.Now().Add(8 * time.Hour)},
		{CreatedAt: time.Now().Add(9 * time.Hour)},
		{CreatedAt: time.Now().Add(10 * time.Hour)},
	}
	s.createItems(items)
	perPage := 3
	total := len(items)
	pageNum := (total / perPage) + 1
	baseQuery := s.db

	var cursor *int64
	var backCursor *int64
	for i := 0; i < pageNum; i++ {
		var itemResults []item
		paginator := NewCursorPaginator(baseQuery, perPage, OrderDirections.Desc, PaginatorDirections.Next, "created_at", "CreatedAt", cursor, TimeModeCursor)
		paginator.Paginate(&itemResults)

		remainingItems := total - (i * perPage)
		expectedItemCount := math.Min(float64(remainingItems), float64(perPage))
		s.Equal(int(expectedItemCount), len(itemResults))

		if i != pageNum-1 {
			nextCursor := paginator.NextCursor.(int64)
			cursor = &nextCursor
		} else {
			cursor = nil
			s.NotNil(paginator.PreviousCursor)
			goBackCursor := paginator.PreviousCursor.(int64)
			backCursor = &goBackCursor
		}
		var prevItem item
		for index, item := range itemResults {
			if index != 0 {
				s.True(item.CreatedAt.Before(prevItem.CreatedAt) || item.CreatedAt.Equal(prevItem.CreatedAt))
			}
			prevItem = item
		}
	}

	// go back
	for i := pageNum - 1; i >= 1; i-- {
		var itemResults []item
		paginator := NewCursorPaginator(baseQuery, perPage, OrderDirections.Desc, PaginatorDirections.Back, "created_at", "CreatedAt", backCursor, TimeModeCursor)
		paginator.Paginate(&itemResults)
		s.Equal(perPage, len(itemResults))

		if i != 1 {
			goBackCursor := paginator.PreviousCursor.(int64)
			backCursor = &goBackCursor
		} else {
			backCursor = nil
		}
		var prevItem item
		for index, item := range itemResults {
			if index != 0 {
				s.True(item.CreatedAt.Before(prevItem.CreatedAt) || item.CreatedAt.Equal(prevItem.CreatedAt))
			}
			prevItem = item
		}
	}

}

func (s *paginatorSuite) createItems(items []item) {
	for i := 0; i < len(items); i++ {
		if err := s.db.Create(&items[i]).Error; err != nil {
			s.FailNow(err.Error())
		}
	}
}
