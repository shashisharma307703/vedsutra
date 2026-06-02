-- ============================================================================
-- TENANT & SUBSCRIPTION
-- ============================================================================

-- name: CreateTenant :one
INSERT INTO tenants (
    id, name, subdomain, slug, timezone, country, state, city, address, pincode,
    contact_email, contact_phone, website, status, subscription_plan_id,
    subscription_start_date, subscription_end_date, trial_ends_at, api_key, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
) RETURNING id;

-- name: GetSubscriptionPlanIDByName :one
SELECT id FROM subscription_plans
WHERE name = $1 AND billing_period = $2
LIMIT 1;

-- name: CountSubscriptionPlans :one
SELECT COUNT(*) FROM subscription_plans;

-- ============================================================================
-- ACADEMIC YEARS
-- Unique: (tenant_id, name)
-- ============================================================================

-- name: GetAcademicYearByName :one
SELECT id FROM academic_years
WHERE tenant_id = $1 AND name = $2
LIMIT 1;

-- name: UpsertAcademicYear :one
INSERT INTO academic_years (id, tenant_id, name, start_date, end_date, is_current)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    is_current = EXCLUDED.is_current
RETURNING id;

-- ============================================================================
-- GRADES
-- Unique: (tenant_id, name)
-- ============================================================================

-- name: GetGradeByName :one
SELECT id FROM grades
WHERE tenant_id = $1 AND name = $2
LIMIT 1;

-- name: UpsertGrade :one
INSERT INTO grades (id, tenant_id, name, sequence, description)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    sequence = EXCLUDED.sequence,
    description = EXCLUDED.description
RETURNING id;

-- ============================================================================
-- DEPARTMENTS
-- Unique: (tenant_id, name)
-- ============================================================================

-- name: GetDepartmentByName :one
SELECT id FROM departments
WHERE tenant_id = $1 AND name = $2
LIMIT 1;

-- name: UpsertDepartment :one
INSERT INTO departments (id, tenant_id, name, description, head_teacher_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    description = EXCLUDED.description,
    head_teacher_id = EXCLUDED.head_teacher_id
RETURNING id;

-- ============================================================================
-- SUBJECTS
-- Unique: (tenant_id, code)
-- ============================================================================

-- name: GetSubjectByCode :one
SELECT id FROM subjects
WHERE tenant_id = $1 AND code = $2
LIMIT 1;

-- name: UpsertSubject :one
INSERT INTO subjects (
    id, tenant_id, name, code, department_id, grade_id, is_elective, credits
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (tenant_id, code) DO UPDATE SET
    name = EXCLUDED.name,
    department_id = EXCLUDED.department_id,
    grade_id = EXCLUDED.grade_id,
    is_elective = EXCLUDED.is_elective,
    credits = EXCLUDED.credits
RETURNING id;

-- ============================================================================
-- CLASSES
-- Unique: (tenant_id, name) – we assume class names are unique per tenant.
-- ============================================================================

-- name: GetClassByName :one
SELECT id FROM classes
WHERE tenant_id = $1 AND name = $2
LIMIT 1;

-- name: GetClassByGradeAndYear :one
SELECT c.id
FROM classes c
JOIN grades g ON g.id = c.grade_id
JOIN academic_years ay ON ay.id = c.academic_year_id
WHERE c.tenant_id = $1 AND g.name = $2 AND ay.name = $3
LIMIT 1;

-- name: UpsertClass :one
INSERT INTO classes (
    id, tenant_id, grade_id, academic_year_id, name, max_students, room_number
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    grade_id = EXCLUDED.grade_id,
    academic_year_id = EXCLUDED.academic_year_id,
    max_students = EXCLUDED.max_students,
    room_number = EXCLUDED.room_number
RETURNING id;

-- ============================================================================
-- SECTIONS
-- Unique: (class_id, name). For upsert we need class_id.
-- ============================================================================

-- name: GetSectionByClassAndName :one
SELECT s.id
FROM sections s
JOIN classes c ON c.id = s.class_id
WHERE c.tenant_id = $1 AND c.name = $2 AND s.name = $3
LIMIT 1;

-- name: UpsertSection :one
INSERT INTO sections (id, class_id, name, teacher_id, max_students)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (class_id, name) DO UPDATE SET
    teacher_id = EXCLUDED.teacher_id,
    max_students = EXCLUDED.max_students
RETURNING id;

-- ============================================================================
-- CLASS SUBJECTS
-- Unique: (class_id, subject_id)
-- ============================================================================

-- name: GetClassSubjectByClassAndSubject :one
SELECT cs.id
FROM class_subjects cs
JOIN classes c ON c.id = cs.class_id
JOIN subjects s ON s.id = cs.subject_id
WHERE c.tenant_id = $1 AND c.name = $2 AND s.code = $3
LIMIT 1;

-- name: UpsertClassSubject :one
INSERT INTO class_subjects (id, class_id, subject_id, teacher_id, periods_per_week)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (class_id, subject_id) DO UPDATE SET
    teacher_id = EXCLUDED.teacher_id,
    periods_per_week = EXCLUDED.periods_per_week
RETURNING id;

-- ============================================================================
-- FEE STRUCTURES
-- Unique: (tenant_id, academic_year_id, class_id, fee_type, name)
-- ============================================================================

-- name: GetFeeStructureByNaturalKey :one
SELECT fs.id
FROM fee_structures fs
JOIN academic_years ay ON ay.id = fs.academic_year_id
JOIN classes c ON c.id = fs.class_id
WHERE fs.tenant_id = $1
  AND ay.name = $2
  AND c.name = $3
  AND fs.fee_type = $4
  AND fs.name = $5
LIMIT 1;

-- name: UpsertFeeStructure :one
INSERT INTO fee_structures (
    id, tenant_id, academic_year_id, class_id, fee_type, name,
    amount, frequency, due_day, due_date, late_fee_per_day, is_mandatory, description
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (tenant_id, academic_year_id, class_id, fee_type, name) DO UPDATE SET
    amount = EXCLUDED.amount,
    frequency = EXCLUDED.frequency,
    due_day = EXCLUDED.due_day,
    due_date = EXCLUDED.due_date,
    late_fee_per_day = EXCLUDED.late_fee_per_day,
    is_mandatory = EXCLUDED.is_mandatory,
    description = EXCLUDED.description
RETURNING id;

-- ============================================================================
-- TENANT SETTINGS
-- Unique: tenant_id
-- ============================================================================

-- name: GetTenantSetting :one
SELECT id FROM tenant_settings WHERE tenant_id = $1 LIMIT 1;

-- name: UpsertTenantSetting :one
INSERT INTO tenant_settings (
    id, tenant_id, session_timeout_min, max_users, max_students,
    academic_year_format, date_format, currency_code, grading_system,
    attendance_mode, sms_gateway, email_provider, features_enabled, theme
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (tenant_id) DO UPDATE SET
    session_timeout_min = EXCLUDED.session_timeout_min,
    max_users = EXCLUDED.max_users,
    max_students = EXCLUDED.max_students,
    academic_year_format = EXCLUDED.academic_year_format,
    date_format = EXCLUDED.date_format,
    currency_code = EXCLUDED.currency_code,
    grading_system = EXCLUDED.grading_system,
    attendance_mode = EXCLUDED.attendance_mode,
    sms_gateway = EXCLUDED.sms_gateway,
    email_provider = EXCLUDED.email_provider,
    features_enabled = EXCLUDED.features_enabled,
    theme = EXCLUDED.theme
RETURNING id;

-- ============================================================================
-- USERS (teacher placeholders)
-- Unique: (tenant_id, email)
-- ============================================================================

-- name: GetUserByEmail :one
SELECT id FROM users
WHERE tenant_id = $1 AND email = $2
LIMIT 1;

-- name: UpsertPlaceholderUser :one
INSERT INTO users (
    id, tenant_id, username, email, password_hash, first_name, last_name,
    status, created_offline_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, 'active', NOW()
)
ON CONFLICT (tenant_id, email) DO UPDATE SET
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    status = EXCLUDED.status
RETURNING id;

-- ============================================================================
-- ROLES & PERMISSIONS (automatic seeding)
-- ============================================================================

-- name: CountTenantRoles :one
SELECT COUNT(*) FROM roles WHERE tenant_id = $1;

-- name: CallSeedTenantRoles :exec
SELECT seed_tenant_roles($1);