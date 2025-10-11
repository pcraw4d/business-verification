# LSTM Risk Prediction Model Training

This directory contains the Python ML training environment for the LSTM-based risk prediction model. The trained model will be exported to ONNX format for deployment in the Go service.

## Overview

The LSTM model provides multi-horizon risk predictions (6, 9, 12 months) with:
- **Attention mechanism** for interpretability
- **Uncertainty estimation** via Monte Carlo dropout
- **Risk-aware loss function** that penalizes high-risk misses more heavily
- **Temporal feature engineering** for time-series analysis

## Quick Start

### 1. Automated Setup

Run the setup script to initialize everything:

```bash
python setup_environment.py
```

This will:
- Create a Python virtual environment
- Install all dependencies
- Generate synthetic training data
- Validate the setup

### 2. Manual Setup

If you prefer manual setup:

```bash
# Create virtual environment
python -m venv venv

# Activate virtual environment
# On Windows:
venv\Scripts\activate
# On Unix/Linux/macOS:
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Generate synthetic data
python data/synthetic_generator.py
```

### 3. Train the Model

```bash
# Activate virtual environment first
source venv/bin/activate  # or venv\Scripts\activate on Windows

# Train the LSTM model
python training/train_lstm.py
```

### 4. Export to ONNX

```bash
# Export trained model to ONNX format
python export/export_to_onnx.py
```

## Directory Structure

```
ml-training/
├── requirements.txt              # Python dependencies
├── setup_environment.py         # Automated setup script
├── README.md                    # This file
├── data/                        # Data generation and preprocessing
│   ├── synthetic_generator.py   # Generate synthetic training data
│   └── feature_preprocessor.py  # Data preprocessing utilities
├── models/                      # Model architecture and configuration
│   ├── lstm_model.py           # LSTM model implementation
│   └── model_config.yaml       # Model configuration
├── training/                    # Training scripts
│   ├── train_lstm.py           # Main training script
│   └── hyperparameter_tuning.py # Hyperparameter optimization (TODO)
├── export/                      # Model export utilities
│   └── export_to_onnx.py       # ONNX export script
├── notebooks/                   # Jupyter notebooks for exploration
│   └── model_exploration.ipynb # Model analysis notebook (TODO)
├── output/                      # Training outputs
│   ├── risk_lstm_v1.pth        # Trained PyTorch model
│   ├── risk_lstm_v1.onnx       # Exported ONNX model
│   └── preprocessor.pkl         # Fitted preprocessor
├── checkpoints/                 # Training checkpoints
├── logs/                        # Training logs
└── venv/                        # Python virtual environment
```

## Configuration

The model configuration is defined in `models/model_config.yaml`. Key settings:

### Data Configuration
- `sequence_length`: 12 months of historical data
- `feature_count`: 20 features per timestep
- `prediction_horizons`: [6, 9, 12] months
- `businesses_per_industry`: 1250 (10k total businesses)

### Model Architecture
- `lstm_layers`: 2 LSTM layers
- `hidden_units`: 128 hidden units
- `attention_heads`: 4 attention heads
- `dropout`: 0.3 dropout rate

### Training Configuration
- `batch_size`: 64
- `epochs`: 100
- `learning_rate`: 0.001
- `early_stopping`: 15 epochs patience

### Performance Targets
- 6-month accuracy: ≥88%
- 9-month accuracy: ≥86%
- 12-month accuracy: ≥85%
- Latency p95: <150ms

## Model Architecture

### LSTM with Attention
```
Input: [batch_size, 12, 20]  # 12 months, 20 features
├── LSTM Layers (2x128 units)
├── Multi-Head Attention (4 heads)
├── Global Average Pooling
└── Output Heads (3 horizons)
    ├── 6-month prediction
    ├── 9-month prediction
    └── 12-month prediction
```

### Key Features
- **Multi-horizon predictions**: Single model predicts multiple time horizons
- **Attention mechanism**: Identifies which time periods are most important
- **Uncertainty estimation**: Monte Carlo dropout for confidence intervals
- **Risk-aware loss**: Penalizes high-risk prediction errors more heavily

## Training Process

### 1. Data Generation
- Generates 10,000+ business time-series
- 2-3 years of historical data per business
- Industry-specific risk patterns
- Seasonal trends and economic cycles

### 2. Feature Engineering
- Time-based features (month, quarter, year)
- Lagged features (1, 3, 6, 12 months)
- Moving averages (3, 6, 12 months)
- Volatility measures
- Statistical features

### 3. Training
- Cross-validation with 5 folds
- Early stopping with patience
- Learning rate scheduling
- Gradient clipping
- Model checkpointing

### 4. Validation
- Separate test set evaluation
- Horizon-specific metrics
- Performance target validation
- Uncertainty calibration

## Export Process

### ONNX Export
- Exports trained PyTorch model to ONNX format
- Optimizes for inference performance
- Validates ONNX output matches PyTorch
- Tests inference performance

### Validation
- Output comparison (PyTorch vs ONNX)
- Performance benchmarking
- Model size optimization
- Deployment readiness check

## Usage Examples

### Training a New Model

```python
from training.train_lstm import LSTMTrainer

# Initialize trainer
trainer = LSTMTrainer()

# Load data
data_loaders, preprocessor = trainer.load_data("data/synthetic_risk_data.parquet")

# Train model
trainer.train(data_loaders, preprocessor)
```

### Exporting to ONNX

```python
from export.export_to_onnx import ONNXExporter

# Initialize exporter
exporter = ONNXExporter()

# Export model
export_summary = exporter.export_complete("output/risk_lstm_v1.pth")
```

### Using the Trained Model

```python
import torch
from models.lstm_model import create_model

# Load trained model
model = create_model(config)
checkpoint = torch.load("output/risk_lstm_v1.pth")
model.load_state_dict(checkpoint['model_state_dict'])

# Make predictions
with torch.no_grad():
    output = model(input_tensor)
    predictions = output['predictions']
    confidence = output['confidence']
```

## Performance Monitoring

### Training Metrics
- Loss (train/validation)
- Accuracy per horizon
- R² score per horizon
- Learning rate schedule
- Gradient norms

### Model Performance
- Inference latency
- Memory usage
- Model size
- Throughput

### Validation Metrics
- MSE, MAE, R² per horizon
- Accuracy vs targets
- Confidence calibration
- Uncertainty quality

## Troubleshooting

### Common Issues

1. **CUDA out of memory**
   - Reduce batch size in config
   - Use CPU training (slower but works)

2. **Training not converging**
   - Check learning rate
   - Verify data quality
   - Adjust model architecture

3. **ONNX export fails**
   - Check PyTorch/ONNX versions
   - Simplify model if needed
   - Verify input shapes

4. **Poor accuracy**
   - Increase training data
   - Adjust model architecture
   - Tune hyperparameters

### Getting Help

- Check the logs in `logs/` directory
- Review training history plots
- Validate data quality
- Test with smaller datasets first

## Next Steps

After training and exporting the model:

1. **Copy ONNX model** to Go service:
   ```bash
   cp output/risk_lstm_v1.onnx ../models/
   ```

2. **Update Go service** to use the real model instead of placeholder

3. **Deploy and test** the enhanced service

4. **Monitor performance** in production

## Dependencies

- Python 3.8+
- PyTorch 2.0+
- ONNX 1.14+
- scikit-learn
- pandas
- numpy
- matplotlib
- seaborn
- tqdm
- pyyaml

See `requirements.txt` for complete list with versions.
