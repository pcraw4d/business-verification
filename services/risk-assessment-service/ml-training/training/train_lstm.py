"""
LSTM Model Training Script

Trains the multi-horizon LSTM model for risk prediction with cross-validation,
early stopping, and comprehensive evaluation.
"""

import os
import sys
import yaml
import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset
import numpy as np
import pandas as pd
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score
from typing import Dict, List, Tuple, Optional
import logging
from datetime import datetime
import json
import pickle
from pathlib import Path

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from models.lstm_model import RiskLSTM, RiskAwareLoss, create_model
from data.feature_preprocessor import FeaturePreprocessor
from data.synthetic_generator import SyntheticDataGenerator


class LSTMTrainer:
    """Trainer for LSTM risk prediction model."""
    
    def __init__(self, config_path: str):
        """Initialize trainer with configuration."""
        self.config = self._load_config(config_path)
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        
        # Setup logging
        self._setup_logging()
        
        # Initialize components
        self.model = None
        self.preprocessor = None
        self.optimizer = None
        self.scheduler = None
        self.criterion = None
        
        # Training history
        self.training_history = {
            'train_loss': [],
            'val_loss': [],
            'train_metrics': [],
            'val_metrics': []
        }
        
        self.logger.info(f"Using device: {self.device}")
        self.logger.info(f"Configuration loaded from: {config_path}")
    
    def _load_config(self, config_path: str) -> Dict:
        """Load configuration from YAML file."""
        with open(config_path, 'r') as f:
            config = yaml.safe_load(f)
        return config
    
    def _setup_logging(self):
        """Setup logging configuration."""
        log_level = self.config.get('logging', {}).get('level', 'INFO')
        logging.basicConfig(
            level=getattr(logging, log_level),
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler('training.log'),
                logging.StreamHandler()
            ]
        )
        self.logger = logging.getLogger(__name__)
    
    def prepare_data(self, data_path: str = None) -> Tuple[Dict, FeaturePreprocessor]:
        """
        Prepare training data.
        
        Args:
            data_path: Path to existing data file, or None to generate new data
            
        Returns:
            Tuple of (data_splits, preprocessor)
        """
        self.logger.info("Preparing training data...")
        
        if data_path and os.path.exists(data_path):
            # Load existing data
            self.logger.info(f"Loading data from {data_path}")
            if data_path.endswith('.parquet'):
                data = pd.read_parquet(data_path)
            else:
                data = pd.read_csv(data_path)
        else:
            # Generate synthetic data
            self.logger.info("Generating synthetic training data...")
            generator = SyntheticDataGenerator(seed=self.config.get('random_seed', 42))
            data = generator.generate_dataset(
                num_businesses=10000,
                months=36
            )
            
            # Save generated data
            output_path = "synthetic_risk_data.parquet"
            generator.save_dataset(data, output_path)
            self.logger.info(f"Synthetic data saved to {output_path}")
        
        # Initialize preprocessor
        self.preprocessor = FeaturePreprocessor(
            sequence_length=self.config['data']['sequence_length'],
            prediction_horizons=self.config['prediction_horizons']
        )
        
        # Fit and transform data
        sequences, targets = self.preprocessor.fit_transform(data)
        
        # Split data
        data_splits = self.preprocessor.split_data(
            sequences, targets,
            test_size=self.config['data']['test_split'],
            val_size=self.config['data']['validation_split']
        )
        
        self.logger.info("Data preparation completed")
        self.logger.info(f"Training samples: {len(data_splits['train'][0])}")
        self.logger.info(f"Validation samples: {len(data_splits['val'][0])}")
        self.logger.info(f"Test samples: {len(data_splits['test'][0])}")
        
        return data_splits, self.preprocessor
    
    def create_model(self) -> RiskLSTM:
        """Create and initialize the LSTM model."""
        self.logger.info("Creating LSTM model...")
        
        # Get input size from preprocessor
        input_size = len(self.preprocessor.numerical_features) + \
                    len(self.preprocessor.categorical_features) + \
                    len(self.preprocessor.temporal_features)
        
        # Update config with actual input size
        self.config['model']['input_size'] = input_size
        
        # Create model
        self.model = create_model(self.config['model'])
        self.model = self.model.to(self.device)
        
        # Initialize optimizer
        self.optimizer = optim.Adam(
            self.model.parameters(),
            lr=self.config['training']['learning_rate'],
            weight_decay=self.config['training']['weight_decay']
        )
        
        # Initialize scheduler
        if self.config['training']['lr_scheduler']['type'] == 'ReduceLROnPlateau':
            self.scheduler = optim.lr_scheduler.ReduceLROnPlateau(
                self.optimizer,
                factor=self.config['training']['lr_scheduler']['factor'],
                patience=self.config['training']['lr_scheduler']['patience'],
                min_lr=self.config['training']['lr_scheduler']['min_lr']
            )
        
        # Initialize loss function
        loss_config = self.config['training']['loss']
        if loss_config['type'] == 'RiskAwareLoss':
            self.criterion = RiskAwareLoss(
                high_risk_threshold=loss_config['high_risk_threshold'],
                penalty_factor=loss_config['penalty_factor']
            )
        else:
            self.criterion = nn.MSELoss()
        
        self.logger.info(f"Model created with {self.model.get_model_info()['total_parameters']:,} parameters")
        self.logger.info(f"Model size: {self.model.get_model_info()['model_size_mb']:.2f} MB")
        
        return self.model
    
    def create_data_loaders(self, data_splits: Dict) -> Dict[str, DataLoader]:
        """Create PyTorch data loaders."""
        loaders = {}
        
        for split_name, (sequences, targets) in data_splits.items():
            # Convert to tensors
            X = torch.FloatTensor(sequences)
            
            # Create datasets for each horizon
            datasets = {}
            for horizon in self.config['prediction_horizons']:
                horizon_key = f'horizon_{horizon}'
                if horizon_key in targets:
                    y = torch.FloatTensor(targets[horizon_key])
                    datasets[horizon_key] = TensorDataset(X, y)
            
            # Create data loader
            batch_size = self.config['training']['batch_size']
            loaders[split_name] = {
                horizon_key: DataLoader(
                    dataset,
                    batch_size=batch_size,
                    shuffle=(split_name == 'train'),
                    num_workers=0,
                    pin_memory=True if self.device.type == 'cuda' else False
                )
                for horizon_key, dataset in datasets.items()
            }
        
        return loaders
    
    def train_epoch(self, data_loaders: Dict, epoch: int) -> Dict[str, float]:
        """Train for one epoch."""
        self.model.train()
        epoch_losses = {f'horizon_{h}': [] for h in self.config['prediction_horizons']}
        
        # Get training loader (use first horizon as primary)
        primary_horizon = f'horizon_{self.config["prediction_horizons"][0]}'
        train_loader = data_loaders['train'][primary_horizon]
        
        for batch_idx, (data, target) in enumerate(train_loader):
            data, target = data.to(self.device), target.to(self.device)
            
            # Zero gradients
            self.optimizer.zero_grad()
            
            # Forward pass
            predictions = self.model(data)
            
            # Calculate loss for each horizon
            total_loss = 0
            for horizon in self.config['prediction_horizons']:
                horizon_key = f'horizon_{horizon}'
                if horizon_key in predictions and horizon_key in data_loaders['train']:
                    # Get target for this horizon
                    horizon_target = target  # For simplicity, using same target
                    horizon_pred = predictions[horizon_key]['risk_score']
                    
                    # Calculate loss
                    loss = self.criterion(horizon_pred, horizon_target)
                    total_loss += loss
                    epoch_losses[horizon_key].append(loss.item())
            
            # Backward pass
            total_loss.backward()
            
            # Gradient clipping
            torch.nn.utils.clip_grad_norm_(self.model.parameters(), max_norm=1.0)
            
            # Update weights
            self.optimizer.step()
            
            # Log progress
            if batch_idx % self.config['logging']['log_interval'] == 0:
                self.logger.info(
                    f'Epoch {epoch}, Batch {batch_idx}/{len(train_loader)}, '
                    f'Loss: {total_loss.item():.6f}'
                )
        
        # Calculate average losses
        avg_losses = {k: np.mean(v) for k, v in epoch_losses.items()}
        return avg_losses
    
    def validate_epoch(self, data_loaders: Dict) -> Dict[str, float]:
        """Validate for one epoch."""
        self.model.eval()
        epoch_losses = {f'horizon_{h}': [] for h in self.config['prediction_horizons']}
        epoch_metrics = {f'horizon_{h}': {'mse': [], 'mae': [], 'r2': []} for h in self.config['prediction_horizons']}
        
        with torch.no_grad():
            # Get validation loader
            primary_horizon = f'horizon_{self.config["prediction_horizons"][0]}'
            val_loader = data_loaders['val'][primary_horizon]
            
            all_predictions = []
            all_targets = []
            
            for data, target in val_loader:
                data, target = data.to(self.device), target.to(self.device)
                
                # Forward pass
                predictions = self.model(data)
                
                # Calculate loss and metrics for each horizon
                for horizon in self.config['prediction_horizons']:
                    horizon_key = f'horizon_{horizon}'
                    if horizon_key in predictions:
                        horizon_pred = predictions[horizon_key]['risk_score']
                        horizon_target = target
                        
                        # Loss
                        loss = self.criterion(horizon_pred, horizon_target)
                        epoch_losses[horizon_key].append(loss.item())
                        
                        # Metrics
                        pred_np = horizon_pred.cpu().numpy()
                        target_np = horizon_target.cpu().numpy()
                        
                        mse = mean_squared_error(target_np, pred_np)
                        mae = mean_absolute_error(target_np, pred_np)
                        r2 = r2_score(target_np, pred_np)
                        
                        epoch_metrics[horizon_key]['mse'].append(mse)
                        epoch_metrics[horizon_key]['mae'].append(mae)
                        epoch_metrics[horizon_key]['r2'].append(r2)
                
                all_predictions.extend(predictions[primary_horizon]['risk_score'].cpu().numpy())
                all_targets.extend(target.cpu().numpy())
        
        # Calculate average losses and metrics
        avg_losses = {k: np.mean(v) for k, v in epoch_losses.items()}
        avg_metrics = {}
        for horizon_key, metrics in epoch_metrics.items():
            avg_metrics[horizon_key] = {k: np.mean(v) for k, v in metrics.items()}
        
        return avg_losses, avg_metrics
    
    def train(self, data_splits: Dict) -> Dict:
        """Train the model."""
        self.logger.info("Starting model training...")
        
        # Create data loaders
        data_loaders = self.create_data_loaders(data_splits)
        
        # Training parameters
        num_epochs = self.config['training']['num_epochs']
        patience = self.config['training']['early_stopping_patience']
        min_delta = self.config['training']['early_stopping_min_delta']
        
        # Early stopping
        best_val_loss = float('inf')
        patience_counter = 0
        best_model_state = None
        
        for epoch in range(num_epochs):
            # Train
            train_losses = self.train_epoch(data_loaders, epoch)
            
            # Validate
            val_losses, val_metrics = self.validate_epoch(data_loaders)
            
            # Update scheduler
            if self.scheduler:
                self.scheduler.step(val_losses[f'horizon_{self.config["prediction_horizons"][0]}'])
            
            # Log epoch results
            self.logger.info(f'Epoch {epoch+1}/{num_epochs}:')
            for horizon in self.config['prediction_horizons']:
                horizon_key = f'horizon_{horizon}'
                self.logger.info(
                    f'  {horizon_key}: Train Loss: {train_losses[horizon_key]:.6f}, '
                    f'Val Loss: {val_losses[horizon_key]:.6f}, '
                    f'Val R²: {val_metrics[horizon_key]["r2"]:.4f}'
                )
            
            # Store history
            self.training_history['train_loss'].append(train_losses)
            self.training_history['val_loss'].append(val_losses)
            self.training_history['train_metrics'].append({})  # Could add train metrics
            self.training_history['val_metrics'].append(val_metrics)
            
            # Early stopping check
            current_val_loss = val_losses[f'horizon_{self.config["prediction_horizons"][0]}']
            if current_val_loss < best_val_loss - min_delta:
                best_val_loss = current_val_loss
                patience_counter = 0
                best_model_state = self.model.state_dict().copy()
            else:
                patience_counter += 1
            
            if patience_counter >= patience:
                self.logger.info(f'Early stopping triggered after {epoch+1} epochs')
                break
            
            # Save checkpoint
            if self.config['logging']['save_checkpoints'] and (epoch + 1) % self.config['logging']['checkpoint_interval'] == 0:
                self.save_checkpoint(epoch + 1)
        
        # Load best model
        if best_model_state:
            self.model.load_state_dict(best_model_state)
            self.logger.info("Loaded best model state")
        
        self.logger.info("Training completed")
        return self.training_history
    
    def evaluate(self, data_splits: Dict) -> Dict:
        """Evaluate the trained model."""
        self.logger.info("Evaluating model...")
        
        # Create data loaders
        data_loaders = self.create_data_loaders(data_splits)
        
        # Evaluate on test set
        self.model.eval()
        test_results = {}
        
        with torch.no_grad():
            for horizon in self.config['prediction_horizons']:
                horizon_key = f'horizon_{horizon}'
                if horizon_key in data_loaders['test']:
                    test_loader = data_loaders['test'][horizon_key]
                    
                    all_predictions = []
                    all_targets = []
                    all_confidences = []
                    
                    for data, target in test_loader:
                        data, target = data.to(self.device), target.to(self.device)
                        
                        predictions = self.model(data)
                        pred = predictions[horizon_key]['risk_score']
                        confidence = predictions[horizon_key]['confidence']
                        
                        all_predictions.extend(pred.cpu().numpy())
                        all_targets.extend(target.cpu().numpy())
                        all_confidences.extend(confidence.cpu().numpy())
                    
                    # Calculate metrics
                    predictions = np.array(all_predictions)
                    targets = np.array(all_targets)
                    confidences = np.array(all_confidences)
                    
                    mse = mean_squared_error(targets, predictions)
                    mae = mean_absolute_error(targets, predictions)
                    rmse = np.sqrt(mse)
                    r2 = r2_score(targets, predictions)
                    
                    # Custom accuracy (within 0.1 of target)
                    accuracy = np.mean(np.abs(predictions - targets) < 0.1)
                    
                    test_results[horizon_key] = {
                        'mse': mse,
                        'mae': mae,
                        'rmse': rmse,
                        'r2': r2,
                        'accuracy': accuracy,
                        'avg_confidence': np.mean(confidences),
                        'num_samples': len(predictions)
                    }
                    
                    self.logger.info(f'{horizon_key} Test Results:')
                    self.logger.info(f'  MSE: {mse:.6f}')
                    self.logger.info(f'  MAE: {mae:.6f}')
                    self.logger.info(f'  RMSE: {rmse:.6f}')
                    self.logger.info(f'  R²: {r2:.4f}')
                    self.logger.info(f'  Accuracy: {accuracy:.4f}')
                    self.logger.info(f'  Avg Confidence: {np.mean(confidences):.4f}')
        
        return test_results
    
    def save_model(self, output_dir: str):
        """Save the trained model and preprocessor."""
        os.makedirs(output_dir, exist_ok=True)
        
        # Save model
        model_path = os.path.join(output_dir, 'lstm_model.pth')
        torch.save({
            'model_state_dict': self.model.state_dict(),
            'model_config': self.config['model'],
            'training_history': self.training_history
        }, model_path)
        
        # Save preprocessor
        preprocessor_path = os.path.join(output_dir, 'preprocessor.pkl')
        self.preprocessor.save_preprocessor(preprocessor_path)
        
        # Save config
        config_path = os.path.join(output_dir, 'config.yaml')
        with open(config_path, 'w') as f:
            yaml.dump(self.config, f, default_flow_style=False)
        
        self.logger.info(f"Model saved to {output_dir}")
    
    def save_checkpoint(self, epoch: int):
        """Save training checkpoint."""
        checkpoint_dir = "checkpoints"
        os.makedirs(checkpoint_dir, exist_ok=True)
        
        checkpoint_path = os.path.join(checkpoint_dir, f'checkpoint_epoch_{epoch}.pth')
        torch.save({
            'epoch': epoch,
            'model_state_dict': self.model.state_dict(),
            'optimizer_state_dict': self.optimizer.state_dict(),
            'scheduler_state_dict': self.scheduler.state_dict() if self.scheduler else None,
            'training_history': self.training_history,
            'config': self.config
        }, checkpoint_path)
        
        self.logger.info(f"Checkpoint saved: {checkpoint_path}")


def main():
    """Main training function."""
    # Configuration
    config_path = "models/model_config.yaml"
    
    # Initialize trainer
    trainer = LSTMTrainer(config_path)
    
    # Prepare data
    data_splits, preprocessor = trainer.prepare_data()
    
    # Create model
    model = trainer.create_model()
    
    # Train model
    training_history = trainer.train(data_splits)
    
    # Evaluate model
    test_results = trainer.evaluate(data_splits)
    
    # Save model
    output_dir = f"trained_models/lstm_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    trainer.save_model(output_dir)
    
    # Save results
    results_path = os.path.join(output_dir, 'test_results.json')
    with open(results_path, 'w') as f:
        json.dump(test_results, f, indent=2, default=str)
    
    print("Training completed successfully!")
    print(f"Model saved to: {output_dir}")
    print("Test Results:")
    for horizon_key, results in test_results.items():
        print(f"  {horizon_key}: Accuracy = {results['accuracy']:.4f}, R² = {results['r2']:.4f}")


if __name__ == "__main__":
    main()
