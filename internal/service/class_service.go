package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shashisharma307703/vedantam/db/dbgen"
	"github.com/shashisharma307703/vedantam/internal/repository"
)

type ClassService struct {
	repo *repository.Repository
}

func NewClassService(repo *repository.Repository) *ClassService {
	return &ClassService{repo: repo}
}

func (s *ClassService) Create(ctx context.Context, p dbgen.UpsertClassParams) (dbgen.ClassLevel, error) {
	p.ClassLevelID = uuid.New()
	return s.repo.UpsertClass(ctx, p)
}

func (s *ClassService) Get(ctx context.Context, orgID, classID uuid.UUID) (dbgen.ClassLevel, error) {
	return s.repo.GetClassByID(ctx, classID)
}

func (s *ClassService) Update(ctx context.Context, p dbgen.ReplaceClassParams) (dbgen.ClassLevel, error) {
	return s.repo.ReplaceClass(ctx, p)
}

func (s *ClassService) Patch(ctx context.Context, orgID, classID uuid.UUID, updates map[string]interface{}) (*dbgen.ClassLevel, error) {
	return s.repo.PatchClass(ctx, orgID, classID, updates)
}

func (s *ClassService) Delete(ctx context.Context, orgID, classID uuid.UUID) error {
	return s.repo.DeleteClass(ctx, classID)
}

func (s *ClassService) List(ctx context.Context, f repository.ClassListAndSearchFilter) ([]dbgen.ClassLevel, error) {
	return s.repo.FindClasses(ctx, f)
}
