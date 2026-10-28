package service

import (
	"context"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/store"
)

// AnalyticsService provides aggregated indicators and data exports for the
// institutional platform. Only institutional roles may access it, and
// organization scope requires read access to that organization.
type AnalyticsService struct{ d Deps }

func (s *AnalyticsService) authorize(ctx context.Context, p Principal, orgID string) error {
	if !p.IsInstitutional() {
		return apperr.Forbidden("institutional role required")
	}
	if orgID != "" && !s.d.canReadOrg(ctx, p, orgID) {
		return apperr.Forbidden("you do not have access to this organization")
	}
	return nil
}

// Overview returns the aggregate dashboard indicators.
func (s *AnalyticsService) Overview(ctx context.Context, p Principal, orgID string) (*models.AnalyticsOverview, error) {
	if err := s.authorize(ctx, p, orgID); err != nil {
		return nil, err
	}
	ov, err := s.d.Store.AnalyticsOverview(ctx, orgID)
	if err != nil {
		return nil, mapErr(err, "analytics")
	}
	return ov, nil
}

// Export returns a whitelisted dataset ready to be encoded as CSV.
func (s *AnalyticsService) Export(ctx context.Context, p Principal, reportType, orgID string) (*store.Report, error) {
	if err := s.authorize(ctx, p, orgID); err != nil {
		return nil, err
	}
	// Non-administrators may only export their own organization's data.
	if orgID == "" && !p.IsAdmin() {
		orgs, err := s.d.Store.ListUserOrganizations(ctx, p.UserID)
		if err != nil {
			return nil, mapErr(err, "organization")
		}
		if len(orgs) == 1 {
			orgID = orgs[0].Organization.ID
		} else {
			return nil, apperr.InvalidField("organization_id", "specify which organization to export")
		}
	}
	report, err := s.d.Store.ExportReport(ctx, store.ReportType(reportType), orgID)
	if err != nil {
		return nil, mapErr(err, "report")
	}
	return report, nil
}

// SupportedReports lists the report identifiers available for export.
func SupportedReports() []string {
	return []string{
		string(store.ReportServices),
		string(store.ReportActivities),
		string(store.ReportEnrollments),
		string(store.ReportAssistance),
		string(store.ReportOpportunities),
		string(store.ReportTerritorial),
	}
}
