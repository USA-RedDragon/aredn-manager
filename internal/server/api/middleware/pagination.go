package middleware

import (
	"math"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/server/api/pagination"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxLimit = 1000
const defaultLimit = 10

func paginateDB(db *gorm.DB, c *gin.Context) *gorm.DB {
	var limit int
	limitStr, exists := c.GetQuery("limit")
	if !exists {
		limit = defaultLimit
	} else {
		if limitStr == "none" {
			limit = math.MaxInt32
		} else {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				// Bad limit, use default
				limit = defaultLimit
			}
		}
	}

	if limitStr != "none" && limit > maxLimit {
		limit = maxLimit
	}
	if limit < 1 {
		limit = 1
	}

	var page int
	pageStr, exists := c.GetQuery("page")
	if !exists {
		page = 1
	} else {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			// Bad page, use default
			page = 1
		}
	}

	if page < 1 {
		page = 1
	}

	return db.WithContext(c.Request.Context()).Scopes(pagination.NewPaginate(limit, page).Paginate)
}
