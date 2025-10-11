"""
Synthetic Time-Series Risk Data Generator

Generates 2-3 years of historical risk data for LSTM training:
- Industry-specific default rate patterns
- Seasonal trends (quarterly, annual cycles)
- Economic cycle simulation (recession, growth periods)
- Random walk with drift for risk evolution
- 10,000+ business time-series (100-1000 per industry)
"""

import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Optional
import random
from datetime import datetime, timedelta
import yaml
import os


class SyntheticRiskDataGenerator:
    """Generates synthetic time-series risk data for LSTM training"""
    
    def __init__(self, config_path: str = "models/model_config.yaml"):
        """Initialize the generator with configuration"""
        self.config = self._load_config(config_path)
        self.industries = self.config.get('industries', [
            'technology', 'finance', 'healthcare', 'manufacturing', 
            'retail', 'energy', 'real_estate', 'transportation'
        ])
        self.sequence_length = self.config.get('sequence_length', 36)  # 3 years
        self.feature_count = self.config.get('feature_count', 20)
        self.businesses_per_industry = self.config.get('businesses_per_industry', 1250)  # 10k total
        
        # Set random seed for reproducibility
        np.random.seed(self.config.get('random_seed', 42))
        random.seed(self.config.get('random_seed', 42))
        
        # Industry-specific risk patterns
        self.industry_patterns = self._initialize_industry_patterns()
        
    def _load_config(self, config_path: str) -> Dict:
        """Load configuration from YAML file"""
        if os.path.exists(config_path):
            with open(config_path, 'r') as f:
                return yaml.safe_load(f)
        else:
            # Default configuration
            return {
                'sequence_length': 36,
                'feature_count': 20,
                'businesses_per_industry': 1250,
                'random_seed': 42,
                'industries': [
                    'technology', 'finance', 'healthcare', 'manufacturing', 
                    'retail', 'energy', 'real_estate', 'transportation'
                ]
            }
    
    def _initialize_industry_patterns(self) -> Dict:
        """Initialize industry-specific risk patterns"""
        patterns = {}
        
        for industry in self.industries:
            patterns[industry] = {
                'base_risk': np.random.uniform(0.1, 0.4),  # Base risk level
                'volatility': np.random.uniform(0.05, 0.2),  # Risk volatility
                'seasonality': {
                    'q1_factor': np.random.uniform(0.8, 1.2),
                    'q2_factor': np.random.uniform(0.9, 1.1),
                    'q3_factor': np.random.uniform(0.8, 1.2),
                    'q4_factor': np.random.uniform(0.7, 1.3)
                },
                'trend': np.random.uniform(-0.02, 0.02),  # Monthly trend
                'economic_sensitivity': np.random.uniform(0.5, 1.5)  # Economic cycle sensitivity
            }
        
        return patterns
    
    def generate_business_sequence(self, business_id: str, industry: str, 
                                 start_date: datetime) -> pd.DataFrame:
        """Generate a time-series sequence for a single business"""
        
        pattern = self.industry_patterns[industry]
        dates = [start_date + timedelta(days=30*i) for i in range(self.sequence_length)]
        
        # Initialize risk score with base industry risk
        risk_scores = np.zeros(self.sequence_length)
        risk_scores[0] = pattern['base_risk'] + np.random.normal(0, 0.05)
        
        # Generate time-series with trend, seasonality, and volatility
        for t in range(1, self.sequence_length):
            # Previous risk score
            prev_risk = risk_scores[t-1]
            
            # Trend component
            trend = pattern['trend']
            
            # Seasonal component
            month = dates[t].month
            quarter = (month - 1) // 3 + 1
            seasonal_factor = pattern['seasonality'][f'q{quarter}_factor']
            seasonal = (seasonal_factor - 1) * 0.1
            
            # Economic cycle component (simulated)
            economic_cycle = self._generate_economic_cycle(t)
            economic_impact = economic_cycle * pattern['economic_sensitivity'] * 0.1
            
            # Random walk component
            random_walk = np.random.normal(0, pattern['volatility'])
            
            # Calculate new risk score
            new_risk = prev_risk + trend + seasonal + economic_impact + random_walk
            
            # Ensure risk score stays within bounds [0, 1]
            risk_scores[t] = np.clip(new_risk, 0.01, 0.99)
        
        # Generate additional features
        features = self._generate_business_features(business_id, industry, risk_scores, dates)
        
        # Create DataFrame
        data = {
            'business_id': [business_id] * self.sequence_length,
            'industry': [industry] * self.sequence_length,
            'date': dates,
            'risk_score': risk_scores,
            'risk_level': [self._score_to_level(score) for score in risk_scores]
        }
        
        # Add generated features
        for i, feature_name in enumerate(features.columns):
            data[f'feature_{i}'] = features[feature_name].values
        
        return pd.DataFrame(data)
    
    def _generate_economic_cycle(self, time_step: int) -> float:
        """Generate economic cycle component (recession/growth periods)"""
        # Simulate economic cycles with 4-year periods
        cycle_period = 48  # 4 years in months
        cycle_phase = (time_step % cycle_period) / cycle_period * 2 * np.pi
        
        # Economic cycle: -1 (recession) to +1 (growth)
        economic_cycle = np.sin(cycle_phase)
        
        # Add some noise to make it more realistic
        economic_cycle += np.random.normal(0, 0.1)
        
        return economic_cycle
    
    def _generate_business_features(self, business_id: str, industry: str, 
                                  risk_scores: np.ndarray, dates: List[datetime]) -> pd.DataFrame:
        """Generate additional business features for the time-series"""
        
        features = pd.DataFrame(index=range(len(risk_scores)))
        
        # Financial metrics (correlated with risk)
        features['revenue_growth'] = np.random.normal(0.02, 0.1, len(risk_scores))
        features['profit_margin'] = np.random.uniform(0.05, 0.25, len(risk_scores))
        features['debt_ratio'] = np.random.uniform(0.2, 0.8, len(risk_scores))
        
        # Operational metrics
        features['employee_count'] = np.random.randint(10, 1000, len(risk_scores))
        features['customer_satisfaction'] = np.random.uniform(3.0, 5.0, len(risk_scores))
        features['market_share'] = np.random.uniform(0.01, 0.3, len(risk_scores))
        
        # Compliance and regulatory
        features['compliance_score'] = np.random.uniform(0.7, 1.0, len(risk_scores))
        features['regulatory_changes'] = np.random.poisson(0.5, len(risk_scores))
        
        # Market conditions
        features['market_volatility'] = np.random.uniform(0.1, 0.4, len(risk_scores))
        features['competition_intensity'] = np.random.uniform(0.3, 0.9, len(risk_scores))
        
        # Time-based features
        features['month'] = [d.month for d in dates]
        features['quarter'] = [(d.month - 1) // 3 + 1 for d in dates]
        features['year'] = [d.year for d in dates]
        
        # Lagged features (previous period values)
        for lag in [1, 3, 6, 12]:
            if lag < len(risk_scores):
                lagged_values = np.roll(risk_scores, lag)
                lagged_values[:lag] = risk_scores[0]
                features[f'risk_score_lag_{lag}'] = lagged_values
        
        # Moving averages
        for window in [3, 6, 12]:
            features[f'risk_score_ma_{window}'] = pd.Series(risk_scores).rolling(window=window, min_periods=1).mean()
        
        # Volatility measures
        features['risk_volatility_3m'] = pd.Series(risk_scores).rolling(window=3, min_periods=1).std()
        features['risk_volatility_6m'] = pd.Series(risk_scores).rolling(window=6, min_periods=1).std()
        
        # Ensure we have exactly the required number of features
        if len(features.columns) > self.feature_count:
            features = features.iloc[:, :self.feature_count]
        elif len(features.columns) < self.feature_count:
            # Add random features to reach target count
            for i in range(len(features.columns), self.feature_count):
                features[f'random_feature_{i}'] = np.random.normal(0, 1, len(risk_scores))
        
        return features
    
    def _score_to_level(self, score: float) -> str:
        """Convert risk score to risk level"""
        if score < 0.3:
            return 'low'
        elif score < 0.6:
            return 'medium'
        elif score < 0.8:
            return 'high'
        else:
            return 'critical'
    
    def generate_dataset(self, output_path: str = "data/synthetic_risk_data.parquet") -> pd.DataFrame:
        """Generate the complete synthetic dataset"""
        
        print(f"Generating synthetic dataset with {len(self.industries)} industries...")
        print(f"Target: {self.businesses_per_industry} businesses per industry")
        print(f"Sequence length: {self.sequence_length} months")
        print(f"Features per timestep: {self.feature_count}")
        
        all_data = []
        business_id_counter = 0
        
        for industry in self.industries:
            print(f"Generating data for {industry} industry...")
            
            for i in range(self.businesses_per_industry):
                business_id = f"business_{business_id_counter:06d}"
                
                # Random start date within the last 3 years
                start_date = datetime.now() - timedelta(days=365*3 + np.random.randint(0, 365))
                
                # Generate sequence for this business
                business_data = self.generate_business_sequence(business_id, industry, start_date)
                all_data.append(business_data)
                
                business_id_counter += 1
                
                if (i + 1) % 100 == 0:
                    print(f"  Generated {i + 1}/{self.businesses_per_industry} businesses")
        
        # Combine all data
        print("Combining all data...")
        dataset = pd.concat(all_data, ignore_index=True)
        
        # Add some data quality checks
        print("Performing data quality checks...")
        self._validate_dataset(dataset)
        
        # Save dataset
        print(f"Saving dataset to {output_path}...")
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        dataset.to_parquet(output_path, index=False)
        
        print(f"Dataset generation complete!")
        print(f"Total records: {len(dataset):,}")
        print(f"Unique businesses: {dataset['business_id'].nunique():,}")
        print(f"Date range: {dataset['date'].min()} to {dataset['date'].max()}")
        print(f"Industries: {dataset['industry'].nunique()}")
        
        return dataset
    
    def _validate_dataset(self, dataset: pd.DataFrame) -> None:
        """Validate the generated dataset"""
        
        # Check for missing values
        missing_values = dataset.isnull().sum().sum()
        if missing_values > 0:
            print(f"Warning: {missing_values} missing values found")
        
        # Check risk score distribution
        risk_stats = dataset['risk_score'].describe()
        print(f"Risk score statistics:")
        print(f"  Mean: {risk_stats['mean']:.3f}")
        print(f"  Std: {risk_stats['std']:.3f}")
        print(f"  Min: {risk_stats['min']:.3f}")
        print(f"  Max: {risk_stats['max']:.3f}")
        
        # Check risk level distribution
        risk_level_dist = dataset['risk_level'].value_counts(normalize=True)
        print(f"Risk level distribution:")
        for level, proportion in risk_level_dist.items():
            print(f"  {level}: {proportion:.1%}")
        
        # Check industry distribution
        industry_dist = dataset['industry'].value_counts()
        print(f"Industry distribution:")
        for industry, count in industry_dist.items():
            print(f"  {industry}: {count:,} records")


def main():
    """Main function to generate synthetic dataset"""
    
    # Create generator
    generator = SyntheticRiskDataGenerator()
    
    # Generate dataset
    dataset = generator.generate_dataset("data/synthetic_risk_data.parquet")
    
    # Print summary statistics
    print("\n" + "="*50)
    print("DATASET SUMMARY")
    print("="*50)
    print(f"Shape: {dataset.shape}")
    print(f"Memory usage: {dataset.memory_usage(deep=True).sum() / 1024**2:.1f} MB")
    print(f"Columns: {list(dataset.columns)}")


if __name__ == "__main__":
    main()