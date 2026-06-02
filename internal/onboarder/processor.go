package onboarder

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shashisharma307703/vedantam/db/dbgen"
)

type TableProcessor interface {
	Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, upsert bool) error
	TableName() string
}

// AcademicYearProcessor
type AcademicYearProcessor struct{}

func (p *AcademicYearProcessor) TableName() string { return "academic_years" }

func (p *AcademicYearProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	name := getString(row, "name")
	startDate, _ := time.Parse("2006-01-02", getString(row, "start_date"))
	endDate, _ := time.Parse("2006-01-02", getString(row, "end_date"))
	isCurrent := getBool(row, "is_current")

	id := uuid.New()
	_, err := resolver.queries.UpsertAcademicYear(ctx, dbgen.UpsertAcademicYearParams{
		ID:        id,
		TenantID:  resolver.tenantID,
		Name:      name,
		StartDate: toPGDate(startDate),
		EndDate:   toPGDate(endDate),
		IsCurrent: isCurrent,
	})
	if err == nil {
		resolver.academicYearCache[name] = id
	}
	return err
}

// GradeProcessor
type GradeProcessor struct{}

func (p *GradeProcessor) TableName() string { return "grades" }

func (p *GradeProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	name := getString(row, "name")
	seq := int16(getInt(row, "sequence"))
	desc := getString(row, "description")

	id := uuid.New()
	_, err := resolver.queries.UpsertGrade(ctx, dbgen.UpsertGradeParams{
		ID:          id,
		TenantID:    resolver.tenantID,
		Name:        name,
		Sequence:    seq,
		Description: &desc,
	})
	if err == nil {
		resolver.gradeCache[name] = id
	}
	return err
}

// DepartmentProcessor
type DepartmentProcessor struct{}

func (p *DepartmentProcessor) TableName() string { return "departments" }

func (p *DepartmentProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	name := getString(row, "name")
	desc := getString(row, "description")
	var headTeacherID *uuid.UUID
	if email, ok := row["head_teacher_email"]; ok && email != nil && email != "" {
		uid, err := resolver.ResolveUserByEmail(ctx, email.(string))
		if err != nil {
			return err
		}
		if uid != uuid.Nil {
			headTeacherID = &uid
		}
	}

	id := uuid.New()
	_, err := resolver.queries.UpsertDepartment(ctx, dbgen.UpsertDepartmentParams{
		ID:            id,
		TenantID:      resolver.tenantID,
		Name:          name,
		Description:   &desc,
		HeadTeacherID: toPGUUIDPtr(headTeacherID),
	})
	if err == nil {
		resolver.departmentCache[name] = id
	}
	return err
}

// SubjectProcessor
type SubjectProcessor struct{}

func (p *SubjectProcessor) TableName() string { return "subjects" }

func (p *SubjectProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	name := getString(row, "name")
	code := getString(row, "code")
	deptName := getString(row, "department_name")
	gradeName := getString(row, "grade_name")
	isElective := getBool(row, "is_elective")
	credits := getFloat32(row, "credits")

	var deptID *uuid.UUID
	if deptName != "" {
		id, err := resolver.ResolveDepartment(ctx, deptName)
		if err != nil {
			return err
		}
		deptID = &id
	}
	var gradeID *uuid.UUID
	if gradeName != "" {
		id, err := resolver.ResolveGrade(ctx, gradeName)
		if err != nil {
			return err
		}
		gradeID = &id
	}

	id := uuid.New()
	_, err := resolver.queries.UpsertSubject(ctx, dbgen.UpsertSubjectParams{
		ID:           id,
		TenantID:     resolver.tenantID,
		Name:         name,
		Code:         code,
		DepartmentID: toPGUUIDPtr(deptID),
		GradeID:      toPGUUIDPtr(gradeID),
		IsElective:   isElective,
		Credits:      toPGNumericFromFloat32(credits),
	})
	if err == nil {
		resolver.subjectCache[code] = id
	}
	return err
}

// ClassProcessor
type ClassProcessor struct{}

func (p *ClassProcessor) TableName() string { return "classes" }

func (p *ClassProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	gradeName := getString(row, "grade_name")
	ayName := getString(row, "academic_year_name")
	className := getString(row, "class_name")
	maxStudents := int32(getInt(row, "max_students"))
	roomNumber := getString(row, "room_number")

	gradeID, err := resolver.ResolveGrade(ctx, gradeName)
	if err != nil {
		return err
	}
	ayID, err := resolver.ResolveAcademicYear(ctx, ayName)
	if err != nil {
		return err
	}

	id := uuid.New()
	_, err = resolver.queries.UpsertClass(ctx, dbgen.UpsertClassParams{
		ID:             id,
		TenantID:       resolver.tenantID,
		GradeID:        gradeID,
		AcademicYearID: ayID,
		Name:           className,
		MaxStudents:    &maxStudents,
		RoomNumber:     &roomNumber,
	})
	if err == nil {
		resolver.classCache[className] = id
	}
	return err
}

// SectionProcessor
type SectionProcessor struct{}

func (p *SectionProcessor) TableName() string { return "sections" }

func (p *SectionProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	className := getString(row, "class_name")
	sectionName := getString(row, "section_name")
	maxStudents := int32(getInt(row, "max_students"))

	classID, err := resolver.ResolveClass(ctx, className)
	if err != nil {
		return err
	}
	var teacherID *uuid.UUID
	if email, ok := row["teacher_email"]; ok && email != nil && email != "" {
		uid, err := resolver.ResolveUserByEmail(ctx, email.(string))
		if err != nil {
			return err
		}
		if uid != uuid.Nil {
			teacherID = &uid
		}
	}

	id := uuid.New()
	_, err = resolver.queries.UpsertSection(ctx, dbgen.UpsertSectionParams{
		ID:          id,
		ClassID:     classID,
		Name:        sectionName,
		TeacherID:   toPGUUIDPtr(teacherID),
		MaxStudents: &maxStudents,
	})
	return err
}

// ClassSubjectProcessor
type ClassSubjectProcessor struct{}

func (p *ClassSubjectProcessor) TableName() string { return "class_subjects" }

func (p *ClassSubjectProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	className := getString(row, "class_name")
	subjectCode := getString(row, "subject_code")
	periods := int16(getInt(row, "periods_per_week"))

	classID, err := resolver.ResolveClass(ctx, className)
	if err != nil {
		return err
	}
	subjectID, err := resolver.ResolveSubject(ctx, subjectCode)
	if err != nil {
		return err
	}
	var teacherID *uuid.UUID
	if email, ok := row["teacher_email"]; ok && email != nil && email != "" {
		uid, err := resolver.ResolveUserByEmail(ctx, email.(string))
		if err != nil {
			return err
		}
		if uid != uuid.Nil {
			teacherID = &uid
		}
	}

	id := uuid.New()
	_, err = resolver.queries.UpsertClassSubject(ctx, dbgen.UpsertClassSubjectParams{
		ID:             id,
		ClassID:        classID,
		SubjectID:      subjectID,
		TeacherID:      toPGUUIDPtr(teacherID),
		PeriodsPerWeek: &periods,
	})
	return err
}

// FeeStructureProcessor
type FeeStructureProcessor struct{}

func (p *FeeStructureProcessor) TableName() string { return "fee_structures" }

func (p *FeeStructureProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	ayName := getString(row, "academic_year_name")
	className := getString(row, "class_name")
	feeType := getString(row, "fee_type")
	name := getString(row, "name")
	amount := getFloat64(row, "amount")
	frequency := getString(row, "frequency")
	dueDay := getInt16Ptr(row, "due_day")
	dueDate := getTimePtr(row, "due_date")
	lateFeePerDay := getFloat64(row, "late_fee_per_day")
	isMandatory := getBool(row, "is_mandatory")
	description := getString(row, "description")

	ayID, err := resolver.ResolveAcademicYear(ctx, ayName)
	if err != nil {
		return err
	}
	classID, err := resolver.ResolveClass(ctx, className)
	if err != nil {
		return err
	}
	
	id := uuid.New()
	_, err = resolver.queries.UpsertFeeStructure(ctx, dbgen.UpsertFeeStructureParams{
		ID:             id,
		TenantID:       resolver.tenantID,
		AcademicYearID: ayID,
		ClassID:        toPGUUID(classID),
		FeeType:        dbgen.FeeType(feeType),
		Name:           name,
		Amount:         toPGNumeric(amount),
		Frequency:      dbgen.FeeFrequency(frequency),
		DueDay:         dueDay,
		DueDate:        toPGDatePtr(dueDate),
		LateFeePerDay:  toPGNumeric(lateFeePerDay),
		IsMandatory:    isMandatory,
		Description:    &description,
	})
	return err
}

// TenantSettingProcessor
type TenantSettingProcessor struct{}

func (p *TenantSettingProcessor) TableName() string { return "tenant_settings" }

func (p *TenantSettingProcessor) Process(ctx context.Context, resolver *Resolver, row map[string]interface{}, _ bool) error {
	sessionTimeout := int32(getInt(row, "session_timeout_min"))
	maxUsers := getInt32Ptr(row, "max_users")
	maxStudents := getInt32Ptr(row, "max_students")
	ayFormat := getString(row, "academic_year_format")
	dateFormat := getString(row, "date_format")
	currency := getString(row, "currency_code")
	grading := getString(row, "grading_system")
	attendanceMode := getString(row, "attendance_mode")
	smsGateway := getString(row, "sms_gateway")
	emailProvider := getString(row, "email_provider")
	featuresJSON := getString(row, "features_enabled")
	themeJSON := getString(row, "theme")

	id := uuid.New()
	_, err := resolver.queries.UpsertTenantSetting(ctx, dbgen.UpsertTenantSettingParams{
		ID:                 id,
		TenantID:           resolver.tenantID,
		SessionTimeoutMin:  sessionTimeout,
		MaxUsers:           maxUsers,
		MaxStudents:        maxStudents,
		AcademicYearFormat: &ayFormat,
		DateFormat:         &dateFormat,
		CurrencyCode:       &currency,
		GradingSystem:      &grading,
		AttendanceMode:     &attendanceMode,
		SmsGateway:         &smsGateway,
		EmailProvider:      &emailProvider,
		FeaturesEnabled:    json.RawMessage(featuresJSON),
		Theme:              json.RawMessage(themeJSON),
	})
	return err
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true" || val == "1" || val == "yes"
		case float64:
			return val != 0
		}
	}
	return false
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		case string:
			var i int
			fmt.Sscanf(val, "%d", &i)
			return i
		}
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		}
	}
	return 0
}

func getFloat32(m map[string]interface{}, key string) *float32 {
	if v, ok := m[key]; ok && v != nil {
		f := float32(getFloat64(m, key))
		return &f
	}
	return nil
}

func getInt16Ptr(m map[string]interface{}, key string) *int16 {
	if v, ok := m[key]; ok && v != nil {
		i := int16(getInt(m, key))
		return &i
	}
	return nil
}

func getInt32Ptr(m map[string]interface{}, key string) *int32 {
	if v, ok := m[key]; ok && v != nil {
		i := int32(getInt(m, key))
		return &i
	}
	return nil
}

func getTimePtr(m map[string]interface{}, key string) *time.Time {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			t, err := time.Parse("2006-01-02", s)
			if err == nil {
				return &t
			}
		}
	}
	return nil
}

// pgtype conversion helpers
func toPGDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func toPGDatePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func toPGText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func toPGTextPtr(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPGUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

func toPGUUIDPtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func toPGInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func toPGInt4Ptr(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func toPGInt2(i int16) pgtype.Int2 {
	return pgtype.Int2{Int16: i, Valid: true}
}

func toPGInt2Ptr(i *int16) pgtype.Int2 {
	if i == nil {
		return pgtype.Int2{Valid: false}
	}
	return pgtype.Int2{Int16: *i, Valid: true}
}

func toPGNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	n.Int = new(big.Int)
	n.Int.SetInt64(int64(f * 100)) // Convert to cents to avoid floating point issues
	n.Exp = -2
	n.Valid = true
	return n
}

func toPGNumericPtr(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	n := pgtype.Numeric{}
	n.Int = new(big.Int)
	n.Int.SetInt64(int64(*f * 100)) // Convert to cents
	n.Exp = -2
	n.Valid = true
	return n
}

func toPGNumericFromFloat32(f *float32) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	f64 := float64(*f)
	return toPGNumericPtr(&f64)
}