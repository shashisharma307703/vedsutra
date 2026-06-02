package onboarder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shashisharma307703/vedantam/db/dbgen"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Loader struct {
	pool       interface{} // Use empty interface, but we know it's pgx.Tx capable
	queries    *dbgen.Queries
	upsert     bool
	dryRun     bool
	logger     Logger
	processors map[string]TableProcessor
}

func NewLoader(pool interface{}, queries *dbgen.Queries, upsert, dryRun bool, logger Logger) *Loader {
	return &Loader{
		pool:       pool,
		queries:    queries,
		upsert:     upsert,
		dryRun:     dryRun,
		logger:     logger,
		processors: map[string]TableProcessor{
			"academic_years":  &AcademicYearProcessor{},
			"grades":          &GradeProcessor{},
			"departments":     &DepartmentProcessor{},
			"subjects":        &SubjectProcessor{},
			"classes":         &ClassProcessor{},
			"sections":        &SectionProcessor{},
			"class_subjects":  &ClassSubjectProcessor{},
			"fee_structures":  &FeeStructureProcessor{},
			"tenant_settings": &TenantSettingProcessor{},
		},
	}
}

// LoadAllFromDirectory reads each table from individual files (CSV or JSON) in a directory.
func (l *Loader) LoadAllFromDirectory(ctx context.Context, tenantID uuid.UUID, dataDir string) error {
	txer, ok := l.pool.(interface{ BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) })
	if !ok {
		return fmt.Errorf("pool does not support transactions")
	}
	tx, err := txer.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := l.queries.WithTx(tx)
	resolver := NewResolver(tenantID, qtx)

	for tableName, processor := range l.processors {
		var foundPath string
		for _, ext := range []string{".csv", ".json"} {
			path := filepath.Join(dataDir, tableName+ext)
			if _, err := os.Stat(path); err == nil {
				foundPath = path
				break
			}
		}
		if foundPath == "" {
			l.logger.Info("No data file for %s, skipping", tableName)
			continue
		}
		if err := l.loadTableFile(ctx, resolver, processor, foundPath); err != nil {
			return fmt.Errorf("load %s: %w", tableName, err)
		}
	}

	// Seed roles if not present
	if !l.dryRun {
		count, err := qtx.CountTenantRoles(ctx, pgtype.UUID{Bytes: tenantID, Valid: true})
		if err != nil {
			return err
		}
		if count == 0 {
			l.logger.Info("Seeding default tenant roles...")
			if err := qtx.CallSeedTenantRoles(ctx, tenantID); err != nil {
				return err
			}
		}
	}

	if l.dryRun {
		l.logger.Info("Dry run completed – no changes made")
		return nil
	}
	return tx.Commit(ctx)
}

// LoadAllFromJSONBundle reads a single JSON file containing all tables.
func (l *Loader) LoadAllFromJSONBundle(ctx context.Context, tenantID uuid.UUID, bundlePath string) error {
	f, err := os.Open(bundlePath)
	if err != nil {
		return err
	}
	defer f.Close()
	var bundle map[string][]map[string]interface{}
	if err := json.NewDecoder(f).Decode(&bundle); err != nil {
		return err
	}

	txer, ok := l.pool.(interface{ BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) })
	if !ok {
		return fmt.Errorf("pool does not support transactions")
	}
	tx, err := txer.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := l.queries.WithTx(tx)
	resolver := NewResolver(tenantID, qtx)

	for tableName, rows := range bundle {
		processor, ok := l.processors[tableName]
		if !ok {
			l.logger.Info("No processor for table %s, skipping", tableName)
			continue
		}
		for i, row := range rows {
			if err := processor.Process(ctx, resolver, row, l.upsert); err != nil {
				return fmt.Errorf("processing %s row %d: %w", tableName, i+1, err)
			}
		}
	}

	// Seed roles
	if !l.dryRun {
		count, err := qtx.CountTenantRoles(ctx, pgtype.UUID{Bytes: tenantID, Valid: true})
		if err != nil {
			return err
		}
		if count == 0 {
			l.logger.Info("Seeding default tenant roles...")
			if err := qtx.CallSeedTenantRoles(ctx, tenantID); err != nil {
				return err
			}
		}
	}

	if l.dryRun {
		l.logger.Info("Dry run completed – no changes made")
		return nil
	}
	return tx.Commit(ctx)
}

func (l *Loader) loadTableFile(ctx context.Context, resolver *Resolver, processor TableProcessor, filePath string) error {
	reader, err := DetectReader(filePath)
	if err != nil {
		return err
	}
	rows, err := ReadRowsAsMaps(reader)
	if err != nil {
		return err
	}
	for i, row := range rows {
		if err := processor.Process(ctx, resolver, row, l.upsert); err != nil {
			return fmt.Errorf("row %d: %w", i+2, err)
		}
	}
	return nil
}