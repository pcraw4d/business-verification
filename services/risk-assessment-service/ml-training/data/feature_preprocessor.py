"""
Feature Preprocessor for LSTM Training

Handles data preprocessing, normalization, and feature engineering
for the LSTM time-series risk prediction model.
"""

import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Optional
from sklearn.preprocessing import StandardScaler, MinMaxScaler, LabelEncoder
from sklearn.model_selection import train_test_split
import pickle
import os


class FeaturePreprocessor:
    """Preprocesses time-series data for LSTM training"""
    
    def __init__(self, config: Dict):
        """Initialize preprocessor with configuration"""
        self.config = config
        self.sequence_length = config.get('sequence_length', 12)
        self.feature_count = config.get('feature_count', 20)
        self.prediction_horizons = config.get('prediction_horizons', [6, 9, 12])
        
        # Scalers for different feature types
        self.feature_scaler = StandardScaler()
        self.target_scaler = MinMaxScaler()
        self.industry_encoder = LabelEncoder()
        
        # Feature statistics for normalization
        self.feature_stats = {}
        self.is_fitted = False
    
    def fit_transform(self, data: pd.DataFrame) -> Tuple[np.ndarray, Dict[str, np.ndarray]]:
        """Fit preprocessors and transform the data"""
        
        print("Fitting preprocessors...")
        
        # Prepare features and targets
        features, targets = self._prepare_features_and_targets(data)
        
        # Fit scalers
        self.feature_scaler.fit(features.reshape(-1, features.shape[-1]))
        self.target_scaler.fit(targets.reshape(-1, len(self.prediction_horizons)))
        
        # Fit industry encoder
        self.industry_encoder.fit(data['industry'].unique())
        
        # Store feature statistics
        self._compute_feature_stats(features)
        
        # Transform data
        print("Transforming data...")
        features_scaled = self._scale_features(features)
        targets_scaled = self._scale_targets(targets)
        
        self.is_fitted = True
        
        return features_scaled, {
            'targets': targets_scaled,
            'business_ids': data['business_id'].unique(),
            'industries': data['industry'].unique()
        }
    
    def transform(self, data: pd.DataFrame) -> Tuple[np.ndarray, Dict[str, np.ndarray]]:
        """Transform new data using fitted preprocessors"""
        
        if not self.is_fitted:
            raise ValueError("Preprocessor must be fitted before transform")
        
        # Prepare features and targets
        features, targets = self._prepare_features_and_targets(data)
        
        # Transform data
        features_scaled = self._scale_features(features)
        targets_scaled = self._scale_targets(targets)
        
        return features_scaled, {
            'targets': targets_scaled,
            'business_ids': data['business_id'].unique(),
            'industries': data['industry'].unique()
        }
    
    def _prepare_features_and_targets(self, data: pd.DataFrame) -> Tuple[np.ndarray, np.ndarray]:
        """Prepare features and targets for LSTM training"""
        
        print("Preparing features and targets...")
        
        # Get unique businesses
        businesses = data['business_id'].unique()
        n_businesses = len(businesses)
        
        # Initialize arrays
        features = np.zeros((n_businesses, self.sequence_length, self.feature_count))
        targets = np.zeros((n_businesses, len(self.prediction_horizons)))
        
        for i, business_id in enumerate(businesses):
            business_data = data[data['business_id'] == business_id].sort_values('date')
            
            if len(business_data) < self.sequence_length:
                # Pad with zeros if insufficient data
                padding_length = self.sequence_length - len(business_data)
                padded_data = pd.concat([
                    pd.DataFrame({
                        'risk_score': [business_data['risk_score'].iloc[0]] * padding_length,
                        **{f'feature_{j}': [0] * padding_length for j in range(self.feature_count)}
                    }),
                    business_data
                ])
            else:
                # Take the last sequence_length records
                padded_data = business_data.tail(self.sequence_length)
            
            # Extract features
            feature_columns = [f'feature_{j}' for j in range(self.feature_count)]
            if all(col in padded_data.columns for col in feature_columns):
                features[i] = padded_data[feature_columns].values
            else:
                # Generate synthetic features if not available
                features[i] = self._generate_synthetic_features(padded_data)
            
            # Extract targets (future risk scores)
            targets[i] = self._extract_targets(business_data)
        
        return features, targets
    
    def _generate_synthetic_features(self, data: pd.DataFrame) -> np.ndarray:
        """Generate synthetic features if not available in data"""
        
        features = np.zeros((len(data), self.feature_count))
        
        # Use risk score and date information to generate features
        risk_scores = data['risk_score'].values
        dates = pd.to_datetime(data['date'])
        
        # Basic features
        features[:, 0] = risk_scores  # Current risk score
        features[:, 1] = np.roll(risk_scores, 1)  # Previous risk score
        features[:, 1][0] = risk_scores[0]  # Handle first value
        
        # Time-based features
        features[:, 2] = dates.month.values / 12.0  # Month normalized
        features[:, 3] = dates.quarter.values / 4.0  # Quarter normalized
        features[:, 4] = (dates.year - dates.year.min()) / (dates.year.max() - dates.year.min())  # Year normalized
        
        # Statistical features
        for i in range(5, min(10, self.feature_count)):
            window = min(i - 4, len(risk_scores))
            if window > 1:
                features[:, i] = pd.Series(risk_scores).rolling(window=window, min_periods=1).mean()
            else:
                features[:, i] = risk_scores
        
        # Volatility features
        for i in range(10, min(15, self.feature_count)):
            window = min(i - 9, len(risk_scores))
            if window > 1:
                features[:, i] = pd.Series(risk_scores).rolling(window=window, min_periods=1).std()
            else:
                features[:, i] = 0
        
        # Random features to fill remaining slots
        for i in range(15, self.feature_count):
            features[:, i] = np.random.normal(0, 1, len(risk_scores))
        
        return features
    
    def _extract_targets(self, business_data: pd.DataFrame) -> np.ndarray:
        """Extract target values for different prediction horizons"""
        
        targets = np.zeros(len(self.prediction_horizons))
        risk_scores = business_data['risk_score'].values
        
        for i, horizon in enumerate(self.prediction_horizons):
            if len(risk_scores) > horizon:
                # Use actual future value if available
                targets[i] = risk_scores[horizon]
            else:
                # Extrapolate based on trend
                if len(risk_scores) > 1:
                    trend = np.mean(np.diff(risk_scores))
                    targets[i] = risk_scores[-1] + trend * horizon
                else:
                    targets[i] = risk_scores[0]
        
        return targets
    
    def _scale_features(self, features: np.ndarray) -> np.ndarray:
        """Scale features using fitted scaler"""
        
        original_shape = features.shape
        features_reshaped = features.reshape(-1, features.shape[-1])
        features_scaled = self.feature_scaler.transform(features_reshaped)
        return features_scaled.reshape(original_shape)
    
    def _scale_targets(self, targets: np.ndarray) -> np.ndarray:
        """Scale targets using fitted scaler"""
        
        original_shape = targets.shape
        targets_reshaped = targets.reshape(-1, targets.shape[-1])
        targets_scaled = self.target_scaler.transform(targets_reshaped)
        return targets_scaled.reshape(original_shape)
    
    def inverse_transform_targets(self, targets_scaled: np.ndarray) -> np.ndarray:
        """Inverse transform scaled targets back to original scale"""
        
        original_shape = targets_scaled.shape
        targets_reshaped = targets_scaled.reshape(-1, targets_scaled.shape[-1])
        targets_original = self.target_scaler.inverse_transform(targets_reshaped)
        return targets_original.reshape(original_shape)
    
    def _compute_feature_stats(self, features: np.ndarray) -> None:
        """Compute and store feature statistics"""
        
        self.feature_stats = {
            'mean': np.mean(features, axis=(0, 1)),
            'std': np.std(features, axis=(0, 1)),
            'min': np.min(features, axis=(0, 1)),
            'max': np.max(features, axis=(0, 1))
        }
    
    def create_sequences(self, features: np.ndarray, targets: np.ndarray, 
                        test_size: float = 0.2, val_size: float = 0.1) -> Dict:
        """Create train/validation/test splits for LSTM training"""
        
        print("Creating train/validation/test splits...")
        
        n_samples = len(features)
        indices = np.arange(n_samples)
        
        # Split indices
        train_idx, temp_idx = train_test_split(indices, test_size=test_size + val_size, random_state=42)
        val_idx, test_idx = train_test_split(temp_idx, test_size=test_size/(test_size + val_size), random_state=42)
        
        # Create splits
        splits = {
            'train': {
                'features': features[train_idx],
                'targets': targets[train_idx]
            },
            'validation': {
                'features': features[val_idx],
                'targets': targets[val_idx]
            },
            'test': {
                'features': features[test_idx],
                'targets': targets[test_idx]
            }
        }
        
        print(f"Train set: {len(train_idx):,} samples")
        print(f"Validation set: {len(val_idx):,} samples")
        print(f"Test set: {len(test_idx):,} samples")
        
        return splits
    
    def save_preprocessor(self, filepath: str) -> None:
        """Save fitted preprocessor to disk"""
        
        if not self.is_fitted:
            raise ValueError("Preprocessor must be fitted before saving")
        
        preprocessor_data = {
            'feature_scaler': self.feature_scaler,
            'target_scaler': self.target_scaler,
            'industry_encoder': self.industry_encoder,
            'feature_stats': self.feature_stats,
            'config': self.config,
            'is_fitted': self.is_fitted
        }
        
        with open(filepath, 'wb') as f:
            pickle.dump(preprocessor_data, f)
        
        print(f"Preprocessor saved to {filepath}")
    
    def load_preprocessor(self, filepath: str) -> None:
        """Load fitted preprocessor from disk"""
        
        with open(filepath, 'rb') as f:
            preprocessor_data = pickle.load(f)
        
        self.feature_scaler = preprocessor_data['feature_scaler']
        self.target_scaler = preprocessor_data['target_scaler']
        self.industry_encoder = preprocessor_data['industry_encoder']
        self.feature_stats = preprocessor_data['feature_stats']
        self.config = preprocessor_data['config']
        self.is_fitted = preprocessor_data['is_fitted']
        
        print(f"Preprocessor loaded from {filepath}")


def main():
    """Main function to test the preprocessor"""
    
    # Load sample data
    import pandas as pd
    
    # Create sample data for testing
    n_businesses = 100
    sequence_length = 12
    feature_count = 20
    
    sample_data = []
    for i in range(n_businesses):
        business_id = f"business_{i:06d}"
        industry = np.random.choice(['technology', 'finance', 'healthcare'])
        
        for t in range(sequence_length):
            sample_data.append({
                'business_id': business_id,
                'industry': industry,
                'date': pd.Timestamp('2023-01-01') + pd.Timedelta(days=30*t),
                'risk_score': np.random.uniform(0.1, 0.9),
                **{f'feature_{j}': np.random.normal(0, 1) for j in range(feature_count)}
            })
    
    sample_df = pd.DataFrame(sample_data)
    
    # Test preprocessor
    config = {
        'sequence_length': sequence_length,
        'feature_count': feature_count,
        'prediction_horizons': [6, 9, 12]
    }
    
    preprocessor = FeaturePreprocessor(config)
    features, targets = preprocessor.fit_transform(sample_df)
    
    print(f"Features shape: {features.shape}")
    print(f"Targets shape: {targets.shape}")
    print("Preprocessor test completed successfully!")


if __name__ == "__main__":
    main()