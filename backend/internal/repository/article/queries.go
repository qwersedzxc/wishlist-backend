package article

import (
	"github.com/KaoriEl/golang-boilerplate/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

// Константы SQL-запросов и имя таблицы.
const (
	table = "articles"

	colID          = "id"
	colTitle       = "title"
	colBody        = "body"
	colAuthorID    = "author_id"
	colPublishedAt = "published_at"
	colCreatedAt   = "created_at"
	colUpdatedAt   = "updated_at"

	allColumns = "id, title, body, author_id, published_at, created_at, updated_at"
)

func buildInsert(b sq.StatementBuilderType, row articleRow) sq.InsertBuilder {
	return b.Insert(table).
		Columns(colTitle, colBody, colAuthorID, colPublishedAt).
		Values(row.Title, row.Body, row.AuthorID, row.PublishedAt).
		Suffix("RETURNING " + allColumns)
}

func buildSelectByID(b sq.StatementBuilderType, id uuid.UUID) sq.SelectBuilder {
	return b.Select(allColumns).From(table).Where(sq.Eq{colID: id})
}

func buildSelectList(b sq.StatementBuilderType, filter dto.ArticleFilter) sq.SelectBuilder {
	offset := uint64((filter.Page - 1) * filter.PerPage) //nolint:gosec

	return b.Select(allColumns).
		From(table).
		OrderBy(colCreatedAt + " DESC").
		Limit(uint64(filter.PerPage)). //nolint:gosec
		Offset(offset)
}

func buildCount(b sq.StatementBuilderType, _ dto.ArticleFilter) sq.SelectBuilder {
	return b.Select("COUNT(*)").From(table)
}

func buildUpdate(b sq.StatementBuilderType, input dto.UpdateArticleInput) sq.UpdateBuilder {
	upd := b.Update(table).Where(sq.Eq{colID: input.ID})
	if input.Title != nil {
		upd = upd.Set(colTitle, *input.Title)
	}
	if input.Body != nil {
		upd = upd.Set(colBody, *input.Body)
	}

	return upd.Suffix("RETURNING " + allColumns)
}

func buildDelete(b sq.StatementBuilderType, id uuid.UUID) sq.DeleteBuilder {
	return b.Delete(table).Where(sq.Eq{colID: id})
}
