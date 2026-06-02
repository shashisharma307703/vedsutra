package onboarder

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shashisharma307703/vedantam/db/dbgen"
	"github.com/shashisharma307703/vedantam/internal/repository"
)

func CreateTenant(ctx context.Context, repo *repository.Repository, meta *TenantMeta) (uuid.UUID, error) {
	apiKey := make([]byte, 32)
	rand.Read(apiKey)
	apiKeyStr := base64.URLEncoding.EncodeToString(apiKey)

	var subPlanID uuid.UUID
	if meta.SubscriptionPlan != "" {
		id, err := repo.Queries.GetSubscriptionPlanIDByName(ctx, dbgen.GetSubscriptionPlanIDByNameParams{
			Name:          meta.SubscriptionPlan,
			BillingPeriod: dbgen.BillingPeriod(meta.BillingPeriod),
		})
		if err == nil {
			subPlanID = id
		}
	}

	trialEnds := time.Now().AddDate(0, 0, meta.TrialDays)
	if meta.TrialDays == 0 && meta.Status == "trial" {
		trialEnds = time.Now().AddDate(0, 0, 30)
	}
	status := meta.Status
	if status == "" {
		status = "trial"
	}

	tenantID := uuid.New()
	_, err := repo.Queries.CreateTenant(ctx, dbgen.CreateTenantParams{
		ID:                     tenantID,
		Name:                   meta.Name,
		Subdomain:              meta.Subdomain,
		Slug:                   meta.Slug,
		Timezone:               meta.Timezone,
		Country:                &meta.Country,
		State:                  &meta.State,
		City:                   &meta.City,
		Address:                &meta.Address,
		Pincode:                &meta.Pincode,
		ContactEmail:           &meta.ContactEmail,
		ContactPhone:           &meta.ContactPhone,
		Website:                &meta.Website,
		Status:                 dbgen.TenantStatus(status),
		SubscriptionPlanID:      pgtype.UUID{Bytes: subPlanID, Valid: true},
		SubscriptionStartDate:  pgtype.Date{Time: time.Now(), Valid: true},
		SubscriptionEndDate:    pgtype.Date{Valid: false},
		TrialEndsAt:            pgtype.Timestamptz{Time: trialEnds, Valid: true},
		ApiKey:                 &apiKeyStr,
		Metadata:               []byte("{}"),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create tenant: %w", err)
	}
	return tenantID, nil
}