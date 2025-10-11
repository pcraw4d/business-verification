"""
LSTM Training Script

Trains the multi-horizon LSTM model for risk prediction with:
- Cross-validation
- Early stopping
- Learning rate scheduling
- Model checkpointing
- Performance monitoring
"""

import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, TensorDataset
import numpy as np
import pandas as pd
import yaml
import os
import time
from typing import Dict, List, Tuple
from sklearn.metrics import mean_squared_error, mean_absolute_error, r2_score
import matplotlib.pyplot as plt
import seaborn as sns
from tqdm import tqdm
import warnings
warnings.filterwarnings('ignore')

# Import our modules
import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from models.lstm_model import create_model, create_loss_function
from data.feature_preprocessor import FeaturePreprocessor


class LSTMTrainer:
    """Trainer class for LSTM risk prediction model"""
    
    def __init__(self, config_path: str = None):
        """Initialize trainer with configuration"""
        
        # Set default config path if not provided
        if config_path is None:
            config_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'models', 'model_config.yaml')
        
        # Load configuration
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        print(f"Using device: {self.device}")
        
        # Training parameters
        self.batch_size = self.config['training']['batch_size']
        self.epochs = self.config['training']['epochs']
        self.learning_rate = self.config['training']['learning_rate']
        self.prediction_horizons = self.config['data']['prediction_horizons']
        
        # Initialize model and loss function
        self.model = None
        self.loss_function = None
        self.optimizer = None
        self.scheduler = None
        
        # Training history
        self.train_history = {
            'loss': [],
            'val_loss': [],
            'accuracy': [],
            'val_accuracy': []
        }
        
        # Create output directories
        self._create_directories()
    
    def _create_directories(self):
        """Create necessary directories"""
        dirs = [
            self.config['paths']['output_dir'],
            self.config['paths']['checkpoint_dir'],
            self.config['paths']['log_dir']
        ]
        
        for dir_path in dirs:
            os.makedirs(dir_path, exist_ok=True)
    
    def load_data(self, data_path: str) -> Dict:
        """Load and preprocess training data"""
        
        print("Loading and preprocessing data...")
        
        # Load data
        data = pd.read_parquet(data_path)
        print(f"Loaded data with shape: {data.shape}")
        
        # Initialize preprocessor
        preprocessor = FeaturePreprocessor(self.config['data'])
        
        # Fit and transform data
        features, targets_data = preprocessor.fit_transform(data)
        
        # Create train/val/test splits
        splits = preprocessor.create_sequences(
            features, 
            targets_data['targets'],
            test_size=self.config['validation']['test_size'],
            val_size=self.config['validation']['val_size']
        )
        
        # Save preprocessor
        preprocessor.save_preprocessor(
            os.path.join(self.config['paths']['output_dir'], 'preprocessor.pkl')
        )
        
        # Create data loaders
        data_loaders = {}
        for split_name, split_data in splits.items():
            dataset = TensorDataset(
                torch.FloatTensor(split_data['features']),
                torch.FloatTensor(split_data['targets'])
            )
            data_loaders[split_name] = DataLoader(
                dataset, 
                batch_size=self.batch_size, 
                shuffle=(split_name == 'train')
            )
        
        print(f"Data loaders created:")
        for name, loader in data_loaders.items():
            print(f"  {name}: {len(loader)} batches")
        
        return data_loaders, preprocessor
    
    def initialize_model(self):
        """Initialize model, optimizer, and scheduler"""
        
        print("Initializing model...")
        
        # Create model
        self.model = create_model(self.config)
        self.model.to(self.device)
        
        # Create loss function
        self.loss_function = create_loss_function(self.config)
        
        # Create optimizer
        optimizer_config = self.config['training']['optimizer']
        self.optimizer = optim.Adam(
            self.model.parameters(),
            lr=self.learning_rate,
            betas=(optimizer_config['beta1'], optimizer_config['beta2']),
            eps=optimizer_config['epsilon'],
            weight_decay=optimizer_config['weight_decay']
        )
        
        # Create scheduler
        if self.config['training']['learning_rate_schedule'] == 'cosine_annealing':
            self.scheduler = optim.lr_scheduler.CosineAnnealingLR(
                self.optimizer, 
                T_max=self.epochs
            )
        else:
            self.scheduler = optim.lr_scheduler.ReduceLROnPlateau(
                self.optimizer, 
                mode='min', 
                factor=0.5, 
                patience=5
            )
        
        print(f"Model initialized with {self.model.count_parameters():,} parameters")
    
    def train_epoch(self, train_loader: DataLoader) -> Dict[str, float]:
        """Train for one epoch"""
        
        self.model.train()
        total_loss = 0
        total_samples = 0
        
        # Metrics for each horizon
        horizon_metrics = {f'{h}mo': {'mse': 0, 'mae': 0, 'r2': 0} for h in self.prediction_horizons}
        
        for batch_idx, (features, targets) in enumerate(tqdm(train_loader, desc="Training")):
            features = features.to(self.device)
            targets = targets.to(self.device)
            
            # Forward pass
            self.optimizer.zero_grad()
            output = self.model(features)
            
            # Calculate loss
            loss = self.loss_function(output['predictions'], targets, output['confidence'])
            
            # Backward pass
            loss.backward()
            
            # Gradient clipping
            if self.config['model']['gradient_clipping'] > 0:
                torch.nn.utils.clip_grad_norm_(
                    self.model.parameters(), 
                    self.config['model']['gradient_clipping']
                )
            
            self.optimizer.step()
            
            # Update metrics
            total_loss += loss.item() * features.size(0)
            total_samples += features.size(0)
            
            # Calculate metrics for each horizon
            with torch.no_grad():
                predictions = output['predictions'].cpu().numpy()
                targets_np = targets.cpu().numpy()
                
                for i, horizon in enumerate(self.prediction_horizons):
                    pred_h = predictions[:, i]
                    target_h = targets_np[:, i]
                    
                    horizon_metrics[f'{horizon}mo']['mse'] += mean_squared_error(target_h, pred_h) * features.size(0)
                    horizon_metrics[f'{horizon}mo']['mae'] += mean_absolute_error(target_h, pred_h) * features.size(0)
                    horizon_metrics[f'{horizon}mo']['r2'] += r2_score(target_h, pred_h) * features.size(0)
        
        # Average metrics
        avg_loss = total_loss / total_samples
        for horizon in horizon_metrics:
            for metric in horizon_metrics[horizon]:
                horizon_metrics[horizon][metric] /= total_samples
        
        return {
            'loss': avg_loss,
            'horizon_metrics': horizon_metrics
        }
    
    def validate_epoch(self, val_loader: DataLoader) -> Dict[str, float]:
        """Validate for one epoch"""
        
        self.model.eval()
        total_loss = 0
        total_samples = 0
        
        # Metrics for each horizon
        horizon_metrics = {f'{h}mo': {'mse': 0, 'mae': 0, 'r2': 0} for h in self.prediction_horizons}
        
        with torch.no_grad():
            for features, targets in tqdm(val_loader, desc="Validation"):
                features = features.to(self.device)
                targets = targets.to(self.device)
                
                # Forward pass
                output = self.model(features)
                
                # Calculate loss
                loss = self.loss_function(output['predictions'], targets, output['confidence'])
                
                # Update metrics
                total_loss += loss.item() * features.size(0)
                total_samples += features.size(0)
                
                # Calculate metrics for each horizon
                predictions = output['predictions'].cpu().numpy()
                targets_np = targets.cpu().numpy()
                
                for i, horizon in enumerate(self.prediction_horizons):
                    pred_h = predictions[:, i]
                    target_h = targets_np[:, i]
                    
                    horizon_metrics[f'{horizon}mo']['mse'] += mean_squared_error(target_h, pred_h) * features.size(0)
                    horizon_metrics[f'{horizon}mo']['mae'] += mean_absolute_error(target_h, pred_h) * features.size(0)
                    horizon_metrics[f'{horizon}mo']['r2'] += r2_score(target_h, pred_h) * features.size(0)
        
        # Average metrics
        avg_loss = total_loss / total_samples
        for horizon in horizon_metrics:
            for metric in horizon_metrics[horizon]:
                horizon_metrics[horizon][metric] /= total_samples
        
        return {
            'loss': avg_loss,
            'horizon_metrics': horizon_metrics
        }
    
    def train(self, data_loaders: Dict, preprocessor: FeaturePreprocessor):
        """Main training loop"""
        
        print("Starting training...")
        
        # Initialize model
        self.initialize_model()
        
        # Early stopping
        best_val_loss = float('inf')
        patience_counter = 0
        early_stopping_config = self.config['training']['early_stopping']
        
        # Training loop
        for epoch in range(self.epochs):
            print(f"\nEpoch {epoch+1}/{self.epochs}")
            print("-" * 50)
            
            # Train
            train_metrics = self.train_epoch(data_loaders['train'])
            
            # Validate
            val_metrics = self.validate_epoch(data_loaders['validation'])
            
            # Update learning rate
            if isinstance(self.scheduler, optim.lr_scheduler.ReduceLROnPlateau):
                self.scheduler.step(val_metrics['loss'])
            else:
                self.scheduler.step()
            
            # Store history
            self.train_history['loss'].append(train_metrics['loss'])
            self.train_history['val_loss'].append(val_metrics['loss'])
            
            # Calculate accuracy (1 - MAE for risk scores)
            train_accuracy = 1 - np.mean([train_metrics['horizon_metrics'][f'{h}mo']['mae'] 
                                        for h in self.prediction_horizons])
            val_accuracy = 1 - np.mean([val_metrics['horizon_metrics'][f'{h}mo']['mae'] 
                                      for h in self.prediction_horizons])
            
            self.train_history['accuracy'].append(train_accuracy)
            self.train_history['val_accuracy'].append(val_accuracy)
            
            # Print metrics
            print(f"Train Loss: {train_metrics['loss']:.4f}, Train Accuracy: {train_accuracy:.4f}")
            print(f"Val Loss: {val_metrics['loss']:.4f}, Val Accuracy: {val_accuracy:.4f}")
            print(f"Learning Rate: {self.optimizer.param_groups[0]['lr']:.6f}")
            
            # Print horizon-specific metrics
            for horizon in self.prediction_horizons:
                train_r2 = train_metrics['horizon_metrics'][f'{horizon}mo']['r2']
                val_r2 = val_metrics['horizon_metrics'][f'{horizon}mo']['r2']
                print(f"  {horizon}mo - Train R²: {train_r2:.4f}, Val R²: {val_r2:.4f}")
            
            # Early stopping
            if val_metrics['loss'] < best_val_loss - early_stopping_config['min_delta']:
                best_val_loss = val_metrics['loss']
                patience_counter = 0
                
                # Save best model
                self.save_checkpoint(epoch, val_metrics['loss'], is_best=True)
            else:
                patience_counter += 1
            
            # Save regular checkpoint
            if (epoch + 1) % self.config['training'].get('checkpoint_interval', 10) == 0:
                self.save_checkpoint(epoch, val_metrics['loss'], is_best=False)
            
            # Early stopping check
            if patience_counter >= early_stopping_config['patience']:
                print(f"\nEarly stopping triggered after {epoch+1} epochs")
                break
        
        print("\nTraining completed!")
        
        # Load best model
        self.load_checkpoint(is_best=True)
        
        # Final evaluation
        self.evaluate(data_loaders['test'], preprocessor)
        
        # Save final model
        self.save_model()
        
        # Plot training history
        self.plot_training_history()
    
    def save_checkpoint(self, epoch: int, val_loss: float, is_best: bool = False):
        """Save model checkpoint"""
        
        checkpoint = {
            'epoch': epoch,
            'model_state_dict': self.model.state_dict(),
            'optimizer_state_dict': self.optimizer.state_dict(),
            'scheduler_state_dict': self.scheduler.state_dict(),
            'val_loss': val_loss,
            'config': self.config
        }
        
        if is_best:
            checkpoint_path = os.path.join(
                self.config['paths']['checkpoint_dir'], 
                'best_model.pth'
            )
        else:
            checkpoint_path = os.path.join(
                self.config['paths']['checkpoint_dir'], 
                f'checkpoint_epoch_{epoch}.pth'
            )
        
        torch.save(checkpoint, checkpoint_path)
        print(f"Checkpoint saved: {checkpoint_path}")
    
    def load_checkpoint(self, is_best: bool = True):
        """Load model checkpoint"""
        
        if is_best:
            checkpoint_path = os.path.join(
                self.config['paths']['checkpoint_dir'], 
                'best_model.pth'
            )
        else:
            # Load latest checkpoint
            checkpoint_files = [f for f in os.listdir(self.config['paths']['checkpoint_dir']) 
                              if f.startswith('checkpoint_epoch_')]
            if not checkpoint_files:
                print("No checkpoint found")
                return
            
            latest_checkpoint = max(checkpoint_files, key=lambda x: int(x.split('_')[-1].split('.')[0]))
            checkpoint_path = os.path.join(self.config['paths']['checkpoint_dir'], latest_checkpoint)
        
        if os.path.exists(checkpoint_path):
            checkpoint = torch.load(checkpoint_path, map_location=self.device)
            self.model.load_state_dict(checkpoint['model_state_dict'])
            self.optimizer.load_state_dict(checkpoint['optimizer_state_dict'])
            self.scheduler.load_state_dict(checkpoint['scheduler_state_dict'])
            print(f"Checkpoint loaded: {checkpoint_path}")
        else:
            print(f"Checkpoint not found: {checkpoint_path}")
    
    def evaluate(self, test_loader: DataLoader, preprocessor: FeaturePreprocessor):
        """Evaluate model on test set"""
        
        print("\nEvaluating on test set...")
        
        self.model.eval()
        all_predictions = []
        all_targets = []
        all_confidence = []
        
        with torch.no_grad():
            for features, targets in tqdm(test_loader, desc="Testing"):
                features = features.to(self.device)
                targets = targets.to(self.device)
                
                # Forward pass
                output = self.model(features)
                
                all_predictions.append(output['predictions'].cpu().numpy())
                all_targets.append(targets.cpu().numpy())
                all_confidence.append(output['confidence'].cpu().numpy())
        
        # Combine all predictions
        predictions = np.concatenate(all_predictions, axis=0)
        targets = np.concatenate(all_targets, axis=0)
        confidence = np.concatenate(all_confidence, axis=0)
        
        # Inverse transform targets to original scale
        targets_original = preprocessor.inverse_transform_targets(targets)
        predictions_original = preprocessor.inverse_transform_targets(predictions)
        
        # Calculate metrics for each horizon
        print("\nTest Set Results:")
        print("=" * 50)
        
        for i, horizon in enumerate(self.prediction_horizons):
            pred_h = predictions_original[:, i]
            target_h = targets_original[:, i]
            conf_h = confidence[:, i]
            
            mse = mean_squared_error(target_h, pred_h)
            mae = mean_absolute_error(target_h, pred_h)
            r2 = r2_score(target_h, pred_h)
            accuracy = 1 - mae  # Approximate accuracy for risk scores
            
            print(f"{horizon} month prediction:")
            print(f"  MSE: {mse:.4f}")
            print(f"  MAE: {mae:.4f}")
            print(f"  R²: {r2:.4f}")
            print(f"  Accuracy: {accuracy:.4f}")
            print(f"  Avg Confidence: {np.mean(conf_h):.4f}")
            print()
        
        # Check if targets are met
        targets_config = self.config['targets']
        for i, horizon in enumerate(self.prediction_horizons):
            pred_h = predictions_original[:, i]
            target_h = targets_original[:, i]
            accuracy = 1 - mean_absolute_error(target_h, pred_h)
            
            target_key = f'accuracy_{horizon}mo'
            if target_key in targets_config:
                target_accuracy = targets_config[target_key]
                if accuracy >= target_accuracy:
                    print(f"✅ {horizon}mo accuracy target met: {accuracy:.3f} >= {target_accuracy:.3f}")
                else:
                    print(f"❌ {horizon}mo accuracy target not met: {accuracy:.3f} < {target_accuracy:.3f}")
    
    def save_model(self):
        """Save the final trained model"""
        
        model_path = os.path.join(
            self.config['paths']['output_dir'], 
            self.config['paths']['model_file']
        )
        
        torch.save({
            'model_state_dict': self.model.state_dict(),
            'config': self.config
        }, model_path)
        
        print(f"Model saved: {model_path}")
    
    def plot_training_history(self):
        """Plot training history"""
        
        fig, axes = plt.subplots(2, 2, figsize=(15, 10))
        
        # Loss
        axes[0, 0].plot(self.train_history['loss'], label='Train Loss')
        axes[0, 0].plot(self.train_history['val_loss'], label='Validation Loss')
        axes[0, 0].set_title('Training and Validation Loss')
        axes[0, 0].set_xlabel('Epoch')
        axes[0, 0].set_ylabel('Loss')
        axes[0, 0].legend()
        axes[0, 0].grid(True)
        
        # Accuracy
        axes[0, 1].plot(self.train_history['accuracy'], label='Train Accuracy')
        axes[0, 1].plot(self.train_history['val_accuracy'], label='Validation Accuracy')
        axes[0, 1].set_title('Training and Validation Accuracy')
        axes[0, 1].set_xlabel('Epoch')
        axes[0, 1].set_ylabel('Accuracy')
        axes[0, 1].legend()
        axes[0, 1].grid(True)
        
        # Learning rate (if available)
        if hasattr(self.scheduler, 'get_last_lr'):
            lr_history = [group['lr'] for group in self.optimizer.param_groups]
            axes[1, 0].plot(lr_history)
            axes[1, 0].set_title('Learning Rate Schedule')
            axes[1, 0].set_xlabel('Epoch')
            axes[1, 0].set_ylabel('Learning Rate')
            axes[1, 0].grid(True)
        
        # Model parameters info
        axes[1, 1].text(0.1, 0.5, f"Model Parameters: {self.model.count_parameters():,}", 
                       transform=axes[1, 1].transAxes, fontsize=12)
        axes[1, 1].text(0.1, 0.4, f"Prediction Horizons: {self.prediction_horizons}", 
                       transform=axes[1, 1].transAxes, fontsize=12)
        axes[1, 1].text(0.1, 0.3, f"Device: {self.device}", 
                       transform=axes[1, 1].transAxes, fontsize=12)
        axes[1, 1].set_title('Model Information')
        axes[1, 1].axis('off')
        
        plt.tight_layout()
        
        # Save plot
        plot_path = os.path.join(self.config['paths']['output_dir'], 'training_history.png')
        plt.savefig(plot_path, dpi=300, bbox_inches='tight')
        plt.show()
        
        print(f"Training history plot saved: {plot_path}")


def main():
    """Main training function"""
    
    # Initialize trainer
    trainer = LSTMTrainer()
    
    # Load data
    data_path = os.path.join(trainer.config['paths']['data_dir'], 
                           trainer.config['paths']['data_file'])
    
    if not os.path.exists(data_path):
        print(f"Data file not found: {data_path}")
        print("Please run the synthetic data generator first.")
        return
    
    data_loaders, preprocessor = trainer.load_data(data_path)
    
    # Train model
    trainer.train(data_loaders, preprocessor)
    
    print("\nTraining completed successfully!")


if __name__ == "__main__":
    main()