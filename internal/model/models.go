package model

import (
	"time"

	"github.com/google/uuid"
)

// ---------- User --------------------------------------------------------------

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Consent      bool      `json:"consent"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------- Profile -----------------------------------------------------------

type SexType string
type BloodType string

const (
	SexFemale      SexType = "female"
	SexMale        SexType = "male"
	SexOther       SexType = "other"
	SexNotInformed SexType = "not_informed"
)

type Profile struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	Sex                SexType    `json:"sex"`
	BloodType          BloodType  `json:"blood_type"`
	PublicSlug         uuid.UUID  `json:"public_slug"`
	PublicPasswordHash string     `json:"-"`
	Active             bool       `json:"active"`
	LockedUntil        *time.Time `json:"locked_until,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ---------- Clinical data -----------------------------------------------------

type EmergencyContact struct {
	ID           uuid.UUID `json:"id"`
	ProfileID    uuid.UUID `json:"profile_id"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship,omitempty"`
	Phone        string    `json:"phone"`
	IsPrimary    bool      `json:"is_primary"`
}

type Allergy struct {
	ID          uuid.UUID `json:"id"`
	ProfileID   uuid.UUID `json:"profile_id"`
	Description string    `json:"description"`
	Severity    string    `json:"severity,omitempty"`
}

type Medication struct {
	ID        uuid.UUID `json:"id"`
	ProfileID uuid.UUID `json:"profile_id"`
	Name      string    `json:"name"`
	Dosage    string    `json:"dosage,omitempty"`
	Frequency string    `json:"frequency,omitempty"`
}

type Condition struct {
	ID          uuid.UUID `json:"id"`
	ProfileID   uuid.UUID `json:"profile_id"`
	Description string    `json:"description"`
	Notes       string    `json:"notes,omitempty"`
}

type Surgery struct {
	ID          uuid.UUID `json:"id"`
	ProfileID   uuid.UUID `json:"profile_id"`
	Description string    `json:"description"`
	Year        *int16    `json:"year,omitempty"`
}

// ---------- Complete profile (API response) -----------------------------------

type FullProfile struct {
	Profile
	EmergencyContacts []EmergencyContact `json:"emergency_contacts"`
	Allergies         []Allergy          `json:"allergies"`
	Medications       []Medication       `json:"medications"`
	Conditions        []Condition        `json:"conditions"`
	Surgeries         []Surgery          `json:"surgeries"`
}

// ---------- Requests ----------------------------------------------------------

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Consent  bool   `json:"consent"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ProfileRequest struct {
	FirstName         string                    `json:"first_name"`
	LastName          string                    `json:"last_name"`
	Sex               SexType                   `json:"sex"`
	BloodType         BloodType                 `json:"blood_type"`
	PublicPassword    string                    `json:"public_password"`
	EmergencyContacts []EmergencyContactRequest `json:"emergency_contacts"`
	Allergies         []AllergyRequest          `json:"allergies"`
	Medications       []MedicationRequest       `json:"medications"`
	Conditions        []ConditionRequest        `json:"conditions"`
	Surgeries         []SurgeryRequest          `json:"surgeries"`
}

type EmergencyContactRequest struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Phone        string `json:"phone"`
	IsPrimary    bool   `json:"is_primary"`
}

type AllergyRequest struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

type MedicationRequest struct {
	Name      string `json:"name"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency"`
}

type ConditionRequest struct {
	Description string `json:"description"`
	Notes       string `json:"notes"`
}

type SurgeryRequest struct {
	Description string `json:"description"`
	Year        *int16 `json:"year"`
}

type PublicAccessRequest struct {
	Password string `json:"password"`
}

// ---------- Responses ---------------------------------------------------------

type TokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
