package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/gin-gonic/gin"
)

// multiSplit is like strings.Split except it splits a slice of strings into a slice of slices.
func multiSplit(strs []string, delimeter string) [][]string {
	res := [][]string{}

	for _, str := range strs {
		res = append(res, strings.Split(str, delimeter))
	}

	return res
}

// requestToCollectionQuery converts a request's parameters into an entries.Query struct.
func requestToCollectionQuery(c *gin.Context) entries.Query {
	delimeter := c.Query("delimeter")
	if delimeter == "" {
		delimeter = " OR "
	}

	dateFormat := c.Query("date-format")
	if dateFormat == "" {
		dateFormat = "2006-01-02 15:04"
	}

	fromStr := c.Query("from")
	untilStr := c.Query("until")
	minLengthStr := c.Query("min-length")
	maxLengthStr := c.Query("max-length")

	tags := c.QueryArray("tag")
	tagsExclude := c.QueryArray("tag-not")

	pathsMatch := c.QueryArray("path")
	pathsExact := c.QueryArray("path-exact")
	pathsMatchNot := c.QueryArray("path-not")
	pathsExactNot := c.QueryArray("path-exact-not")
	titlesMatch := c.QueryArray("title")
	titlesExact := c.QueryArray("title-exact")
	titlesMatchNot := c.QueryArray("title-not")
	titlesExactNot := c.QueryArray("title-exact-not")
	contentsMatch := c.QueryArray("contents")
	contentsExact := c.QueryArray("contents-exact")
	contentsMatchNot := c.QueryArray("contents-not")
	contentsExactNot := c.QueryArray("contents-exact-not")

	var from, until time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(dateFormat, fromStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_type": "error parsing date",
				"error":      err.Error(),
			})
			return entries.Query{}
		}
	}

	if untilStr != "" {
		until, err = time.Parse(dateFormat, untilStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_type": "error parsing date",
				"error":      err.Error(),
			})
			return entries.Query{}
		}
	}

	var minLength, maxLength int

	if minLengthStr != "" {
		minLength, err = strconv.Atoi(minLengthStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_type": "error parsing length",
				"error":      err.Error(),
			})
			return entries.Query{}
		}
	}

	if maxLengthStr != "" {
		maxLength, err = strconv.Atoi(maxLengthStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_type": "error parsing length",
				"error":      err.Error(),
			})
			return entries.Query{}
		}
	}

	return entries.Query{
		From:  from,
		Until: until,

		MinLength: minLength,
		MaxLength: maxLength,

		Tags:        tags,
		TagsExclude: tagsExclude,

		ContentsExact:        multiSplit(contentsExact, delimeter),
		ContentsMatch:        multiSplit(contentsMatch, delimeter),
		ContentsExactExclude: multiSplit(contentsExactNot, delimeter),
		ContentsMatchExclude: multiSplit(contentsMatchNot, delimeter),

		PathsExact:        multiSplit(pathsExact, delimeter),
		PathsMatch:        multiSplit(pathsMatch, delimeter),
		PathsExactExclude: multiSplit(pathsExactNot, delimeter),
		PathsMatchExclude: multiSplit(pathsMatchNot, delimeter),

		TitlesExact:        multiSplit(titlesExact, delimeter),
		TitlesMatch:        multiSplit(titlesMatch, delimeter),
		TitlesExactExclude: multiSplit(titlesExactNot, delimeter),
		TitlesMatchExclude: multiSplit(titlesMatchNot, delimeter),
	}
}

// searchHandler handles requests for searching.
func (s *Server) searchHandler(c *gin.Context) {
	query := requestToCollectionQuery(c)
	if c.IsAborted() {
		return
	}

	fmt.Println(query)

	filter := query.Filter()

	filtered, err := s.collection.Filter(filter)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error_type": "error filtering collection",
			"error":      err.Error(),
		})
		return
	}

	if filtered.Len() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"matched": 0,
			"entries": []string{},
		})

		return
	}

	number := c.Query("number")
	rev := c.Query("rev")
	sort := c.Query("sort")

	list := filtered.List()

	switch sort {
	case "alpha":
		list = list.Sort(entries.SortAlpha)
	case "date":
		list = list.Sort(entries.SortDate)
	}

	if rev == "true" {
		list = list.Reverse()
	}

	if number != "" {
		num, err := strconv.Atoi(number)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error_type": "error parsing number",
				"error":      err.Error(),
			})
			return
		}

		list = list.First(num)
	}

	c.JSON(http.StatusOK, gin.H{
		"matched": filtered.Len(),
		"entries": list.Slice(),
	})

	return
}
