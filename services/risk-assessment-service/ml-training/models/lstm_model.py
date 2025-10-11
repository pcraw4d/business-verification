"""
LSTM Model for Risk Prediction

Multi-horizon LSTM model with attention mechanism for 6-12 month risk forecasting.
Includes uncertainty estimation via Monte Carlo dropout.
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
import numpy as np
from typing import Dict, List, Tuple, Optional
import math


class MultiHeadAttention(nn.Module):
    """Multi-head attention mechanism for LSTM outputs"""
    
    def __init__(self, hidden_dim: int, num_heads: int, dropout: float = 0.1):
        super().__init__()
        assert hidden_dim % num_heads == 0
        
        self.hidden_dim = hidden_dim
        self.num_heads = num_heads
        self.head_dim = hidden_dim // num_heads
        
        self.query = nn.Linear(hidden_dim, hidden_dim)
        self.key = nn.Linear(hidden_dim, hidden_dim)
        self.value = nn.Linear(hidden_dim, hidden_dim)
        self.dropout = nn.Dropout(dropout)
        self.out_proj = nn.Linear(hidden_dim, hidden_dim)
        
    def forward(self, x: torch.Tensor) -> Tuple[torch.Tensor, torch.Tensor]:
        batch_size, seq_len, hidden_dim = x.size()
        
        # Linear projections
        Q = self.query(x).view(batch_size, seq_len, self.num_heads, self.head_dim).transpose(1, 2)
        K = self.key(x).view(batch_size, seq_len, self.num_heads, self.head_dim).transpose(1, 2)
        V = self.value(x).view(batch_size, seq_len, self.num_heads, self.head_dim).transpose(1, 2)
        
        # Scaled dot-product attention
        scores = torch.matmul(Q, K.transpose(-2, -1)) / math.sqrt(self.head_dim)
        attention_weights = F.softmax(scores, dim=-1)
        attention_weights = self.dropout(attention_weights)
        
        # Apply attention to values
        context = torch.matmul(attention_weights, V)
        context = context.transpose(1, 2).contiguous().view(batch_size, seq_len, hidden_dim)
        
        # Output projection
        output = self.out_proj(context)
        
        return output, attention_weights.mean(dim=1)  # Average across heads


class RiskLSTM(nn.Module):
    """Multi-horizon LSTM model with attention for risk prediction"""
    
    def __init__(self, config: Dict):
        super().__init__()
        
        self.config = config
        self.sequence_length = config['data']['sequence_length']
        self.feature_count = config['data']['feature_count']
        self.prediction_horizons = config['data']['prediction_horizons']
        self.hidden_units = config['model']['hidden_units']
        self.lstm_layers = config['model']['lstm_layers']
        self.dropout = config['model']['dropout']
        self.attention_heads = config['model']['attention_heads']
        self.attention_dim = config['model']['attention_dim']
        
        # LSTM layers
        self.lstm = nn.LSTM(
            input_size=self.feature_count,
            hidden_size=self.hidden_units,
            num_layers=self.lstm_layers,
            dropout=self.dropout if self.lstm_layers > 1 else 0,
            batch_first=True,
            bidirectional=False
        )
        
        # Attention mechanism
        self.attention = MultiHeadAttention(
            hidden_dim=self.hidden_units,
            num_heads=self.attention_heads,
            dropout=self.dropout
        )
        
        # Dropout for uncertainty estimation
        self.mc_dropout = nn.Dropout(self.dropout)
        
        # Horizon-specific output heads
        self.output_heads = nn.ModuleDict()
        for horizon in self.prediction_horizons:
            self.output_heads[str(horizon)] = nn.Sequential(
                nn.Linear(self.hidden_units, self.hidden_units // 2),
                nn.ReLU(),
                nn.Dropout(self.dropout),
                nn.Linear(self.hidden_units // 2, 1),
                nn.Sigmoid()
            )
        
        # Confidence estimation head
        self.confidence_head = nn.Sequential(
            nn.Linear(self.hidden_units, self.hidden_units // 2),
            nn.ReLU(),
            nn.Dropout(self.dropout),
            nn.Linear(self.hidden_units // 2, len(self.prediction_horizons)),
            nn.Sigmoid()
        )
        
        # Initialize weights
        self._initialize_weights()
    
    def _initialize_weights(self):
        """Initialize model weights"""
        for name, param in self.named_parameters():
            if 'weight' in name:
                if 'lstm' in name:
                    nn.init.xavier_uniform_(param)
                else:
                    nn.init.kaiming_uniform_(param)
            elif 'bias' in name:
                nn.init.constant_(param, 0)
    
    def forward(self, x: torch.Tensor, return_attention: bool = False, 
                mc_dropout: bool = False) -> Dict[str, torch.Tensor]:
        """
        Forward pass through the model
        
        Args:
            x: Input tensor of shape (batch_size, sequence_length, feature_count)
            return_attention: Whether to return attention weights
            mc_dropout: Whether to use Monte Carlo dropout for uncertainty estimation
        
        Returns:
            Dictionary containing predictions and confidence scores
        """
        batch_size, seq_len, features = x.size()
        
        # LSTM forward pass
        lstm_out, (hidden, cell) = self.lstm(x)
        
        # Apply attention
        attended_out, attention_weights = self.attention(lstm_out)
        
        # Global average pooling
        pooled = torch.mean(attended_out, dim=1)
        
        # Apply Monte Carlo dropout if requested
        if mc_dropout:
            pooled = self.mc_dropout(pooled)
        
        # Generate predictions for each horizon
        predictions = {}
        for horizon in self.prediction_horizons:
            predictions[f'risk_score_{horizon}mo'] = self.output_heads[str(horizon)](pooled)
        
        # Generate confidence scores
        confidence_scores = self.confidence_head(pooled)
        
        # Combine predictions and confidence
        output = {
            'predictions': torch.cat([predictions[f'risk_score_{h}mo'] for h in self.prediction_horizons], dim=1),
            'confidence': confidence_scores,
            'hidden_state': pooled
        }
        
        if return_attention:
            output['attention_weights'] = attention_weights
        
        return output
    
    def predict_with_uncertainty(self, x: torch.Tensor, n_samples: int = 10) -> Dict[str, torch.Tensor]:
        """
        Predict with uncertainty estimation using Monte Carlo dropout
        
        Args:
            x: Input tensor
            n_samples: Number of Monte Carlo samples
        
        Returns:
            Dictionary with mean predictions, uncertainty, and confidence intervals
        """
        self.train()  # Enable dropout
        
        predictions_list = []
        confidence_list = []
        
        with torch.no_grad():
            for _ in range(n_samples):
                output = self.forward(x, mc_dropout=True)
                predictions_list.append(output['predictions'])
                confidence_list.append(output['confidence'])
        
        # Stack predictions
        predictions_stack = torch.stack(predictions_list, dim=0)  # (n_samples, batch_size, n_horizons)
        confidence_stack = torch.stack(confidence_list, dim=0)
        
        # Calculate statistics
        mean_predictions = torch.mean(predictions_stack, dim=0)
        std_predictions = torch.std(predictions_stack, dim=0)
        mean_confidence = torch.mean(confidence_stack, dim=0)
        
        # Calculate confidence intervals
        lower_bound = mean_predictions - 1.96 * std_predictions
        upper_bound = mean_predictions + 1.96 * std_predictions
        
        return {
            'mean_predictions': mean_predictions,
            'uncertainty': std_predictions,
            'confidence': mean_confidence,
            'lower_bound': torch.clamp(lower_bound, 0, 1),
            'upper_bound': torch.clamp(upper_bound, 0, 1)
        }
    
    def get_attention_weights(self, x: torch.Tensor) -> torch.Tensor:
        """Get attention weights for interpretability"""
        with torch.no_grad():
            output = self.forward(x, return_attention=True)
            return output['attention_weights']
    
    def count_parameters(self) -> int:
        """Count the number of trainable parameters"""
        return sum(p.numel() for p in self.parameters() if p.requires_grad)


class RiskAwareLoss(nn.Module):
    """Custom loss function that penalizes high-risk misses more heavily"""
    
    def __init__(self, risk_weight: float = 2.0, uncertainty_weight: float = 0.1):
        super().__init__()
        self.risk_weight = risk_weight
        self.uncertainty_weight = uncertainty_weight
        self.mse_loss = nn.MSELoss()
    
    def forward(self, predictions: torch.Tensor, targets: torch.Tensor, 
                confidence: torch.Tensor) -> torch.Tensor:
        """
        Calculate risk-aware loss
        
        Args:
            predictions: Model predictions (batch_size, n_horizons)
            targets: Ground truth targets (batch_size, n_horizons)
            confidence: Model confidence scores (batch_size, n_horizons)
        """
        # Base MSE loss
        mse_loss = self.mse_loss(predictions, targets)
        
        # Risk-aware weighting: penalize high-risk misses more
        risk_weights = 1 + self.risk_weight * targets  # Higher weight for higher risk
        weighted_mse = torch.mean(risk_weights * (predictions - targets) ** 2)
        
        # Uncertainty penalty: encourage higher confidence for accurate predictions
        accuracy = 1 - torch.abs(predictions - targets)
        uncertainty_penalty = torch.mean((1 - confidence) * (1 - accuracy))
        
        total_loss = weighted_mse + self.uncertainty_weight * uncertainty_penalty
        
        return total_loss


def create_model(config: Dict) -> RiskLSTM:
    """Create and initialize the LSTM model"""
    
    model = RiskLSTM(config)
    
    print(f"Created LSTM model with {model.count_parameters():,} parameters")
    print(f"Model architecture:")
    print(f"  - LSTM layers: {config['model']['lstm_layers']}")
    print(f"  - Hidden units: {config['model']['hidden_units']}")
    print(f"  - Attention heads: {config['model']['attention_heads']}")
    print(f"  - Prediction horizons: {config['data']['prediction_horizons']}")
    print(f"  - Dropout: {config['model']['dropout']}")
    
    return model


def create_loss_function(config: Dict) -> RiskAwareLoss:
    """Create the risk-aware loss function"""
    
    loss_config = config['training']['loss']
    return RiskAwareLoss(
        risk_weight=loss_config['risk_weight'],
        uncertainty_weight=loss_config['uncertainty_weight']
    )


def main():
    """Test the model creation and forward pass"""
    
    # Load config
    import yaml
    with open('models/model_config.yaml', 'r') as f:
        config = yaml.safe_load(f)
    
    # Create model
    model = create_model(config)
    
    # Create sample input
    batch_size = 32
    sequence_length = config['data']['sequence_length']
    feature_count = config['data']['feature_count']
    
    sample_input = torch.randn(batch_size, sequence_length, feature_count)
    
    # Test forward pass
    print("\nTesting forward pass...")
    output = model(sample_input)
    
    print(f"Input shape: {sample_input.shape}")
    print(f"Predictions shape: {output['predictions'].shape}")
    print(f"Confidence shape: {output['confidence'].shape}")
    print(f"Hidden state shape: {output['hidden_state'].shape}")
    
    # Test uncertainty estimation
    print("\nTesting uncertainty estimation...")
    uncertainty_output = model.predict_with_uncertainty(sample_input, n_samples=5)
    
    print(f"Mean predictions shape: {uncertainty_output['mean_predictions'].shape}")
    print(f"Uncertainty shape: {uncertainty_output['uncertainty'].shape}")
    print(f"Lower bound shape: {uncertainty_output['lower_bound'].shape}")
    print(f"Upper bound shape: {uncertainty_output['upper_bound'].shape}")
    
    # Test attention weights
    print("\nTesting attention weights...")
    attention_weights = model.get_attention_weights(sample_input)
    print(f"Attention weights shape: {attention_weights.shape}")
    
    print("\nModel test completed successfully!")


if __name__ == "__main__":
    main()