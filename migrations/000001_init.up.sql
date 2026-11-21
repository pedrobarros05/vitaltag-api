-- VitalTag — Database schema (PostgreSQL 16)
-- Phase 1 — Integrated Project Development Course

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerated types enforce consistency at the database level
CREATE TYPE sexo_tipo           AS ENUM ('feminino', 'masculino', 'outro', 'nao_informado');
CREATE TYPE tipo_sanguineo_tipo AS ENUM ('A+','A-','B+','B-','AB+','AB-','O+','O-','nao_sei');

-- ---------------------------------------------------------------- user
CREATE TABLE usuario (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    senha_hash    VARCHAR(72)  NOT NULL,          -- bcrypt custo 12 (RNF01)
    consentimento BOOLEAN      NOT NULL DEFAULT FALSE,  -- LGPD art. 11 (RNF04)
    criado_em     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ----------------------------------------------------------------- profile
-- One-to-one relationship with user, enforced by UNIQUE on usuario_id
CREATE TABLE perfil (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id          UUID NOT NULL UNIQUE
                        REFERENCES usuario(id) ON DELETE CASCADE,
    nome                VARCHAR(100) NOT NULL,
    sobrenome           VARCHAR(100) NOT NULL,
    sexo                sexo_tipo           NOT NULL DEFAULT 'nao_informado',
    tipo_sanguineo      tipo_sanguineo_tipo NOT NULL DEFAULT 'nao_sei',
    slug_publico        UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), -- random: prevents enumeration
    senha_publica_hash  VARCHAR(72) NOT NULL,     -- distinta da senha da conta (RF05)
    ativo               BOOLEAN     NOT NULL DEFAULT TRUE,
    bloqueado_ate       TIMESTAMPTZ,              -- bloqueio por tentativas (RF09)
    criado_em           TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_perfil_slug ON perfil(slug_publico);

-- --------------------------------------------------- emergency_contact
CREATE TABLE contato_emergencia (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    perfil_id  UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    nome       VARCHAR(150) NOT NULL,
    parentesco VARCHAR(60),
    telefone   VARCHAR(20)  NOT NULL,
    principal  BOOLEAN      NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_contato_perfil ON contato_emergencia(perfil_id);

-- ------------------------------------------------------ clinical lists
-- Normalized (3NF) instead of free-text fields: enables querying,
-- counting, and structured display on the public page.
CREATE TABLE alergia (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    perfil_id   UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    descricao   VARCHAR(150) NOT NULL,
    gravidade   VARCHAR(20) CHECK (gravidade IN ('leve','moderada','grave'))
);

CREATE TABLE medicamento (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    perfil_id   UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    nome        VARCHAR(150) NOT NULL,
    dosagem     VARCHAR(80),
    frequencia  VARCHAR(80)
);

CREATE TABLE doenca (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    perfil_id   UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    descricao   VARCHAR(150) NOT NULL,
    observacao  TEXT
);

CREATE TABLE cirurgia (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    perfil_id       UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    descricao       VARCHAR(150) NOT NULL,
    ano_realizacao  SMALLINT CHECK (ano_realizacao BETWEEN 1900 AND 2100)
);

CREATE INDEX idx_alergia_perfil     ON alergia(perfil_id);
CREATE INDEX idx_medicamento_perfil ON medicamento(perfil_id);
CREATE INDEX idx_doenca_perfil      ON doenca(perfil_id);
CREATE INDEX idx_cirurgia_perfil    ON cirurgia(perfil_id);

-- ----------------------------------------------------- access_attempt
-- Supports RF09 lockout and the audit trail required by Brazil's LGPD
CREATE TABLE tentativa_acesso (
    id         BIGSERIAL PRIMARY KEY,
    perfil_id  UUID NOT NULL REFERENCES perfil(id) ON DELETE CASCADE,
    ip         INET,
    sucesso    BOOLEAN     NOT NULL,
    data_hora  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tentativa_perfil_data ON tentativa_acesso(perfil_id, data_hora DESC);
