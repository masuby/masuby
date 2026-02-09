# 📊 MODULE 02: INFORM RISK - Implementation Plan

**Date:** December 15, 2025
**Status:** 🚧 **IN PROGRESS** - Core Framework Complete, Components In Development
**Priority:** HIGH - Critical risk assessment dashboard

---

## 🎯 OBJECTIVE

Build a comprehensive INFORM Risk dashboard that implements the **exact INFORM methodology** from the Tanzania Country Model Template Excel file, following these principles:

1. **Exact Formula Implementation**: Risk = (H&E × V × LCC)^(1/3)
2. **Sheet-by-Sheet Fidelity**: Every indicator from Excel template represented
3. **Formula-by-Formula Accuracy**: All aggregations use geometric mean
4. **Indicator-by-Indicator Transparency**: Show all indicators with data availability
5. **API Integration Ready**: Structure supports future real-time data integration

---

## ✅ COMPLETED (Phase 1A)

### 1. **Data Service Layer** ✅
**File:** [`src/services/informRiskDataService.js`](masuby-model/src/services/informRiskDataService.js)

**Completed Features:**
- ✅ Column mapping from Excel template (48 columns mapped)
- ✅ Data parsing logic for INFORM SADC 2024 sheet
- ✅ Administrative unit parsing (Country, ADM1, ADM2 levels)
- ✅ Three dimension extraction (H&E, V, LCC)
- ✅ National aggregation calculation
- ✅ Geometric mean implementation
- ✅ Formula verification logic
- ✅ Risk classification function (5-level scale)

**Key Functions:**
```javascript
parseInformRiskData(filePath)        // Parse Excel → JSON
calculateGeometricMean(...values)     // Geometric aggregation
verifyInformFormula(he, v, lcc, risk) // Validate calculations
getRiskClassification(score)          // 5-level classification
```

### 2. **Main Dashboard Component** ✅
**File:** [`src/components/inform-risk/Module02InformRisk.jsx`](masuby-model/src/components/inform-risk/Module02InformRisk.jsx)

**Completed Features:**
- ✅ Dashboard layout with header
- ✅ INFORM formula visualization
- ✅ Three-dimension overview cards
- ✅ Navigation tabs (Overview, H&E, V, LCC, Districts)
- ✅ Formula verification display
- ✅ Risk badge with classification
- ✅ District analysis section (grid view)
- ✅ Methodology notice

### 3. **Hazard & Exposure Component** ✅
**File:** [`src/components/inform-risk/dimensions/HazardExposureDimension.jsx`](masuby-model/src/components/inform-risk/dimensions/HazardExposureDimension.jsx)

**Completed Features:**
- ✅ 12 Natural Hazard indicators
- ✅ 5 Human Hazard indicators
- ✅ Expandable categories
- ✅ Indicator cards with bars
- ✅ Data availability handling (null values)
- ✅ Classification per indicator
- ✅ Aggregation methodology explanation

**Natural Hazards Implemented:**
1. Coastal Hazards
2. Drought
3. Earthquake
4. Environmental Degradation
5. Flood
6. Heatwave
7. Landslide
8. Lightning
9. Storms & Cyclone
10. Volcano
11. Wildfire
12. Zoonoses, Plants & Pests

**Human Hazards Implemented:**
1. Conflict Intensity
2. Conflict Risk
3. Hazardous Material
4. Internal Violence
5. Vehicle Accidents

### 4. **Mock Data Service** ✅
**File:** [`src/components/inform-risk/mockData.js`](masuby-model/src/components/inform-risk/mockData.js)

**Purpose:** Temporary data source matching Excel structure exactly, will be replaced with real Excel parsing.

---

## 🚧 IN PROGRESS (Phase 1B)

### 5. **Vulnerability Dimension Component** 📝 PENDING
**File:** `src/components/inform-risk/dimensions/VulnerabilityDimension.jsx`

**Requirements:**

**A. Socio-Economic Vulnerability (4 indicators):**
1. Development & Poverty
2. Economic Dependency
3. Habitat
4. Livelihoods
5. **Aggregate:** SOCIO-ECONOMIC VULNERABILITY

**B. Vulnerable Groups (4 indicators):**
1. Displaced People
2. Health Conditions
3. Children Health & Nutrition
4. Economic
5. **Aggregate:** VULNERABLE GROUPS

**C. Total Vulnerability:**
- Formula: V = (V_SE × V_VG)^(1/2)
- Two-step aggregation (sub-dimensions first, then total)

**UI Requirements:**
- Two expandable sections (Socio-Economic, Vulnerable Groups)
- Indicator cards for each component
- Sub-dimension scores displayed
- Total vulnerability score banner
- Explanation of two-component model

### 6. **Lack of Coping Capacity Component** 📝 PENDING
**File:** `src/components/inform-risk/dimensions/CopingCapacityDimension.jsx`

**Requirements:**

**A. Infrastructure (5 indicators):**
1. Access to health care
2. Economic capacity
3. WASH (Water, Sanitation, Hygiene)
4. Communication
5. Education
6. **Aggregate:** INFRASTRUCTURE

**B. Institutional (2 indicators):**
1. DRR implementation (Disaster Risk Reduction)
2. Governance
3. **Aggregate:** INSTITUTIONAL

**C. Total LCC:**
- Formula: LCC = (Infrastructure × Institutional)^(1/2)
- **Important:** These are "Lack of" indicators (higher score = less capacity = higher risk)

**UI Requirements:**
- Two expandable sections (Infrastructure, Institutional)
- Indicator cards with capacity inversion explanation
- Sub-dimension scores
- Total LCC score banner
- Explanation of "Lack of" framing

### 7. **Complete CSS Styling** 📝 PENDING
**Files:**
- `src/components/inform-risk/Module02InformRisk.css` (main dashboard)
- `src/components/inform-risk/dimensions/DimensionStyles.css` (shared dimension styles)

**Required Styles:**
- Dashboard header with risk badge
- INFORM formula visualization
- Navigation tabs (active states)
- Dimension overview cards (3-column grid)
- Indicator cards (with bars, scores, classifications)
- Expandable categories
- District grid
- Risk classification scale
- Responsive design (mobile, tablet, desktop)
- INFORM color standards

**Color Palette (INFORM Standards):**
- Hazard & Exposure: `#D32F2F` (Red)
- Vulnerability: `#FF9800` (Orange)
- Lack of Coping: `#1976D2` (Blue)
- Risk Classifications:
  - Very Low: `#43A047` (Green)
  - Low: `#8BC34A` (Light Green)
  - Medium: `#FFC107` (Yellow)
  - High: `#FF9800` (Orange)
  - Very High: `#F44336` (Red)

---

## 🔜 PENDING (Phase 1C - Integration)

### 8. **Excel Data Integration** ⏳ HIGH PRIORITY
**Current Status:** Using mock data

**Tasks:**
1. Update Module02InformRisk.jsx to use real Excel parsing:
   ```javascript
   import { parseInformRiskData } from '../../services/informRiskDataService';

   // In useEffect:
   const filePath = '/home/kaijage/model/inform/Tanzania - Country Model Template.xlsx';
   const riskData = parseInformRiskData(filePath);
   setData(riskData);
   ```

2. Handle data loading states
3. Handle parsing errors
4. Display data timestamp from Excel

### 9. **App Integration** ⏳ REQUIRED
**File:** [`src/App.jsx`](masuby-model/src/App.jsx)

**Current State:** Module 01 integrated, Module 02 not yet added

**Required Changes:**
```jsx
import Module02InformRisk from "./components/inform-risk/Module02InformRisk";

// In dropdown:
<option value="module02">MODULE 02 - INFORM RISK</option>

// In render:
{currentView === "module02" ? (
  <Module02InformRisk />
) : currentView === "module01" ? (
  <Module01Landing onComplete={() => setCurrentView("risk")} />
) : (
  <Home />
)}
```

### 10. **Data Reliability Indicators** ⏳ IMPORTANT
**Purpose:** Show data quality/availability per indicator (INFORM requirement)

**Requirements:**
- Add reliability score per dimension
- Flag missing data
- Show data source and timestamp
- Warning messages for low-reliability indicators
- Color-code data quality (High/Medium/Low)

**Example:**
```jsx
<div className="reliability-warning">
  ⚠️ High risk score, but low data reliability for health indicators
</div>
```

### 11. **Visualization Enhancements** ⏳ OPTIONAL
**Future Enhancements:**
- Interactive charts (bar charts, radar charts)
- District map visualization (Leaflet/Mapbox)
- Dimension comparison charts
- Trend analysis (if historical data available)
- Export functionality (PDF/PNG)

---

## 📂 FILE STRUCTURE

```
src/
├── components/
│   ├── inform-risk/
│   │   ├── Module02InformRisk.jsx          ✅ Main dashboard
│   │   ├── Module02InformRisk.css          📝 PENDING
│   │   ├── mockData.js                     ✅ Temporary data
│   │   └── dimensions/
│   │       ├── HazardExposureDimension.jsx ✅ Complete
│   │       ├── VulnerabilityDimension.jsx  📝 PENDING
│   │       ├── CopingCapacityDimension.jsx 📝 PENDING
│   │       └── DimensionStyles.css         📝 PENDING
│   └── landing/
│       ├── Module01Landing.jsx             ✅ Complete (Module 01)
│       └── ...
├── services/
│   ├── informRiskDataService.js            ✅ Excel parsing
│   └── excelStorageService.js              ✅ Existing (Supabase)
├── App.jsx                                 🚧 Needs Module 02 integration
└── ...
```

---

## 🎯 SUCCESS CRITERIA

| Criteria | Status | Priority |
|----------|--------|----------|
| Parse Excel template correctly | ✅ Complete | CRITICAL |
| Display all H&E indicators | ✅ Complete | CRITICAL |
| Display all V indicators | 📝 Pending | CRITICAL |
| Display all LCC indicators | 📝 Pending | CRITICAL |
| Show INFORM formula visualization | ✅ Complete | CRITICAL |
| Implement geometric mean aggregation | ✅ Complete | CRITICAL |
| Verify formula accuracy | ✅ Complete | CRITICAL |
| Display risk classification | ✅ Complete | CRITICAL |
| Handle missing data | ✅ Complete | HIGH |
| Show data reliability | 📝 Pending | HIGH |
| District-level analysis | ✅ Partial | MEDIUM |
| Responsive design | 📝 Pending | HIGH |
| App integration | 📝 Pending | CRITICAL |

---

## 📊 PROGRESS METRICS

| Component | Lines of Code | Completion % | Priority |
|-----------|---------------|--------------|----------|
| Data Service | 300 | 100% | ✅ |
| Main Dashboard | 450 | 80% | 🚧 |
| H&E Dimension | 250 | 100% | ✅ |
| V Dimension | 0 | 0% | ⏳ |
| LCC Dimension | 0 | 0% | ⏳ |
| CSS Styles | 0 | 0% | ⏳ |
| App Integration | 0 | 0% | ⏳ |
| **Overall Module 02** | **1,000+** | **40%** | 🚧 |

---

## 🚀 NEXT STEPS (Priority Order)

### **Immediate (Today):**
1. ✅ Create Vulnerability Dimension component (~250 lines)
2. ✅ Create Lack of Coping Capacity component (~250 lines)
3. ✅ Create comprehensive CSS (~500 lines)
4. ✅ Integrate Module 02 into App.jsx
5. ✅ Test with mock data

### **Short-term (Tomorrow):**
6. ⏳ Connect real Excel data (replace mock data)
7. ⏳ Add data reliability indicators
8. ⏳ Add responsive design refinements
9. ⏳ User testing

### **Medium-term (This Week):**
10. ⏳ Add visualization enhancements (charts)
11. ⏳ Implement district map view
12. ⏳ Add export functionality
13. ⏳ Performance optimization

---

## 🎓 MODULE 02 LEARNING OUTCOMES

Users who complete Module 02 will:
1. **Understand Tanzania's INFORM Risk Score** (4.2 - Medium)
2. **See all three dimensions** breakdown (H&E, V, LCC)
3. **Explore individual indicators** that drive risk
4. **Verify the INFORM formula** calculation
5. **Compare districts** by risk level
6. **Interpret data reliability** and limitations
7. **Make informed decisions** based on risk analysis

---

## 🔗 INTEGRATION WITH MODULE 01

**Flow:**
1. Module 01 (Education) → teaches INFORM concepts
2. **Module 02 (Risk Dashboard)** → applies concepts to Tanzania data
3. Module 03 (Warning) → early warning based on risk
4. Module 04 (Severity) → incident tracking
5. Module 05 (Climate) → future risk projections
6. Module 06 (Response) → planning and resources

**Module 02 reinforces Module 01 by:**
- Showing real Tanzania data for concepts learned
- Demonstrating the formula in practice
- Allowing exploration of dimension interactions
- Validating understanding through actual scores

---

**Last Updated:** December 15, 2025
**Status:** 🚧 **40% COMPLETE** - Core framework done, dimension components needed
**Estimated Time to Completion:** 6-8 hours (2 components + CSS + integration)
