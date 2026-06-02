package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/shashisharma307703/vedantam/db/dbgen"
)

type ClassListAndSearchFilter struct {
	OrgID      uuid.UUID // Enforces mandatory multi-tenant isolation boundaries
	SearchTerm string    // Matches against class names or description terms
	IsActive   *bool
	Limit      uint64
	Offset     uint64
}

// FindClasses executes "Get All" list operations and search term matching
func (r *Repository) FindClasses(ctx context.Context, f ClassListAndSearchFilter) ([]dbgen.ClassLevel, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	qb := psql.Select("class_level_id", "class_name", "class_code", "display_order", "description", "is_lower_primary", "is_upper_primary", "is_secondary", "is_higher_secondary", "created_at").
		From("class_levels").
		Where(squirrel.Eq{"org_id": f.OrgID})

	if f.SearchTerm != "" {
		wildcard := fmt.Sprintf("%%%s%%", f.SearchTerm)
		qb = qb.Where(squirrel.Or{
			squirrel.ILike{"class_name": wildcard},
			squirrel.ILike{"description": wildcard},
		})
	}

	if f.IsActive != nil {
		qb = qb.Where(squirrel.Eq{"is_active": *f.IsActive})
	}

	if f.Limit > 0 {
		qb = qb.Limit(f.Limit)
	} else {
		qb = qb.Limit(20)
	}
	if f.Offset > 0 {
		qb = qb.Offset(f.Offset)
	}

	sqlStr, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build list class query: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed executing list class query: %w", err)
	}
	defer rows.Close()

	var classes []dbgen.ClassLevel
	for rows.Next() {
		var c dbgen.ClassLevel
		err := rows.Scan(
			&c.ClassLevelID, &c.ClassName, &c.ClassCode, &c.DisplayOrder, &c.Description,
			&c.IsLowerPrimary, &c.IsUpperPrimary, &c.IsSecondary, &c.IsHigherSecondary, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

// PatchClass patches changing variables using a dynamic multi-tenant update map
func (r *Repository) PatchClass(ctx context.Context, orgID, classID uuid.UUID, updates map[string]interface{}) (*dbgen.ClassLevel, error) {
	if len(updates) == 0 {
		// Uses the static sqlc generated function if no dynamic changes are passed
		c, err := r.Queries.GetClassByID(ctx, classID)
		return &c, err
	}

	updates["updated_at"] = squirrel.Expr("CURRENT_TIMESTAMP")

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	sqlStr, args, err := psql.Update("class_levels").
		SetMap(updates).
		Where(squirrel.Eq{"class_level_id": classID, "org_id": orgID}).
		Suffix("RETURNING class_level_id, class_name, class_code, display_order, description, is_lower_primary, is_upper_primary, is_secondary, is_higher_secondary, created_at").
		ToSql()

	if err != nil {
		return nil, err
	}

	var c dbgen.ClassLevel
	err = r.Pool.QueryRow(ctx, sqlStr, args...).Scan(
		&c.ClassLevelID, &c.ClassName, &c.ClassCode, &c.DisplayOrder, &c.Description,
		&c.IsLowerPrimary, &c.IsUpperPrimary, &c.IsSecondary, &c.IsHigherSecondary, &c.CreatedAt,
	)
	return &c, err
}
