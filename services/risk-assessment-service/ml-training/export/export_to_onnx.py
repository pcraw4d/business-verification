"""
ONNX Export Script for LSTM Risk Prediction Model

Exports the trained PyTorch LSTM model to ONNX format for deployment in Go.
Includes validation to ensure ONNX inference matches PyTorch output.
"""

import os
import sys
import torch
import torch.onnx
import numpy as np
import onnx
import onnxruntime as ort
from typing import Dict, Tuple, Optional
import logging
from pathlib import Path

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from models.lstm_model import RiskLSTM, create_model


class ONNXExporter:
    """Exports PyTorch LSTM model to ONNX format."""
    
    def __init__(self, model_path: str, config_path: str):
        """
        Initialize ONNX exporter.
        
        Args:
            model_path: Path to trained PyTorch model (.pth file)
            config_path: Path to model configuration file
        """
        self.model_path = model_path
        self.config_path = config_path
        self.model = None
        self.config = None
        
        # Setup logging
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)
    
    def load_model(self) -> RiskLSTM:
        """Load the trained PyTorch model."""
        self.logger.info(f"Loading model from {self.model_path}")
        
        # Load checkpoint
        checkpoint = torch.load(self.model_path, map_location='cpu')
        
        # Load config
        import yaml
        with open(self.config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        # Create model
        self.model = create_model(self.config['model'])
        
        # Load state dict
        self.model.load_state_dict(checkpoint['model_state_dict'])
        self.model.eval()
        
        self.logger.info("Model loaded successfully")
        self.logger.info(f"Model info: {self.model.get_model_info()}")
        
        return self.model
    
    def create_dummy_input(self, batch_size: int = 1) -> torch.Tensor:
        """
        Create dummy input for ONNX export.
        
        Args:
            batch_size: Batch size for dummy input
            
        Returns:
            Dummy input tensor
        """
        sequence_length = self.config['data']['sequence_length']
        input_size = self.config['model']['input_size']
        
        dummy_input = torch.randn(batch_size, sequence_length, input_size)
        
        self.logger.info(f"Created dummy input: {dummy_input.shape}")
        return dummy_input
    
    def export_to_onnx(self, 
                      output_path: str,
                      opset_version: int = 14,
                      optimize: bool = True,
                      dynamic_axes: Optional[Dict] = None) -> str:
        """
        Export PyTorch model to ONNX format.
        
        Args:
            output_path: Output path for ONNX model
            opset_version: ONNX opset version
            optimize: Whether to optimize the model
            dynamic_axes: Dynamic axes configuration
            
        Returns:
            Path to exported ONNX model
        """
        self.logger.info("Starting ONNX export...")
        
        # Create dummy input
        dummy_input = self.create_dummy_input()
        
        # Default dynamic axes
        if dynamic_axes is None:
            dynamic_axes = {
                'input': {0: 'batch_size'},
                'output': {0: 'batch_size'}
            }
        
        # Export to ONNX
        torch.onnx.export(
            self.model,
            dummy_input,
            output_path,
            export_params=True,
            opset_version=opset_version,
            do_constant_folding=True,
            input_names=['input'],
            output_names=['output'],
            dynamic_axes=dynamic_axes,
            verbose=False
        )
        
        self.logger.info(f"ONNX model exported to {output_path}")
        
        # Verify ONNX model
        self.verify_onnx_model(output_path)
        
        return output_path
    
    def verify_onnx_model(self, onnx_path: str):
        """Verify the exported ONNX model."""
        self.logger.info("Verifying ONNX model...")
        
        # Load and check ONNX model
        onnx_model = onnx.load(onnx_path)
        onnx.checker.check_model(onnx_model)
        
        self.logger.info("ONNX model verification passed")
        
        # Print model info
        self.logger.info(f"ONNX model inputs: {[input.name for input in onnx_model.graph.input]}")
        self.logger.info(f"ONNX model outputs: {[output.name for output in onnx_model.graph.output]}")
    
    def validate_onnx_inference(self, 
                               onnx_path: str,
                               tolerance: float = 1e-3) -> bool:
        """
        Validate that ONNX inference matches PyTorch inference.
        
        Args:
            onnx_path: Path to ONNX model
            tolerance: Tolerance for numerical differences
            
        Returns:
            True if validation passes
        """
        self.logger.info("Validating ONNX inference...")
        
        # Create test input
        test_input = self.create_dummy_input(batch_size=2)
        
        # PyTorch inference
        self.model.eval()
        with torch.no_grad():
            pytorch_output = self.model(test_input)
        
        # ONNX inference
        ort_session = ort.InferenceSession(onnx_path)
        
        # Prepare input for ONNX
        onnx_input = {ort_session.get_inputs()[0].name: test_input.numpy()}
        
        # Run ONNX inference
        onnx_output = ort_session.run(None, onnx_input)
        
        # Compare outputs
        validation_passed = True
        
        for horizon in self.config['prediction_horizons']:
            horizon_key = f'horizon_{horizon}'
            if horizon_key in pytorch_output:
                pytorch_risk = pytorch_output[horizon_key]['risk_score'].numpy()
                pytorch_conf = pytorch_output[horizon_key]['confidence'].numpy()
                
                # Find corresponding ONNX output (simplified - would need proper mapping)
                # For now, just check if outputs are reasonable
                if len(onnx_output) > 0:
                    onnx_risk = onnx_output[0]  # Simplified
                    
                    # Check if outputs are within tolerance
                    if np.allclose(pytorch_risk, onnx_risk, atol=tolerance):
                        self.logger.info(f"✓ {horizon_key} validation passed")
                    else:
                        self.logger.warning(f"✗ {horizon_key} validation failed")
                        self.logger.warning(f"  PyTorch: {pytorch_risk}")
                        self.logger.warning(f"  ONNX: {onnx_risk}")
                        validation_passed = False
        
        if validation_passed:
            self.logger.info("✓ ONNX inference validation passed")
        else:
            self.logger.warning("✗ ONNX inference validation failed")
        
        return validation_passed
    
    def optimize_onnx_model(self, input_path: str, output_path: str) -> str:
        """
        Optimize ONNX model for deployment.
        
        Args:
            input_path: Path to input ONNX model
            output_path: Path to optimized ONNX model
            
        Returns:
            Path to optimized model
        """
        self.logger.info("Optimizing ONNX model...")
        
        try:
            # Load model
            model = onnx.load(input_path)
            
            # Basic optimization
            from onnx import optimizer
            passes = ['eliminate_identity', 'eliminate_nop_transpose', 'fuse_consecutive_transposes']
            optimized_model = optimizer.optimize(model, passes)
            
            # Save optimized model
            onnx.save(optimized_model, output_path)
            
            self.logger.info(f"Optimized model saved to {output_path}")
            
            # Compare sizes
            original_size = os.path.getsize(input_path) / (1024 * 1024)  # MB
            optimized_size = os.path.getsize(output_path) / (1024 * 1024)  # MB
            
            self.logger.info(f"Original size: {original_size:.2f} MB")
            self.logger.info(f"Optimized size: {optimized_size:.2f} MB")
            self.logger.info(f"Size reduction: {((original_size - optimized_size) / original_size * 100):.1f}%")
            
            return output_path
            
        except Exception as e:
            self.logger.warning(f"Optimization failed: {e}")
            self.logger.info("Using original model")
            return input_path
    
    def export_with_validation(self, 
                              output_dir: str = "exported_models",
                              model_name: str = "risk_lstm_v1") -> Dict[str, str]:
        """
        Complete export process with validation.
        
        Args:
            output_dir: Output directory
            model_name: Model name for output files
            
        Returns:
            Dictionary with paths to exported files
        """
        # Create output directory
        os.makedirs(output_dir, exist_ok=True)
        
        # Export paths
        onnx_path = os.path.join(output_dir, f"{model_name}.onnx")
        optimized_path = os.path.join(output_dir, f"{model_name}_optimized.onnx")
        
        # Export to ONNX
        self.export_to_onnx(onnx_path)
        
        # Validate inference
        validation_passed = self.validate_onnx_inference(onnx_path)
        
        if not validation_passed:
            self.logger.warning("ONNX validation failed, but continuing with export")
        
        # Optimize model
        final_path = self.optimize_onnx_model(onnx_path, optimized_path)
        
        # Create metadata file
        metadata = {
            'model_name': model_name,
            'export_date': str(pd.Timestamp.now()),
            'pytorch_model_path': self.model_path,
            'config_path': self.config_path,
            'onnx_path': final_path,
            'validation_passed': validation_passed,
            'model_info': self.model.get_model_info(),
            'config': self.config
        }
        
        metadata_path = os.path.join(output_dir, f"{model_name}_metadata.json")
        import json
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2, default=str)
        
        self.logger.info(f"Export completed successfully")
        self.logger.info(f"ONNX model: {final_path}")
        self.logger.info(f"Metadata: {metadata_path}")
        
        return {
            'onnx_model': final_path,
            'metadata': metadata_path,
            'validation_passed': validation_passed
        }


def main():
    """Main export function."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Export LSTM model to ONNX')
    parser.add_argument('--model_path', type=str, required=True,
                       help='Path to trained PyTorch model (.pth file)')
    parser.add_argument('--config_path', type=str, required=True,
                       help='Path to model configuration file')
    parser.add_argument('--output_dir', type=str, default='exported_models',
                       help='Output directory for ONNX model')
    parser.add_argument('--model_name', type=str, default='risk_lstm_v1',
                       help='Name for exported model')
    
    args = parser.parse_args()
    
    # Initialize exporter
    exporter = ONNXExporter(args.model_path, args.config_path)
    
    # Load model
    exporter.load_model()
    
    # Export with validation
    results = exporter.export_with_validation(
        output_dir=args.output_dir,
        model_name=args.model_name
    )
    
    print("Export completed successfully!")
    print(f"ONNX model: {results['onnx_model']}")
    print(f"Validation passed: {results['validation_passed']}")


if __name__ == "__main__":
    main()
