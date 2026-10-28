package models

// CountByLabel is a generic label/count aggregation result.
type CountByLabel struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// TerritorialDistribution aggregates resources by comuna/region.
type TerritorialDistribution struct {
	Comuna string `json:"comuna"`
	Region string `json:"region"`
	Count  int    `json:"count"`
}

// AnalyticsOverview is the aggregate dashboard for the institutional web app.
type AnalyticsOverview struct {
	ServicesTotal       int                       `json:"services_total"`
	ServicesActive      int                       `json:"services_active"`
	ActivitiesTotal     int                       `json:"activities_total"`
	ActivitiesPublished int                       `json:"activities_published"`
	ActivitiesUpcoming  int                       `json:"activities_upcoming"`
	EnrollmentsTotal    int                       `json:"enrollments_total"`
	EnrollmentsActive   int                       `json:"enrollments_active"`
	UsersTotal          int                       `json:"users_total"`
	UsersMobile         int                       `json:"users_mobile"`
	OrganizationsTotal  int                       `json:"organizations_total"`
	OpportunitiesOpen   int                       `json:"opportunities_open"`
	AssistanceTotal     int                       `json:"assistance_total"`
	AssistancePending   int                       `json:"assistance_pending"`
	AverageOccupancy    float64                   `json:"average_occupancy"`
	ServicesByCategory  []CountByLabel            `json:"services_by_category"`
	ActivitiesByType    []CountByLabel            `json:"activities_by_type"`
	Territorial         []TerritorialDistribution `json:"territorial_distribution"`
}

// Scope optionally restricts analytics to one organization.
type Scope struct {
	OrganizationID string
}
