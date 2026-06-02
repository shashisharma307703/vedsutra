package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/shashisharma307703/vedantam/config"
	"github.com/shashisharma307703/vedantam/internal/onboarder"
	"github.com/shashisharma307703/vedantam/internal/repository"
	"github.com/spf13/cobra"
)

var (
	tenantFile  string
	dataDir     string
	bundleFile  string
	upsertFlag  bool
	dryRunFlag  bool

	name      string
	subdomain string
	slug      string
	plan      string
	period    string
	trialDays int
)

func main() {
	rootCmd := &cobra.Command{Use: "schoolctl"}

	createCmd := &cobra.Command{
		Use:   "create-tenant",
		Short: "Create a new tenant and load initial data",
		RunE:  runCreateTenant,
	}
	createCmd.Flags().StringVar(&tenantFile, "tenant-file", "", "JSON file with tenant metadata")
	createCmd.Flags().StringVar(&name, "name", "", "Tenant name")
	createCmd.Flags().StringVar(&subdomain, "subdomain", "", "Subdomain")
	createCmd.Flags().StringVar(&slug, "slug", "", "Slug")
	createCmd.Flags().StringVar(&plan, "plan", "Basic", "Subscription plan")
	createCmd.Flags().StringVar(&period, "period", "monthly", "Billing period")
	createCmd.Flags().IntVar(&trialDays, "trial-days", 30, "Trial days")
	createCmd.Flags().StringVar(&dataDir, "data-dir", "", "Directory with table files")
	createCmd.Flags().StringVar(&bundleFile, "bundle", "", "JSON bundle file")
	createCmd.Flags().BoolVar(&upsertFlag, "upsert", false, "Update existing records")
	createCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Validate only")
	rootCmd.AddCommand(createCmd)

	onboardCmd := &cobra.Command{
		Use:   "onboard",
		Short: "Load data into an existing tenant",
		RunE:  runOnboard,
	}
	onboardCmd.Flags().String("tenant-id", "", "Tenant UUID")
	onboardCmd.Flags().StringVar(&dataDir, "data-dir", "", "Directory with table files")
	onboardCmd.Flags().StringVar(&bundleFile, "bundle", "", "JSON bundle file")
	onboardCmd.Flags().BoolVar(&upsertFlag, "upsert", false, "Update existing records")
	onboardCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry run")
	onboardCmd.MarkFlagRequired("tenant-id")
	rootCmd.AddCommand(onboardCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runCreateTenant(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	pool, err := repository.InitPool(context.Background(), cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)
	logger := &cliLogger{}

	var meta onboarder.TenantMeta
	if tenantFile != "" {
		f, err := os.Open(tenantFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&meta); err != nil {
			return err
		}
	} else {
		meta = onboarder.TenantMeta{
			Name:              name,
			Subdomain:         subdomain,
			Slug:              slug,
			SubscriptionPlan:  plan,
			BillingPeriod:     period,
			TrialDays:         trialDays,
			Status:            "trial",
		}
		if meta.Slug == "" {
			meta.Slug = meta.Subdomain
		}
	}

	tenantID, err := onboarder.CreateTenant(context.Background(), repo, &meta)
	if err != nil {
		return err
	}
	logger.Info("Created tenant ID: %s", tenantID)

	if dataDir != "" || bundleFile != "" {
		loader := onboarder.NewLoader(repo.Pool, repo.Queries, upsertFlag, dryRunFlag, logger)
		if dataDir != "" {
			err = loader.LoadAllFromDirectory(context.Background(), tenantID, dataDir)
		} else {
			err = loader.LoadAllFromJSONBundle(context.Background(), tenantID, bundleFile)
		}
		if err != nil {
			return err
		}
	}
	fmt.Printf("Tenant onboarded: %s\n", tenantID)
	return nil
}

func runOnboard(cmd *cobra.Command, args []string) error {
	tenantIDStr, _ := cmd.Flags().GetString("tenant-id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return err
	}
	cfg := config.Load()
	pool, err := repository.InitPool(context.Background(), cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()
	repo := repository.NewRepository(pool)
	logger := &cliLogger{}

	loader := onboarder.NewLoader(repo.Pool, repo.Queries, upsertFlag, dryRunFlag, logger)
	if dataDir != "" {
		return loader.LoadAllFromDirectory(context.Background(), tenantID, dataDir)
	} else if bundleFile != "" {
		return loader.LoadAllFromJSONBundle(context.Background(), tenantID, bundleFile)
	}
	return fmt.Errorf("no data source provided")
}

type cliLogger struct{}

func (l *cliLogger) Info(msg string, args ...interface{}) { log.Printf("INFO: "+msg, args...) }
func (l *cliLogger) Debug(msg string, args ...interface{}) { log.Printf("DEBUG: "+msg, args...) }
func (l *cliLogger) Error(msg string, args ...interface{}) { log.Printf("ERROR: "+msg, args...) }