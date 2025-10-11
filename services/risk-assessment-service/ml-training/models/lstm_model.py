"""
LSTM Model for Risk Prediction

Implements a multi-horizon LSTM model with attention mechanism for
time-series risk prediction with uncertainty estimation.
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
import numpy as np
from typing import Dict, List, Tuple, Optional
import math


class AttentionLayer(nn.Module):
    """Attention mechanism for LSTM outputs."""
    
    def __init__(self, hidden_size: int):
        super(AttentionLayer, self).__init__()
        self.hidden_size = hidden_size
        self.attention = nn.Linear(hidden_size, 1)
        
    def forward(self, lstm_outputs: torch.Tensor) -> Tuple[torch.Tensor, torch.Tensor]:
        """
        Apply attention to LSTM outputs.
        
        Args:
            lstm_outputs: [batch_size, seq_len, hidden_size]
            
        Returns:
            Tuple of (weighted_output, attention_weights)
        """
        # Calculate attention scores
        attention_scores = self.attention(lstm_outputs)  # [batch_size, seq_len, 1]
        attention_weights = F.softmax(attention_scores, dim=1)  # [batch_size, seq_len, 1]
        
        # Apply attention weights
        weighted_output = torch.sum(lstm_outputs * attention_weights, dim=1)  # [batch_size, hidden_size]
        
        return weighted_output, attention_weights.squeeze(-1)


class MultiHorizonHead(nn.Module):
    """Multi-horizon prediction head."""
    
    def __init__(self, input_size: int, hidden_size: int, dropout_rate: float = 0.3):
        super(MultiHorizonHead, self).__init__()
        self.hidden_size = hidden_size
        
        # Shared layers
        self.shared_layers = nn.Sequential(
            nn.Linear(input_size, hidden_size),
            nn.ReLU(),
            nn.Dropout(dropout_rate),
            nn.Linear(hidden_size, hidden_size // 2),
            nn.ReLU(),
            nn.Dropout(dropout_rate)
        )
        
        # Risk score prediction
        self.risk_score_head = nn.Linear(hidden_size // 2, 1)
        
        # Confidence prediction
        self.confidence_head = nn.Linear(hidden_size // 2, 1)
        
    def forward(self, x: torch.Tensor) -> Tuple[torch.Tensor, torch.Tensor]:
        """
        Forward pass for multi-horizon head.
        
        Args:
            x: [batch_size, input_size]
            
        Returns:
            Tuple of (risk_score, confidence)
        """
        shared_features = self.shared_layers(x)
        
        # Risk score (sigmoid to [0, 1])
        risk_score = torch.sigmoid(self.risk_score_head(shared_features))
        
        # Confidence (sigmoid to [0, 1])
        confidence = torch.sigmoid(self.confidence_head(shared_features))
        
        return risk_score, confidence


class RiskLSTM(nn.Module):
    """
    Multi-horizon LSTM model for risk prediction with attention mechanism.
    """
    
    def __init__(self, 
                 input_size: int,
                 hidden_size: int = 128,
                 num_layers: int = 2,
                 dropout_rate: float = 0.3,
                 prediction_horizons: List[int] = [6, 9, 12],
                 use_attention: bool = True,
                 use_uncertainty: bool = True):
        super(RiskLSTM, self).__init__()
        
        self.input_size = input_size
        self.hidden_size = hidden_size
        self.num_layers = num_layers
        self.dropout_rate = dropout_rate
        self.prediction_horizons = prediction_horizons
        self.use_attention = use_attention
        self.use_uncertainty = use_uncertainty
        
        # Input normalization
        self.input_norm = nn.LayerNorm(input_size)
        
        # LSTM layers
        self.lstm = nn.LSTM(
            input_size=input_size,
            hidden_size=hidden_size,
            num_layers=num_layers,
            dropout=dropout_rate if num_layers > 1 else 0,
            batch_first=True,
            bidirectional=False
        )
        
        # Attention mechanism
        if use_attention:
            self.attention = AttentionLayer(hidden_size)
        
        # Multi-horizon prediction heads
        self.horizon_heads = nn.ModuleDict({
            f'horizon_{h}': MultiHorizonHead(
                input_size=hidden_size,
                hidden_size=hidden_size,
                dropout_rate=dropout_rate
            )
            for h in prediction_horizons
        })
        
        # Uncertainty estimation (Monte Carlo dropout)
        self.mc_dropout = nn.Dropout(dropout_rate)
        
        # Initialize weights
        self._initialize_weights()
    
    def _initialize_weights(self):
        """Initialize model weights."""
        for name, param in self.named_parameters():
            if 'weight' in name:
                if 'lstm' in name:
                    # LSTM weight initialization
                    nn.init.xavier_uniform_(param)
                else:
                    # Linear layer weight initialization
                    nn.init.xavier_uniform_(param)
            elif 'bias' in name:
                nn.init.constant_(param, 0)
    
    def forward(self, x: torch.Tensor, mc_samples: int = 1) -> Dict[str, torch.Tensor]:
        """
        Forward pass of the model.
        
        Args:
            x: Input sequences [batch_size, seq_len, input_size]
            mc_samples: Number of Monte Carlo samples for uncertainty estimation
            
        Returns:
            Dictionary with predictions for each horizon
        """
        batch_size, seq_len, _ = x.shape
        
        # Input normalization
        x = self.input_norm(x)
        
        # LSTM forward pass
        lstm_outputs, (hidden, cell) = self.lstm(x)  # [batch_size, seq_len, hidden_size]
        
        # Apply attention if enabled
        if self.use_attention:
            context_vector, attention_weights = self.attention(lstm_outputs)
        else:
            # Use last output
            context_vector = lstm_outputs[:, -1, :]  # [batch_size, hidden_size]
            attention_weights = None
        
        # Multi-horizon predictions
        predictions = {}
        
        for horizon in self.prediction_horizons:
            horizon_key = f'horizon_{horizon}'
            
            if self.use_uncertainty and mc_samples > 1:
                # Monte Carlo dropout for uncertainty estimation
                risk_scores = []
                confidences = []
                
                for _ in range(mc_samples):
                    # Apply dropout
                    mc_context = self.mc_dropout(context_vector)
                    
                    # Get predictions
                    risk_score, confidence = self.horizon_heads[horizon_key](mc_context)
                    risk_scores.append(risk_score)
                    confidences.append(confidence)
                
                # Stack and compute statistics
                risk_scores = torch.stack(risk_scores, dim=1)  # [batch_size, mc_samples, 1]
                confidences = torch.stack(confidences, dim=1)  # [batch_size, mc_samples, 1]
                
                # Mean and standard deviation
                mean_risk = torch.mean(risk_scores, dim=1)
                std_risk = torch.std(risk_scores, dim=1)
                mean_confidence = torch.mean(confidences, dim=1)
                std_confidence = torch.std(confidences, dim=1)
                
                predictions[horizon_key] = {
                    'risk_score': mean_risk,
                    'risk_uncertainty': std_risk,
                    'confidence': mean_confidence,
                    'confidence_uncertainty': std_confidence
                }
            else:
                # Single prediction
                risk_score, confidence = self.horizon_heads[horizon_key](context_vector)
                predictions[horizon_key] = {
                    'risk_score': risk_score,
                    'confidence': confidence
                }
        
        # Add attention weights if available
        if attention_weights is not None:
            predictions['attention_weights'] = attention_weights
        
        return predictions
    
    def predict_single(self, x: torch.Tensor) -> Dict[str, float]:
        """
        Make a single prediction (no uncertainty estimation).
        
        Args:
            x: Input sequence [1, seq_len, input_size]
            
        Returns:
            Dictionary with predictions
        """
        self.eval()
        with torch.no_grad():
            predictions = self.forward(x, mc_samples=1)
            
            result = {}
            for horizon in self.prediction_horizons:
                horizon_key = f'horizon_{horizon}'
                result[f'{horizon}_month_risk'] = predictions[horizon_key]['risk_score'].item()
                result[f'{horizon}_month_confidence'] = predictions[horizon_key]['confidence'].item()
            
            return result
    
    def predict_with_uncertainty(self, x: torch.Tensor, mc_samples: int = 10) -> Dict[str, Dict[str, float]]:
        """
        Make predictions with uncertainty estimation.
        
        Args:
            x: Input sequence [1, seq_len, input_size]
            mc_samples: Number of Monte Carlo samples
            
        Returns:
            Dictionary with predictions and uncertainties
        """
        self.eval()
        with torch.no_grad():
            predictions = self.forward(x, mc_samples=mc_samples)
            
            result = {}
            for horizon in self.prediction_horizons:
                horizon_key = f'horizon_{horizon}'
                result[f'{horizon}_month'] = {
                    'risk_score': predictions[horizon_key]['risk_score'].item(),
                    'risk_uncertainty': predictions[horizon_key]['risk_uncertainty'].item(),
                    'confidence': predictions[horizon_key]['confidence'].item(),
                    'confidence_uncertainty': predictions[horizon_key]['confidence_uncertainty'].item()
                }
            
            return result
    
    def get_attention_weights(self, x: torch.Tensor) -> np.ndarray:
        """
        Get attention weights for interpretability.
        
        Args:
            x: Input sequence [1, seq_len, input_size]
            
        Returns:
            Attention weights as numpy array
        """
        if not self.use_attention:
            return None
        
        self.eval()
        with torch.no_grad():
            predictions = self.forward(x, mc_samples=1)
            return predictions['attention_weights'].cpu().numpy()
    
    def get_model_info(self) -> Dict:
        """Get model information."""
        total_params = sum(p.numel() for p in self.parameters())
        trainable_params = sum(p.numel() for p in self.parameters() if p.requires_grad)
        
        return {
            'model_type': 'RiskLSTM',
            'input_size': self.input_size,
            'hidden_size': self.hidden_size,
            'num_layers': self.num_layers,
            'dropout_rate': self.dropout_rate,
            'prediction_horizons': self.prediction_horizons,
            'use_attention': self.use_attention,
            'use_uncertainty': self.use_uncertainty,
            'total_parameters': total_params,
            'trainable_parameters': trainable_params,
            'model_size_mb': total_params * 4 / (1024 * 1024)  # Assuming float32
        }


class RiskAwareLoss(nn.Module):
    """
    Custom loss function that penalizes high-risk prediction errors more heavily.
    """
    
    def __init__(self, high_risk_threshold: float = 0.7, penalty_factor: float = 2.0):
        super(RiskAwareLoss, self).__init__()
        self.high_risk_threshold = high_risk_threshold
        self.penalty_factor = penalty_factor
        self.mse_loss = nn.MSELoss()
    
    def forward(self, predictions: torch.Tensor, targets: torch.Tensor) -> torch.Tensor:
        """
        Calculate risk-aware loss.
        
        Args:
            predictions: Model predictions [batch_size, 1]
            targets: Ground truth targets [batch_size, 1]
            
        Returns:
            Weighted loss
        """
        # Base MSE loss
        base_loss = self.mse_loss(predictions, targets)
        
        # Calculate prediction errors
        errors = torch.abs(predictions - targets)
        
        # Apply higher penalty for high-risk cases
        high_risk_mask = targets > self.high_risk_threshold
        penalty_weights = torch.where(high_risk_mask, 
                                    torch.tensor(self.penalty_factor, device=targets.device),
                                    torch.tensor(1.0, device=targets.device))
        
        # Weighted loss
        weighted_errors = errors * penalty_weights
        weighted_loss = torch.mean(weighted_errors)
        
        return base_loss + weighted_loss


def create_model(config: Dict) -> RiskLSTM:
    """
    Create LSTM model from configuration.
    
    Args:
        config: Model configuration dictionary
        
    Returns:
        Initialized RiskLSTM model
    """
    model = RiskLSTM(
        input_size=config['input_size'],
        hidden_size=config.get('hidden_size', 128),
        num_layers=config.get('num_layers', 2),
        dropout_rate=config.get('dropout_rate', 0.3),
        prediction_horizons=config.get('prediction_horizons', [6, 9, 12]),
        use_attention=config.get('use_attention', True),
        use_uncertainty=config.get('use_uncertainty', True)
    )
    
    return model


def main():
    """Test the LSTM model."""
    # Test configuration
    config = {
        'input_size': 20,
        'hidden_size': 128,
        'num_layers': 2,
        'dropout_rate': 0.3,
        'prediction_horizons': [6, 9, 12],
        'use_attention': True,
        'use_uncertainty': True
    }
    
    # Create model
    model = create_model(config)
    
    # Test forward pass
    batch_size, seq_len, input_size = 32, 12, 20
    x = torch.randn(batch_size, seq_len, input_size)
    
    print(f"Model created with {model.get_model_info()['total_parameters']:,} parameters")
    
    # Forward pass
    predictions = model(x)
    
    print("Model output shapes:")
    for horizon in config['prediction_horizons']:
        horizon_key = f'horizon_{horizon}'
        print(f"  {horizon_key}: {predictions[horizon_key]['risk_score'].shape}")
    
    print("Model test completed successfully!")


if __name__ == "__main__":
    main()
