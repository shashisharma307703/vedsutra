package onboarder

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shashisharma307703/vedantam/db/dbgen"
	"golang.org/x/crypto/bcrypt"
)

type Resolver struct {
	tenantID          uuid.UUID
	queries           *dbgen.Queries
	gradeCache        map[string]uuid.UUID
	academicYearCache map[string]uuid.UUID
	classCache        map[string]uuid.UUID
	departmentCache   map[string]uuid.UUID
	subjectCache      map[string]uuid.UUID
	userCache         map[string]uuid.UUID
	feeCache          map[string]uuid.UUID
}

func NewResolver(tenantID uuid.UUID, queries *dbgen.Queries) *Resolver {
	return &Resolver{
		tenantID:          tenantID,
		queries:           queries,
		gradeCache:        make(map[string]uuid.UUID),
		academicYearCache: make(map[string]uuid.UUID),
		classCache:        make(map[string]uuid.UUID),
		departmentCache:   make(map[string]uuid.UUID),
		subjectCache:      make(map[string]uuid.UUID),
		userCache:         make(map[string]uuid.UUID),
		feeCache:          make(map[string]uuid.UUID),
	}
}

func (r *Resolver) ResolveGrade(ctx context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.gradeCache[name]; ok {
		return id, nil
	}
	id, err := r.queries.GetGradeByName(ctx, dbgen.GetGradeByNameParams{
		TenantID: r.tenantID,
		Name:     name,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("grade %q not found: %w", name, err)
	}
	r.gradeCache[name] = id
	return id, nil
}

func (r *Resolver) ResolveAcademicYear(ctx context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.academicYearCache[name]; ok {
		return id, nil
	}
	id, err := r.queries.GetAcademicYearByName(ctx, dbgen.GetAcademicYearByNameParams{
		TenantID: r.tenantID,
		Name:     name,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("academic year %q not found: %w", name, err)
	}
	r.academicYearCache[name] = id
	return id, nil
}

func (r *Resolver) ResolveDepartment(ctx context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.departmentCache[name]; ok {
		return id, nil
	}
	id, err := r.queries.GetDepartmentByName(ctx, dbgen.GetDepartmentByNameParams{
		TenantID: r.tenantID,
		Name:     name,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("department %q not found: %w", name, err)
	}
	r.departmentCache[name] = id
	return id, nil
}

func (r *Resolver) ResolveSubject(ctx context.Context, code string) (uuid.UUID, error) {
	if id, ok := r.subjectCache[code]; ok {
		return id, nil
	}
	id, err := r.queries.GetSubjectByCode(ctx, dbgen.GetSubjectByCodeParams{
		TenantID: r.tenantID,
		Code:     code,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("subject code %q not found: %w", code, err)
	}
	r.subjectCache[code] = id
	return id, nil
}

func (r *Resolver) ResolveClass(ctx context.Context, className string) (uuid.UUID, error) {
	if id, ok := r.classCache[className]; ok {
		return id, nil
	}
	id, err := r.queries.GetClassByName(ctx, dbgen.GetClassByNameParams{
		TenantID: r.tenantID,
		Name:     className,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("class %q not found: %w", className, err)
	}
	r.classCache[className] = id
	return id, nil
}

// ResolveUserByEmail finds or creates a placeholder user (for teachers).
func (r *Resolver) ResolveUserByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	if email == "" {
		return uuid.Nil, nil
	}
	if id, ok := r.userCache[email]; ok {
		return id, nil
	}
	// Try to find existing user
	id, err := r.queries.GetUserByEmail(ctx, dbgen.GetUserByEmailParams{
		TenantID: pgtype.UUID{Bytes: r.tenantID, Valid: true},
		Email:    &email,
	})
	if err == nil {
		r.userCache[email] = id
		return id, nil
	}
	// Create placeholder user
	randomPwd := generateRandomPassword()
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(randomPwd), bcrypt.DefaultCost)
	username := strings.Split(email, "@")[0]
	newID := uuid.New()
	_, err = r.queries.UpsertPlaceholderUser(ctx, dbgen.UpsertPlaceholderUserParams{
		ID:           newID,
		TenantID:     pgtype.UUID{Bytes: r.tenantID, Valid: true},
		Username:     username,
		Email:        &email,
		PasswordHash: string(hashedPwd),
		FirstName:    username,
		LastName:     nil,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create placeholder user for %s: %w", email, err)
	}
	r.userCache[email] = newID
	return newID, nil
}

func generateRandomPassword() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}