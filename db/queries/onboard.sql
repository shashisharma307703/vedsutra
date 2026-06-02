-- name: GetAcademicYearByName :one
SELECT id FROM academic_years 
WHERE tenant_id = $1 AND name = $2 LIMIT 1;

-- name: GetGradeByName :one
SELECT id FROM grades 
WHERE tenant_id = $1 AND name = $2 LIMIT 1;

-- name: GetDepartmentByName :one
SELECT id FROM departments 
WHERE tenant_id = $1 AND name = $2 LIMIT 1;

-- name: GetClassByName :one
SELECT id FROM classes 
WHERE tenant_id = $1 AND name = $2 LIMIT 1;

-- name: GetSubjectByCode :one
SELECT id FROM subjects 
WHERE tenant_id = $1 AND code = $2 LIMIT 1;

-- name: FindUserByName :one
SELECT id FROM users 
WHERE tenant_id = $1 
  AND LOWER(CONCAT(first_name, ' ', last_name)) = LOWER($2) 
LIMIT 1;

-- name: CreatePlaceholderUser :one
INSERT INTO users (id, tenant_id, first_name, last_name, email, password_hash, role_id, is_active)
VALUES ($1, $2, $3, $4, $5, 'PLACEHOLDER_HASH', $6, false)
RETURNING id;

-- name: GetSystemRoleBySlug :one
SELECT id FROM roles 
WHERE (tenant_id = $1 OR tenant_id IS NULL) AND slug = $2 LIMIT 1;

-- name: UpsertAcademicYear :one
INSERT INTO academic_years (id, tenant_id, name, start_date, end_date, is_current, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (tenant_id, name) 
DO UPDATE SET 
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    is_current = EXCLUDED.is_current,
    updated_at = NOW()
RETURNING id;

-- name: UpsertGrade :one
INSERT INTO grades (id, tenant_id, name, sequence, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (tenant_id, name) 
DO UPDATE SET 
    sequence = EXCLUDED.sequence,
    description = EXCLUDED.description,
    updated_at = NOW()
RETURNING id;

-- name: UpsertDepartment :one
INSERT INTO departments (id, tenant_id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (tenant_id, name) 
DO UPDATE SET 
    description = EXCLUDED.description,
    updated_at = NOW()
RETURNING id;

-- name: UpsertClass :one
INSERT INTO classes (id, tenant_id, grade_id, academic_year_id, name, max_students, room_number, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT (tenant_id, academic_year_id, name) 
DO UPDATE SET 
    grade_id = EXCLUDED.grade_id,
    max_students = EXCLUDED.max_students,
    room_number = EXCLUDED.room_number,
    updated_at = NOW()
RETURNING id;

-- name: UpsertSection :one
INSERT INTO sections (id, tenant_id, class_id, name, teacher_id, max_students, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (tenant_id, class_id, name) 
DO UPDATE SET 
    teacher_id = EXCLUDED.teacher_id,
    max_students = EXCLUDED.max_students,
    updated_at = NOW()
RETURNING id;

-- name: UpsertSubject :one
INSERT INTO subjects (id, tenant_id, name, code, department_id, grade_id, is_elective, credits, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
ON CONFLICT (tenant_id, code) 
DO UPDATE SET 
    name = EXCLUDED.name,
    department_id = EXCLUDED.department_id,
    grade_id = EXCLUDED.grade_id,
    is_elective = EXCLUDED.is_elective,
    credits = EXCLUDED.credits,
    updated_at = NOW()
RETURNING id;

-- name: UpsertClassSubject :exec
INSERT INTO class_subjects (tenant_id, class_id, subject_id, teacher_id, periods_per_week, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (tenant_id, class_id, subject_id) 
DO UPDATE SET 
    teacher_id = EXCLUDED.teacher_id,
    periods_per_week = EXCLUDED.periods_per_week,
    updated_at = NOW();

-- name: UpsertFeeStructure :one
INSERT INTO fee_structures (id, tenant_id, academic_year_id, class_id, fee_type, name, amount, frequency, due_day, is_mandatory, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
ON CONFLICT (tenant_id, academic_year_id, class_id, name) 
DO UPDATE SET 
    fee_type = EXCLUDED.fee_type,
    amount = EXCLUDED.amount,
    frequency = EXCLUDED.frequency,
    due_day = EXCLUDED.due_day,
    is_mandatory = EXCLUDED.is_mandatory,
    updated_at = NOW()
RETURNING id;

-- name: CallSeedTenantRoles :exec
SELECT seed_tenant_roles($1::UUID);

-- name: PopulateModuleAccessFromPlan :exec
INSERT INTO tenant_module_access (tenant_id, module_name, is_enabled, created_at, updated_at)
SELECT $1::UUID, plan_modules.module_name, true, NOW(), NOW()
FROM tenants t
JOIN subscription_plans sp ON t.subscription_plan_id = sp.id
JOIN plan_modules ON sp.id = plan_modules.plan_id
WHERE t.id = $1::UUID
ON CONFLICT (tenant_id, module_name) DO NOTHING;