package service

import (
	"time"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/validate"
)

// maxAgeYears is the oldest plausible age accepted for a date of birth.
const maxAgeYears = 120

func validateBirthDate(v *validate.Errors, field string, birthDate *time.Time) {
	if birthDate == nil {
		return
	}
	now := time.Now()
	if birthDate.After(now) {
		v.Add(field, "cannot be in the future")
		return
	}
	if birthDate.Before(now.AddDate(-maxAgeYears, 0, 0)) {
		v.Add(field, "must be a plausible date of birth")
	}
}

func validateHealthSystem(v *validate.Errors, field string, value models.HealthSystem) {
	if !value.Valid() {
		v.Add(field, "must be one of: fonasa, isapre, capredena, dipreca, private, other")
	}
}

func validateSex(v *validate.Errors, field string, value models.Sex) {
	if !value.Valid() {
		v.Add(field, "must be one of: female, male, other, prefer_not_to_say")
	}
}

func validateLivingSituation(v *validate.Errors, field string, value models.LivingSituation) {
	if !value.Valid() {
		v.Add(field, "must be one of: alone, with_partner, with_children, with_partner_and_children, with_family, with_caregiver, other")
	}
}
