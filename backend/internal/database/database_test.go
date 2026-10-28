package database

import "testing"

func TestIsDemoMigration(t *testing.T) {
	demo := []string{"0007_seed_demo_data.sql", "0008_seed_web_demo_data.sql", "0011_demo_data_refresh.sql"}
	real := []string{"0001_identity.sql", "0006_seed_categories.sql", "0009_institutional_approval.sql", "0010_user_profile.sql"}
	for _, n := range demo {
		if !IsDemoMigration(n) {
			t.Errorf("%s should be a demo migration", n)
		}
	}
	for _, n := range real {
		if IsDemoMigration(n) {
			t.Errorf("%s must not be a demo migration", n)
		}
	}
}
