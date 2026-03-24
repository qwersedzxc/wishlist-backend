package dto_test

import (
	"testing"

	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	"github.com/stretchr/testify/assert"
)

func TestCreateArticleInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		input := dto.CreateArticleInput{
			Title: "Hello World",
			Body:  "Some body text",
		}
		assert.NotEmpty(t, input.Title)
		assert.NotEmpty(t, input.Body)
	})

	t.Run("empty title", func(t *testing.T) {
		input := dto.CreateArticleInput{Body: "Some body"}
		assert.Empty(t, input.Title)
	})
}

func TestArticleFilter_Defaults(t *testing.T) {
	f := dto.ArticleFilter{Page: 1, PerPage: 20}
	assert.Equal(t, 1, f.Page)
	assert.Equal(t, 20, f.PerPage)
}
