package service

import (
	"github.com/shashisharma307703/vedantam/internal/repository"
)

type OrgService struct {
	repo *repository.Repository
}

func NewOrgService(repo *repository.Repository) *OrgService {
	return &OrgService{repo: repo}
}

// func (s *OrgService) Create(ctx context.Context, p dbgen.UpsertOrganizationParams) (dbgen.Organization, error) {
// 	p.OrgID = uuid.New() // Enforce Application-Generated UUID Validation Boundaries
// 	return s.repo.UpsertOrganization(ctx, p)
// }

// func (s *OrgService) Get(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
// 	return s.repo.GetOrganizationByID(ctx, id)
// }

// func (s *OrgService) Update(ctx context.Context, p dbgen.ReplaceOrganizationParams) (dbgen.Organization, error) {
// 	return s.repo.ReplaceOrganization(ctx, p)
// }

// func (s *OrgService) Patch(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*dbgen.Organization, error) {
// 	return s.repo.PatchOrganization(ctx, id, updates)
// }

// func (s *OrgService) Delete(ctx context.Context, id uuid.UUID) error {
// 	return s.repo.DeleteOrganization(ctx, id)
// }

// func (s *OrgService) List(ctx context.Context, f repository.OrgListAndSearchFilter) ([]dbgen.Organization, error) {
// 	return s.repo.FindOrganizations(ctx, f)
// }
