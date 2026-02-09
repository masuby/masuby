package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect(databaseURL string) error {
	var err error

	// Use SQLite database file in the project directory
	dbPath := filepath.Join(".", "inform.db")
	if databaseURL != "" && databaseURL != "sqlite" {
		dbPath = databaseURL
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Enable foreign keys
	DB.Exec("PRAGMA foreign_keys = ON")

	log.Printf("Database connected: %s", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func InitSchema() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			phone TEXT,
			role TEXT NOT NULL DEFAULT 'viewer',
			committee_id INTEGER,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS committees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			adm1_code TEXT NOT NULL,
			adm1_name TEXT NOT NULL,
			adm2_code TEXT,
			adm2_name TEXT,
			contact_person TEXT,
			contact_phone TEXT,
			contact_email TEXT,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS indicators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			dimension TEXT NOT NULL,
			category TEXT NOT NULL,
			component TEXT,
			unit TEXT,
			data_type TEXT NOT NULL,
			min_value REAL DEFAULT 0,
			max_value REAL DEFAULT 10,
			resolution TEXT NOT NULL,
			transformation TEXT DEFAULT 'none',
			invert_scale INTEGER DEFAULT 0,
			weight REAL DEFAULT 1.0,
			data_source TEXT,
			update_frequency TEXT,
			is_active INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS data_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			indicator_id INTEGER NOT NULL,
			country TEXT NOT NULL,
			iso3 TEXT NOT NULL,
			adm1_code TEXT,
			adm1_name TEXT,
			adm2_code TEXT,
			adm2_name TEXT,
			raw_value REAL NOT NULL,
			normalized_value REAL,
			year INTEGER NOT NULL,
			quarter INTEGER,
			data_source TEXT,
			notes TEXT,
			entered_by_id INTEGER NOT NULL,
			verified_by_id INTEGER,
			status TEXT DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (indicator_id) REFERENCES indicators(id),
			FOREIGN KEY (entered_by_id) REFERENCES users(id),
			UNIQUE(indicator_id, iso3, adm1_code, adm2_code, year, quarter)
		)`,
		`CREATE TABLE IF NOT EXISTS risk_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			country TEXT NOT NULL,
			iso3 TEXT NOT NULL,
			adm1_code TEXT,
			adm1_name TEXT,
			adm2_code TEXT,
			adm2_name TEXT,
			resolution TEXT NOT NULL,
			year INTEGER NOT NULL,
			hazard_natural REAL,
			hazard_human REAL,
			hazard_total REAL,
			vulnerability_socio_econ REAL,
			vulnerability_vuln_group REAL,
			vulnerability_total REAL,
			coping_infrastructure REAL,
			coping_institutional REAL,
			lack_of_coping_capacity REAL,
			risk_score REAL,
			risk_class TEXT,
			calculated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(iso3, adm1_code, adm2_code, year)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			user_email TEXT,
			action TEXT NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			old_value TEXT,
			new_value TEXT,
			ip_address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_indicator ON data_entries(indicator_id)`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_location ON data_entries(iso3, adm1_code, adm2_code)`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_year ON data_entries(year)`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_status ON data_entries(status)`,
		`CREATE INDEX IF NOT EXISTS idx_risk_scores_location ON risk_scores(iso3, adm1_code, adm2_code)`,
		`CREATE INDEX IF NOT EXISTS idx_risk_scores_year ON risk_scores(year)`,

		// Module 03: Early Warning System Tables
		`CREATE TABLE IF NOT EXISTS hazard_forecasts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hazard_type TEXT NOT NULL,
			institution TEXT NOT NULL,
			institution_name TEXT,
			intensity_level TEXT NOT NULL,
			intensity_value REAL,
			intensity_unit TEXT,
			confidence TEXT DEFAULT 'medium',
			spatial_extent TEXT NOT NULL,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			forecast_day INTEGER NOT NULL,
			description TEXT,
			data_source TEXT,
			issued_by_id INTEGER NOT NULL,
			issued_by_name TEXT,
			status TEXT DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (issued_by_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS warnings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			warning_id TEXT UNIQUE NOT NULL,
			hazard_forecast_id INTEGER NOT NULL,
			hazard_type TEXT NOT NULL,
			adm1_code TEXT NOT NULL,
			adm1_name TEXT NOT NULL,
			adm2_code TEXT,
			adm2_name TEXT,
			hazard_intensity REAL NOT NULL,
			vulnerability_score REAL NOT NULL,
			coping_capacity_score REAL NOT NULL,
			baseline_risk_score REAL NOT NULL,
			baseline_risk_class TEXT NOT NULL,
			risk_sensitivity REAL NOT NULL,
			warning_score REAL NOT NULL,
			warning_level TEXT NOT NULL,
			warning_color TEXT NOT NULL,
			population_affected INTEGER DEFAULT 0,
			vulnerable_population INTEGER DEFAULT 0,
			infrastructure_at_risk TEXT,
			recommended_actions TEXT,
			response_level TEXT NOT NULL,
			validated_by_id INTEGER,
			validated_by_name TEXT,
			validation_notes TEXT,
			validated_at DATETIME,
			status TEXT DEFAULT 'pending',
			disseminated_at DATETIME,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (hazard_forecast_id) REFERENCES hazard_forecasts(id),
			FOREIGN KEY (validated_by_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS warning_disseminations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			warning_id INTEGER NOT NULL,
			audience TEXT NOT NULL,
			channel TEXT NOT NULL,
			message TEXT NOT NULL,
			recipients TEXT,
			sent_at DATETIME,
			delivery_status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (warning_id) REFERENCES warnings(id)
		)`,
		`CREATE TABLE IF NOT EXISTS warning_feedback (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			warning_id INTEGER NOT NULL,
			actual_impacts TEXT,
			population_actual INTEGER,
			casualties_actual INTEGER,
			economic_loss_actual REAL,
			response_time_hours REAL,
			warning_effectiveness TEXT,
			lessons_learned TEXT,
			threshold_adjustments TEXT,
			submitted_by_id INTEGER NOT NULL,
			submitted_by_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (warning_id) REFERENCES warnings(id),
			FOREIGN KEY (submitted_by_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_hazard_forecasts_institution ON hazard_forecasts(institution)`,
		`CREATE INDEX IF NOT EXISTS idx_hazard_forecasts_type ON hazard_forecasts(hazard_type)`,
		`CREATE INDEX IF NOT EXISTS idx_hazard_forecasts_day ON hazard_forecasts(forecast_day)`,
		`CREATE INDEX IF NOT EXISTS idx_warnings_level ON warnings(warning_level)`,
		`CREATE INDEX IF NOT EXISTS idx_warnings_status ON warnings(status)`,
		`CREATE INDEX IF NOT EXISTS idx_warnings_location ON warnings(adm1_code, adm2_code)`,
		`CREATE INDEX IF NOT EXISTS idx_warnings_expires ON warnings(expires_at)`,
	}

	for _, stmt := range tables {
		if _, err := DB.Exec(stmt); err != nil {
			log.Printf("Warning executing: %v", err)
		}
	}

	log.Println("Database schema initialized")
	return nil
}

func SeedIndicators() error {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM indicators").Scan(&count)
	if count > 0 {
		log.Println("Indicators already seeded")
		return nil
	}

	indicators := []struct {
		Code, Name, Desc, Dim, Cat, Comp, Unit, DType, Res, Trans, Src, Freq string
		Min, Max, Wt float64
		Inv int
	}{
		{"HA.NAT.EQ-EXP", "Earthquake Exposure", "Population exposed to earthquake", "HAZARD", "Natural", "Earthquake", "MMI", "numeric", "adm2", "none", "USGS", "5y", 0, 10, 1, 0},
		{"HA.NAT.FL-EXP", "Flood Exposure", "Population exposed to floods", "HAZARD", "Natural", "Flood", "%", "percentage", "adm2", "none", "JRC", "annual", 0, 100, 1, 0},
		{"HA.NAT.LS-EXP", "Landslide Exposure", "Population exposed to landslides", "HAZARD", "Natural", "Landslide", "%", "percentage", "adm2", "none", "NASA", "5y", 0, 100, 1, 0},
		{"HA.NAT.DR-FRE", "Drought Frequency", "Drought events per decade", "HAZARD", "Natural", "Drought", "events", "count", "adm1", "none", "EM-DAT", "annual", 0, 20, 1, 0},
		{"HA.NAT.ST-TC", "Cyclone Tracks", "Tropical cyclone tracks", "HAZARD", "Natural", "Cyclone", "tracks", "count", "adm1", "none", "IBTrACS", "annual", 0, 50, 1, 0},
		{"HA.NAT.WF-BURN", "Wildfire Area", "Area burned by wildfires", "HAZARD", "Natural", "Wildfire", "km2", "numeric", "adm1", "log", "MODIS", "annual", 0, 10000, 1, 0},
		{"HA.HUM.VIO-EVE", "Violence Events", "Number of violent events", "HAZARD", "Human", "Violence", "events", "count", "adm1", "log", "ACLED", "monthly", 0, 1000, 1, 0},
		{"HA.HUM.VIO-FAT", "Violence Fatalities", "Conflict fatalities", "HAZARD", "Human", "Violence", "deaths", "count", "adm1", "log", "ACLED", "monthly", 0, 5000, 1, 0},
		{"VU.SE.POV-HDI", "Human Development Index", "HDI score", "VULNERABILITY", "Socio-Economic", "Development", "index", "index", "adm1", "none", "UNDP", "5y", 0, 1, 1, 1},
		{"VU.SE.POV-MPI", "Poverty Index", "Multidimensional Poverty", "VULNERABILITY", "Socio-Economic", "Poverty", "index", "index", "adm1", "none", "OPHI", "5y", 0, 1, 1, 0},
		{"VU.SE.LV-FS", "Food Security", "Food insecurity prevalence", "VULNERABILITY", "Socio-Economic", "Food", "%", "percentage", "adm1", "none", "WFP", "quarterly", 0, 100, 1, 0},
		{"VU.SE.LV-IPC", "IPC Phase", "Food security phase", "VULNERABILITY", "Socio-Economic", "Food", "phase", "numeric", "adm1", "none", "IPC", "quarterly", 1, 5, 1, 0},
		{"VU.VG.DP-IDP", "Displaced Persons", "Internally displaced", "VULNERABILITY", "Vulnerable Groups", "Displaced", "count", "count", "adm1", "log", "UNHCR", "monthly", 0, 1000000, 1, 0},
		{"VU.VG.CH-MORTINF", "Infant Mortality", "Infant mortality rate", "VULNERABILITY", "Vulnerable Groups", "Child Health", "per 1000", "numeric", "adm1", "none", "UNICEF", "annual", 0, 200, 1, 0},
		{"VU.VG.CH-UW", "Child Underweight", "Underweight children", "VULNERABILITY", "Vulnerable Groups", "Child Health", "%", "percentage", "adm1", "none", "UNICEF", "annual", 0, 50, 1, 0},
		{"CC.INF.HC-BCG", "BCG Immunization", "BCG coverage", "COPING_CAPACITY", "Infrastructure", "Health", "%", "percentage", "adm1", "none", "WHO", "annual", 0, 100, 1, 0},
		{"CC.INF.HC-DTP", "DTP Immunization", "DTP3 coverage", "COPING_CAPACITY", "Infrastructure", "Health", "%", "percentage", "adm1", "none", "WHO", "annual", 0, 100, 1, 0},
		{"CC.INF.COM-ELEC", "Electricity Access", "Electricity access", "COPING_CAPACITY", "Infrastructure", "Communication", "%", "percentage", "adm1", "none", "DHS", "annual", 0, 100, 1, 0},
		{"CC.INF.COM-PHONE", "Mobile Phone", "Phone ownership", "COPING_CAPACITY", "Infrastructure", "Communication", "%", "percentage", "adm1", "none", "DHS", "annual", 0, 100, 1, 0},
		{"CC.INF.EDU-YRS", "Years of Schooling", "Mean schooling years", "COPING_CAPACITY", "Infrastructure", "Education", "years", "numeric", "adm1", "none", "UNESCO", "annual", 0, 15, 1, 0},
		{"CC.INS.GOV-SCI", "Social Cohesion", "Social cohesion index", "COPING_CAPACITY", "Institutional", "Governance", "index", "index", "adm1", "none", "Afrobarometer", "5y", 0, 1, 1, 0},
		{"CC.INS.GOV-EFF", "Gov Effectiveness", "Governance effectiveness", "COPING_CAPACITY", "Institutional", "Governance", "index", "index", "national", "none", "WB", "annual", -2.5, 2.5, 1, 0},
	}

	for _, i := range indicators {
		DB.Exec(`INSERT INTO indicators (code, name, description, dimension, category, component, unit, data_type, resolution, transformation, min_value, max_value, weight, invert_scale, data_source, update_frequency, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
			i.Code, i.Name, i.Desc, i.Dim, i.Cat, i.Comp, i.Unit, i.DType, i.Res, i.Trans, i.Min, i.Max, i.Wt, i.Inv, i.Src, i.Freq)
	}

	log.Println("Indicators seeded")
	return nil
}

func SeedAdminUser() error {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if count > 0 {
		return nil
	}

	// Generate bcrypt hash for admin123
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash admin password: %v", err)
		return err
	}

	DB.Exec(`INSERT INTO users (email, password_hash, full_name, phone, role, is_active)
		VALUES (?, ?, ?, ?, ?, 1)`,
		"admin@inform.go.tz", string(hash),
		"System Administrator", "+255700000000", "admin")

	log.Println("Admin created: admin@inform.go.tz / admin123")
	return nil
}

func SeedTanzaniaCommittees() error {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM committees").Scan(&count)
	if count > 0 {
		return nil
	}

	// All 31 Tanzania regions (26 Mainland + 5 Zanzibar)
	regions := [][2]string{
		// Mainland regions
		{"Dodoma", "TZ01"}, {"Arusha", "TZ02"}, {"Kilimanjaro", "TZ03"},
		{"Tanga", "TZ04"}, {"Morogoro", "TZ05"}, {"Pwani", "TZ06"},
		{"Dar es Salaam", "TZ07"}, {"Lindi", "TZ08"}, {"Mtwara", "TZ09"},
		{"Ruvuma", "TZ10"}, {"Iringa", "TZ11"}, {"Mbeya", "TZ12"},
		{"Singida", "TZ13"}, {"Tabora", "TZ14"}, {"Rukwa", "TZ15"},
		{"Kigoma", "TZ16"}, {"Shinyanga", "TZ17"}, {"Kagera", "TZ18"},
		{"Mwanza", "TZ19"}, {"Mara", "TZ20"}, {"Manyara", "TZ21"},
		{"Njombe", "TZ22"}, {"Katavi", "TZ23"}, {"Simiyu", "TZ24"},
		{"Geita", "TZ25"}, {"Songwe", "TZ26"},
		// Zanzibar regions
		{"Kaskazini Unguja", "TZ27"}, {"Kusini Unguja", "TZ28"},
		{"Mjini Magharibi", "TZ29"}, {"Kaskazini Pemba", "TZ30"},
		{"Kusini Pemba", "TZ31"},
	}

	for _, r := range regions {
		DB.Exec(`INSERT INTO committees (name, type, adm1_code, adm1_name, is_active)
			VALUES (?, ?, ?, ?, 1)`, r[0]+" Regional Disaster Committee", "regional", r[1], r[0])
	}

	log.Println("Tanzania committees seeded")
	return nil
}
