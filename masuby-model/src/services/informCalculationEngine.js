/**
 * INFORM CALCULATION ENGINE - COMPLETE IMPLEMENTATION
 *
 * Based on Tanzania - Country Model Template.xlsx
 *
 * Implements full INFORM methodology:
 * 1. Raw Indicators (84) → Normalize to 0-10 scale
 * 2. Components (32) → Aggregate indicators using MAX/MEAN
 * 3. Categories (6) → Aggregate components
 * 4. Dimensions (3) → Aggregate categories
 * 5. RISK = (HAZARD × VULNERABILITY × LCC)^(1/3)
 */

import { COMPLETE_HIERARCHY, ALL_INDICATORS, ALL_COMPONENTS } from './informIndicatorDefinitions';

// ============================================================================
// RISK CLASSIFICATION (INFORM Standard)
// ============================================================================

export const RISK_CLASSES = [
  { min: 0.0, max: 2.0, label: 'Very Low', labelSw: 'Hatari Ndogo Sana', color: '#2E7D32', level: 1 },
  { min: 2.0, max: 3.5, label: 'Low', labelSw: 'Hatari Ndogo', color: '#8BC34A', level: 2 },
  { min: 3.5, max: 5.0, label: 'Medium', labelSw: 'Hatari ya Wastani', color: '#FFC107', level: 3 },
  { min: 5.0, max: 6.5, label: 'High', labelSw: 'Hatari Kubwa', color: '#FF9800', level: 4 },
  { min: 6.5, max: 10.0, label: 'Very High', labelSw: 'Hatari Kubwa Sana', color: '#D32F2F', level: 5 }
];

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

/**
 * Round to specified decimal places
 */
export function roundTo(value, decimals = 2) {
  if (value === null || value === undefined || isNaN(value)) return null;
  return Math.round(value * Math.pow(10, decimals)) / Math.pow(10, decimals);
}

/**
 * Clamp value to range
 */
export function clamp(value, min = 0, max = 10) {
  if (value === null || value === undefined || isNaN(value)) return null;
  return Math.min(max, Math.max(min, value));
}

/**
 * Normalize value to 0-10 scale based on polarity
 * NEGATIVE: higher raw value = higher risk (keep as-is if already 0-10)
 * POSITIVE: higher raw value = lower risk (invert: 10 - value)
 */
export function normalizeValue(value, indicator) {
  if (value === null || value === undefined || isNaN(value)) return null;

  const indDef = ALL_INDICATORS[indicator];
  if (!indDef) return clamp(value);

  // If value is already in 0-10 range, just clamp it
  let normalized = clamp(value);

  // For POSITIVE polarity (like coping capacity), invert
  // Actually, we handle inversion at dimension level for CC
  return normalized;
}

/**
 * Calculate MAX aggregation
 */
export function maxAggregation(values) {
  const valid = values.filter(v => v !== null && v !== undefined && !isNaN(v));
  if (valid.length === 0) return null;
  return Math.max(...valid);
}

/**
 * Calculate MEAN aggregation
 */
export function meanAggregation(values) {
  const valid = values.filter(v => v !== null && v !== undefined && !isNaN(v));
  if (valid.length === 0) return null;
  return valid.reduce((sum, v) => sum + v, 0) / valid.length;
}

/**
 * Aggregate values using specified method
 */
export function aggregate(values, method = 'MEAN') {
  switch (method.toUpperCase()) {
    case 'MAX':
      return maxAggregation(values);
    case 'MEAN':
    case 'ARITHMETIC_MEAN':
    case 'AVG':
      return meanAggregation(values);
    default:
      return meanAggregation(values);
  }
}

/**
 * Calculate geometric mean
 */
export function geometricMean(values) {
  const valid = values.filter(v => v !== null && v !== undefined && !isNaN(v) && v > 0);
  if (valid.length === 0) return null;

  const logSum = valid.reduce((acc, val) => acc + Math.log(val), 0);
  return Math.exp(logSum / valid.length);
}

/**
 * INFORM Scaled Geometric Mean (matches Excel formula exactly)
 * Excel: =(10-GEOMEAN(((10-val1)/10*9+1),((10-val2)/10*9+1)))/9*10
 *
 * This converts 0-10 scale to 1-10, takes GEOMEAN, then converts back.
 * Used for combining categories into dimensions.
 */
export function informScaledGeomean(values) {
  const valid = values.filter(v => v !== null && v !== undefined && !isNaN(v));
  if (valid.length === 0) return null;

  // Step 1: Convert from 0-10 to 1-10 scale: (10-val)/10*9+1
  const scaled = valid.map(v => ((10 - v) / 10 * 9) + 1);

  // Step 2: Calculate geometric mean
  const logSum = scaled.reduce((acc, val) => acc + Math.log(val), 0);
  const geomean = Math.exp(logSum / scaled.length);

  // Step 3: Convert back to 0-10 scale: (10-geomean)/9*10
  const result = (10 - geomean) / 9 * 10;

  return result;
}

/**
 * Classify risk score
 */
export function classifyRisk(score) {
  if (score === null || score === undefined || isNaN(score)) return null;

  for (const level of RISK_CLASSES) {
    if (score >= level.min && score < level.max) {
      return { ...level };
    }
  }

  // Edge case: exactly 10
  if (score >= 10) {
    return { ...RISK_CLASSES[RISK_CLASSES.length - 1] };
  }

  return null;
}

/**
 * Get risk color
 */
export function getRiskColor(score) {
  const classification = classifyRisk(score);
  return classification?.color || '#666';
}

// ============================================================================
// COMPLETE INFORM CALCULATION
// ============================================================================

/**
 * Calculate component score from raw indicator values
 *
 * @param {string} componentId - Component ID
 * @param {Object} indicatorValues - { indicator_id: { value: number, ... } }
 * @returns {Object} { score, indicatorScores, coverage }
 */
export function calculateComponent(componentId, indicatorValues) {
  const component = ALL_COMPONENTS[componentId];
  if (!component) return { score: null, indicatorScores: {}, coverage: 0 };

  const indicatorScores = {};
  const values = [];

  for (const indId of component.indicators) {
    const indData = indicatorValues[indId];
    const value = indData?.value ?? indData; // Handle both { value: x } and plain number

    if (value !== null && value !== undefined && !isNaN(value)) {
      const normalized = normalizeValue(value, indId);
      indicatorScores[indId] = normalized;
      values.push(normalized);
    }
  }

  const score = aggregate(values, component.aggregation);
  const coverage = component.indicators.length > 0
    ? (values.length / component.indicators.length) * 100
    : 0;

  return {
    score: roundTo(score, 2),
    indicatorScores,
    coverage: roundTo(coverage, 0),
    indicatorCount: values.length,
    totalIndicators: component.indicators.length
  };
}

/**
 * Calculate category score from components
 *
 * @param {string} dimensionId - Dimension ID
 * @param {string} categoryId - Category ID
 * @param {Object} indicatorValues - Raw indicator values
 * @returns {Object} { score, components, coverage }
 */
export function calculateCategory(dimensionId, categoryId, indicatorValues) {
  const dim = COMPLETE_HIERARCHY[dimensionId];
  if (!dim) return { score: null, components: {}, coverage: 0 };

  const cat = dim.categories[categoryId];
  if (!cat) return { score: null, components: {}, coverage: 0 };

  const componentResults = {};
  const componentScores = [];

  for (const [compId, comp] of Object.entries(cat.components)) {
    const result = calculateComponent(compId, indicatorValues);
    componentResults[compId] = {
      name: comp.name,
      code: comp.code,
      ...result
    };
    if (result.score !== null) {
      componentScores.push(result.score);
    }
  }

  const score = aggregate(componentScores, cat.aggregation);
  const totalComponents = Object.keys(cat.components).length;
  const coverage = totalComponents > 0
    ? (componentScores.length / totalComponents) * 100
    : 0;

  return {
    score: roundTo(score, 2),
    name: cat.name,
    components: componentResults,
    coverage: roundTo(coverage, 0),
    componentCount: componentScores.length,
    totalComponents
  };
}

/**
 * Calculate dimension score from categories
 *
 * @param {string} dimensionId - Dimension ID
 * @param {Object} indicatorValues - Raw indicator values
 * @returns {Object} Full dimension calculation result
 */
export function calculateDimension(dimensionId, indicatorValues) {
  const dim = COMPLETE_HIERARCHY[dimensionId];
  if (!dim) return { score: null, categories: {} };

  const categoryResults = {};
  const categoryScores = [];

  for (const [catId, cat] of Object.entries(dim.categories)) {
    const result = calculateCategory(dimensionId, catId, indicatorValues);
    categoryResults[catId] = result;
    if (result.score !== null) {
      categoryScores.push({
        score: result.score,
        weight: cat.weight || 0.5
      });
    }
  }

  // Calculate dimension score using INFORM methodology
  // Excel uses scaled geometric mean for combining categories into dimensions
  let dimensionScore = null;
  if (categoryScores.length > 0) {
    const scores = categoryScores.map(c => c.score);

    if (categoryScores.length >= 2) {
      // Use INFORM scaled geometric mean (matches Excel formula)
      dimensionScore = informScaledGeomean(scores);
    } else if (categoryScores.length === 1) {
      // Single category - use its score directly
      dimensionScore = scores[0];
    }
  }

  // For Coping Capacity, invert to get Lack of Coping Capacity
  // Note: The scaled geomean already handles the "lower is better" logic,
  // but CC needs explicit inversion for the final risk calculation
  if (dim.invert && dimensionScore !== null) {
    dimensionScore = 10 - dimensionScore;
  }

  return {
    score: roundTo(dimensionScore, 2),
    name: dim.name,
    code: dim.code,
    color: dim.color,
    categories: categoryResults,
    inverted: dim.invert || false
  };
}

/**
 * Calculate complete INFORM Risk from indicator values
 *
 * @param {Object} indicatorValues - { indicator_id: { value: number, confidence?: string } }
 * @returns {Object} Complete risk calculation result
 */
export function calculateINFORMRisk(indicatorValues) {
  const result = {
    dimensions: {},
    risk: null,
    classification: null,
    formula: null,
    metadata: {
      calculatedAt: new Date().toISOString(),
      methodology: 'INFORM SADC 2024',
      template: 'Tanzania - Country Model Template.xlsx',
      indicatorCount: 0,
      componentCount: 0,
      coverage: 0
    }
  };

  // Calculate each dimension
  const dimensionScores = [];
  let totalIndicators = 0;
  let providedIndicators = 0;

  for (const dimId of Object.keys(COMPLETE_HIERARCHY)) {
    const dimResult = calculateDimension(dimId, indicatorValues);
    result.dimensions[dimId] = dimResult;

    if (dimResult.score !== null) {
      dimensionScores.push({
        id: dimId,
        score: dimResult.score
      });
    }

    // Count indicators
    for (const cat of Object.values(dimResult.categories)) {
      for (const comp of Object.values(cat.components)) {
        totalIndicators += comp.totalIndicators || 0;
        providedIndicators += comp.indicatorCount || 0;
      }
    }
  }

  result.metadata.indicatorCount = providedIndicators;
  result.metadata.totalIndicators = totalIndicators;
  result.metadata.coverage = totalIndicators > 0
    ? roundTo((providedIndicators / totalIndicators) * 100, 0)
    : 0;

  // Calculate final INFORM Risk using geometric mean
  if (dimensionScores.length === 3) {
    const H = result.dimensions.HAZARD.score;
    const V = result.dimensions.VULNERABILITY.score;
    const LCC = result.dimensions.COPING_CAPACITY.score; // Already inverted

    if (H > 0 && V > 0 && LCC > 0) {
      result.risk = roundTo(Math.pow(H * V * LCC, 1/3), 2);
      result.classification = classifyRisk(result.risk);
      result.formula = {
        expression: `Risk = (${H} × ${V} × ${LCC})^(1/3)`,
        H, V, LCC,
        result: result.risk
      };
    }
  }

  return result;
}

/**
 * Validate indicator values
 */
export function validateIndicatorValues(indicatorValues) {
  const errors = [];
  const warnings = [];
  const coverage = { HAZARD: 0, VULNERABILITY: 0, COPING_CAPACITY: 0 };

  for (const [indId, data] of Object.entries(indicatorValues)) {
    const indDef = ALL_INDICATORS[indId];
    const value = data?.value ?? data;

    if (!indDef) {
      warnings.push(`Unknown indicator: ${indId}`);
      continue;
    }

    if (value === null || value === undefined) {
      continue;
    }

    if (isNaN(value)) {
      errors.push(`Invalid value for ${indDef.name}: not a number`);
      continue;
    }

    if (value < 0 || value > 10) {
      warnings.push(`Value out of typical range for ${indDef.name}: ${value}`);
    }

    coverage[indDef.dimension] = (coverage[indDef.dimension] || 0) + 1;
  }

  const hasHazard = coverage.HAZARD > 0;
  const hasVulnerability = coverage.VULNERABILITY > 0;
  const hasCoping = coverage.COPING_CAPACITY > 0;

  if (!hasHazard) warnings.push('No Hazard indicators provided');
  if (!hasVulnerability) warnings.push('No Vulnerability indicators provided');
  if (!hasCoping) warnings.push('No Coping Capacity indicators provided');

  return {
    isValid: errors.length === 0 && hasHazard && hasVulnerability && hasCoping,
    canCalculate: hasHazard && hasVulnerability && hasCoping,
    errors,
    warnings,
    coverage: {
      hazard: coverage.HAZARD,
      vulnerability: coverage.VULNERABILITY,
      copingCapacity: coverage.COPING_CAPACITY,
      total: Object.values(coverage).reduce((a, b) => a + b, 0)
    }
  };
}

// ============================================================================
// EXPORTS
// ============================================================================

export default {
  // Main calculation
  calculateINFORMRisk,
  calculateDimension,
  calculateCategory,
  calculateComponent,
  validateIndicatorValues,

  // Utilities
  roundTo,
  clamp,
  normalizeValue,
  aggregate,
  geometricMean,
  informScaledGeomean,
  classifyRisk,
  getRiskColor,

  // Constants
  RISK_CLASSES,
  COMPLETE_HIERARCHY,
  ALL_INDICATORS,
  ALL_COMPONENTS
};
