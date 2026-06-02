-- ============================================================================
-- ENHANCED ERP SCHEMA WITH UUID PRIMARY KEYS (application-generated)
-- Use UUID v6/v7. No auto-generation in DB.
-- ============================================================================

DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;

-- ----------------------------------------------------------------------------
-- ORGANIZATIONS (tenant) – UUID for multi-tenancy support
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS organizations (
    org_id UUID PRIMARY KEY,
    org_name VARCHAR(200) NOT NULL,
    org_code VARCHAR(50) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    email VARCHAR(100),
    website VARCHAR(200),
    logo_url TEXT,
    address TEXT,
    city VARCHAR(50),
    state VARCHAR(50),
    country VARCHAR(50),
    postal_code VARCHAR(10),
    established_date DATE,
    affiliation_number VARCHAR(100),
    license_number VARCHAR(100),
    license_expiry DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ----------------------------------------------------------------------------
-- GLOBAL LOOKUP TABLES (UUID primary keys)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS class_levels (
    class_level_id UUID PRIMARY KEY,   -- application-generated UUID v6
    class_name VARCHAR(50) NOT NULL,
    class_code VARCHAR(20) UNIQUE NOT NULL,
    display_order INT NOT NULL,
    description TEXT,
    is_lower_primary BOOLEAN DEFAULT FALSE,
    is_upper_primary BOOLEAN DEFAULT FALSE,
    is_secondary BOOLEAN DEFAULT FALSE,
    is_higher_secondary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS streams (
    stream_id UUID PRIMARY KEY,
    stream_name VARCHAR(50) UNIQUE NOT NULL,
    stream_code VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subjects (
    subject_id UUID PRIMARY KEY,
    subject_name VARCHAR(100) NOT NULL,
    subject_code VARCHAR(20) UNIQUE NOT NULL,
    description TEXT,
    is_mandatory BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ----------------------------------------------------------------------------
-- TENANT-SPECIFIC TABLES (foreign keys use UUID)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS academic_years (
    academic_year_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    year VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, year),
    CONSTRAINT chk_year_dates CHECK (start_date < end_date)
);

CREATE TABLE IF NOT EXISTS sections (
    section_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    stream_id UUID NULL,
    section_name VARCHAR(20) NOT NULL,
    section_code VARCHAR(20) NOT NULL,
    class_teacher_id UUID,  -- references users.user_id
    capacity INT DEFAULT 50,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id) ON DELETE CASCADE,
    FOREIGN KEY (stream_id) REFERENCES streams(stream_id) ON DELETE SET NULL,
    UNIQUE(org_id, class_level_id, stream_id, section_code)
);

CREATE TABLE IF NOT EXISTS class_subjects (
    class_subject_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    stream_id UUID NULL,
    subject_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id) ON DELETE CASCADE,
    FOREIGN KEY (stream_id) REFERENCES streams(stream_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    UNIQUE(org_id, class_level_id, stream_id, subject_id, academic_year_id)
);

CREATE TABLE IF NOT EXISTS subject_syllabus (
    syllabus_id UUID PRIMARY KEY,
    subject_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    curriculum TEXT,
    learning_outcomes TEXT,
    teaching_method TEXT,
    assessment_method TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- USERS & ROLES (user_id as UUID)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    date_of_birth DATE,
    gender VARCHAR(10) CHECK (gender IN ('Male', 'Female', 'Other')),
    address TEXT,
    city VARCHAR(50),
    state VARCHAR(50),
    country VARCHAR(50),
    postal_code VARCHAR(10),
    profile_picture_url TEXT,
    status VARCHAR(20) DEFAULT 'Active' CHECK (status IN ('Active', 'Inactive', 'Suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    CONSTRAINT chk_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
    UNIQUE(org_id, username),
    UNIQUE(org_id, email),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS roles (
    role_id SERIAL PRIMARY KEY,   -- still integer, roles are global and few
    role_name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    permission_id SERIAL PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    module VARCHAR(50),
    action VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    org_id UUID,
    PRIMARY KEY (role_id, permission_id, org_id),
    FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(permission_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id INT NOT NULL,
    org_id UUID NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id, org_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_logs (
    audit_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    org_id UUID,
    action VARCHAR(100) NOT NULL,
    module VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id VARCHAR(255),   -- can store UUID or other identifier as string
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS system_settings (
    setting_id UUID PRIMARY KEY,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(20),
    description TEXT,
    is_editable BOOLEAN DEFAULT TRUE,
    org_id UUID,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(setting_key, org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- STUDENT MANAGEMENT
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS students (
    student_id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    org_id UUID NOT NULL,
    admission_number VARCHAR(50) NOT NULL,
    admission_date DATE NOT NULL,
    admission_class_level_id UUID,
    admission_stream_id UUID,
    roll_number VARCHAR(20),
    mother_name VARCHAR(100),
    father_name VARCHAR(100),
    guardian_name VARCHAR(100),
    guardian_phone VARCHAR(20),
    guardian_email VARCHAR(100),
    emergency_contact_1 VARCHAR(20),
    emergency_contact_2 VARCHAR(20),
    blood_group VARCHAR(5),
    aadhar_number VARCHAR(12),
    birth_place VARCHAR(100),
    nationality VARCHAR(50),
    religion VARCHAR(50),
    caste VARCHAR(50),
    status VARCHAR(20) DEFAULT 'Active' CHECK (status IN ('Active', 'Inactive', 'Transferred', 'Passed Out', 'Suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (admission_class_level_id) REFERENCES class_levels(class_level_id) ON DELETE SET NULL,
    FOREIGN KEY (admission_stream_id) REFERENCES streams(stream_id) ON DELETE SET NULL,
    UNIQUE(org_id, admission_number)
);

CREATE TABLE IF NOT EXISTS student_enrollments (
    enrollment_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    section_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    enrollment_date DATE DEFAULT CURRENT_DATE,
    status VARCHAR(20) DEFAULT 'Active' CHECK (status IN ('Active', 'Inactive', 'Transferred', 'Dropped')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    UNIQUE(student_id, section_id, academic_year_id)
);

CREATE TABLE IF NOT EXISTS student_health_records (
    health_record_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    blood_group VARCHAR(5),
    height DECIMAL(5,2),
    weight DECIMAL(5,2),
    vision_left VARCHAR(10),
    vision_right VARCHAR(10),
    hearing_status VARCHAR(20),
    allergies TEXT,
    medical_conditions TEXT,
    vaccinations JSONB,
    last_checkup_date DATE,
    notes TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS student_documents (
    document_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    document_type VARCHAR(50),
    document_name VARCHAR(100),
    file_path TEXT,
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date DATE,
    verified BOOLEAN DEFAULT FALSE,
    verified_by UUID,
    verification_date TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (verified_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS student_promotions (
    promotion_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    from_class_level_id UUID NOT NULL,
    to_class_level_id UUID NOT NULL,
    from_section_id UUID,
    to_section_id UUID,
    from_stream_id UUID,
    to_stream_id UUID,
    academic_year_id UUID NOT NULL,
    promotion_date DATE,
    promotion_status VARCHAR(20) NOT NULL CHECK (promotion_status IN ('Promoted', 'Retained', 'Transferred')),
    remarks TEXT,
    approved_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (from_class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (to_class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (from_stream_id) REFERENCES streams(stream_id),
    FOREIGN KEY (to_stream_id) REFERENCES streams(stream_id),
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id),
    FOREIGN KEY (approved_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- ----------------------------------------------------------------------------
-- ADMISSION MODULE
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admission_forms (
    admission_form_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    applicant_name VARCHAR(100) NOT NULL,
    applicant_email VARCHAR(100),
    applicant_phone VARCHAR(20),
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('Male', 'Female', 'Other')),
    address TEXT,
    city VARCHAR(50),
    state VARCHAR(50),
    postal_code VARCHAR(10),
    applied_class_level_id UUID NOT NULL,
    applied_stream_id UUID,
    previous_school VARCHAR(200),
    previous_marks DECIMAL(5,2),
    parent_name VARCHAR(100),
    parent_email VARCHAR(100),
    parent_phone VARCHAR(20),
    form_status VARCHAR(20) DEFAULT 'Draft' CHECK (form_status IN ('Draft', 'Submitted', 'Under Review', 'Approved', 'Rejected')),
    application_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submitted_date TIMESTAMP,
    approval_date TIMESTAMP,
    approved_by UUID,
    remarks TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (applied_class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (applied_stream_id) REFERENCES streams(stream_id),
    FOREIGN KEY (approved_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS admission_documents (
    admission_doc_id UUID PRIMARY KEY,
    admission_form_id UUID NOT NULL,
    document_type VARCHAR(50),
    file_path TEXT,
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admission_form_id) REFERENCES admission_forms(admission_form_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- STAFF MANAGEMENT
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS staff (
    staff_id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,
    org_id UUID NOT NULL,
    employee_id VARCHAR(50) NOT NULL,
    designation VARCHAR(100),
    department VARCHAR(100),
    qualification VARCHAR(255),
    specialization VARCHAR(255),
    date_of_joining DATE,
    date_of_birth DATE,
    employee_status VARCHAR(20) DEFAULT 'Active' CHECK (employee_status IN ('Active', 'Inactive', 'On Leave', 'Retired')),
    bank_account_number VARCHAR(50),
    bank_name VARCHAR(100),
    ifsc_code VARCHAR(20),
    pan_number VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, employee_id)
);

CREATE TABLE IF NOT EXISTS staff_subject_assignment (
    assignment_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    section_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    assigned_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id) ON DELETE CASCADE,
    FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    UNIQUE(staff_id, subject_id, section_id, academic_year_id)
);

-- ----------------------------------------------------------------------------
-- ATTENDANCE MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS student_attendance (
    attendance_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    section_id UUID NOT NULL,
    attendance_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('Present', 'Absent', 'Late', 'Leave', 'Excused')),
    marked_by UUID,
    marked_time TIMESTAMP,
    remarks TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE,
    FOREIGN KEY (marked_by) REFERENCES users(user_id) ON DELETE SET NULL,
    UNIQUE(student_id, attendance_date)
);

CREATE TABLE IF NOT EXISTS staff_attendance (
    staff_attendance_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    attendance_date DATE NOT NULL,
    check_in_time TIME,
    check_out_time TIME,
    status VARCHAR(20) NOT NULL CHECK (status IN ('Present', 'Absent', 'Leave', 'Late')),
    biometric_id VARCHAR(50),
    marked_by UUID,
    remarks TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (marked_by) REFERENCES users(user_id) ON DELETE SET NULL,
    UNIQUE(staff_id, attendance_date)
);

CREATE TABLE IF NOT EXISTS biometric_devices (
    device_id SERIAL PRIMARY KEY,  -- integer, fine for device enumeration
    org_id UUID NOT NULL,
    device_name VARCHAR(100),
    device_serial_number VARCHAR(100) UNIQUE NOT NULL,
    device_type VARCHAR(50),
    location VARCHAR(100),
    ip_address VARCHAR(45),
    is_active BOOLEAN DEFAULT TRUE,
    last_sync_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS biometric_mappings (
    mapping_id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    biometric_id VARCHAR(50) NOT NULL,
    biometric_type VARCHAR(20) CHECK (biometric_type IN ('Fingerprint', 'Face', 'RFID', 'PIN')),
    device_id INT NOT NULL,
    mapped_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (device_id) REFERENCES biometric_devices(device_id) ON DELETE CASCADE,
    UNIQUE(biometric_id, device_id)
);

CREATE TABLE IF NOT EXISTS leave_types (
    leave_type_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    leave_type_name VARCHAR(50) NOT NULL,
    description TEXT,
    annual_limit INT,
    carry_forward_allowed BOOLEAN DEFAULT FALSE,
    requires_approval BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS leave_applications (
    leave_application_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    leave_type_id INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    number_of_days INT,
    reason TEXT,
    application_status VARCHAR(20) DEFAULT 'Draft' CHECK (application_status IN ('Draft', 'Submitted', 'Approved', 'Rejected', 'Cancelled')),
    approved_by UUID,
    approval_date TIMESTAMP,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (leave_type_id) REFERENCES leave_types(leave_type_id),
    FOREIGN KEY (approved_by) REFERENCES users(user_id) ON DELETE SET NULL,
    CONSTRAINT chk_leave_dates CHECK (start_date <= end_date)
);

-- ----------------------------------------------------------------------------
-- EXAMINATION MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS exam_types (
    exam_type_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    exam_type_name VARCHAR(100) NOT NULL,
    exam_code VARCHAR(20) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, exam_code)
);

CREATE TABLE IF NOT EXISTS exams (
    exam_id UUID PRIMARY KEY,
    exam_type_id INT NOT NULL,
    academic_year_id UUID NOT NULL,
    exam_name VARCHAR(100) NOT NULL,
    description TEXT,
    exam_start_date DATE,
    exam_end_date DATE,
    exam_status VARCHAR(20) DEFAULT 'Draft' CHECK (exam_status IN ('Draft', 'Scheduled', 'Ongoing', 'Completed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exam_type_id) REFERENCES exam_types(exam_type_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS exam_schedule (
    schedule_id UUID PRIMARY KEY,
    exam_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    stream_id UUID NULL,
    exam_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    total_marks INT DEFAULT 100,
    passing_marks INT,
    duration_minutes INT,
    location VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exam_id) REFERENCES exams(exam_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id),
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (stream_id) REFERENCES streams(stream_id),
    CONSTRAINT chk_exam_times CHECK (start_time < end_time)
);

CREATE TABLE IF NOT EXISTS marks_entry (
    marks_entry_id UUID PRIMARY KEY,
    exam_schedule_id UUID NOT NULL,
    student_id UUID NOT NULL,
    marks_obtained DECIMAL(5,2),
    entry_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    entered_by UUID,
    is_absent BOOLEAN DEFAULT FALSE,
    remarks TEXT,
    FOREIGN KEY (exam_schedule_id) REFERENCES exam_schedule(schedule_id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (entered_by) REFERENCES users(user_id) ON DELETE SET NULL,
    UNIQUE(exam_schedule_id, student_id)
);

CREATE TABLE IF NOT EXISTS grade_scales (
    grade_scale_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    scale_name VARCHAR(50),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS grades (
    grade_id SERIAL PRIMARY KEY,
    grade_scale_id INT NOT NULL,
    grade_letter VARCHAR(5),
    min_percentage DECIMAL(5,2),
    max_percentage DECIMAL(5,2),
    grade_point DECIMAL(3,2),
    description VARCHAR(50),
    FOREIGN KEY (grade_scale_id) REFERENCES grade_scales(grade_scale_id) ON DELETE CASCADE,
    UNIQUE(grade_scale_id, grade_letter),
    CONSTRAINT chk_grade_range CHECK (min_percentage < max_percentage)
);

CREATE TABLE IF NOT EXISTS exam_results (
    result_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    exam_id UUID NOT NULL,
    section_id UUID NOT NULL,
    total_marks DECIMAL(5,2),
    obtained_marks DECIMAL(5,2),
    percentage DECIMAL(5,2),
    grade_id INT,
    status VARCHAR(10) NOT NULL CHECK (status IN ('Pass', 'Fail', 'Absent')),
    result_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published BOOLEAN DEFAULT FALSE,
    published_date TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (exam_id) REFERENCES exams(exam_id) ON DELETE CASCADE,
    FOREIGN KEY (section_id) REFERENCES sections(section_id),
    FOREIGN KEY (grade_id) REFERENCES grades(grade_id) ON DELETE SET NULL,
    UNIQUE(student_id, exam_id)
);

CREATE TABLE IF NOT EXISTS report_cards (
    report_card_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    exam_id UUID,
    section_id UUID NOT NULL,
    generated_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    generated_by UUID,
    report_card_data JSONB,
    is_printed BOOLEAN DEFAULT FALSE,
    printed_date TIMESTAMP,
    signed_by UUID,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id),
    FOREIGN KEY (exam_id) REFERENCES exams(exam_id) ON DELETE SET NULL,
    FOREIGN KEY (section_id) REFERENCES sections(section_id),
    FOREIGN KEY (generated_by) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (signed_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS exam_analytics (
    analytics_id UUID PRIMARY KEY,
    exam_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    total_students INT,
    present_students INT,
    absent_students INT,
    passed_students INT,
    failed_students INT,
    average_marks DECIMAL(5,2),
    highest_marks DECIMAL(5,2),
    lowest_marks DECIMAL(5,2),
    standard_deviation DECIMAL(5,2),
    analytics_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (exam_id) REFERENCES exams(exam_id) ON DELETE CASCADE,
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id)
);

-- ----------------------------------------------------------------------------
-- FEE & FINANCE MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS fee_structures (
    fee_structure_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    stream_id UUID NULL,
    fee_type VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    description TEXT,
    is_mandatory BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (stream_id) REFERENCES streams(stream_id)
);

CREATE TABLE IF NOT EXISTS fee_concessions (
    concession_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    concession_type VARCHAR(50),
    concession_percentage DECIMAL(5,2),
    concession_amount DECIMAL(10,2),
    reason TEXT,
    approved_by UUID,
    approval_date TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS fee_transactions (
    transaction_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    transaction_date DATE DEFAULT CURRENT_DATE,
    amount_due DECIMAL(10,2),
    amount_paid DECIMAL(10,2),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('Cash', 'Check', 'Online', 'Bank Transfer')),
    transaction_reference VARCHAR(100),
    payment_status VARCHAR(20) DEFAULT 'Pending' CHECK (payment_status IN ('Pending', 'Completed', 'Failed', 'Refunded')),
    notes TEXT,
    collected_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (collected_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS payment_gateways (
    gateway_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    gateway_name VARCHAR(50),
    api_key VARCHAR(255),
    api_secret VARCHAR(255),
    merchant_id VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS online_payments (
    payment_id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    gateway_id INT NOT NULL,
    gateway_transaction_id VARCHAR(100),
    payment_gateway_response JSONB,
    payment_status VARCHAR(20) DEFAULT 'Initiated' CHECK (payment_status IN ('Initiated', 'Successful', 'Failed', 'Cancelled')),
    amount DECIMAL(10,2),
    payment_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (transaction_id) REFERENCES fee_transactions(transaction_id) ON DELETE CASCADE,
    FOREIGN KEY (gateway_id) REFERENCES payment_gateways(gateway_id)
);

CREATE TABLE IF NOT EXISTS fee_alerts (
    alert_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    due_amount DECIMAL(10,2),
    due_date DATE,
    alert_status VARCHAR(20) DEFAULT 'Active' CHECK (alert_status IN ('Active', 'Cleared', 'Overdue')),
    alert_count INT DEFAULT 0,
    last_alert_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS invoices (
    invoice_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    invoice_number VARCHAR(50) NOT NULL,
    invoice_date DATE DEFAULT CURRENT_DATE,
    total_amount DECIMAL(10,2),
    tax_amount DECIMAL(10,2),
    discount_amount DECIMAL(10,2),
    net_amount DECIMAL(10,2),
    due_date DATE,
    payment_status VARCHAR(20) DEFAULT 'Pending' CHECK (payment_status IN ('Pending', 'Partial', 'Paid')),
    invoice_data JSONB,
    generated_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by) REFERENCES users(user_id) ON DELETE SET NULL,
    UNIQUE(org_id, invoice_number)
);

CREATE TABLE IF NOT EXISTS financial_reports (
    report_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    report_type VARCHAR(50),
    report_name VARCHAR(200),
    start_date DATE,
    end_date DATE,
    total_revenue DECIMAL(12,2),
    total_expense DECIMAL(12,2),
    net_profit DECIMAL(12,2),
    report_data JSONB,
    generated_by UUID,
    generated_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (generated_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS account_heads (
    account_head_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    head_name VARCHAR(100),
    head_code VARCHAR(20),
    head_type VARCHAR(20) CHECK (head_type IN ('Income', 'Expense', 'Asset', 'Liability')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    ledger_entry_id UUID PRIMARY KEY,
    account_head_id INT NOT NULL,
    transaction_date DATE,
    debit_amount DECIMAL(12,2),
    credit_amount DECIMAL(12,2),
    reference_id VARCHAR(255),
    reference_type VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_head_id) REFERENCES account_heads(account_head_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- COMMUNICATION MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS announcements (
    announcement_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    announcement_type VARCHAR(50),
    target_audience JSONB,
    created_by UUID NOT NULL,
    publish_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date TIMESTAMP,
    is_published BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS announcements_read (
    read_id UUID PRIMARY KEY,
    announcement_id UUID NOT NULL,
    user_id UUID NOT NULL,
    read_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (announcement_id) REFERENCES announcements(announcement_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE(announcement_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    message_id UUID PRIMARY KEY,
    sender_id UUID NOT NULL,
    recipient_id UUID NOT NULL,
    org_id UUID NOT NULL,
    subject VARCHAR(200),
    message_content TEXT NOT NULL,
    attachment_url TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    read_date TIMESTAMP,
    message_type VARCHAR(20) DEFAULT 'Personal' CHECK (message_type IN ('Personal', 'Academic', 'Administrative')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sms_gateway (
    sms_gateway_id SERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    gateway_name VARCHAR(50),
    api_key VARCHAR(255),
    api_url TEXT,
    sender_id VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sms_notifications (
    sms_id UUID PRIMARY KEY,
    org_id UUID,
    recipient_phone VARCHAR(20),
    message_content TEXT,
    notification_type VARCHAR(50),
    sent_by UUID,
    sent_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_status VARCHAR(20) DEFAULT 'Pending' CHECK (delivery_status IN ('Pending', 'Sent', 'Failed', 'Delivered')),
    delivery_response TEXT,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE SET NULL,
    FOREIGN KEY (sent_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS email_notifications (
    email_id UUID PRIMARY KEY,
    org_id UUID,
    recipient_email VARCHAR(100),
    email_subject VARCHAR(200),
    email_content TEXT,
    email_template VARCHAR(100),
    notification_type VARCHAR(50),
    sent_by UUID,
    sent_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivery_status VARCHAR(20) DEFAULT 'Pending' CHECK (delivery_status IN ('Pending', 'Sent', 'Failed', 'Bounced')),
    delivery_response TEXT,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE SET NULL,
    FOREIGN KEY (sent_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- ----------------------------------------------------------------------------
-- TRANSPORT MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transport_routes (
    route_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    route_name VARCHAR(100) NOT NULL,
    route_code VARCHAR(20) NOT NULL,
    start_location VARCHAR(200),
    end_location VARCHAR(200),
    total_stops INT,
    estimated_duration INT,
    route_map JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, route_code)
);

CREATE TABLE IF NOT EXISTS transport_stops (
    stop_id UUID PRIMARY KEY,
    route_id UUID NOT NULL,
    stop_name VARCHAR(100),
    stop_sequence INT,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    stop_time TIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES transport_routes(route_id) ON DELETE CASCADE,
    UNIQUE(route_id, stop_sequence)
);

CREATE TABLE IF NOT EXISTS vehicles (
    vehicle_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    vehicle_number VARCHAR(20) NOT NULL,
    vehicle_type VARCHAR(50),
    vehicle_model VARCHAR(100),
    registration_number VARCHAR(50),
    vehicle_capacity INT,
    purchase_date DATE,
    last_maintenance_date DATE,
    next_maintenance_date DATE,
    insurance_expiry DATE,
    pollution_certificate_expiry DATE,
    vehicle_status VARCHAR(20) DEFAULT 'Active' CHECK (vehicle_status IN ('Active', 'Inactive', 'Under Maintenance')),
    gps_device_id VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, vehicle_number)
);

CREATE TABLE IF NOT EXISTS drivers (
    driver_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    staff_id UUID UNIQUE,
    driver_name VARCHAR(100) NOT NULL,
    license_number VARCHAR(50) NOT NULL,
    license_expiry DATE,
    phone_number VARCHAR(20),
    address TEXT,
    date_of_joining DATE,
    background_check_status VARCHAR(50),
    driver_status VARCHAR(20) DEFAULT 'Active' CHECK (driver_status IN ('Active', 'Inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE SET NULL,
    UNIQUE(org_id, license_number)
);

CREATE TABLE IF NOT EXISTS vehicle_drivers (
    assignment_id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    assigned_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id) ON DELETE CASCADE,
    FOREIGN KEY (driver_id) REFERENCES drivers(driver_id) ON DELETE CASCADE,
    UNIQUE(vehicle_id, driver_id)
);

CREATE TABLE IF NOT EXISTS student_route_allocation (
    allocation_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    route_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    stop_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    allocation_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (route_id) REFERENCES transport_routes(route_id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id),
    FOREIGN KEY (stop_id) REFERENCES transport_stops(stop_id),
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id),
    UNIQUE(student_id, academic_year_id)
);

CREATE TABLE IF NOT EXISTS transport_fees (
    transport_fee_id UUID PRIMARY KEY,
    route_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    fee_amount DECIMAL(10,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (route_id) REFERENCES transport_routes(route_id),
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id)
);

CREATE TABLE IF NOT EXISTS vehicle_gps_tracking (
    tracking_id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    speed DECIMAL(5,2),
    heading INT,
    tracking_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- LIBRARY MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS library_books (
    book_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    isbn VARCHAR(20),
    title VARCHAR(200) NOT NULL,
    author VARCHAR(100),
    publisher VARCHAR(100),
    publication_year INT,
    edition VARCHAR(50),
    category VARCHAR(50),
    subject VARCHAR(100),
    total_copies INT,
    available_copies INT,
    book_price DECIMAL(10,2),
    added_date DATE,
    book_status VARCHAR(20) DEFAULT 'Active' CHECK (book_status IN ('Active', 'Damaged', 'Lost', 'Discarded')),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, isbn)
);

CREATE TABLE IF NOT EXISTS library_book_copies (
    copy_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    book_id UUID NOT NULL,
    barcode_number VARCHAR(50) NOT NULL,
    copy_number INT,
    acquisition_date DATE,
    copy_status VARCHAR(20) DEFAULT 'Available' CHECK (copy_status IN ('Available', 'Issued', 'Damaged', 'Lost')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES library_books(book_id) ON DELETE CASCADE,
    UNIQUE(org_id, barcode_number)
);

CREATE TABLE IF NOT EXISTS book_issue_return (
    transaction_id UUID PRIMARY KEY,
    copy_id UUID NOT NULL,
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    issue_date DATE NOT NULL,
    expected_return_date DATE NOT NULL,
    actual_return_date DATE,
    is_returned BOOLEAN DEFAULT FALSE,
    issued_by UUID,
    returned_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (copy_id) REFERENCES library_book_copies(copy_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (issued_by) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (returned_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS library_fines (
    fine_id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    fine_amount DECIMAL(10,2),
    fine_date DATE,
    is_paid BOOLEAN DEFAULT FALSE,
    paid_date DATE,
    FOREIGN KEY (transaction_id) REFERENCES book_issue_return(transaction_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- ACADEMIC PLANNING
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS timetables (
    timetable_id UUID PRIMARY KEY,
    section_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    timetable_name VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (section_id) REFERENCES sections(section_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS timetable_entries (
    entry_id UUID PRIMARY KEY,
    timetable_id UUID NOT NULL,
    day_of_week INT,
    period_number INT,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_id UUID NOT NULL,
    staff_id UUID,
    room_number VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (timetable_id) REFERENCES timetables(timetable_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id),
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS lesson_plans (
    lesson_plan_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    class_level_id UUID NOT NULL,
    stream_id UUID NULL,
    academic_year_id UUID NOT NULL,
    lesson_title VARCHAR(200) NOT NULL,
    lesson_date DATE,
    lesson_content TEXT,
    learning_outcomes TEXT,
    teaching_methodology VARCHAR(255),
    resources_required TEXT,
    assessment_method TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id),
    FOREIGN KEY (class_level_id) REFERENCES class_levels(class_level_id),
    FOREIGN KEY (stream_id) REFERENCES streams(stream_id),
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id)
);

CREATE TABLE IF NOT EXISTS assignments (
    assignment_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    subject_id UUID NOT NULL,
    section_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    assignment_date DATE,
    submission_date DATE,
    total_marks INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (subject_id) REFERENCES subjects(subject_id),
    FOREIGN KEY (section_id) REFERENCES sections(section_id)
);

CREATE TABLE IF NOT EXISTS assignment_submissions (
    submission_id UUID PRIMARY KEY,
    assignment_id UUID NOT NULL,
    student_id UUID NOT NULL,
    submission_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    submission_file_url TEXT,
    marks_obtained INT,
    feedback TEXT,
    submitted_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(assignment_id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (submitted_by) REFERENCES users(user_id) ON DELETE SET NULL,
    UNIQUE(assignment_id, student_id)
);

CREATE TABLE IF NOT EXISTS courses (
    course_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    course_name VARCHAR(200) NOT NULL,
    course_code VARCHAR(50) NOT NULL,
    description TEXT,
    instructor_id UUID,
    start_date DATE,
    end_date DATE,
    course_status VARCHAR(20) DEFAULT 'Draft' CHECK (course_status IN ('Draft', 'Active', 'Completed', 'Archived')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (instructor_id) REFERENCES staff(staff_id) ON DELETE SET NULL,
    UNIQUE(org_id, course_code)
);

CREATE TABLE IF NOT EXISTS course_modules (
    module_id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    module_name VARCHAR(200),
    module_sequence INT,
    module_content TEXT,
    content_file_url TEXT,
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS course_enrollments (
    enrollment_id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    user_id UUID NOT NULL,
    enrollment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completion_percentage INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'Active' CHECK (status IN ('Active', 'Completed', 'Dropped')),
    FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE(course_id, user_id)
);

CREATE TABLE IF NOT EXISTS course_assessments (
    assessment_id UUID PRIMARY KEY,
    course_id UUID NOT NULL,
    assessment_name VARCHAR(200),
    assessment_type VARCHAR(50),
    total_marks INT,
    passing_marks INT,
    due_date DATE,
    FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS course_assessment_results (
    result_id UUID PRIMARY KEY,
    assessment_id UUID NOT NULL,
    user_id UUID NOT NULL,
    marks_obtained INT,
    submission_date TIMESTAMP,
    feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assessment_id) REFERENCES course_assessments(assessment_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- ----------------------------------------------------------------------------
-- HR MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payroll (
    payroll_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    org_id UUID NOT NULL,
    academic_year_id UUID,
    month_of_payroll INT,
    year_of_payroll INT,
    basic_salary DECIMAL(10,2),
    allowances JSONB,
    deductions JSONB,
    gross_salary DECIMAL(10,2),
    net_salary DECIMAL(10,2),
    payment_date DATE,
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'Draft' CHECK (payment_status IN ('Draft', 'Submitted', 'Approved', 'Paid')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS staff_performance_evaluation (
    evaluation_id UUID PRIMARY KEY,
    staff_id UUID NOT NULL,
    org_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    evaluation_period VARCHAR(50),
    overall_rating DECIMAL(3,2),
    attendance_rating DECIMAL(3,2),
    teaching_quality DECIMAL(3,2),
    student_feedback DECIMAL(3,2),
    professional_development DECIMAL(3,2),
    comments TEXT,
    evaluated_by UUID,
    evaluation_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (staff_id) REFERENCES staff(staff_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (evaluated_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- ----------------------------------------------------------------------------
-- ADMINISTRATIVE MODULES
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hostel_blocks (
    block_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    block_name VARCHAR(100) NOT NULL,
    block_type VARCHAR(10) CHECK (block_type IN ('Boys', 'Girls', 'Mixed')),
    total_rooms INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS hostel_rooms (
    room_id UUID PRIMARY KEY,
    block_id UUID NOT NULL,
    room_number VARCHAR(20) NOT NULL,
    room_type VARCHAR(50),
    capacity INT,
    room_status VARCHAR(20) DEFAULT 'Available' CHECK (room_status IN ('Available', 'Occupied', 'Maintenance')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (block_id) REFERENCES hostel_blocks(block_id) ON DELETE CASCADE,
    UNIQUE(block_id, room_number)
);

CREATE TABLE IF NOT EXISTS hostel_allocations (
    allocation_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    room_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    allocation_date DATE,
    checkout_date DATE,
    allocation_status VARCHAR(20) DEFAULT 'Active' CHECK (allocation_status IN ('Active', 'Inactive', 'Completed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES hostel_rooms(room_id),
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id)
);

CREATE TABLE IF NOT EXISTS inventory_items (
    item_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    item_name VARCHAR(200) NOT NULL,
    item_code VARCHAR(50) NOT NULL,
    category VARCHAR(100),
    quantity INT,
    reorder_level INT,
    unit_price DECIMAL(10,2),
    supplier_name VARCHAR(200),
    last_purchase_date DATE,
    item_status VARCHAR(20) DEFAULT 'Active' CHECK (item_status IN ('Active', 'Inactive')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, item_code)
);

CREATE TABLE IF NOT EXISTS inventory_transactions (
    transaction_id UUID PRIMARY KEY,
    item_id UUID NOT NULL,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('Purchase', 'Usage', 'Damage', 'Return')),
    quantity INT,
    transaction_date DATE,
    reference_id VARCHAR(100),
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES inventory_items(item_id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS assets (
    asset_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    asset_name VARCHAR(200),
    asset_code VARCHAR(50) NOT NULL,
    asset_category VARCHAR(100),
    asset_location VARCHAR(200),
    purchase_date DATE,
    purchase_value DECIMAL(12,2),
    current_value DECIMAL(12,2),
    depreciation_rate DECIMAL(5,2),
    asset_status VARCHAR(20) DEFAULT 'Active' CHECK (asset_status IN ('Active', 'Inactive', 'Damaged')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(org_id, asset_code)
);

CREATE TABLE IF NOT EXISTS visitor_management (
    visitor_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    visitor_name VARCHAR(100),
    visitor_phone VARCHAR(20),
    visitor_email VARCHAR(100),
    purpose_of_visit VARCHAR(200),
    visit_date DATE,
    check_in_time TIME,
    check_out_time TIME,
    host_id UUID,
    id_proof_type VARCHAR(50),
    id_proof_number VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (host_id) REFERENCES users(user_id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS complaints (
    complaint_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    complainant_id UUID NOT NULL,
    complaint_type VARCHAR(50),
    complaint_description TEXT,
    complaint_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    priority VARCHAR(20) DEFAULT 'Medium' CHECK (priority IN ('Low', 'Medium', 'High', 'Critical')),
    status VARCHAR(20) DEFAULT 'Open' CHECK (status IN ('Open', 'In Progress', 'Resolved', 'Closed')),
    assigned_to UUID,
    resolution_description TEXT,
    resolved_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    FOREIGN KEY (complainant_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES users(user_id) ON DELETE SET NULL
);

-- ----------------------------------------------------------------------------
-- PARENT/STUDENT PORTAL
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS parent_student_mapping (
    mapping_id UUID PRIMARY KEY,
    parent_id UUID NOT NULL,
    student_id UUID NOT NULL,
    relationship VARCHAR(50),
    primary_contact BOOLEAN DEFAULT FALSE,
    org_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(parent_id, student_id)
);

CREATE TABLE IF NOT EXISTS student_performance_summary (
    summary_id UUID PRIMARY KEY,
    student_id UUID NOT NULL,
    academic_year_id UUID NOT NULL,
    org_id UUID NOT NULL,
    total_classes INT,
    total_present INT,
    attendance_percentage DECIMAL(5,2),
    average_marks DECIMAL(5,2),
    performance_status VARCHAR(50),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(student_id) ON DELETE CASCADE,
    FOREIGN KEY (academic_year_id) REFERENCES academic_years(academic_year_id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE,
    UNIQUE(student_id, academic_year_id)
);

-- ----------------------------------------------------------------------------
-- INDEXES (add indexes for UUID columns used in joins)
-- ----------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_org ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_students_user ON students(user_id);
CREATE INDEX IF NOT EXISTS idx_students_org ON students(org_id);
CREATE INDEX IF NOT EXISTS idx_staff_user ON staff(user_id);
CREATE INDEX IF NOT EXISTS idx_staff_org ON staff(org_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_student ON student_enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_section ON student_enrollments(section_id);
CREATE INDEX IF NOT EXISTS idx_enrollment_year ON student_enrollments(academic_year_id);
CREATE INDEX IF NOT EXISTS idx_marks_exam ON marks_entry(exam_schedule_id);
CREATE INDEX IF NOT EXISTS idx_marks_student ON marks_entry(student_id);
CREATE INDEX IF NOT EXISTS idx_transaction_student ON fee_transactions(student_id);
CREATE INDEX IF NOT EXISTS idx_exam_schedule_date ON exam_schedule(exam_date);
CREATE INDEX IF NOT EXISTS idx_attendance_date_section ON student_attendance(attendance_date, section_id);
CREATE INDEX IF NOT EXISTS idx_audit_user_timestamp ON audit_logs(user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_admission_status ON admission_forms(form_status, org_id);
CREATE INDEX IF NOT EXISTS idx_attendance_date ON student_attendance(attendance_date);
CREATE INDEX IF NOT EXISTS idx_staff_attendance_date ON staff_attendance(attendance_date);
CREATE INDEX IF NOT EXISTS idx_fee_transaction_status ON fee_transactions(student_id, payment_status);
CREATE INDEX IF NOT EXISTS idx_announcement_publish ON announcements(org_id, is_published, publish_date);
CREATE INDEX IF NOT EXISTS idx_message_recipient ON messages(recipient_id, is_read);
CREATE INDEX IF NOT EXISTS idx_book_title ON library_books(title, org_id);
CREATE INDEX IF NOT EXISTS idx_issue_user ON book_issue_return(user_id, is_returned);

-- ============================================================================
-- END OF SCHEMA (UUID PRIMARY KEYS, APPLICATION-GENERATED)
-- ============================================================================