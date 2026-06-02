package onboarder

import "time"

// TenantMeta for tenant creation (JSON or CSV)
type TenantMeta struct {
	Name             string `json:"name" csv:"name"`
	Subdomain        string `json:"subdomain" csv:"subdomain"`
	Slug             string `json:"slug" csv:"slug"`
	Timezone         string `json:"timezone" csv:"timezone"`
	Country          string `json:"country" csv:"country"`
	State            string `json:"state" csv:"state"`
	City             string `json:"city" csv:"city"`
	Address          string `json:"address" csv:"address"`
	Pincode          string `json:"pincode" csv:"pincode"`
	ContactEmail     string `json:"contact_email" csv:"contact_email"`
	ContactPhone     string `json:"contact_phone" csv:"contact_phone"`
	Website          string `json:"website" csv:"website"`
	Status           string `json:"status" csv:"status"`
	SubscriptionPlan string `json:"subscription_plan" csv:"subscription_plan"`
	BillingPeriod    string `json:"billing_period" csv:"billing_period"`
	TrialDays        int    `json:"trial_days" csv:"trial_days"`
	APIPrefix        string `json:"api_key_prefix" csv:"api_key_prefix"`
}

// CSV/JSON row structs for each table
type AcademicYearRow struct {
	Name      string    `json:"name" csv:"name"`
	StartDate time.Time `json:"start_date" csv:"start_date"`
	EndDate   time.Time `json:"end_date" csv:"end_date"`
	IsCurrent bool      `json:"is_current" csv:"is_current"`
}

type GradeRow struct {
	Name        string `json:"name" csv:"name"`
	Sequence    int16  `json:"sequence" csv:"sequence"`
	Description string `json:"description" csv:"description"`
}

type DepartmentRow struct {
	Name             string  `json:"name" csv:"name"`
	Description      string  `json:"description" csv:"description"`
	HeadTeacherEmail *string `json:"head_teacher_email" csv:"head_teacher_email"`
}

type SubjectRow struct {
	Name       string   `json:"name" csv:"name"`
	Code       string   `json:"code" csv:"code"`
	Department string   `json:"department_name" csv:"department_name"`
	GradeName  string   `json:"grade_name" csv:"grade_name"`
	IsElective bool     `json:"is_elective" csv:"is_elective"`
	Credits    *float32 `json:"credits" csv:"credits"`
}

type ClassRow struct {
	GradeName        string `json:"grade_name" csv:"grade_name"`
	AcademicYearName string `json:"academic_year_name" csv:"academic_year_name"`
	ClassName        string `json:"class_name" csv:"class_name"`
	MaxStudents      int32  `json:"max_students" csv:"max_students"`
	RoomNumber       string `json:"room_number" csv:"room_number"`
}

type SectionRow struct {
	ClassName    string  `json:"class_name" csv:"class_name"`
	SectionName  string  `json:"section_name" csv:"section_name"`
	TeacherEmail *string `json:"teacher_email" csv:"teacher_email"`
	MaxStudents  int32   `json:"max_students" csv:"max_students"`
}

type ClassSubjectRow struct {
	ClassName      string  `json:"class_name" csv:"class_name"`
	SubjectCode    string  `json:"subject_code" csv:"subject_code"`
	TeacherEmail   *string `json:"teacher_email" csv:"teacher_email"`
	PeriodsPerWeek int16   `json:"periods_per_week" csv:"periods_per_week"`
}

type FeeStructureRow struct {
	AcademicYearName string     `json:"academic_year_name" csv:"academic_year_name"`
	ClassName        string     `json:"class_name" csv:"class_name"`
	FeeType          string     `json:"fee_type" csv:"fee_type"`
	Name             string     `json:"name" csv:"name"`
	Amount           float64    `json:"amount" csv:"amount"`
	Frequency        string     `json:"frequency" csv:"frequency"`
	DueDay           *int16     `json:"due_day" csv:"due_day"`
	DueDate          *time.Time `json:"due_date" csv:"due_date"`
	LateFeePerDay    float64    `json:"late_fee_per_day" csv:"late_fee_per_day"`
	IsMandatory      bool       `json:"is_mandatory" csv:"is_mandatory"`
	Description      string     `json:"description" csv:"description"`
}

type TenantSettingRow struct {
	SessionTimeoutMin   int32   `json:"session_timeout_min" csv:"session_timeout_min"`
	MaxUsers            *int32  `json:"max_users" csv:"max_users"`
	MaxStudents         *int32  `json:"max_students" csv:"max_students"`
	AcademicYearFormat  string  `json:"academic_year_format" csv:"academic_year_format"`
	DateFormat          string  `json:"date_format" csv:"date_format"`
	CurrencyCode        string  `json:"currency_code" csv:"currency_code"`
	GradingSystem       string  `json:"grading_system" csv:"grading_system"`
	AttendanceMode      string  `json:"attendance_mode" csv:"attendance_mode"`
	SMSGateway          string  `json:"sms_gateway" csv:"sms_gateway"`
	EmailProvider       string  `json:"email_provider" csv:"email_provider"`
	FeaturesEnabledJSON string  `json:"features_enabled" csv:"features_enabled"`
	ThemeJSON           string  `json:"theme" csv:"theme"`
}