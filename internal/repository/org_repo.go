package repository

// OrgListAndSearchFilter combines "Get All" pagination with optional "Search" criteria
type OrgListAndSearchFilter struct {
	SearchTerm string // For specific terms matching name or code
	City       string // Optional filter
	Limit      uint64
	Offset     uint64
}

// FindOrganizations handles List, Pagination, and Target Term Searching dynamically
// func (r *Repository) FindOrganizations(ctx context.Context, f OrgListAndSearchFilter) ([]dbgen.Organization, error) {
// 	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
// 	qb := psql.Select("org_id", "org_name", "org_code", "phone_number", "email", "website", "logo_url", "address", "city", "state", "country", "postal_code", "established_date", "affiliation_number", "license_number", "license_expiry", "created_at", "updated_at").
// 		From("organizations")

// 	// Apply specific search term criteria if requested
// 	if f.SearchTerm != "" {
// 		wildcard := fmt.Sprintf("%%%s%%", f.SearchTerm)
// 		qb = qb.Where(squirrel.Or{
// 			squirrel.ILike{"org_name": wildcard},
// 			squirrel.ILike{"org_code": wildcard},
// 			squirrel.ILike{"email": wildcard},
// 		})
// 	}

// 	// Apply exact structural filters if present
// 	if f.City != "" {
// 		qb = qb.Where(squirrel.Eq{"city": f.City})
// 	}

// 	// Apply safe default boundary pagination limits
// 	if f.Limit > 0 {
// 		qb = qb.Limit(f.Limit)
// 	} else {
// 		qb = qb.Limit(20)
// 	}
// 	if f.Offset > 0 {
// 		qb = qb.Offset(f.Offset)
// 	}

// 	sqlStr, args, err := qb.ToSql()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to build list organization query: %w", err)
// 	}

// 	rows, err := r.Pool.Query(ctx, sqlStr, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed executing list organization query: %w", err)
// 	}
// 	defer rows.Close()

// 	var orgs []dbgen.Organization
// 	for rows.Next() {
// 		var o dbgen.Organization
// 		err := rows.Scan(
// 			&o.OrgID, &o.OrgName, &o.OrgCode, &o.PhoneNumber, &o.Email, &o.Website,
// 			&o.LogoUrl, &o.Address, &o.City, &o.State, &o.Country, &o.PostalCode,
// 			&o.EstablishedDate, &o.AffiliationNumber, &o.LicenseNumber, &o.LicenseExpiry,
// 			&o.CreatedAt, &o.UpdatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		orgs = append(orgs, o)
// 	}
// 	return orgs, nil
// }

// // PatchOrganization performs a dynamic partial field update using a clean map input
// func (r *Repository) PatchOrganization(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*dbgen.Organization, error) {
// 	if len(updates) == 0 {
// 		org, err := r.Queries.GetOrganizationByID(ctx, id)
// 		return &org, err
// 	}

// 	updates["updated_at"] = squirrel.Expr("CURRENT_TIMESTAMP")

// 	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
// 	sqlStr, args, err := psql.Update("organizations").
// 		SetMap(updates).
// 		Where(squirrel.Eq{"org_id": id}).
// 		Suffix("RETURNING org_id, org_name, org_code, phone_number, email, website, logo_url, address, city, state, country, postal_code, established_date, affiliation_number, license_number, license_expiry, created_at, updated_at").
// 		ToSql()

// 	if err != nil {
// 		return nil, err
// 	}

// 	var o dbgen.Organization
// 	err = r.Pool.QueryRow(ctx, sqlStr, args...).Scan(
// 		&o.OrgID, &o.OrgName, &o.OrgCode, &o.PhoneNumber, &o.Email, &o.Website,
// 		&o.LogoUrl, &o.Address, &o.City, &o.State, &o.Country, &o.PostalCode,
// 		&o.EstablishedDate, &o.AffiliationNumber, &o.LicenseNumber, &o.LicenseExpiry,
// 		&o.CreatedAt, &o.UpdatedAt,
// 	)
// 	return &o, err
// }
