"""
Synthetic Time-Series Data Generator for Risk Assessment Training

Generates realistic historical risk data for businesses across different industries,
including seasonal patterns, economic cycles, and industry-specific risk factors.
"""

import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Optional
import random
from datetime import datetime, timedelta
import json


class RiskPattern:
    """Represents risk patterns for a specific industry or business type."""
    
    def __init__(self, 
                 base_risk: float,
                 volatility: float,
                 seasonal_amplitude: float,
                 trend_drift: float,
                 economic_sensitivity: float,
                 compliance_events: List[Dict]):
        self.base_risk = base_risk
        self.volatility = volatility
        self.seasonal_amplitude = seasonal_amplitude
        self.trend_drift = trend_drift
        self.economic_sensitivity = economic_sensitivity
        self.compliance_events = compliance_events


class SyntheticDataGenerator:
    """Generates synthetic time-series data for risk assessment training."""
    
    def __init__(self, seed: int = 42):
        self.seed = seed
        np.random.seed(seed)
        random.seed(seed)
        
        # Industry-specific risk patterns
        self.industry_patterns = self._initialize_industry_patterns()
        
        # Economic cycle parameters
        self.economic_cycle_length = 84  # 7 years in months
        self.current_cycle_phase = 0.3  # 0-1, where 0.5 is peak
        
    def _initialize_industry_patterns(self) -> Dict[str, RiskPattern]:
        """Initialize risk patterns for different industries."""
        return {
            "technology": RiskPattern(
                base_risk=0.15,
                volatility=0.08,
                seasonal_amplitude=0.05,
                trend_drift=0.02,
                economic_sensitivity=0.3,
                compliance_events=[
                    {"type": "data_breach", "frequency": 0.02, "impact": 0.3},
                    {"type": "regulatory_change", "frequency": 0.01, "impact": 0.2}
                ]
            ),
            "financial": RiskPattern(
                base_risk=0.25,
                volatility=0.12,
                seasonal_amplitude=0.08,
                trend_drift=0.01,
                economic_sensitivity=0.8,
                compliance_events=[
                    {"type": "regulatory_audit", "frequency": 0.05, "impact": 0.4},
                    {"type": "market_volatility", "frequency": 0.03, "impact": 0.6}
                ]
            ),
            "healthcare": RiskPattern(
                base_risk=0.20,
                volatility=0.06,
                seasonal_amplitude=0.03,
                trend_drift=0.01,
                economic_sensitivity=0.2,
                compliance_events=[
                    {"type": "regulatory_inspection", "frequency": 0.04, "impact": 0.3},
                    {"type": "patient_safety", "frequency": 0.02, "impact": 0.5}
                ]
            ),
            "manufacturing": RiskPattern(
                base_risk=0.18,
                volatility=0.10,
                seasonal_amplitude=0.06,
                trend_drift=0.02,
                economic_sensitivity=0.6,
                compliance_events=[
                    {"type": "safety_incident", "frequency": 0.03, "impact": 0.4},
                    {"type": "supply_chain", "frequency": 0.02, "impact": 0.3}
                ]
            ),
            "retail": RiskPattern(
                base_risk=0.22,
                volatility=0.15,
                seasonal_amplitude=0.12,
                trend_drift=0.03,
                economic_sensitivity=0.7,
                compliance_events=[
                    {"type": "seasonal_demand", "frequency": 0.08, "impact": 0.2},
                    {"type": "consumer_complaint", "frequency": 0.05, "impact": 0.1}
                ]
            ),
            "real_estate": RiskPattern(
                base_risk=0.30,
                volatility=0.20,
                seasonal_amplitude=0.08,
                trend_drift=0.04,
                economic_sensitivity=0.9,
                compliance_events=[
                    {"type": "market_crash", "frequency": 0.01, "impact": 0.8},
                    {"type": "interest_rate", "frequency": 0.02, "impact": 0.4}
                ]
            )
        }
    
    def generate_business_time_series(self, 
                                    business_id: str,
                                    industry: str,
                                    start_date: datetime,
                                    months: int = 36,
                                    business_size: str = "medium") -> pd.DataFrame:
        """
        Generate time-series data for a single business.
        
        Args:
            business_id: Unique identifier for the business
            industry: Industry classification
            start_date: Start date for the time series
            months: Number of months to generate
            business_size: Size category (small, medium, large)
            
        Returns:
            DataFrame with columns: date, business_id, industry, risk_score, 
            financial_health, compliance_score, market_conditions, features
        """
        if industry not in self.industry_patterns:
            industry = "technology"  # Default fallback
            
        pattern = self.industry_patterns[industry]
        
        # Generate base time series
        dates = [start_date + timedelta(days=30*i) for i in range(months)]
        
        # Base risk evolution (random walk with drift)
        base_risk = self._generate_base_risk_series(months, pattern)
        
        # Seasonal component
        seasonal = self._generate_seasonal_component(months, pattern)
        
        # Economic cycle component
        economic = self._generate_economic_cycle(months, pattern)
        
        # Compliance events
        compliance_events = self._generate_compliance_events(months, pattern)
        
        # Business size adjustment
        size_adjustment = self._get_size_adjustment(business_size)
        
        # Combine all components
        risk_scores = base_risk + seasonal + economic + compliance_events + size_adjustment
        
        # Ensure risk scores are in valid range [0, 1]
        risk_scores = np.clip(risk_scores, 0.0, 1.0)
        
        # Generate additional features
        features = self._generate_additional_features(months, industry, business_size)
        
        # Create DataFrame
        data = {
            'date': dates,
            'business_id': [business_id] * months,
            'industry': [industry] * months,
            'risk_score': risk_scores,
            'financial_health': features['financial_health'],
            'compliance_score': features['compliance_score'],
            'market_conditions': features['market_conditions'],
            'revenue_trend': features['revenue_trend'],
            'employee_count': features['employee_count'],
            'debt_ratio': features['debt_ratio'],
            'profit_margin': features['profit_margin'],
            'customer_satisfaction': features['customer_satisfaction'],
            'regulatory_compliance': features['regulatory_compliance'],
            'market_share': features['market_share'],
            'innovation_index': features['innovation_index']
        }
        
        return pd.DataFrame(data)
    
    def _generate_base_risk_series(self, months: int, pattern: RiskPattern) -> np.ndarray:
        """Generate base risk evolution using random walk with drift."""
        # Random walk with drift
        drift = pattern.trend_drift / 12  # Monthly drift
        volatility = pattern.volatility / np.sqrt(12)  # Monthly volatility
        
        # Generate random walk
        innovations = np.random.normal(0, volatility, months)
        innovations[0] = 0  # Start at base risk
        
        # Apply drift
        for i in range(1, months):
            innovations[i] += drift
        
        # Cumulative sum to get the series
        risk_series = pattern.base_risk + np.cumsum(innovations)
        
        return risk_series
    
    def _generate_seasonal_component(self, months: int, pattern: RiskPattern) -> np.ndarray:
        """Generate seasonal component with quarterly and annual cycles."""
        t = np.arange(months)
        
        # Quarterly cycle (every 3 months)
        quarterly = pattern.seasonal_amplitude * 0.6 * np.sin(2 * np.pi * t / 3)
        
        # Annual cycle (every 12 months)
        annual = pattern.seasonal_amplitude * 0.4 * np.sin(2 * np.pi * t / 12)
        
        return quarterly + annual
    
    def _generate_economic_cycle(self, months: int, pattern: RiskPattern) -> np.ndarray:
        """Generate economic cycle component."""
        t = np.arange(months)
        
        # Long-term economic cycle
        cycle_phase = (t / self.economic_cycle_length + self.current_cycle_phase) % 1.0
        economic_cycle = pattern.economic_sensitivity * 0.1 * np.sin(2 * np.pi * cycle_phase)
        
        # Add some economic shocks
        shock_probability = 0.02  # 2% chance per month
        shocks = np.random.choice([0, 1], size=months, p=[1-shock_probability, shock_probability])
        shock_magnitude = np.random.normal(0, 0.05, months)
        economic_shocks = shocks * shock_magnitude * pattern.economic_sensitivity
        
        return economic_cycle + economic_shocks
    
    def _generate_compliance_events(self, months: int, pattern: RiskPattern) -> np.ndarray:
        """Generate compliance events and their impact."""
        events = np.zeros(months)
        
        for event_type in pattern.compliance_events:
            # Generate event occurrences
            event_occurrences = np.random.poisson(event_type['frequency'] * months)
            
            for _ in range(event_occurrences):
                # Random month for the event
                event_month = np.random.randint(0, months)
                
                # Event impact (can last multiple months)
                impact_duration = np.random.randint(1, 4)  # 1-3 months
                impact_magnitude = event_type['impact'] * np.random.uniform(0.5, 1.5)
                
                # Apply impact
                for i in range(impact_duration):
                    if event_month + i < months:
                        # Decay impact over time
                        decay_factor = 1.0 - (i / impact_duration) * 0.5
                        events[event_month + i] += impact_magnitude * decay_factor
        
        return events
    
    def _get_size_adjustment(self, business_size: str) -> float:
        """Get risk adjustment based on business size."""
        size_adjustments = {
            "small": 0.05,    # Small businesses are riskier
            "medium": 0.0,    # Baseline
            "large": -0.03    # Large businesses are less risky
        }
        return size_adjustments.get(business_size, 0.0)
    
    def _generate_additional_features(self, months: int, industry: str, business_size: str) -> Dict[str, np.ndarray]:
        """Generate additional features for the time series."""
        features = {}
        
        # Financial health (correlated with risk score)
        base_health = 0.7 if business_size == "large" else 0.5
        features['financial_health'] = np.random.normal(base_health, 0.1, months)
        features['financial_health'] = np.clip(features['financial_health'], 0.0, 1.0)
        
        # Compliance score
        base_compliance = 0.8
        features['compliance_score'] = np.random.normal(base_compliance, 0.05, months)
        features['compliance_score'] = np.clip(features['compliance_score'], 0.0, 1.0)
        
        # Market conditions
        features['market_conditions'] = np.random.normal(0.5, 0.15, months)
        features['market_conditions'] = np.clip(features['market_conditions'], 0.0, 1.0)
        
        # Revenue trend
        base_revenue = 1000000 if business_size == "large" else 100000
        features['revenue_trend'] = np.random.normal(base_revenue, base_revenue * 0.2, months)
        features['revenue_trend'] = np.maximum(features['revenue_trend'], base_revenue * 0.1)
        
        # Employee count
        base_employees = 1000 if business_size == "large" else 50
        features['employee_count'] = np.random.normal(base_employees, base_employees * 0.1, months)
        features['employee_count'] = np.maximum(features['employee_count'], 1)
        
        # Debt ratio
        features['debt_ratio'] = np.random.normal(0.3, 0.1, months)
        features['debt_ratio'] = np.clip(features['debt_ratio'], 0.0, 1.0)
        
        # Profit margin
        features['profit_margin'] = np.random.normal(0.15, 0.05, months)
        features['profit_margin'] = np.clip(features['profit_margin'], -0.1, 0.4)
        
        # Customer satisfaction
        features['customer_satisfaction'] = np.random.normal(0.75, 0.1, months)
        features['customer_satisfaction'] = np.clip(features['customer_satisfaction'], 0.0, 1.0)
        
        # Regulatory compliance
        features['regulatory_compliance'] = np.random.normal(0.85, 0.05, months)
        features['regulatory_compliance'] = np.clip(features['regulatory_compliance'], 0.0, 1.0)
        
        # Market share
        base_share = 0.1 if business_size == "large" else 0.01
        features['market_share'] = np.random.normal(base_share, base_share * 0.2, months)
        features['market_share'] = np.clip(features['market_share'], 0.0, 1.0)
        
        # Innovation index
        features['innovation_index'] = np.random.normal(0.6, 0.1, months)
        features['innovation_index'] = np.clip(features['innovation_index'], 0.0, 1.0)
        
        return features
    
    def generate_dataset(self, 
                        num_businesses: int = 10000,
                        start_date: datetime = None,
                        months: int = 36) -> pd.DataFrame:
        """
        Generate a complete dataset of business time series.
        
        Args:
            num_businesses: Number of businesses to generate
            start_date: Start date for all time series
            months: Number of months per business
            
        Returns:
            Combined DataFrame with all business time series
        """
        if start_date is None:
            start_date = datetime.now() - timedelta(days=months * 30)
        
        industries = list(self.industry_patterns.keys())
        business_sizes = ["small", "medium", "large"]
        
        all_data = []
        
        for i in range(num_businesses):
            # Random selection
            industry = np.random.choice(industries)
            business_size = np.random.choice(business_sizes, p=[0.4, 0.4, 0.2])  # More small/medium
            
            # Generate business ID
            business_id = f"biz_{i:06d}"
            
            # Generate time series
            business_data = self.generate_business_time_series(
                business_id=business_id,
                industry=industry,
                start_date=start_date,
                months=months,
                business_size=business_size
            )
            
            all_data.append(business_data)
            
            if (i + 1) % 1000 == 0:
                print(f"Generated {i + 1}/{num_businesses} businesses")
        
        # Combine all data
        combined_data = pd.concat(all_data, ignore_index=True)
        
        # Add some additional derived features
        combined_data['risk_level'] = combined_data['risk_score'].apply(self._risk_score_to_level)
        combined_data['month'] = combined_data['date'].dt.month
        combined_data['quarter'] = combined_data['date'].dt.quarter
        combined_data['year'] = combined_data['date'].dt.year
        
        return combined_data
    
    def _risk_score_to_level(self, score: float) -> str:
        """Convert risk score to risk level."""
        if score < 0.25:
            return "low"
        elif score < 0.5:
            return "medium"
        elif score < 0.75:
            return "high"
        else:
            return "critical"
    
    def save_dataset(self, dataset: pd.DataFrame, filepath: str):
        """Save dataset to file."""
        if filepath.endswith('.parquet'):
            dataset.to_parquet(filepath, index=False)
        elif filepath.endswith('.csv'):
            dataset.to_csv(filepath, index=False)
        else:
            raise ValueError("Unsupported file format. Use .csv or .parquet")
        
        print(f"Dataset saved to {filepath}")
        print(f"Shape: {dataset.shape}")
        print(f"Date range: {dataset['date'].min()} to {dataset['date'].max()}")
        print(f"Industries: {dataset['industry'].unique()}")
        print(f"Risk level distribution:")
        print(dataset['risk_level'].value_counts())


def main():
    """Generate synthetic dataset for training."""
    generator = SyntheticDataGenerator(seed=42)
    
    print("Generating synthetic risk assessment dataset...")
    print("This may take a few minutes for large datasets...")
    
    # Generate dataset
    dataset = generator.generate_dataset(
        num_businesses=10000,
        months=36  # 3 years of data
    )
    
    # Save dataset
    output_path = "synthetic_risk_data.parquet"
    generator.save_dataset(dataset, output_path)
    
    print("\nDataset generation complete!")
    print(f"Total records: {len(dataset):,}")
    print(f"Unique businesses: {dataset['business_id'].nunique():,}")
    print(f"Time span: {dataset['date'].min()} to {dataset['date'].max()}")


if __name__ == "__main__":
    main()
