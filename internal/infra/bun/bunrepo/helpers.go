package bunrepo

import (
	"context"

	"github.com/lynx-go/lynx-clean-template/pkg/crud"
	"github.com/lynx-go/lynx-clean-template/pkg/types"
	"github.com/uptrace/bun"
)

// mapFromMetadata converts types.Metadata to map[string]any for database storage
func mapFromMetadata(m types.Metadata) map[string]any {
	if m == nil {
		return nil
	}
	return map[string]any(m)
}

// mapToMetadata converts map[string]any to types.Metadata
func mapToMetadata(m map[string]any) types.Metadata {
	if len(m) == 0 {
		return types.Metadata{}
	}
	return types.Metadata(m)
}

// applyFilter applies AIP-160 filter to a bun query
func applyFilter(query *bun.SelectQuery, filterStr string, fieldMappings map[string]string) (*bun.SelectQuery, error) {
	if filterStr == "" {
		return query, nil
	}

	filter, err := crud.ParseFilter(filterStr)
	if err != nil {
		return nil, err
	}

	// Build validator from field mappings
	validator := crud.NewFilterValidator()
	for field := range fieldMappings {
		validator.AddMapping(field, fieldMappings[field])
	}

	if err := validator.Validate(filter); err != nil {
		return nil, err
	}

	var args []interface{}
	whereClause := crud.BuildSQLWhere(filter, fieldMappings, &args)
	if whereClause != "" {
		query = query.Where(whereClause, args...)
	}

	return query, nil
}

// applyOrderBy applies AIP-132 order by to a bun query
func applyOrderBy(query *bun.SelectQuery, orderByStr string, fieldMappings map[string]string) (*bun.SelectQuery, error) {
	if orderByStr == "" {
		return query, nil
	}

	// Build validator from field mappings
	validator := crud.NewOrderByValidator()
	for field := range fieldMappings {
		validator.AddFields(field)
	}

	orderByClauses, err := validator.Parse(orderByStr)
	if err != nil {
		return nil, err
	}

	orderBySQL := crud.BuildOrderByClause(orderByClauses, fieldMappings)
	if orderBySQL != "" {
		query = query.OrderExpr(orderBySQL)
	}

	return query, nil
}

// ListQueryConfig contains configuration for list query execution
type ListQueryConfig[M any, D any] struct {
	// FieldMappings maps API field names to SQL column names for filter/order by
	FieldMappings map[string]string

	// DefaultOrder is the default order by clause when no order is specified (e.g., "t.created_at DESC")
	DefaultOrder string

	// Converter converts model type to domain type
	Converter func(M) D
}

// ExecuteListQuery executes a standard list query with filter, order by, and pagination
// M: model type (Bun struct), D: domain type
//
// Example:
//
//	query := r.db.NewSelect().Where("tm.status = ?", 1)
//	config := ListQueryConfig[model.TeamMember, *teamsrepo.TeamMember]{
//	    FieldMappings: teamMemberFieldMappings,
//	    DefaultOrder:  "tm.created_at DESC",
//	    Converter:     toDomainTeamMember,
//	}
//	return ExecuteListQuery(ctx, query, params, config)
//
// Returns: (results, totalCount, nextPageToken, error)
func ExecuteListQuery[M any, D any](
	ctx context.Context,
	db *bun.DB,
	params crud.ListParams,
	config ListQueryConfig[M, D],
	queryModifier func(*bun.SelectQuery) *bun.SelectQuery,
) ([]D, int, string, error) {
	var models []M

	// Build base query with model
	query := db.NewSelect().Model(&models)

	// Apply custom query modifications (where clauses, joins, etc.)
	if queryModifier != nil {
		query = queryModifier(query)
	}

	// Apply filter (before count)
	query, err := applyFilter(query, params.Filter, config.FieldMappings)
	if err != nil {
		return nil, 0, "", err
	}

	// Count total (no order by needed)
	totalCount, err := query.Count(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	// Apply order by (after count, before scan)
	if params.OrderBy != "" {
		query, err = applyOrderBy(query, params.OrderBy, config.FieldMappings)
		if err != nil {
			return nil, 0, "", err
		}
	} else if config.DefaultOrder != "" {
		query = query.Order(config.DefaultOrder)
	}

	// Apply pagination
	offset, limit := params.OffsetLimit()
	query = query.Offset(offset).Limit(limit)

	// Execute query
	err = query.Scan(ctx)
	if err != nil {
		return nil, 0, "", err
	}

	// Convert to domain types
	results := make([]D, len(models))
	for i, m := range models {
		results[i] = config.Converter(m)
	}

	// Calculate next page token
	var nextPageToken string
	if offset+len(results) < totalCount {
		nextPageToken = crud.EncodePageToken(offset + len(results))
	}

	return results, totalCount, nextPageToken, nil
}
