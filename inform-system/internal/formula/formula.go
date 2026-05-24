package formula

import (
	"math"
)

// INFORM Formula Engine
// Implements all calculations for the INFORM Risk Index

// =====================================
// CORE FORMULAS
// =====================================

// CalculateRisk computes the final INFORM Risk Score
// Formula: RISK = HAZARD^(1/3) × VULNERABILITY^(1/3) × LOCC^(1/3)
// This is a cubic geometric mean of the three dimensions
func CalculateRisk(hazard, vulnerability, lackOfCopingCapacity float64) float64 {
	if hazard <= 0 || vulnerability <= 0 || lackOfCopingCapacity <= 0 {
		return 0
	}
	risk := math.Pow(hazard, 1.0/3.0) * math.Pow(vulnerability, 1.0/3.0) * math.Pow(lackOfCopingCapacity, 1.0/3.0)
	return Round(risk, 1)
}

// CalculateDimension computes a dimension score from two categories
// Formula: (10 - GEOMEAN(((10-CAT1)/10*9+1), ((10-CAT2)/10*9+1))) / 9 * 10
// This is an adjusted geometric mean that handles the 0-10 scale
func CalculateDimension(category1, category2 float64) float64 {
	if category1 < 0 || category2 < 0 {
		return 0
	}

	// Adjust values for geometric mean calculation
	adj1 := ((10.0 - category1) / 10.0 * 9.0) + 1.0
	adj2 := ((10.0 - category2) / 10.0 * 9.0) + 1.0

	// Calculate geometric mean
	geomean := math.Sqrt(adj1 * adj2)

	// Convert back to 0-10 scale
	result := (10.0 - geomean) / 9.0 * 10.0

	return Round(result, 1)
}

// CalculateDimensionMultiple computes a dimension score from multiple categories
// Uses the same adjusted geometric mean formula but for n categories
func CalculateDimensionMultiple(categories []float64) float64 {
	if len(categories) == 0 {
		return 0
	}

	if len(categories) == 1 {
		return Round(categories[0], 1)
	}

	// Adjust all values
	product := 1.0
	for _, cat := range categories {
		if cat < 0 {
			continue
		}
		adj := ((10.0 - cat) / 10.0 * 9.0) + 1.0
		product *= adj
	}

	// Calculate geometric mean
	geomean := math.Pow(product, 1.0/float64(len(categories)))

	// Convert back to 0-10 scale
	result := (10.0 - geomean) / 9.0 * 10.0

	return Round(result, 1)
}

// CalculateCategory computes a category score from multiple components/indicators
// Formula: Simple arithmetic average (AVERAGE)
func CalculateCategory(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	count := 0
	for _, v := range values {
		if !math.IsNaN(v) && v >= 0 {
			sum += v
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return Round(sum/float64(count), 1)
}

// CalculateCategoryGeomean aggregates components into a category using the
// INFORM scaled geometric mean (PDF Box 6 / footnote 33). Used by Natural
// Hazard (Table 5) and Vulnerable Groups (Table 13) per the INFORM 2017
// methodology.
func CalculateCategoryGeomean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	valid := make([]float64, 0, len(values))
	for _, v := range values {
		if !math.IsNaN(v) && v >= 0 {
			valid = append(valid, v)
		}
	}
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return Round(valid[0], 1)
	}

	// Step 1: rescale 0–10 → 1–10 via inversion (Box 6)
	product := 1.0
	for _, v := range valid {
		adj := ((10.0-v)/10.0*9.0) + 1.0
		product *= adj
	}
	// Step 2: geometric mean
	geo := math.Pow(product, 1.0/float64(len(valid)))
	// Step 3: rescale back to 0–10
	result := (10.0 - geo) / 9.0 * 10.0
	return Round(result, 1)
}

// CalculateCategoryMax returns the maximum of valid component scores. Used by
// Human Hazard category (PDF Table 7: MAXIMUM of Current vs Projected).
func CalculateCategoryMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := math.Inf(-1)
	found := false
	for _, v := range values {
		if !math.IsNaN(v) && v >= 0 {
			if v > max {
				max = v
			}
			found = true
		}
	}
	if !found {
		return 0
	}
	return Round(max, 1)
}

// CalculateWeightedCategory computes a weighted category score
func CalculateWeightedCategory(values, weights []float64) float64 {
	if len(values) == 0 || len(values) != len(weights) {
		return 0
	}

	weightedSum := 0.0
	totalWeight := 0.0

	for i, v := range values {
		if !math.IsNaN(v) && v >= 0 {
			weightedSum += v * weights[i]
			totalWeight += weights[i]
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return Round(weightedSum/totalWeight, 1)
}

// =====================================
// NORMALIZATION FUNCTIONS
// =====================================

// NormalizeMinMax normalizes a value to 0-10 scale using min-max normalization
// Formula: (value - min) / (max - min) * 10
func NormalizeMinMax(value, minVal, maxVal float64, invertScale bool) float64 {
	if maxVal == minVal {
		return 5.0 // Return middle value if range is zero
	}

	// Clamp value to range
	if value < minVal {
		value = minVal
	}
	if value > maxVal {
		value = maxVal
	}

	normalized := ((value - minVal) / (maxVal - minVal)) * 10.0

	if invertScale {
		normalized = 10.0 - normalized
	}

	return Round(normalized, 2)
}

// NormalizeWithLog applies log transformation before normalization
// Used for highly skewed data like population counts
func NormalizeWithLog(value, minVal, maxVal float64, invertScale bool) float64 {
	if value <= 0 {
		value = 0.001 // Small positive value for log
	}
	if minVal <= 0 {
		minVal = 0.001
	}

	logValue := math.Log(value)
	logMin := math.Log(minVal)
	logMax := math.Log(maxVal)

	return NormalizeMinMax(logValue, logMin, logMax, invertScale)
}

// NormalizeWithSqrt applies square root transformation before normalization
func NormalizeWithSqrt(value, minVal, maxVal float64, invertScale bool) float64 {
	if value < 0 {
		value = 0
	}

	sqrtValue := math.Sqrt(value)
	sqrtMin := math.Sqrt(math.Max(0, minVal))
	sqrtMax := math.Sqrt(math.Max(0, maxVal))

	return NormalizeMinMax(sqrtValue, sqrtMin, sqrtMax, invertScale)
}

// =====================================
// OUTLIER DETECTION
// =====================================

// DetectOutliersIQR identifies outliers using the Interquartile Range method
// Returns lower and upper bounds for valid values
func DetectOutliersIQR(values []float64, k float64) (lowerBound, upperBound float64) {
	if len(values) < 4 {
		return math.Inf(-1), math.Inf(1)
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sortFloat64s(sorted)

	// Calculate quartiles
	q1 := percentile(sorted, 25)
	q3 := percentile(sorted, 75)
	iqr := q3 - q1

	// Calculate bounds (k is typically 1.5 for moderate outliers, 3.0 for extreme)
	lowerBound = q1 - (k * iqr)
	upperBound = q3 + (k * iqr)

	return lowerBound, upperBound
}

// ClampOutliers replaces outliers with the nearest valid bound
func ClampOutliers(value, lowerBound, upperBound float64) float64 {
	if value < lowerBound {
		return lowerBound
	}
	if value > upperBound {
		return upperBound
	}
	return value
}

// =====================================
// RISK CLASSIFICATION
// =====================================

// ClassifyRisk returns the risk class based on the INFORM methodology (PDF
// Table 20: thresholds for the final RISK score).
func ClassifyRisk(riskScore float64) string {
	switch {
	case riskScore < 2.0:
		return "Very Low"
	case riskScore < 3.5:
		return "Low"
	case riskScore < 5.0:
		return "Medium"
	case riskScore < 6.5:
		return "High"
	default:
		return "Very High"
	}
}

// PerDimensionThresholds mirrors PDF Tables 21 (dimensions) and the operational
// Tanzania-tuned thresholds embedded in TZ_INFORM_model.xlsx "Thresholds" sheet.
// Each entry is the upper bound (exclusive) of the named class.
type ClassThresholds struct {
	VeryLowMax  float64
	LowMax      float64
	MediumMax   float64
	HighMax     float64
}

// INFORMGlobalThresholds (PDF Tables 20 + 21)
var INFORMGlobalThresholds = map[string]ClassThresholds{
	"RISK":           {VeryLowMax: 2.0, LowMax: 3.5, MediumMax: 5.0, HighMax: 6.5},
	"HAZARD":         {VeryLowMax: 1.4, LowMax: 2.6, MediumMax: 4.0, HighMax: 6.0},
	"VULNERABILITY":  {VeryLowMax: 1.9, LowMax: 3.2, MediumMax: 4.7, HighMax: 6.3},
	"COPING_CAPACITY":{VeryLowMax: 3.1, LowMax: 4.6, MediumMax: 5.9, HighMax: 7.3},
}

// TanzaniaThresholds (TZ_INFORM_model.xlsx Thresholds sheet, E1:G5)
var TanzaniaThresholds = map[string]ClassThresholds{
	"RISK":           {VeryLowMax: 2.5, LowMax: 3.4, MediumMax: 4.3, HighMax: 5.9},
	"HAZARD":         {VeryLowMax: 1.3, LowMax: 2.0, MediumMax: 3.3, HighMax: 4.7},
	"VULNERABILITY":  {VeryLowMax: 2.5, LowMax: 3.3, MediumMax: 4.1, HighMax: 5.0},
	"COPING_CAPACITY":{VeryLowMax: 4.1, LowMax: 5.3, MediumMax: 6.7, HighMax: 7.7},
}

// ClassifyByThresholds returns the named class given a score and a thresholds
// table. Used for per-dimension or per-category classification.
func ClassifyByThresholds(score float64, t ClassThresholds) string {
	switch {
	case score < t.VeryLowMax:
		return "Very Low"
	case score < t.LowMax:
		return "Low"
	case score < t.MediumMax:
		return "Medium"
	case score < t.HighMax:
		return "High"
	default:
		return "Very High"
	}
}

// ClassifyDimension classifies a dimension score using the given scheme
// ("TANZANIA" or "GLOBAL").
func ClassifyDimension(score float64, dimension string, scheme string) string {
	table := TanzaniaThresholds
	if scheme == "GLOBAL" {
		table = INFORMGlobalThresholds
	}
	t, ok := table[dimension]
	if !ok {
		t = table["RISK"]
	}
	return ClassifyByThresholds(score, t)
}

// LRIResult holds the components of the Lack of Reliability Index per PDF §3.6.1.
type LRIResult struct {
	Score              float64 `json:"score"`
	Classification     string  `json:"classification"`
	MissingPct         float64 `json:"missing_pct"`
	MissingScore       float64 `json:"missing_score"`
	AvgYearGap         float64 `json:"avg_year_gap"`
	StalenessScore     float64 `json:"staleness_score"`
	ConflictAggravator bool    `json:"conflict_aggravator"`
}

// CalculateLRI computes the Lack of Reliability Index (PDF §3.6.1, Figure 4).
// Three components:
//   1. Missing data — count of indicators absent (incl. estimated)
//   2. Out-of-date data — average years older than the reference year
//   3. Conflict status — HIIK level 4/5 multiplies the score by 1.3
//
// Higher LRI = less reliable.
func CalculateLRI(totalIndicators, missingCount int, yearGaps []float64, inConflict bool) LRIResult {
	if totalIndicators <= 0 {
		return LRIResult{}
	}

	missingPct := float64(missingCount) / float64(totalIndicators) * 100.0
	missingScore := missingPct / 50.0 * 10.0
	if missingScore > 10.0 {
		missingScore = 10.0
	}
	if missingScore < 0 {
		missingScore = 0
	}

	avgGap := 0.0
	if len(yearGaps) > 0 {
		sum := 0.0
		n := 0
		for _, g := range yearGaps {
			if !math.IsNaN(g) {
				sum += g
				n++
			}
		}
		if n > 0 {
			avgGap = sum / float64(n)
		}
	}
	stalenessScore := avgGap / 10.0 * 10.0
	if stalenessScore > 10.0 {
		stalenessScore = 10.0
	}

	score := (missingScore + stalenessScore) / 2.0
	if inConflict {
		score *= 1.3
		if score > 10.0 {
			score = 10.0
		}
	}

	classification := "Very High (least reliable)"
	switch {
	case score < 2.0:
		classification = "Very Low (most reliable)"
	case score < 4.0:
		classification = "Low"
	case score < 6.0:
		classification = "Medium"
	case score < 8.0:
		classification = "High"
	}

	return LRIResult{
		Score:              Round(score, 2),
		Classification:     classification,
		MissingPct:         Round(missingPct, 1),
		MissingScore:       Round(missingScore, 2),
		AvgYearGap:         Round(avgGap, 1),
		StalenessScore:     Round(stalenessScore, 2),
		ConflictAggravator: inConflict,
	}
}

// GetRiskColor returns a color code for the risk class
func GetRiskColor(riskClass string) string {
	switch riskClass {
	case "Very Low":
		return "#22c55e" // Green
	case "Low":
		return "#84cc16" // Lime
	case "Medium":
		return "#eab308" // Yellow
	case "High":
		return "#f97316" // Orange
	case "Very High":
		return "#ef4444" // Red
	default:
		return "#6b7280" // Gray
	}
}

// =====================================
// HELPER FUNCTIONS
// =====================================

// Round rounds a float to specified decimal places
func Round(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// sortFloat64s sorts a slice of float64 in ascending order
func sortFloat64s(values []float64) {
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

// percentile calculates the pth percentile of sorted values
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	k := (p / 100.0) * float64(len(sorted)-1)
	f := math.Floor(k)
	c := math.Ceil(k)

	if f == c {
		return sorted[int(k)]
	}

	return sorted[int(f)]*(c-k) + sorted[int(c)]*(k-f)
}

// GeometricMean calculates the geometric mean of positive values
func GeometricMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	product := 1.0
	count := 0

	for _, v := range values {
		if v > 0 {
			product *= v
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return math.Pow(product, 1.0/float64(count))
}

// =====================================
// INFORM CALCULATION PIPELINE
// =====================================

// INFORMCalculation holds all scores for a location
type INFORMCalculation struct {
	// Location
	ISO3     string
	ADM1Code string
	ADM2Code string
	Year     int

	// Hazard Components
	HazardNaturalComponents map[string]float64 // e.g., "Drought": 5.2, "Flood": 3.1
	HazardHumanComponents   map[string]float64 // e.g., "Conflict": 2.0

	// Vulnerability Components
	VulnerabilitySocioEconComponents map[string]float64
	VulnerabilityVulnGroupComponents map[string]float64

	// Coping Capacity Components
	CopingInfrastructureComponents map[string]float64
	CopingInstitutionalComponents  map[string]float64

	// Aggregated Scores
	HazardNatural         float64
	HazardHuman           float64
	HazardTotal           float64
	VulnerabilitySocioEcon float64
	VulnerabilityVulnGroup float64
	VulnerabilityTotal    float64
	CopingInfrastructure  float64
	CopingInstitutional   float64
	CopingCapacity        float64
	LackOfCopingCapacity  float64

	// Final Score
	RiskScore float64
	RiskClass string
}

// NewINFORMCalculation creates a new calculation instance
func NewINFORMCalculation(iso3, adm1, adm2 string, year int) *INFORMCalculation {
	return &INFORMCalculation{
		ISO3:     iso3,
		ADM1Code: adm1,
		ADM2Code: adm2,
		Year:     year,
		HazardNaturalComponents:          make(map[string]float64),
		HazardHumanComponents:            make(map[string]float64),
		VulnerabilitySocioEconComponents: make(map[string]float64),
		VulnerabilityVulnGroupComponents: make(map[string]float64),
		CopingInfrastructureComponents:   make(map[string]float64),
		CopingInstitutionalComponents:    make(map[string]float64),
	}
}

// Calculate performs the full INFORM calculation pipeline
func (c *INFORMCalculation) Calculate() {
	// Step 1: Calculate category scores (arithmetic average)
	c.HazardNatural = CalculateCategory(mapValues(c.HazardNaturalComponents))
	c.HazardHuman = CalculateCategory(mapValues(c.HazardHumanComponents))
	c.VulnerabilitySocioEcon = CalculateCategory(mapValues(c.VulnerabilitySocioEconComponents))
	c.VulnerabilityVulnGroup = CalculateCategory(mapValues(c.VulnerabilityVulnGroupComponents))
	c.CopingInfrastructure = CalculateCategory(mapValues(c.CopingInfrastructureComponents))
	c.CopingInstitutional = CalculateCategory(mapValues(c.CopingInstitutionalComponents))

	// Step 2: Calculate dimension scores (adjusted geometric mean)
	c.HazardTotal = CalculateDimension(c.HazardNatural, c.HazardHuman)
	c.VulnerabilityTotal = CalculateDimension(c.VulnerabilitySocioEcon, c.VulnerabilityVulnGroup)
	c.CopingCapacity = CalculateDimension(c.CopingInfrastructure, c.CopingInstitutional)

	// Step 3: Invert coping capacity (higher is better) to lack of coping capacity
	c.LackOfCopingCapacity = Round(10.0 - c.CopingCapacity, 1)

	// Step 4: Calculate final INFORM Risk Score
	c.RiskScore = CalculateRisk(c.HazardTotal, c.VulnerabilityTotal, c.LackOfCopingCapacity)

	// Step 5: Classify risk
	c.RiskClass = ClassifyRisk(c.RiskScore)
}

// mapValues extracts values from a map
func mapValues(m map[string]float64) []float64 {
	values := make([]float64, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// =====================================
// FORMULA DOCUMENTATION FOR TRANSPARENCY
// =====================================

// GetAllFormulas returns documentation for all INFORM formulas
func GetAllFormulas() []FormulaDoc {
	return []FormulaDoc{
		{
			Name:        "INFORM Risk Score",
			Formula:     "RISK = HAZARD^(1/3) × VULNERABILITY^(1/3) × LACK_OF_COPING_CAPACITY^(1/3)",
			Description: "The final INFORM Risk score is calculated as the cubic geometric mean of the three main dimensions. This ensures that high risk in any dimension significantly impacts the overall score.",
			Inputs:      []string{"Hazard Total (0-10)", "Vulnerability Total (0-10)", "Lack of Coping Capacity (0-10)"},
			Output:      "Risk Score (0-10)",
			Example:     "RISK = 5.2^(1/3) × 4.8^(1/3) × 6.1^(1/3) = 1.73 × 1.69 × 1.83 = 5.3",
		},
		{
			Name:        "Dimension Score (Adjusted Geometric Mean)",
			Formula:     "(10 - GEOMEAN(((10-CAT1)/10×9+1), ((10-CAT2)/10×9+1))) / 9 × 10",
			Description: "Dimension scores (Hazard, Vulnerability, Coping Capacity) are calculated using an adjusted geometric mean. This transformation ensures proper handling of the 0-10 scale and gives appropriate weight to both categories.",
			Inputs:      []string{"Category 1 Score (0-10)", "Category 2 Score (0-10)"},
			Output:      "Dimension Score (0-10)",
			Example:     "For Natural=6.5, Human=3.2: adj1=((10-6.5)/10×9+1)=4.15, adj2=((10-3.2)/10×9+1)=7.12, GEOMEAN=5.44, Result=(10-5.44)/9×10=5.1",
		},
		{
			Name:        "Category Score (Arithmetic Average)",
			Formula:     "CATEGORY = AVERAGE(Component1, Component2, ..., ComponentN)",
			Description: "Category scores are calculated as the simple arithmetic average of their component indicators. This ensures equal weighting among components within a category.",
			Inputs:      []string{"Component scores (0-10 each)"},
			Output:      "Category Score (0-10)",
			Example:     "For components [5.2, 6.1, 4.8]: AVERAGE = (5.2 + 6.1 + 4.8) / 3 = 5.4",
		},
		{
			Name:        "Min-Max Normalization",
			Formula:     "NORMALIZED = ((VALUE - MIN) / (MAX - MIN)) × 10",
			Description: "Raw indicator values are normalized to a 0-10 scale using min-max normalization. The MIN and MAX values are either theoretical bounds or empirical bounds from the data.",
			Inputs:      []string{"Raw Value", "Minimum Value", "Maximum Value"},
			Output:      "Normalized Score (0-10)",
			Example:     "For HDI=0.54 with min=0.2, max=0.95: NORMALIZED = ((0.54-0.2)/(0.95-0.2))×10 = 4.5",
		},
		{
			Name:        "Log Transformation",
			Formula:     "TRANSFORMED = LN(VALUE + 1)",
			Description: "Log transformation is applied to highly skewed data (e.g., population counts, fatalities) before normalization to reduce the impact of extreme values.",
			Inputs:      []string{"Raw Value (positive)"},
			Output:      "Transformed Value",
			Example:     "For population=500000: LN(500001) = 13.12",
		},
		{
			Name:        "Lack of Coping Capacity",
			Formula:     "LOCC = 10 - COPING_CAPACITY",
			Description: "Coping Capacity is inverted because higher capacity means lower risk. The inversion transforms it to 'Lack of Coping Capacity' for use in the risk formula.",
			Inputs:      []string{"Coping Capacity (0-10)"},
			Output:      "Lack of Coping Capacity (0-10)",
			Example:     "If Coping Capacity = 6.2, then LOCC = 10 - 6.2 = 3.8",
		},
		{
			Name:        "Outlier Detection (IQR Method)",
			Formula:     "Lower Bound = Q1 - 1.5×IQR, Upper Bound = Q3 + 1.5×IQR",
			Description: "Outliers are detected using the Interquartile Range method. Values outside the bounds are clamped to prevent extreme values from skewing normalization.",
			Inputs:      []string{"Dataset values", "Q1 (25th percentile)", "Q3 (75th percentile)"},
			Output:      "Lower and Upper bounds",
			Example:     "For Q1=2.5, Q3=7.5, IQR=5.0: Lower=2.5-7.5=-5, Upper=7.5+7.5=15",
		},
		{
			Name:        "Risk Classification",
			Formula:     "IF RISK<2 THEN 'Very Low', IF RISK<3.5 THEN 'Low', IF RISK<5 THEN 'Medium', IF RISK<6.5 THEN 'High', ELSE 'Very High'",
			Description: "The numeric risk score is classified into five risk categories based on INFORM thresholds.",
			Inputs:      []string{"Risk Score (0-10)"},
			Output:      "Risk Class (Very Low / Low / Medium / High / Very High)",
			Example:     "Risk Score 5.3 → 'High' (5.0 ≤ 5.3 < 6.5)",
		},
	}
}

// FormulaDoc represents documentation for a single formula
type FormulaDoc struct {
	Name        string   `json:"name"`
	Formula     string   `json:"formula"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
	Example     string   `json:"example"`
}

// GetDataFlow returns the data processing pipeline documentation
func GetDataFlow() []DataFlowStep {
	return []DataFlowStep{
		{
			Step:        1,
			Name:        "Raw Data Collection",
			Description: "Data is collected from various sources (field surveys, satellite imagery, statistical offices)",
			Input:       "Original measurements in various units",
			Output:      "Raw indicator values",
			Sheets:      []string{"Data Entry Forms", "Indicator Data"},
		},
		{
			Step:        2,
			Name:        "Data Validation",
			Description: "Entries are validated for completeness, format, and plausibility",
			Input:       "Raw indicator values",
			Output:      "Validated values with status flags",
			Sheets:      []string{"Indicator Data", "Validation Rules"},
		},
		{
			Step:        3,
			Name:        "Outlier Detection",
			Description: "Statistical outliers are identified using IQR method and clamped",
			Input:       "Validated values",
			Output:      "Outlier-adjusted values",
			Sheets:      []string{"Indicator Processing - 12"},
		},
		{
			Step:        4,
			Name:        "Transformation",
			Description: "Log or sqrt transformation applied to skewed distributions",
			Input:       "Outlier-adjusted values",
			Output:      "Transformed values",
			Sheets:      []string{"Indicator Processing - 12"},
		},
		{
			Step:        5,
			Name:        "Normalization",
			Description: "Values normalized to 0-10 scale using min-max normalization",
			Input:       "Transformed values",
			Output:      "Normalized scores (0-10)",
			Sheets:      []string{"Indicator Processing - 22", "Indicator - Processed"},
		},
		{
			Step:        6,
			Name:        "Component Aggregation",
			Description: "Indicators aggregated to component scores using arithmetic average",
			Input:       "Normalized scores",
			Output:      "Component scores (0-10)",
			Sheets:      []string{"INFORM SADC 2024"},
		},
		{
			Step:        7,
			Name:        "Category Aggregation",
			Description: "Components aggregated to category scores using arithmetic average",
			Input:       "Component scores",
			Output:      "Category scores (0-10)",
			Sheets:      []string{"INFORM SADC 2024"},
		},
		{
			Step:        8,
			Name:        "Dimension Calculation",
			Description: "Categories aggregated to dimensions using adjusted geometric mean",
			Input:       "Category scores",
			Output:      "Dimension scores (0-10)",
			Sheets:      []string{"INFORM SADC 2024"},
		},
		{
			Step:        9,
			Name:        "Risk Calculation",
			Description: "Final INFORM Risk calculated from three dimensions",
			Input:       "Dimension scores",
			Output:      "INFORM Risk Score (0-10)",
			Sheets:      []string{"INFORM SADC 2024"},
		},
		{
			Step:        10,
			Name:        "Classification & Visualization",
			Description: "Risk scores classified and visualized on maps and dashboards",
			Input:       "INFORM Risk Score",
			Output:      "Risk Class, Maps, Reports",
			Sheets:      []string{"Dashboard", "Maps"},
		},
	}
}

// DataFlowStep represents a step in the data processing pipeline
type DataFlowStep struct {
	Step        int      `json:"step"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Input       string   `json:"input"`
	Output      string   `json:"output"`
	Sheets      []string `json:"sheets"`
}

// GetSheetLinkages returns documentation of how sheets are linked
func GetSheetLinkages() []SheetLinkage {
	return []SheetLinkage{
		{Source: "Data Entry Forms", Target: "Indicator Data", LinkType: "data_input", Description: "User-entered data flows to indicator storage"},
		{Source: "Metadata", Target: "Indicator Data", LinkType: "reference", Description: "Indicator definitions and validation rules"},
		{Source: "Indicator Data", Target: "Indicator Processing - 12", LinkType: "data_flow", Description: "Raw values for outlier detection"},
		{Source: "Indicator Processing - 12", Target: "Indicator Processing - 22", LinkType: "data_flow", Description: "Outlier-adjusted values for transformation"},
		{Source: "Indicator Processing - 22", Target: "Indicator - Processed", LinkType: "data_flow", Description: "Transformed values for normalization"},
		{Source: "Indicator - Processed", Target: "INFORM SADC 2024", LinkType: "aggregation", Description: "Normalized values for dimension calculation"},
		{Source: "INFORM SADC 2024", Target: "Dashboard", LinkType: "visualization", Description: "Risk scores for display"},
		{Source: "INFORM SADC 2024", Target: "Maps", LinkType: "visualization", Description: "Geographic risk visualization"},
	}
}

// SheetLinkage represents a connection between two sheets/components
type SheetLinkage struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	LinkType    string `json:"link_type"`
	Description string `json:"description"`
}
