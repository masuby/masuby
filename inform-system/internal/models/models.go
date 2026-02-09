package models

import (
	"time"
)

// User represents a registered user (Regional/Ward Committee member)
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"` // admin, regional_committee, ward_committee, viewer
	CommitteeID  *int64    `json:"committee_id,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Committee represents Regional or Ward Disaster Committee
type Committee struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // regional, ward
	ADM1Code      string    `json:"adm1_code"`
	ADM1Name      string    `json:"adm1_name"`
	ADM2Code      *string   `json:"adm2_code,omitempty"`
	ADM2Name      *string   `json:"adm2_name,omitempty"`
	ContactPerson string    `json:"contact_person"`
	ContactPhone  string    `json:"contact_phone"`
	ContactEmail  string    `json:"contact_email"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Indicator represents an INFORM indicator definition
type Indicator struct {
	ID              int64   `json:"id"`
	Code            string  `json:"code"`            // e.g., HA.NAT.DR-FRE
	Name            string  `json:"name"`            // e.g., Drought Frequency
	Description     string  `json:"description"`
	Dimension       string  `json:"dimension"`       // HAZARD, VULNERABILITY, COPING_CAPACITY
	Category        string  `json:"category"`        // e.g., Natural Hazards, Socio-Economic
	Component       string  `json:"component"`       // e.g., Drought, Flood
	Unit            string  `json:"unit"`            // e.g., events/year, percentage
	DataType        string  `json:"data_type"`       // numeric, percentage, index, count, binary
	MinValue        float64 `json:"min_value"`
	MaxValue        float64 `json:"max_value"`
	Resolution      string  `json:"resolution"`      // national, adm1, adm2
	Transformation  string  `json:"transformation"`  // none, log, sqrt
	InvertScale     bool    `json:"invert_scale"`    // true if higher is better
	Weight          float64 `json:"weight"`
	DataSource      string  `json:"data_source"`
	UpdateFrequency string  `json:"update_frequency"`
	IsActive        bool    `json:"is_active"`
}

// DataEntry represents a single data entry for an indicator
type DataEntry struct {
	ID           int64      `json:"id"`
	IndicatorID  int64      `json:"indicator_id"`
	IndicatorCode string    `json:"indicator_code"`
	Country      string     `json:"country"`
	ISO3         string     `json:"iso3"`
	ADM1Code     *string    `json:"adm1_code,omitempty"`
	ADM1Name     *string    `json:"adm1_name,omitempty"`
	ADM2Code     *string    `json:"adm2_code,omitempty"`
	ADM2Name     *string    `json:"adm2_name,omitempty"`
	RawValue     float64    `json:"raw_value"`
	NormalizedValue *float64 `json:"normalized_value,omitempty"`
	Year         int        `json:"year"`
	Quarter      *int       `json:"quarter,omitempty"`
	DataSource   string     `json:"data_source"`
	Notes        string     `json:"notes"`
	EnteredByID  int64      `json:"entered_by_id"`
	EnteredBy    string     `json:"entered_by"`
	VerifiedByID *int64     `json:"verified_by_id,omitempty"`
	Status       string     `json:"status"` // draft, submitted, verified, rejected
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RiskScore represents calculated INFORM risk scores
type RiskScore struct {
	ID                    int64     `json:"id"`
	Country               string    `json:"country"`
	ISO3                  string    `json:"iso3"`
	ADM1Code              *string   `json:"adm1_code,omitempty"`
	ADM1Name              *string   `json:"adm1_name,omitempty"`
	ADM2Code              *string   `json:"adm2_code,omitempty"`
	ADM2Name              *string   `json:"adm2_name,omitempty"`
	Resolution            string    `json:"resolution"` // national, adm1, adm2
	Year                  int       `json:"year"`

	// Hazard Dimension
	HazardNatural         float64   `json:"hazard_natural"`
	HazardHuman           float64   `json:"hazard_human"`
	HazardTotal           float64   `json:"hazard_total"`

	// Vulnerability Dimension
	VulnerabilitySocioEcon float64  `json:"vulnerability_socio_econ"`
	VulnerabilityVulnGroup float64  `json:"vulnerability_vuln_group"`
	VulnerabilityTotal     float64  `json:"vulnerability_total"`

	// Coping Capacity Dimension (inverted to Lack of)
	CopingInfrastructure  float64   `json:"coping_infrastructure"`
	CopingInstitutional   float64   `json:"coping_institutional"`
	LackOfCopingCapacity  float64   `json:"lack_of_coping_capacity"`

	// Final INFORM Risk Score
	RiskScore             float64   `json:"risk_score"`
	RiskClass             string    `json:"risk_class"` // Very Low, Low, Medium, High, Very High

	CalculatedAt          time.Time `json:"calculated_at"`
}

// AuditLog tracks all data changes
type AuditLog struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	Action      string    `json:"action"` // create, update, delete, verify
	EntityType  string    `json:"entity_type"`
	EntityID    int64     `json:"entity_id"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

// WebSocket message types
type WSMessage struct {
	Type      string      `json:"type"` // data_entry, risk_update, notification
	Action    string      `json:"action"` // created, updated, deleted
	Data      interface{} `json:"data"`
	UserID    int64       `json:"user_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// API Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Pagination
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Formula documentation for transparency
type FormulaDoc struct {
	Name        string   `json:"name"`
	Formula     string   `json:"formula"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
	Example     string   `json:"example"`
}

// Sheet linkage for transparency
type SheetLinkage struct {
	SourceSheet string `json:"source_sheet"`
	TargetSheet string `json:"target_sheet"`
	LinkType    string `json:"link_type"` // data_flow, formula_reference, aggregation
	Description string `json:"description"`
}

// ============================================================================
// MODULE 03: EARLY WARNING SYSTEM MODELS
// ============================================================================

// HazardForecast represents a hazard forecast from an institution (TMA, MoW, MoH, MoA, GST)
type HazardForecast struct {
	ID               int64     `json:"id"`
	HazardType       string    `json:"hazard_type"`       // flood, drought, cyclone, disease_outbreak, earthquake, etc.
	Institution      string    `json:"institution"`       // TMA, MoW, MoH, MoA, GST
	InstitutionName  string    `json:"institution_name"`
	IntensityLevel   string    `json:"intensity_level"`   // low, moderate, high, very_high
	IntensityValue   float64   `json:"intensity_value"`   // Quantitative value (e.g., mm rainfall)
	IntensityUnit    string    `json:"intensity_unit"`    // mm, m/s, cases, etc.
	Confidence       string    `json:"confidence"`        // low, medium, high
	SpatialExtent    string    `json:"spatial_extent"`    // JSON array of affected regions/districts
	ValidFrom        time.Time `json:"valid_from"`
	ValidTo          time.Time `json:"valid_to"`
	ForecastDay      int       `json:"forecast_day"`      // 1-5 (Today, Tomorrow, Day 3, etc.)
	Description      string    `json:"description"`
	DataSource       string    `json:"data_source"`
	IssuedByID       int64     `json:"issued_by_id"`
	IssuedByName     string    `json:"issued_by_name"`
	Status           string    `json:"status"`            // draft, submitted, validated, expired
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Warning represents a risk-informed warning generated from hazard + risk context
type Warning struct {
	ID                   int64     `json:"id"`
	WarningID            string    `json:"warning_id"`          // e.g., TZA-2026-02-09-001
	HazardForecastID     int64     `json:"hazard_forecast_id"`
	HazardType           string    `json:"hazard_type"`
	ADM1Code             string    `json:"adm1_code"`
	ADM1Name             string    `json:"adm1_name"`
	ADM2Code             *string   `json:"adm2_code,omitempty"`
	ADM2Name             *string   `json:"adm2_name,omitempty"`

	// Hazard component
	HazardIntensity      float64   `json:"hazard_intensity"`    // 0-10 normalized

	// Risk context from Module 02
	VulnerabilityScore   float64   `json:"vulnerability_score"`
	CopingCapacityScore  float64   `json:"coping_capacity_score"`
	BaselineRiskScore    float64   `json:"baseline_risk_score"`
	BaselineRiskClass    string    `json:"baseline_risk_class"`

	// Risk-informed warning calculation
	RiskSensitivity      float64   `json:"risk_sensitivity"`    // sqrt(V * LCC)
	WarningScore         float64   `json:"warning_score"`       // sqrt(hazard * sensitivity)
	WarningLevel         string    `json:"warning_level"`       // monitor, advisory, warning, major_warning, emergency
	WarningColor         string    `json:"warning_color"`       // green, yellow, orange, red, purple

	// Impact assessment
	PopulationAffected   int64     `json:"population_affected"`
	VulnerablePopulation int64     `json:"vulnerable_population"`
	InfrastructureAtRisk string    `json:"infrastructure_at_risk"` // JSON object

	// Response information
	RecommendedActions   string    `json:"recommended_actions"`  // JSON array
	ResponseLevel        string    `json:"response_level"`       // none, preparedness, partial, full

	// PMO validation
	ValidatedByID        *int64    `json:"validated_by_id,omitempty"`
	ValidatedByName      *string   `json:"validated_by_name,omitempty"`
	ValidationNotes      *string   `json:"validation_notes,omitempty"`
	ValidatedAt          *time.Time `json:"validated_at,omitempty"`

	// Status
	Status               string    `json:"status"`              // pending, validated, disseminated, expired
	DisseminatedAt       *time.Time `json:"disseminated_at,omitempty"`
	ExpiresAt            time.Time `json:"expires_at"`

	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// WarningDissemination tracks warning dissemination to different audiences
type WarningDissemination struct {
	ID           int64     `json:"id"`
	WarningID    int64     `json:"warning_id"`
	Audience     string    `json:"audience"`     // pmo, sector_ministry, regional_authority, public
	Channel      string    `json:"channel"`      // sms, email, dashboard, radio, social_media
	Message      string    `json:"message"`
	Recipients   string    `json:"recipients"`   // JSON array of recipient contacts
	SentAt       time.Time `json:"sent_at"`
	DeliveryStatus string  `json:"delivery_status"` // pending, sent, delivered, failed
	CreatedAt    time.Time `json:"created_at"`
}

// WarningFeedback tracks actual impacts vs predicted for learning
type WarningFeedback struct {
	ID                   int64     `json:"id"`
	WarningID            int64     `json:"warning_id"`
	ActualImpacts        string    `json:"actual_impacts"`        // JSON object
	PopulationActual     int64     `json:"population_actual"`
	CasualtiesActual     int64     `json:"casualties_actual"`
	EconomicLossActual   float64   `json:"economic_loss_actual"`
	ResponseTimeHours    float64   `json:"response_time_hours"`
	WarningEffectiveness string    `json:"warning_effectiveness"` // poor, adequate, good, excellent
	LessonsLearned       string    `json:"lessons_learned"`
	ThresholdAdjustments string    `json:"threshold_adjustments"` // JSON object
	SubmittedByID        int64     `json:"submitted_by_id"`
	SubmittedByName      string    `json:"submitted_by_name"`
	CreatedAt            time.Time `json:"created_at"`
}
