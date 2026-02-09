# INFORM Risk Management System

A comprehensive Go-based web application for subnational INFORM risk assessment in Tanzania. This system allows Regional and Ward Disaster Committees to collect, manage, and analyze risk indicators with real-time data synchronization.

## Features

- **User Registration & Authentication**: Regional and Ward Disaster Committee members can register and access the system
- **Multi-Level Data Entry**: Support for National, ADM1 (Regional), and ADM2 (District) level indicators
- **Real-Time Updates**: WebSocket-based live data synchronization
- **Automated Risk Calculation**: Implements all INFORM formulas exactly as specified in the Tanzania Country Model Template
- **Transparency Portal**: Full documentation of formulas, data flows, API endpoints, and methodology
- **Role-Based Access Control**: Admin, Regional Committee, Ward Committee, and Viewer roles

## INFORM Risk Formulas

The system implements the exact formulas from the INFORM SADC methodology:

### Risk Score (Cubic Geometric Mean)
```
RISK = HAZARD^(1/3) × VULNERABILITY^(1/3) × LACK_OF_COPING_CAPACITY^(1/3)
```

### Dimension Score (Adjusted Geometric Mean)
```
DIMENSION = (10 - GEOMEAN(((10-CAT1)/10×9+1), ((10-CAT2)/10×9+1))) / 9 × 10
```

### Category Score (Arithmetic Average)
```
CATEGORY = AVERAGE(Component1, Component2, ..., ComponentN)
```

### Normalization (Min-Max)
```
NORMALIZED = ((VALUE - MIN) / (MAX - MIN)) × 10
```

With optional LOG transformation for skewed data:
```
TRANSFORMED = LN(VALUE + 1)
```

## Project Structure

```
inform-system/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── database/
│   │   └── database.go          # Database connection & migrations
│   ├── formula/
│   │   └── formula.go           # INFORM calculation engine
│   ├── handlers/
│   │   └── handlers.go          # HTTP API handlers
│   ├── middleware/
│   │   └── auth.go              # JWT authentication
│   ├── models/
│   │   └── models.go            # Data structures
│   └── websocket/
│       └── hub.go               # Real-time updates
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   └── style.css        # Application styles
│   │   └── js/
│   │       └── app.js           # Client-side utilities
│   └── templates/
│       ├── index.html           # Landing page
│       ├── login.html           # Login page
│       ├── register.html        # Registration page
│       ├── dashboard.html       # Main dashboard
│       ├── data_entry.html      # Data entry form
│       ├── risk_scores.html     # Risk scores display
│       ├── committees.html      # Committee management
│       └── transparency.html    # Documentation portal
├── .env.example                 # Environment template
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Modern web browser

## Installation

1. **Clone the repository**
   ```bash
   cd /home/kaijage/model/inform/inform-system
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   ```

3. **Set up PostgreSQL database**
   ```bash
   # Create database
   createdb inform_db

   # Or using psql
   psql -c "CREATE DATABASE inform_db;"
   ```

4. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

6. **Access the application**
   - Open http://localhost:8080 in your browser
   - Default admin: `admin@inform.go.tz` / `admin123`

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/logout` | User logout |

### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user |

### Committees
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/committees` | List all committees |
| POST | `/api/v1/committees` | Create committee (admin) |

### Indicators
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/indicators` | List indicators |
| GET | `/api/v1/indicators/:id` | Get indicator by ID or code |

### Data Entry
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/data` | Create data entry |
| GET | `/api/v1/data` | List data entries |
| PUT | `/api/v1/data/:id/verify` | Verify/reject entry |

### Risk Calculation
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/risk/calculate` | Calculate risk scores |
| GET | `/api/v1/risk/scores` | Get risk scores |

### Transparency (Public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/transparency/formulas` | Get all formulas |
| GET | `/api/v1/transparency/dataflow` | Get data flow |
| GET | `/api/v1/transparency/linkages` | Get sheet linkages |
| GET | `/api/v1/transparency/api` | Get API documentation |

### WebSocket
| Endpoint | Description |
|----------|-------------|
| `/api/v1/ws` | Real-time updates connection |

## Indicators

The system includes 60+ pre-defined indicators across three dimensions:

### HAZARD
- **Natural**: Drought, Flood, Earthquake, Landslide, Cyclone, Wildfire, etc.
- **Human**: Conflict, Violence Events

### VULNERABILITY
- **Socio-Economic**: HDI, MPI, GINI, Food Security, Dependency
- **Vulnerable Groups**: Displaced Populations, Health Conditions, Child Mortality

### COPING CAPACITY
- **Infrastructure**: Health Services, Economic Capacity, Communication, Education
- **Institutional**: Governance, DRR Capacity, Early Warning Systems

## Data Flow

1. **Raw Data Collection** → Regional/Ward committees enter data
2. **Data Validation** → System validates format and range
3. **Outlier Detection** → IQR method identifies outliers
4. **Transformation** → LOG/SQRT applied if configured
5. **Normalization** → Min-Max to 0-10 scale
6. **Component Aggregation** → Arithmetic average
7. **Category Aggregation** → Arithmetic average
8. **Dimension Calculation** → Adjusted geometric mean
9. **Risk Calculation** → Cubic geometric mean
10. **Classification** → Very Low to Very High

## Risk Classification

| Score Range | Risk Class |
|-------------|------------|
| 0.0 - 1.9 | Very Low |
| 2.0 - 3.4 | Low |
| 3.5 - 4.9 | Medium |
| 5.0 - 6.4 | High |
| 6.5 - 10.0 | Very High |

## User Roles

| Role | Permissions |
|------|-------------|
| Admin | Full system access, manage users/committees |
| Regional Committee | Enter/verify data for their region |
| Ward Committee | Enter data for their ward |
| Viewer | View-only access to public data |

## Development

### Run in development mode
```bash
go run cmd/server/main.go
```

### Build for production
```bash
go build -o inform-system cmd/server/main.go
```

### Run tests
```bash
go test ./...
```

## License

This project is developed for the Tanzania Disaster Management Department.

## Support

For support, contact the system administrator or submit an issue.
