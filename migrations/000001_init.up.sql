-- VitalTag — Database schema (PostgreSQL 16)
-- Phase 1 — Integrated Project Development Course

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerated types enforce consistency at the database level
CREATE TYPE sex_type   AS ENUM ('female', 'male', 'other', 'not_informed');
CREATE TYPE blood_type AS ENUM ('A+','A-','B+','B-','AB+','AB-','O+','O-','unknown');

-- ---------------------------------------------------------------- users
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(72)  NOT NULL,          -- bcrypt cost 12 (RNF01)
    consent       BOOLEAN      NOT NULL DEFAULT FALSE,  -- LGPD art. 11 (RNF04)
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ------------------------------------------------------------- profiles
-- One-to-one relationship with user, enforced by UNIQUE on user_id
CREATE TABLE profiles (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID NOT NULL UNIQUE
                          REFERENCES users(id) ON DELETE CASCADE,
    first_name            VARCHAR(100) NOT NULL,
    last_name             VARCHAR(100) NOT NULL,
    sex                   sex_type    NOT NULL DEFAULT 'not_informed',
    blood_type            blood_type  NOT NULL DEFAULT 'unknown',
    public_slug           UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), -- random: prevents enumeration
    public_password_hash  VARCHAR(72) NOT NULL,     -- distinct from the account password (RF05)
    active                BOOLEAN     NOT NULL DEFAULT TRUE,
    locked_until          TIMESTAMPTZ,              -- lockout after failed attempts (RF09)
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_profiles_slug ON profiles(public_slug);

-- --------------------------------------------------- emergency_contacts
CREATE TABLE emergency_contacts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id    UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name          VARCHAR(150) NOT NULL,
    relationship  VARCHAR(60),
    phone         VARCHAR(20)  NOT NULL,
    is_primary    BOOLEAN      NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_emergency_contacts_profile ON emergency_contacts(profile_id);

-- ------------------------------------------------------ clinical lists
-- Normalized (3NF) instead of free-text fields: enables querying,
-- counting, and structured display on the public page.
CREATE TABLE allergies (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    description  VARCHAR(150) NOT NULL,
    severity     VARCHAR(20) CHECK (severity IN ('mild','moderate','severe'))
);

CREATE TABLE medications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name         VARCHAR(150) NOT NULL,
    dosage       VARCHAR(80),
    frequency    VARCHAR(80)
);

CREATE TABLE conditions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    description  VARCHAR(150) NOT NULL,
    notes        TEXT
);

CREATE TABLE surgeries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    description  VARCHAR(150) NOT NULL,
    year         SMALLINT CHECK (year BETWEEN 1900 AND 2100)
);

CREATE INDEX idx_allergies_profile   ON allergies(profile_id);
CREATE INDEX idx_medications_profile ON medications(profile_id);
CREATE INDEX idx_conditions_profile  ON conditions(profile_id);
CREATE INDEX idx_surgeries_profile   ON surgeries(profile_id);

-- ----------------------------------------------------- access_attempts
-- Supports RF09 lockout and the audit trail required by Brazil's LGPD
CREATE TABLE access_attempts (
    id            BIGSERIAL PRIMARY KEY,
    profile_id    UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    ip            INET,
    success       BOOLEAN     NOT NULL,
    attempted_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_access_attempts_profile_time ON access_attempts(profile_id, attempted_at DESC);
