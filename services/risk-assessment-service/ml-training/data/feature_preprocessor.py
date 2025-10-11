"""
Feature Preprocessor for LSTM Risk Prediction

Handles data preprocessing, feature engineering, and sequence preparation
for LSTM time-series risk prediction models.
"""

import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Optional
from sklearn.preprocessing import StandardScaler, MinMaxScaler, LabelEncoder
from sklearn.model_selection import train_test_split
import pickle
import json


class FeaturePreprocessor:
    """Preprocesses features for LSTM risk prediction."""
    
    def __init__(self, sequence_length: int = 12, prediction_horizons: List[int] = [6, 9, 12]):
        self.sequence_length = sequence_length
        self.prediction_horizons = prediction_horizons
        
        # Scalers for different feature types
        self.feature_scalers = {}
        self.label_encoders = {}
        
        # Feature groups
        self.numerical_features = [
            'risk_score', 'financial_health', 'compliance_score', 'market_conditions',
            'revenue_trend', 'employee_count', 'debt_ratio', 'profit_margin',
            'customer_satisfaction', 'regulatory_compliance', 'market_share', 'innovation_index'
        ]
        
        self.categorical_features = ['industry', 'risk_level']
        
        self.temporal_features = ['month', 'quarter', 'year']
        
        # All features combined
        self.all_features = self.numerical_features + self.categorical_features + self.temporal_features
    
    def fit_transform(self, data: pd.DataFrame) -> Tuple[np.ndarray, Dict[str, np.ndarray]]:
        """
        Fit scalers and transform the data.
        
        Args:
            data: Raw dataset
            
        Returns:
            Tuple of (processed_data, targets_dict)
        """
        print("Fitting scalers and preprocessing data...")
        
        # Prepare features
        processed_data = self._prepare_features(data)
        
        # Fit scalers
        self._fit_scalers(processed_data)
        
        # Transform data
        transformed_data = self._transform_data(processed_data)
        
        # Create sequences and targets
        sequences, targets = self._create_sequences_and_targets(transformed_data)
        
        print(f"Created {len(sequences)} sequences of length {self.sequence_length}")
        print(f"Target shapes: {[(k, v.shape) for k, v in targets.items()]}")
        
        return sequences, targets
    
    def transform(self, data: pd.DataFrame) -> Tuple[np.ndarray, Dict[str, np.ndarray]]:
        """
        Transform data using fitted scalers.
        
        Args:
            data: Raw dataset
            
        Returns:
            Tuple of (processed_data, targets_dict)
        """
        if not self.feature_scalers:
            raise ValueError("Scalers not fitted. Call fit_transform first.")
        
        # Prepare features
        processed_data = self._prepare_features(data)
        
        # Transform data
        transformed_data = self._transform_data(processed_data)
        
        # Create sequences and targets
        sequences, targets = self._create_sequences_and_targets(transformed_data)
        
        return sequences, targets
    
    def _prepare_features(self, data: pd.DataFrame) -> pd.DataFrame:
        """Prepare and clean features."""
        processed = data.copy()
        
        # Handle missing values
        processed = self._handle_missing_values(processed)
        
        # Create derived features
        processed = self._create_derived_features(processed)
        
        # Ensure all required features exist
        for feature in self.all_features:
            if feature not in processed.columns:
                if feature in self.numerical_features:
                    processed[feature] = 0.0
                elif feature in self.categorical_features:
                    processed[feature] = 'unknown'
                else:
                    processed[feature] = 0
        
        return processed
    
    def _handle_missing_values(self, data: pd.DataFrame) -> pd.DataFrame:
        """Handle missing values in the dataset."""
        processed = data.copy()
        
        # Fill numerical features with median
        for feature in self.numerical_features:
            if feature in processed.columns:
                processed[feature] = processed[feature].fillna(processed[feature].median())
        
        # Fill categorical features with mode
        for feature in self.categorical_features:
            if feature in processed.columns:
                processed[feature] = processed[feature].fillna(processed[feature].mode()[0] if not processed[feature].mode().empty else 'unknown')
        
        return processed
    
    def _create_derived_features(self, data: pd.DataFrame) -> pd.DataFrame:
        """Create additional derived features."""
        processed = data.copy()
        
        # Risk score momentum (change over time)
        processed['risk_momentum'] = processed.groupby('business_id')['risk_score'].diff().fillna(0)
        
        # Financial health trend
        processed['financial_trend'] = processed.groupby('business_id')['financial_health'].diff().fillna(0)
        
        # Revenue growth rate
        processed['revenue_growth'] = processed.groupby('business_id')['revenue_trend'].pct_change().fillna(0)
        
        # Employee growth rate
        processed['employee_growth'] = processed.groupby('business_id')['employee_count'].pct_change().fillna(0)
        
        # Risk volatility (rolling standard deviation)
        processed['risk_volatility'] = processed.groupby('business_id')['risk_score'].rolling(window=3, min_periods=1).std().fillna(0).values
        
        # Compliance trend
        processed['compliance_trend'] = processed.groupby('business_id')['compliance_score'].diff().fillna(0)
        
        # Market position (relative to industry)
        industry_avg = processed.groupby(['industry', 'date'])['market_share'].transform('mean')
        processed['market_position'] = processed['market_share'] / (industry_avg + 1e-8)
        
        # Add derived features to feature lists
        derived_features = ['risk_momentum', 'financial_trend', 'revenue_growth', 
                          'employee_growth', 'risk_volatility', 'compliance_trend', 'market_position']
        self.numerical_features.extend(derived_features)
        
        return processed
    
    def _fit_scalers(self, data: pd.DataFrame):
        """Fit scalers for different feature types."""
        # Fit numerical feature scaler
        numerical_data = data[self.numerical_features].values
        self.feature_scalers['numerical'] = StandardScaler()
        self.feature_scalers['numerical'].fit(numerical_data)
        
        # Fit label encoders for categorical features
        for feature in self.categorical_features:
            if feature in data.columns:
                self.label_encoders[feature] = LabelEncoder()
                self.label_encoders[feature].fit(data[feature].astype(str))
        
        # Fit temporal feature scaler
        temporal_data = data[self.temporal_features].values
        self.feature_scalers['temporal'] = MinMaxScaler()
        self.feature_scalers['temporal'].fit(temporal_data)
    
    def _transform_data(self, data: pd.DataFrame) -> pd.DataFrame:
        """Transform data using fitted scalers."""
        processed = data.copy()
        
        # Transform numerical features
        numerical_data = processed[self.numerical_features].values
        processed[self.numerical_features] = self.feature_scalers['numerical'].transform(numerical_data)
        
        # Transform categorical features
        for feature in self.categorical_features:
            if feature in processed.columns and feature in self.label_encoders:
                processed[feature] = self.label_encoders[feature].transform(processed[feature].astype(str))
        
        # Transform temporal features
        temporal_data = processed[self.temporal_features].values
        processed[self.temporal_features] = self.feature_scalers['temporal'].transform(temporal_data)
        
        return processed
    
    def _create_sequences_and_targets(self, data: pd.DataFrame) -> Tuple[np.ndarray, Dict[str, np.ndarray]]:
        """Create sequences and targets for LSTM training."""
        sequences = []
        targets = {f'horizon_{h}': [] for h in self.prediction_horizons}
        
        # Group by business
        for business_id, business_data in data.groupby('business_id'):
            business_data = business_data.sort_values('date')
            
            if len(business_data) < self.sequence_length + max(self.prediction_horizons):
                continue  # Skip businesses with insufficient data
            
            # Create sequences
            for i in range(len(business_data) - self.sequence_length - max(self.prediction_horizons) + 1):
                # Input sequence
                sequence = business_data.iloc[i:i + self.sequence_length]
                
                # Extract features for the sequence
                feature_columns = self.numerical_features + self.categorical_features + self.temporal_features
                sequence_features = sequence[feature_columns].values
                sequences.append(sequence_features)
                
                # Create targets for each horizon
                for horizon in self.prediction_horizons:
                    target_idx = i + self.sequence_length + horizon - 1
                    if target_idx < len(business_data):
                        target_risk_score = business_data.iloc[target_idx]['risk_score']
                        targets[f'horizon_{horizon}'].append(target_risk_score)
                    else:
                        # If target is beyond available data, use last available risk score
                        targets[f'horizon_{horizon}'].append(business_data.iloc[-1]['risk_score'])
        
        # Convert to numpy arrays
        sequences = np.array(sequences)
        for horizon in self.prediction_horizons:
            targets[f'horizon_{horizon}'] = np.array(targets[f'horizon_{horizon}'])
        
        return sequences, targets
    
    def split_data(self, sequences: np.ndarray, targets: Dict[str, np.ndarray], 
                   test_size: float = 0.2, val_size: float = 0.1) -> Dict[str, Tuple[np.ndarray, np.ndarray]]:
        """
        Split data into train, validation, and test sets.
        
        Args:
            sequences: Input sequences
            targets: Target values for each horizon
            test_size: Fraction for test set
            val_size: Fraction for validation set (from remaining data)
            
        Returns:
            Dictionary with train/val/test splits
        """
        # First split: train+val vs test
        train_val_sequences, test_sequences, train_val_targets, test_targets = train_test_split(
            sequences, targets, test_size=test_size, random_state=42
        )
        
        # Second split: train vs val
        val_size_adjusted = val_size / (1 - test_size)  # Adjust val_size for remaining data
        train_sequences, val_sequences, train_targets, val_targets = train_test_split(
            train_val_sequences, train_val_targets, test_size=val_size_adjusted, random_state=42
        )
        
        return {
            'train': (train_sequences, train_targets),
            'val': (val_sequences, val_targets),
            'test': (test_sequences, test_targets)
        }
    
    def save_preprocessor(self, filepath: str):
        """Save the fitted preprocessor."""
        preprocessor_data = {
            'sequence_length': self.sequence_length,
            'prediction_horizons': self.prediction_horizons,
            'numerical_features': self.numerical_features,
            'categorical_features': self.categorical_features,
            'temporal_features': self.temporal_features,
            'feature_scalers': self.feature_scalers,
            'label_encoders': self.label_encoders
        }
        
        with open(filepath, 'wb') as f:
            pickle.dump(preprocessor_data, f)
        
        print(f"Preprocessor saved to {filepath}")
    
    def load_preprocessor(self, filepath: str):
        """Load a fitted preprocessor."""
        with open(filepath, 'rb') as f:
            preprocessor_data = pickle.load(f)
        
        self.sequence_length = preprocessor_data['sequence_length']
        self.prediction_horizons = preprocessor_data['prediction_horizons']
        self.numerical_features = preprocessor_data['numerical_features']
        self.categorical_features = preprocessor_data['categorical_features']
        self.temporal_features = preprocessor_data['temporal_features']
        self.feature_scalers = preprocessor_data['feature_scalers']
        self.label_encoders = preprocessor_data['label_encoders']
        
        print(f"Preprocessor loaded from {filepath}")
    
    def get_feature_info(self) -> Dict:
        """Get information about features and preprocessing."""
        return {
            'sequence_length': self.sequence_length,
            'prediction_horizons': self.prediction_horizons,
            'total_features': len(self.numerical_features) + len(self.categorical_features) + len(self.temporal_features),
            'numerical_features': len(self.numerical_features),
            'categorical_features': len(self.categorical_features),
            'temporal_features': len(self.temporal_features),
            'feature_names': {
                'numerical': self.numerical_features,
                'categorical': self.categorical_features,
                'temporal': self.temporal_features
            }
        }


def main():
    """Test the feature preprocessor."""
    # This would typically be called with real data
    print("Feature preprocessor ready for use.")
    print("Use with synthetic data generator to create training sequences.")


if __name__ == "__main__":
    main()
