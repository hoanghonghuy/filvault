package postgres

import (
	"context"
	"fmt"
	"strings"

	"filvault/internal/search"

	"github.com/jackc/pgx/v5"
)

const searchFolderSelect = `
	SELECT id, parent_id, name, created_at, updated_at
	FROM folders
	WHERE owner_id = $1 AND deleted_at IS NULL
`

const searchFileSelect = `
	SELECT id, name, mime_type, size_bytes, created_at, updated_at
	FROM files
	WHERE owner_id = $1 AND deleted_at IS NULL AND status = 'READY'
`

// folderSubtreeSQL returns a recursive CTE selecting the folder and all its
// descendants; ph is the placeholder holding the root folder id.
func folderSubtreeSQL(ph string) string {
	return fmt.Sprintf(`WITH RECURSIVE sub(id) AS (
			SELECT id FROM folders WHERE id = %s AND deleted_at IS NULL
			UNION ALL
			SELECT f.id FROM folders f JOIN sub ON f.parent_id = sub.id WHERE f.deleted_at IS NULL
		) SELECT id FROM sub`, ph)
}

func mimeGroupsFor(t search.TypeFilter) []string {
	switch t {
	case search.TypeImage:
		return []string{"image/%"}
	case search.TypeVideo:
		return []string{"video/%"}
	case search.TypeDocument:
		return []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"text/%",
		}
	case search.TypeArchive:
		return []string{"application/zip"}
	default:
		return nil
	}
}

func orderClause(column string, order search.Order) string {
	if order == search.OrderAsc {
		return fmt.Sprintf(" ORDER BY %s ASC", column)
	}
	return fmt.Sprintf(" ORDER BY %s DESC", column)
}

func folderSortColumn(sort search.SortField) (string, bool) {
	switch sort {
	case search.SortName:
		return "name", true
	case search.SortDate:
		return "updated_at", true
	default:
		return "", false
	}
}

func fileSortColumn(sort search.SortField) (string, bool) {
	switch sort {
	case search.SortName:
		return "name", true
	case search.SortDate:
		return "updated_at", true
	case search.SortSize:
		return "size_bytes", true
	default:
		return "", false
	}
}

type searchSQL struct {
	sb   strings.Builder
	args []any
}

func newSearchSQL(base string, ownerID string) *searchSQL {
	q := &searchSQL{args: []any{ownerID}}
	q.sb.WriteString(base)
	return q
}

func (q *searchSQL) addArg(value any) string {
	q.args = append(q.args, value)
	return fmt.Sprintf("$%d", len(q.args))
}

func (q *searchSQL) appendLimit(limit int) {
	fmt.Fprintf(&q.sb, " LIMIT %d", limit)
}

func (s *Store) SearchFolders(ctx context.Context, ownerID string, q search.Query) ([]search.FolderHit, error) {
	sql := newSearchSQL(searchFolderSelect, ownerID)
	textArg := sql.addArg(ilikePatternArg(q.Text))
	fmt.Fprintf(&sql.sb, " AND name ILIKE '%%' || %s || '%%' ESCAPE '\\'", textArg)
	if q.FolderID != nil {
		fmt.Fprintf(&sql.sb, " AND id IN (%s)", folderSubtreeSQL(sql.addArg(*q.FolderID)))
	}
	if col, ok := folderSortColumn(q.Sort); ok {
		sql.sb.WriteString(orderClause(col, q.Order))
	} else {
		sql.sb.WriteString(" ORDER BY name ASC")
	}
	sql.appendLimit(q.Limit)

	rows, err := s.pool.Query(ctx, sql.sb.String(), sql.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []search.FolderHit
	for rows.Next() {
		var f search.FolderHit
		if err := rows.Scan(&f.ID, &f.ParentID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Store) SearchFiles(ctx context.Context, ownerID string, q search.Query) ([]search.FileHit, error) {
	sql := newSearchSQL(searchFileSelect, ownerID)
	textArg := sql.addArg(ilikePatternArg(q.Text))
	fmt.Fprintf(&sql.sb, " AND name ILIKE '%%' || %s || '%%' ESCAPE '\\'", textArg)

	if groups := mimeGroupsFor(q.Type); len(groups) > 0 {
		patterns := make([]string, 0, len(groups))
		for _, g := range groups {
			patterns = append(patterns, fmt.Sprintf("mime_type LIKE %s", sql.addArg(g)))
		}
		fmt.Fprintf(&sql.sb, " AND (%s)", strings.Join(patterns, " OR "))
	}
	if q.FolderID != nil {
		fmt.Fprintf(&sql.sb, " AND folder_id IN (%s)", folderSubtreeSQL(sql.addArg(*q.FolderID)))
	}
	if q.From != nil {
		fmt.Fprintf(&sql.sb, " AND created_at >= %s", sql.addArg(*q.From))
	}
	if q.To != nil {
		fmt.Fprintf(&sql.sb, " AND created_at < (%s::timestamptz + interval '1 day')", sql.addArg(*q.To))
	}
	col, ok := fileSortColumn(q.Sort)
	if !ok {
		col = "name"
	}
	sql.sb.WriteString(orderClause(col, q.Order))
	sql.appendLimit(q.Limit)

	rows, err := s.pool.Query(ctx, sql.sb.String(), sql.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []search.FileHit
	for rows.Next() {
		var f search.FileHit
		if err := rows.Scan(&f.ID, &f.Name, &f.MimeType, &f.SizeBytes, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func ilikePatternArg(text string) string {
	p := text
	p = strings.ReplaceAll(p, `\`, `\\`)
	p = strings.ReplaceAll(p, `%`, `\%`)
	p = strings.ReplaceAll(p, `_`, `\_`)
	return p
}

// FolderExists reports whether an alive folder with this id is owned by ownerID.
func (s *Store) FolderExists(ctx context.Context, ownerID, folderID string) (bool, error) {
	var id string
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM folders
		WHERE owner_id = $1 AND id = $2 AND deleted_at IS NULL
	`, ownerID, folderID).Scan(&id)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type searchRepo struct {
	store *Store
}

func NewSearchRepository(store *Store) search.Repository {
	return searchRepo{store: store}
}

func (r searchRepo) FolderExists(ctx context.Context, ownerID, folderID string) (bool, error) {
	return r.store.FolderExists(ctx, ownerID, folderID)
}

func (r searchRepo) SearchFolders(ctx context.Context, ownerID string, q search.Query) ([]search.FolderHit, error) {
	return r.store.SearchFolders(ctx, ownerID, q)
}

func (r searchRepo) SearchFiles(ctx context.Context, ownerID string, q search.Query) ([]search.FileHit, error) {
	return r.store.SearchFiles(ctx, ownerID, q)
}

var _ search.Repository = searchRepo{}
