package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrobarros/vitaltag-api/internal/model"
)

type ProfileRepo struct {
	db *pgxpool.Pool
}

func NewProfileRepo(db *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{db: db}
}

// FindByUser returns the user's complete profile.
func (r *ProfileRepo) FindByUser(ctx context.Context, userID uuid.UUID) (*model.FullProfile, error) {
	p := &model.FullProfile{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, first_name, last_name, sex, blood_type,
		        public_slug, public_password_hash, active, locked_until,
		        created_at, updated_at
		 FROM profiles WHERE user_id = $1`,
		userID,
	).Scan(
		&p.ID, &p.UserID, &p.FirstName, &p.LastName,
		&p.Sex, &p.BloodType, &p.PublicSlug,
		&p.PublicPasswordHash, &p.Active, &p.LockedUntil,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find profile: %w", err)
	}

	if err := r.loadClinicalData(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// FindBySlug returns a profile by public slug (for emergency access).
func (r *ProfileRepo) FindBySlug(ctx context.Context, slug uuid.UUID) (*model.FullProfile, error) {
	p := &model.FullProfile{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, first_name, last_name, sex, blood_type,
		        public_slug, public_password_hash, active, locked_until,
		        created_at, updated_at
		 FROM profiles WHERE public_slug = $1`,
		slug,
	).Scan(
		&p.ID, &p.UserID, &p.FirstName, &p.LastName,
		&p.Sex, &p.BloodType, &p.PublicSlug,
		&p.PublicPasswordHash, &p.Active, &p.LockedUntil,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find profile by slug: %w", err)
	}

	if err := r.loadClinicalData(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Upsert creates or replaces the profile and all clinical data in a transaction.
func (r *ProfileRepo) Upsert(ctx context.Context, userID uuid.UUID, req *model.ProfileRequest, publicPasswordHash string) (*model.FullProfile, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert the main profile
	p := &model.FullProfile{}
	err = tx.QueryRow(ctx,
		`INSERT INTO profiles (user_id, first_name, last_name, sex, blood_type, public_password_hash)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (user_id) DO UPDATE SET
		   first_name = EXCLUDED.first_name,
		   last_name = EXCLUDED.last_name,
		   sex = EXCLUDED.sex,
		   blood_type = EXCLUDED.blood_type,
		   public_password_hash = EXCLUDED.public_password_hash,
		   updated_at = now()
		 RETURNING id, user_id, first_name, last_name, sex, blood_type,
		           public_slug, public_password_hash, active, locked_until,
		           created_at, updated_at`,
		userID, req.FirstName, req.LastName, req.Sex, req.BloodType, publicPasswordHash,
	).Scan(
		&p.ID, &p.UserID, &p.FirstName, &p.LastName,
		&p.Sex, &p.BloodType, &p.PublicSlug,
		&p.PublicPasswordHash, &p.Active, &p.LockedUntil,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert profile: %w", err)
	}

	// Delete and recreate clinical data (full replacement — RF12)
	tables := []string{"emergency_contacts", "allergies", "medications", "conditions", "surgeries"}
	for _, t := range tables {
		if _, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE profile_id = $1", t), p.ID); err != nil {
			return nil, fmt.Errorf("clear %s: %w", t, err)
		}
	}

	// Emergency contacts
	for _, c := range req.EmergencyContacts {
		if _, err := tx.Exec(ctx,
			`INSERT INTO emergency_contacts (profile_id, name, relationship, phone, is_primary)
			 VALUES ($1,$2,$3,$4,$5)`,
			p.ID, c.Name, c.Relationship, c.Phone, c.IsPrimary,
		); err != nil {
			return nil, fmt.Errorf("insert emergency contact: %w", err)
		}
	}

	// Allergies
	for _, a := range req.Allergies {
		if _, err := tx.Exec(ctx,
			`INSERT INTO allergies (profile_id, description, severity) VALUES ($1,$2,$3)`,
			p.ID, a.Description, nullIfEmpty(a.Severity),
		); err != nil {
			return nil, fmt.Errorf("insert allergy: %w", err)
		}
	}

	// Medications
	for _, m := range req.Medications {
		if _, err := tx.Exec(ctx,
			`INSERT INTO medications (profile_id, name, dosage, frequency) VALUES ($1,$2,$3,$4)`,
			p.ID, m.Name, nullIfEmpty(m.Dosage), nullIfEmpty(m.Frequency),
		); err != nil {
			return nil, fmt.Errorf("insert medication: %w", err)
		}
	}

	// Conditions
	for _, c := range req.Conditions {
		if _, err := tx.Exec(ctx,
			`INSERT INTO conditions (profile_id, description, notes) VALUES ($1,$2,$3)`,
			p.ID, c.Description, nullIfEmpty(c.Notes),
		); err != nil {
			return nil, fmt.Errorf("insert condition: %w", err)
		}
	}

	// Surgeries
	for _, s := range req.Surgeries {
		if _, err := tx.Exec(ctx,
			`INSERT INTO surgeries (profile_id, description, year) VALUES ($1,$2,$3)`,
			p.ID, s.Description, s.Year,
		); err != nil {
			return nil, fmt.Errorf("insert surgery: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Reload clinical data for the response
	if err := r.loadClinicalData(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// RecordAttempt records a public-access attempt (RF09 / LGPD).
func (r *ProfileRepo) RecordAttempt(ctx context.Context, profileID uuid.UUID, ip string, success bool) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO access_attempts (profile_id, ip, success) VALUES ($1, $2::inet, $3)`,
		profileID, ip, success,
	)
	return err
}

// CountRecentFailures counts failed attempts in the last 15 minutes (RF09).
func (r *ProfileRepo) CountRecentFailures(ctx context.Context, profileID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM access_attempts
		 WHERE profile_id = $1 AND success = false AND attempted_at > now() - interval '15 minutes'`,
		profileID,
	).Scan(&count)
	return count, err
}

// Lock sets locked_until for the profile (RF09).
func (r *ProfileRepo) Lock(ctx context.Context, profileID uuid.UUID, until time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE profiles SET locked_until = $1 WHERE id = $2`,
		until, profileID,
	)
	return err
}

// ---------- internal helpers --------------------------------------------------

func (r *ProfileRepo) loadClinicalData(ctx context.Context, p *model.FullProfile) error {
	// Emergency contacts
	rows, err := r.db.Query(ctx,
		`SELECT id, profile_id, name, COALESCE(relationship,''), phone, is_primary
		 FROM emergency_contacts WHERE profile_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("find emergency contacts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c model.EmergencyContact
		if err := rows.Scan(&c.ID, &c.ProfileID, &c.Name, &c.Relationship, &c.Phone, &c.IsPrimary); err != nil {
			return err
		}
		p.EmergencyContacts = append(p.EmergencyContacts, c)
	}

	// Allergies
	aRows, err := r.db.Query(ctx,
		`SELECT id, profile_id, description, COALESCE(severity,'') FROM allergies WHERE profile_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("find allergies: %w", err)
	}
	defer aRows.Close()
	for aRows.Next() {
		var a model.Allergy
		if err := aRows.Scan(&a.ID, &a.ProfileID, &a.Description, &a.Severity); err != nil {
			return err
		}
		p.Allergies = append(p.Allergies, a)
	}

	// Medications
	mRows, err := r.db.Query(ctx,
		`SELECT id, profile_id, name, COALESCE(dosage,''), COALESCE(frequency,'')
		 FROM medications WHERE profile_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("find medications: %w", err)
	}
	defer mRows.Close()
	for mRows.Next() {
		var m model.Medication
		if err := mRows.Scan(&m.ID, &m.ProfileID, &m.Name, &m.Dosage, &m.Frequency); err != nil {
			return err
		}
		p.Medications = append(p.Medications, m)
	}

	// Conditions
	cRows, err := r.db.Query(ctx,
		`SELECT id, profile_id, description, COALESCE(notes,'') FROM conditions WHERE profile_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("find conditions: %w", err)
	}
	defer cRows.Close()
	for cRows.Next() {
		var c model.Condition
		if err := cRows.Scan(&c.ID, &c.ProfileID, &c.Description, &c.Notes); err != nil {
			return err
		}
		p.Conditions = append(p.Conditions, c)
	}

	// Surgeries
	sRows, err := r.db.Query(ctx,
		`SELECT id, profile_id, description, year FROM surgeries WHERE profile_id = $1`, p.ID)
	if err != nil {
		return fmt.Errorf("find surgeries: %w", err)
	}
	defer sRows.Close()
	for sRows.Next() {
		var s model.Surgery
		if err := sRows.Scan(&s.ID, &s.ProfileID, &s.Description, &s.Year); err != nil {
			return err
		}
		p.Surgeries = append(p.Surgeries, s)
	}

	// Ensure slices are not null in JSON
	if p.EmergencyContacts == nil {
		p.EmergencyContacts = []model.EmergencyContact{}
	}
	if p.Allergies == nil {
		p.Allergies = []model.Allergy{}
	}
	if p.Medications == nil {
		p.Medications = []model.Medication{}
	}
	if p.Conditions == nil {
		p.Conditions = []model.Condition{}
	}
	if p.Surgeries == nil {
		p.Surgeries = []model.Surgery{}
	}

	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
