package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"inform-system/internal/database"
	"inform-system/internal/formula"
	"inform-system/internal/middleware"
	"inform-system/internal/models"
	"inform-system/internal/websocket"
)

var wsHub *websocket.Hub

func SetWebSocketHub(hub *websocket.Hub) {
	wsHub = hub
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	Phone       string `json:"phone"`
	CommitteeID *int64 `json:"committee_id"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	var user models.User
	var adm1Code sql.NullString
	err := database.DB.QueryRow(`
		SELECT u.id, u.email, u.password_hash, u.full_name, u.phone, u.role, u.committee_id, u.is_active,
		       COALESCE(c.adm1_code, '') as adm1_code
		FROM users u
		LEFT JOIN committees c ON u.committee_id = c.id
		WHERE u.email = ?
	`, req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Phone, &user.Role, &user.CommitteeID, &user.IsActive, &adm1Code)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Error: "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Error: "Account is deactivated"})
		return
	}

	if !middleware.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Error: "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.FullName, user.Role, user.CommitteeID, adm1Code.String)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to generate token"})
		return
	}

	c.SetCookie("auth_token", token, 86400, "/", "", false, false)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"token": token,
			"user": gin.H{
				"id":        user.ID,
				"email":     user.Email,
				"full_name": user.FullName,
				"role":      user.Role,
				"adm1_code": adm1Code.String,
			},
		},
	})
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	passwordHash, err := middleware.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to hash password"})
		return
	}

	role := "viewer"
	if req.CommitteeID != nil {
		role = "regional_committee"
	}

	result, err := database.DB.Exec(`
		INSERT INTO users (email, password_hash, full_name, phone, role, committee_id, is_active)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`, req.Email, passwordHash, req.FullName, req.Phone, role, req.CommitteeID)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Email already registered"})
		return
	}

	userID, _ := result.LastInsertId()

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Registration successful. Please wait for admin approval.",
		Data:    gin.H{"user_id": userID},
	})
}

func Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Logged out successfully"})
}

func GetCurrentUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, email, full_name, phone, role, committee_id, is_active, created_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Email, &user.FullName, &user.Phone, &user.Role, &user.CommitteeID, &user.IsActive, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: user})
}

func ListCommittees(c *gin.Context) {
	committeeType := c.Query("type")

	query := `SELECT id, name, type, adm1_code, adm1_name, adm2_code, adm2_name,
	          contact_person, contact_phone, contact_email, is_active, created_at
	          FROM committees WHERE is_active = 1`
	args := []interface{}{}

	if committeeType != "" {
		query += " AND type = ?"
		args = append(args, committeeType)
	}

	query += " ORDER BY adm1_name, adm2_name"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var committees []models.Committee
	for rows.Next() {
		var comm models.Committee
		rows.Scan(&comm.ID, &comm.Name, &comm.Type, &comm.ADM1Code, &comm.ADM1Name,
			&comm.ADM2Code, &comm.ADM2Name, &comm.ContactPerson, &comm.ContactPhone,
			&comm.ContactEmail, &comm.IsActive, &comm.CreatedAt)
		committees = append(committees, comm)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: committees})
}

func CreateCommittee(c *gin.Context) {
	var comm models.Committee
	if err := c.ShouldBindJSON(&comm); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	result, err := database.DB.Exec(`
		INSERT INTO committees (name, type, adm1_code, adm1_name, adm2_code, adm2_name,
		                       contact_person, contact_phone, contact_email, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
	`, comm.Name, comm.Type, comm.ADM1Code, comm.ADM1Name, comm.ADM2Code, comm.ADM2Name,
		comm.ContactPerson, comm.ContactPhone, comm.ContactEmail)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Failed to create committee"})
		return
	}

	comm.ID, _ = result.LastInsertId()
	comm.CreatedAt = time.Now()

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: comm})
}

func ListIndicators(c *gin.Context) {
	dimension := c.Query("dimension")
	resolution := c.Query("resolution")

	query := `SELECT id, code, name, description, dimension, category, component,
	          unit, data_type, min_value, max_value, resolution, transformation,
	          invert_scale, weight, data_source, update_frequency, is_active
	          FROM indicators WHERE is_active = 1`
	args := []interface{}{}

	if dimension != "" {
		query += " AND dimension = ?"
		args = append(args, dimension)
	}

	if resolution != "" {
		query += " AND resolution = ?"
		args = append(args, resolution)
	}

	query += " ORDER BY dimension, category, component, code"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var indicators []models.Indicator
	for rows.Next() {
		var ind models.Indicator
		rows.Scan(&ind.ID, &ind.Code, &ind.Name, &ind.Description, &ind.Dimension,
			&ind.Category, &ind.Component, &ind.Unit, &ind.DataType, &ind.MinValue,
			&ind.MaxValue, &ind.Resolution, &ind.Transformation, &ind.InvertScale,
			&ind.Weight, &ind.DataSource, &ind.UpdateFrequency, &ind.IsActive)
		indicators = append(indicators, ind)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: indicators})
}

func GetIndicator(c *gin.Context) {
	idOrCode := c.Param("id")

	var ind models.Indicator
	var err error

	if id, parseErr := strconv.ParseInt(idOrCode, 10, 64); parseErr == nil {
		err = database.DB.QueryRow(`
			SELECT id, code, name, description, dimension, category, component,
			       unit, data_type, min_value, max_value, resolution, transformation,
			       invert_scale, weight, data_source, update_frequency, is_active
			FROM indicators WHERE id = ?
		`, id).Scan(&ind.ID, &ind.Code, &ind.Name, &ind.Description, &ind.Dimension,
			&ind.Category, &ind.Component, &ind.Unit, &ind.DataType, &ind.MinValue,
			&ind.MaxValue, &ind.Resolution, &ind.Transformation, &ind.InvertScale,
			&ind.Weight, &ind.DataSource, &ind.UpdateFrequency, &ind.IsActive)
	} else {
		err = database.DB.QueryRow(`
			SELECT id, code, name, description, dimension, category, component,
			       unit, data_type, min_value, max_value, resolution, transformation,
			       invert_scale, weight, data_source, update_frequency, is_active
			FROM indicators WHERE code = ?
		`, idOrCode).Scan(&ind.ID, &ind.Code, &ind.Name, &ind.Description, &ind.Dimension,
			&ind.Category, &ind.Component, &ind.Unit, &ind.DataType, &ind.MinValue,
			&ind.MaxValue, &ind.Resolution, &ind.Transformation, &ind.InvertScale,
			&ind.Weight, &ind.DataSource, &ind.UpdateFrequency, &ind.IsActive)
	}

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Error: "Indicator not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: ind})
}

type DataEntryRequest struct {
	IndicatorCode string  `json:"indicator_code" binding:"required"`
	Country       string  `json:"country" binding:"required"`
	ISO3          string  `json:"iso3" binding:"required"`
	ADM1Code      *string `json:"adm1_code"`
	ADM1Name      *string `json:"adm1_name"`
	ADM2Code      *string `json:"adm2_code"`
	ADM2Name      *string `json:"adm2_name"`
	RawValue      float64 `json:"raw_value" binding:"required"`
	Year          int     `json:"year" binding:"required"`
	Quarter       *int    `json:"quarter"`
	DataSource    string  `json:"data_source"`
	Notes         string  `json:"notes"`
}

func CreateDataEntry(c *gin.Context) {
	var req DataEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	userName, _ := c.Get("user_name")

	var indicatorID int64
	var minVal, maxVal float64
	var transformation string
	var invertScale bool

	err := database.DB.QueryRow(`
		SELECT id, min_value, max_value, transformation, invert_scale
		FROM indicators WHERE code = ? AND is_active = 1
	`, req.IndicatorCode).Scan(&indicatorID, &minVal, &maxVal, &transformation, &invertScale)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid indicator code"})
		return
	}

	var normalizedValue float64
	switch transformation {
	case "log":
		normalizedValue = formula.NormalizeWithLog(req.RawValue, minVal, maxVal, invertScale)
	case "sqrt":
		normalizedValue = formula.NormalizeWithSqrt(req.RawValue, minVal, maxVal, invertScale)
	default:
		normalizedValue = formula.NormalizeMinMax(req.RawValue, minVal, maxVal, invertScale)
	}

	result, err := database.DB.Exec(`
		INSERT INTO data_entries (indicator_id, country, iso3, adm1_code, adm1_name, adm2_code, adm2_name,
		                         raw_value, normalized_value, year, quarter, data_source, notes,
		                         entered_by_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'submitted')
	`, indicatorID, req.Country, req.ISO3, req.ADM1Code, req.ADM1Name, req.ADM2Code, req.ADM2Name,
		req.RawValue, normalizedValue, req.Year, req.Quarter, req.DataSource, req.Notes, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to save data entry"})
		return
	}

	entryID, _ := result.LastInsertId()

	userNameStr := ""
	if userName != nil {
		userNameStr = userName.(string)
	}

	entry := &models.DataEntry{
		ID:              entryID,
		IndicatorID:     indicatorID,
		IndicatorCode:   req.IndicatorCode,
		Country:         req.Country,
		ISO3:            req.ISO3,
		ADM1Code:        req.ADM1Code,
		ADM1Name:        req.ADM1Name,
		ADM2Code:        req.ADM2Code,
		ADM2Name:        req.ADM2Name,
		RawValue:        req.RawValue,
		NormalizedValue: &normalizedValue,
		Year:            req.Year,
		Quarter:         req.Quarter,
		DataSource:      req.DataSource,
		Notes:           req.Notes,
		EnteredByID:     userID,
		EnteredBy:       userNameStr,
		Status:          "submitted",
		CreatedAt:       time.Now(),
	}

	if wsHub != nil {
		wsHub.BroadcastDataEntry(entry, "created")
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: entry})
}

func ListDataEntries(c *gin.Context) {
	indicatorCode := c.Query("indicator_code")
	iso3 := c.Query("iso3")
	adm1Code := c.Query("adm1_code")
	year := c.Query("year")
	status := c.Query("status")

	query := `SELECT de.id, de.indicator_id, i.code, de.country, de.iso3, de.adm1_code, de.adm1_name,
	          de.adm2_code, de.adm2_name, de.raw_value, de.normalized_value, de.year, de.quarter,
	          de.data_source, de.notes, de.entered_by_id, u.full_name, de.status, de.created_at, de.updated_at
	          FROM data_entries de
	          JOIN indicators i ON de.indicator_id = i.id
	          JOIN users u ON de.entered_by_id = u.id
	          WHERE 1=1`
	args := []interface{}{}

	if indicatorCode != "" {
		query += " AND i.code = ?"
		args = append(args, indicatorCode)
	}
	if iso3 != "" {
		query += " AND de.iso3 = ?"
		args = append(args, iso3)
	}
	if adm1Code != "" {
		query += " AND de.adm1_code = ?"
		args = append(args, adm1Code)
	}
	if year != "" {
		query += " AND de.year = ?"
		args = append(args, year)
	}
	if status != "" {
		query += " AND de.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY de.updated_at DESC LIMIT 500"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var entries []models.DataEntry
	for rows.Next() {
		var entry models.DataEntry
		rows.Scan(&entry.ID, &entry.IndicatorID, &entry.IndicatorCode, &entry.Country, &entry.ISO3,
			&entry.ADM1Code, &entry.ADM1Name, &entry.ADM2Code, &entry.ADM2Name, &entry.RawValue,
			&entry.NormalizedValue, &entry.Year, &entry.Quarter, &entry.DataSource, &entry.Notes,
			&entry.EnteredByID, &entry.EnteredBy, &entry.Status, &entry.CreatedAt, &entry.UpdatedAt)
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: entries})
}

func VerifyDataEntry(c *gin.Context) {
	entryID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid entry ID"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	status := "verified"
	if req.Action == "reject" {
		status = "rejected"
	}

	database.DB.Exec(`
		UPDATE data_entries SET status = ?, verified_by_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, userID, entryID)

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Data entry " + status})
}

func CalculateRiskScores(c *gin.Context) {
	iso3 := c.Query("iso3")
	year := c.Query("year")

	if iso3 == "" || year == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "iso3 and year are required"})
		return
	}

	yearInt, _ := strconv.Atoi(year)

	rows, err := database.DB.Query(`
		SELECT i.code, i.dimension, i.category, i.component, de.normalized_value,
		       de.adm1_code, de.adm1_name, de.adm2_code, de.adm2_name
		FROM data_entries de
		JOIN indicators i ON de.indicator_id = i.id
		WHERE de.iso3 = ? AND de.year = ? AND de.status = 'verified'
		ORDER BY de.adm1_code, de.adm2_code
	`, iso3, yearInt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	locationData := make(map[string]*formula.INFORMCalculation)

	for rows.Next() {
		var code, dimension, category, component string
		var normalizedValue sql.NullFloat64
		var adm1Code, adm1Name, adm2Code, adm2Name sql.NullString

		rows.Scan(&code, &dimension, &category, &component, &normalizedValue,
			&adm1Code, &adm1Name, &adm2Code, &adm2Name)

		if !normalizedValue.Valid {
			continue
		}

		locKey := iso3
		if adm1Code.Valid && adm1Code.String != "" {
			locKey += "_" + adm1Code.String
		}
		if adm2Code.Valid && adm2Code.String != "" {
			locKey += "_" + adm2Code.String
		}

		if _, ok := locationData[locKey]; !ok {
			calc := formula.NewINFORMCalculation(iso3, adm1Code.String, adm2Code.String, yearInt)
			locationData[locKey] = calc
		}

		calc := locationData[locKey]
		value := normalizedValue.Float64

		switch dimension {
		case "HAZARD":
			if category == "Natural" {
				calc.HazardNaturalComponents[component] = value
			} else {
				calc.HazardHumanComponents[component] = value
			}
		case "VULNERABILITY":
			if category == "Socio-Economic" {
				calc.VulnerabilitySocioEconComponents[component] = value
			} else {
				calc.VulnerabilityVulnGroupComponents[component] = value
			}
		case "COPING_CAPACITY":
			if category == "Infrastructure" {
				calc.CopingInfrastructureComponents[component] = value
			} else {
				calc.CopingInstitutionalComponents[component] = value
			}
		}
	}

	var results []models.RiskScore
	for _, calc := range locationData {
		calc.Calculate()

		resolution := "national"
		if calc.ADM1Code != "" {
			resolution = "adm1"
		}
		if calc.ADM2Code != "" {
			resolution = "adm2"
		}

		score := models.RiskScore{
			Country:               "United Republic of Tanzania",
			ISO3:                  calc.ISO3,
			ADM1Code:              stringPtr(calc.ADM1Code),
			ADM1Name:              stringPtr(calc.ADM1Code),
			ADM2Code:              stringPtr(calc.ADM2Code),
			ADM2Name:              stringPtr(calc.ADM2Code),
			Resolution:            resolution,
			Year:                  calc.Year,
			HazardNatural:         calc.HazardNatural,
			HazardHuman:           calc.HazardHuman,
			HazardTotal:           calc.HazardTotal,
			VulnerabilitySocioEcon: calc.VulnerabilitySocioEcon,
			VulnerabilityVulnGroup: calc.VulnerabilityVulnGroup,
			VulnerabilityTotal:    calc.VulnerabilityTotal,
			CopingInfrastructure:  calc.CopingInfrastructure,
			CopingInstitutional:   calc.CopingInstitutional,
			LackOfCopingCapacity:  calc.LackOfCopingCapacity,
			RiskScore:             calc.RiskScore,
			RiskClass:             calc.RiskClass,
			CalculatedAt:          time.Now(),
		}

		database.DB.Exec(`
			INSERT OR REPLACE INTO risk_scores (country, iso3, adm1_code, adm1_name, adm2_code, adm2_name,
			                        resolution, year, hazard_natural, hazard_human, hazard_total,
			                        vulnerability_socio_econ, vulnerability_vuln_group, vulnerability_total,
			                        coping_infrastructure, coping_institutional, lack_of_coping_capacity,
			                        risk_score, risk_class, calculated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, score.Country, score.ISO3, score.ADM1Code, score.ADM1Name, score.ADM2Code, score.ADM2Name,
			score.Resolution, score.Year, score.HazardNatural, score.HazardHuman, score.HazardTotal,
			score.VulnerabilitySocioEcon, score.VulnerabilityVulnGroup, score.VulnerabilityTotal,
			score.CopingInfrastructure, score.CopingInstitutional, score.LackOfCopingCapacity,
			score.RiskScore, score.RiskClass, score.CalculatedAt)

		results = append(results, score)

		if wsHub != nil {
			wsHub.BroadcastRiskUpdate(&score, "calculated")
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Risk scores calculated",
		Data:    results,
	})
}

func GetRiskScores(c *gin.Context) {
	iso3 := c.Query("iso3")
	adm1Code := c.Query("adm1_code")
	year := c.Query("year")
	resolution := c.Query("resolution")

	query := `SELECT id, country, iso3, adm1_code, adm1_name, adm2_code, adm2_name,
	          resolution, year, hazard_natural, hazard_human, hazard_total,
	          vulnerability_socio_econ, vulnerability_vuln_group, vulnerability_total,
	          coping_infrastructure, coping_institutional, lack_of_coping_capacity,
	          risk_score, risk_class, calculated_at
	          FROM risk_scores WHERE 1=1`
	args := []interface{}{}

	if iso3 != "" {
		query += " AND iso3 = ?"
		args = append(args, iso3)
	}
	if adm1Code != "" {
		query += " AND adm1_code = ?"
		args = append(args, adm1Code)
	}
	if year != "" {
		query += " AND year = ?"
		args = append(args, year)
	}
	if resolution != "" {
		query += " AND resolution = ?"
		args = append(args, resolution)
	}

	query += " ORDER BY risk_score DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var scores []models.RiskScore
	for rows.Next() {
		var score models.RiskScore
		rows.Scan(&score.ID, &score.Country, &score.ISO3, &score.ADM1Code, &score.ADM1Name,
			&score.ADM2Code, &score.ADM2Name, &score.Resolution, &score.Year,
			&score.HazardNatural, &score.HazardHuman, &score.HazardTotal,
			&score.VulnerabilitySocioEcon, &score.VulnerabilityVulnGroup, &score.VulnerabilityTotal,
			&score.CopingInfrastructure, &score.CopingInstitutional, &score.LackOfCopingCapacity,
			&score.RiskScore, &score.RiskClass, &score.CalculatedAt)
		scores = append(scores, score)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: scores})
}

func GetFormulas(c *gin.Context) {
	formulas := formula.GetAllFormulas()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: formulas})
}

func GetDataFlow(c *gin.Context) {
	dataFlow := formula.GetDataFlow()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: dataFlow})
}

func GetSheetLinkages(c *gin.Context) {
	linkages := formula.GetSheetLinkages()
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: linkages})
}

func GetAPIDocumentation(c *gin.Context) {
	docs := map[string]interface{}{
		"version": "1.0.0",
		"title":   "INFORM Risk Management System API",
		"base_url": "/api/v1",
		"endpoints": []map[string]interface{}{
			{"method": "POST", "path": "/auth/login", "description": "User login", "auth": false},
			{"method": "POST", "path": "/auth/register", "description": "User registration", "auth": false},
			{"method": "GET", "path": "/users/me", "description": "Get current user", "auth": true},
			{"method": "GET", "path": "/committees", "description": "List committees", "auth": true},
			{"method": "GET", "path": "/indicators", "description": "List indicators", "auth": true},
			{"method": "POST", "path": "/data", "description": "Create data entry", "auth": true},
			{"method": "GET", "path": "/data", "description": "List data entries", "auth": true},
			{"method": "GET", "path": "/risk/calculate", "description": "Calculate risk scores", "auth": true},
			{"method": "GET", "path": "/risk/scores", "description": "Get risk scores", "auth": true},
			{"method": "GET", "path": "/transparency/formulas", "description": "Get all formulas", "auth": false},
			{"method": "GET", "path": "/transparency/dataflow", "description": "Get data flow", "auth": false},
			{"method": "GET", "path": "/transparency/linkages", "description": "Get sheet linkages", "auth": false},
		},
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: docs})
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ============================================================================
// MODULE 03: EARLY WARNING SYSTEM HANDLERS
// ============================================================================

type HazardForecastRequest struct {
	HazardType      string  `json:"hazard_type" binding:"required"`
	Institution     string  `json:"institution" binding:"required"`
	InstitutionName string  `json:"institution_name"`
	IntensityLevel  string  `json:"intensity_level" binding:"required"`
	IntensityValue  float64 `json:"intensity_value"`
	IntensityUnit   string  `json:"intensity_unit"`
	Confidence      string  `json:"confidence"`
	SpatialExtent   string  `json:"spatial_extent" binding:"required"` // JSON array
	ValidFrom       string  `json:"valid_from" binding:"required"`
	ValidTo         string  `json:"valid_to" binding:"required"`
	ForecastDay     int     `json:"forecast_day" binding:"required"`
	Description     string  `json:"description"`
	DataSource      string  `json:"data_source"`
}

// CreateHazardForecast creates a new hazard forecast from an institution
func CreateHazardForecast(c *gin.Context) {
	var req HazardForecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	userName, _ := c.Get("user_name")
	userNameStr := ""
	if userName != nil {
		userNameStr = userName.(string)
	}

	validFrom, _ := time.Parse(time.RFC3339, req.ValidFrom)
	validTo, _ := time.Parse(time.RFC3339, req.ValidTo)

	result, err := database.DB.Exec(`
		INSERT INTO hazard_forecasts (hazard_type, institution, institution_name, intensity_level,
		                              intensity_value, intensity_unit, confidence, spatial_extent,
		                              valid_from, valid_to, forecast_day, description, data_source,
		                              issued_by_id, issued_by_name, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'submitted')
	`, req.HazardType, req.Institution, req.InstitutionName, req.IntensityLevel,
		req.IntensityValue, req.IntensityUnit, req.Confidence, req.SpatialExtent,
		validFrom, validTo, req.ForecastDay, req.Description, req.DataSource,
		userID, userNameStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to save hazard forecast"})
		return
	}

	forecastID, _ := result.LastInsertId()

	forecast := &models.HazardForecast{
		ID:              forecastID,
		HazardType:      req.HazardType,
		Institution:     req.Institution,
		InstitutionName: req.InstitutionName,
		IntensityLevel:  req.IntensityLevel,
		IntensityValue:  req.IntensityValue,
		IntensityUnit:   req.IntensityUnit,
		Confidence:      req.Confidence,
		SpatialExtent:   req.SpatialExtent,
		ValidFrom:       validFrom,
		ValidTo:         validTo,
		ForecastDay:     req.ForecastDay,
		Description:     req.Description,
		DataSource:      req.DataSource,
		IssuedByID:      userID,
		IssuedByName:    userNameStr,
		Status:          "submitted",
		CreatedAt:       time.Now(),
	}

	if wsHub != nil {
		wsHub.BroadcastMessage("hazard_forecast", "created", forecast)
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: forecast})
}

// ListHazardForecasts lists hazard forecasts with filters
func ListHazardForecasts(c *gin.Context) {
	institution := c.Query("institution")
	hazardType := c.Query("hazard_type")
	forecastDay := c.Query("forecast_day")
	status := c.Query("status")

	query := `SELECT id, hazard_type, institution, institution_name, intensity_level,
	          intensity_value, intensity_unit, confidence, spatial_extent,
	          valid_from, valid_to, forecast_day, description, data_source,
	          issued_by_id, issued_by_name, status, created_at, updated_at
	          FROM hazard_forecasts WHERE 1=1`
	args := []interface{}{}

	if institution != "" {
		query += " AND institution = ?"
		args = append(args, institution)
	}
	if hazardType != "" {
		query += " AND hazard_type = ?"
		args = append(args, hazardType)
	}
	if forecastDay != "" {
		query += " AND forecast_day = ?"
		args = append(args, forecastDay)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var forecasts []models.HazardForecast
	for rows.Next() {
		var f models.HazardForecast
		rows.Scan(&f.ID, &f.HazardType, &f.Institution, &f.InstitutionName, &f.IntensityLevel,
			&f.IntensityValue, &f.IntensityUnit, &f.Confidence, &f.SpatialExtent,
			&f.ValidFrom, &f.ValidTo, &f.ForecastDay, &f.Description, &f.DataSource,
			&f.IssuedByID, &f.IssuedByName, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		forecasts = append(forecasts, f)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: forecasts})
}

type WarningRequest struct {
	HazardForecastID    int64   `json:"hazard_forecast_id" binding:"required"`
	ADM1Code            string  `json:"adm1_code" binding:"required"`
	ADM1Name            string  `json:"adm1_name" binding:"required"`
	ADM2Code            *string `json:"adm2_code"`
	ADM2Name            *string `json:"adm2_name"`
	HazardIntensity     float64 `json:"hazard_intensity"`
	VulnerabilityScore  float64 `json:"vulnerability_score"`
	CopingCapacityScore float64 `json:"coping_capacity_score"`
	BaselineRiskScore   float64 `json:"baseline_risk_score"`
	BaselineRiskClass   string  `json:"baseline_risk_class"`
	PopulationAffected  int64   `json:"population_affected"`
	VulnerablePopulation int64  `json:"vulnerable_population"`
	InfrastructureAtRisk string `json:"infrastructure_at_risk"`
	RecommendedActions  string  `json:"recommended_actions"`
	ExpiresAt           string  `json:"expires_at"`
}

// CreateWarning creates a risk-informed warning
func CreateWarning(c *gin.Context) {
	var req WarningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	// Get hazard forecast info
	var hazardType string
	err := database.DB.QueryRow(`SELECT hazard_type FROM hazard_forecasts WHERE id = ?`,
		req.HazardForecastID).Scan(&hazardType)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid hazard forecast ID"})
		return
	}

	// Calculate risk sensitivity and warning score
	riskSensitivity := formula.GeometricMean([]float64{req.VulnerabilityScore, req.CopingCapacityScore})
	warningScore := formula.GeometricMean([]float64{req.HazardIntensity, riskSensitivity})

	// Classify warning level
	warningLevel, warningColor := classifyWarningLevel(warningScore)

	// Determine response level
	responseLevel := determineResponseLevel(warningLevel)

	// Generate warning ID
	warningID := generateWarningID(hazardType)

	expiresAt, _ := time.Parse(time.RFC3339, req.ExpiresAt)

	result, err := database.DB.Exec(`
		INSERT INTO warnings (warning_id, hazard_forecast_id, hazard_type, adm1_code, adm1_name,
		                     adm2_code, adm2_name, hazard_intensity, vulnerability_score,
		                     coping_capacity_score, baseline_risk_score, baseline_risk_class,
		                     risk_sensitivity, warning_score, warning_level, warning_color,
		                     population_affected, vulnerable_population, infrastructure_at_risk,
		                     recommended_actions, response_level, status, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)
	`, warningID, req.HazardForecastID, hazardType, req.ADM1Code, req.ADM1Name,
		req.ADM2Code, req.ADM2Name, req.HazardIntensity, req.VulnerabilityScore,
		req.CopingCapacityScore, req.BaselineRiskScore, req.BaselineRiskClass,
		riskSensitivity, warningScore, warningLevel, warningColor,
		req.PopulationAffected, req.VulnerablePopulation, req.InfrastructureAtRisk,
		req.RecommendedActions, responseLevel, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Failed to save warning"})
		return
	}

	id, _ := result.LastInsertId()

	warning := &models.Warning{
		ID:                   id,
		WarningID:            warningID,
		HazardForecastID:     req.HazardForecastID,
		HazardType:           hazardType,
		ADM1Code:             req.ADM1Code,
		ADM1Name:             req.ADM1Name,
		ADM2Code:             req.ADM2Code,
		ADM2Name:             req.ADM2Name,
		HazardIntensity:      req.HazardIntensity,
		VulnerabilityScore:   req.VulnerabilityScore,
		CopingCapacityScore:  req.CopingCapacityScore,
		BaselineRiskScore:    req.BaselineRiskScore,
		BaselineRiskClass:    req.BaselineRiskClass,
		RiskSensitivity:      riskSensitivity,
		WarningScore:         warningScore,
		WarningLevel:         warningLevel,
		WarningColor:         warningColor,
		PopulationAffected:   req.PopulationAffected,
		VulnerablePopulation: req.VulnerablePopulation,
		InfrastructureAtRisk: req.InfrastructureAtRisk,
		RecommendedActions:   req.RecommendedActions,
		ResponseLevel:        responseLevel,
		Status:               "pending",
		ExpiresAt:            expiresAt,
		CreatedAt:            time.Now(),
	}

	if wsHub != nil {
		wsHub.BroadcastMessage("warning", "created", warning)
	}

	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Data: warning})
}

// ListWarnings lists warnings with filters
func ListWarnings(c *gin.Context) {
	adm1Code := c.Query("adm1_code")
	warningLevel := c.Query("warning_level")
	status := c.Query("status")
	hazardType := c.Query("hazard_type")

	query := `SELECT id, warning_id, hazard_forecast_id, hazard_type, adm1_code, adm1_name,
	          adm2_code, adm2_name, hazard_intensity, vulnerability_score, coping_capacity_score,
	          baseline_risk_score, baseline_risk_class, risk_sensitivity, warning_score,
	          warning_level, warning_color, population_affected, vulnerable_population,
	          infrastructure_at_risk, recommended_actions, response_level,
	          validated_by_id, validated_by_name, validation_notes, validated_at,
	          status, disseminated_at, expires_at, created_at, updated_at
	          FROM warnings WHERE 1=1`
	args := []interface{}{}

	if adm1Code != "" {
		query += " AND adm1_code = ?"
		args = append(args, adm1Code)
	}
	if warningLevel != "" {
		query += " AND warning_level = ?"
		args = append(args, warningLevel)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if hazardType != "" {
		query += " AND hazard_type = ?"
		args = append(args, hazardType)
	}

	query += " ORDER BY warning_score DESC, created_at DESC LIMIT 100"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var warnings []models.Warning
	for rows.Next() {
		var w models.Warning
		rows.Scan(&w.ID, &w.WarningID, &w.HazardForecastID, &w.HazardType, &w.ADM1Code, &w.ADM1Name,
			&w.ADM2Code, &w.ADM2Name, &w.HazardIntensity, &w.VulnerabilityScore, &w.CopingCapacityScore,
			&w.BaselineRiskScore, &w.BaselineRiskClass, &w.RiskSensitivity, &w.WarningScore,
			&w.WarningLevel, &w.WarningColor, &w.PopulationAffected, &w.VulnerablePopulation,
			&w.InfrastructureAtRisk, &w.RecommendedActions, &w.ResponseLevel,
			&w.ValidatedByID, &w.ValidatedByName, &w.ValidationNotes, &w.ValidatedAt,
			&w.Status, &w.DisseminatedAt, &w.ExpiresAt, &w.CreatedAt, &w.UpdatedAt)
		warnings = append(warnings, w)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: warnings})
}

// ValidateWarning validates a warning (PMO action)
func ValidateWarning(c *gin.Context) {
	warningID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "Invalid warning ID"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // validate, reject
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	userName, _ := c.Get("user_name")
	userNameStr := ""
	if userName != nil {
		userNameStr = userName.(string)
	}

	status := "validated"
	if req.Action == "reject" {
		status = "rejected"
	}

	database.DB.Exec(`
		UPDATE warnings SET status = ?, validated_by_id = ?, validated_by_name = ?,
		                   validation_notes = ?, validated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, userID, userNameStr, req.Notes, warningID)

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Warning " + status})
}

// GetActiveWarnings gets all current active warnings
func GetActiveWarnings(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT id, warning_id, hazard_forecast_id, hazard_type, adm1_code, adm1_name,
		       adm2_code, adm2_name, warning_score, warning_level, warning_color,
		       population_affected, response_level, status, expires_at
		FROM warnings
		WHERE status IN ('pending', 'validated', 'disseminated')
		  AND expires_at > CURRENT_TIMESTAMP
		ORDER BY warning_score DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Database error"})
		return
	}
	defer rows.Close()

	var warnings []gin.H
	for rows.Next() {
		var w struct {
			ID, HazardForecastID, PopulationAffected int64
			WarningID, HazardType, ADM1Code, ADM1Name string
			ADM2Code, ADM2Name                       sql.NullString
			WarningScore                              float64
			WarningLevel, WarningColor, ResponseLevel, Status string
			ExpiresAt                                 time.Time
		}
		rows.Scan(&w.ID, &w.WarningID, &w.HazardForecastID, &w.HazardType, &w.ADM1Code, &w.ADM1Name,
			&w.ADM2Code, &w.ADM2Name, &w.WarningScore, &w.WarningLevel, &w.WarningColor,
			&w.PopulationAffected, &w.ResponseLevel, &w.Status, &w.ExpiresAt)

		warning := gin.H{
			"id":                  w.ID,
			"warning_id":          w.WarningID,
			"hazard_type":         w.HazardType,
			"adm1_code":           w.ADM1Code,
			"adm1_name":           w.ADM1Name,
			"warning_score":       w.WarningScore,
			"warning_level":       w.WarningLevel,
			"warning_color":       w.WarningColor,
			"population_affected": w.PopulationAffected,
			"response_level":      w.ResponseLevel,
			"status":              w.Status,
			"expires_at":          w.ExpiresAt,
		}
		if w.ADM2Code.Valid {
			warning["adm2_code"] = w.ADM2Code.String
		}
		if w.ADM2Name.Valid {
			warning["adm2_name"] = w.ADM2Name.String
		}
		warnings = append(warnings, warning)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: warnings})
}

// Helper functions for warning classification

func classifyWarningLevel(score float64) (string, string) {
	switch {
	case score < 2.5:
		return "monitor", "#4CAF50"      // Green
	case score < 5.0:
		return "advisory", "#FFEB3B"     // Yellow
	case score < 7.5:
		return "warning", "#FF9800"      // Orange
	case score < 9.0:
		return "major_warning", "#F44336" // Red
	default:
		return "emergency", "#9C27B0"    // Purple
	}
}

func determineResponseLevel(warningLevel string) string {
	switch warningLevel {
	case "monitor":
		return "none"
	case "advisory":
		return "preparedness"
	case "warning":
		return "partial"
	case "major_warning", "emergency":
		return "full"
	default:
		return "none"
	}
}

func generateWarningID(hazardType string) string {
	now := time.Now()
	return "TZA-" + now.Format("2006-01-02") + "-" + hazardType[:3] + "-" + strconv.FormatInt(now.UnixNano()%10000, 10)
}
