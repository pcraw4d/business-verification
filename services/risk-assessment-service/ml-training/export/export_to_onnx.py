"""
ONNX Export Script

Exports the trained LSTM model to ONNX format for deployment in Go service.
Includes validation to ensure ONNX inference matches PyTorch output.
"""

import torch
import torch.onnx
import numpy as np
import yaml
import os
import sys
from typing import Dict, Tuple
import onnx
import onnxruntime as ort

# Add parent directory to path for imports
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from models.lstm_model import create_model


class ONNXExporter:
    """Exports PyTorch LSTM model to ONNX format"""
    
    def __init__(self, config_path: str = None):
        """Initialize exporter with configuration"""
        
        # Set default config path if not provided
        if config_path is None:
            config_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'models', 'model_config.yaml')
        
        # Load configuration
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        self.model = None
        self.onnx_model_path = None
        
    def load_trained_model(self, model_path: str):
        """Load the trained PyTorch model"""
        
        print(f"Loading trained model from {model_path}...")
        
        # Create model architecture
        self.model = create_model(self.config)
        
        # Load trained weights
        checkpoint = torch.load(model_path, map_location=self.device)
        self.model.load_state_dict(checkpoint['model_state_dict'])
        
        # Set to evaluation mode
        self.model.eval()
        
        print(f"Model loaded successfully with {self.model.count_parameters():,} parameters")
    
    def create_dummy_input(self, batch_size: int = 1) -> torch.Tensor:
        """Create dummy input tensor for ONNX export"""
        
        sequence_length = self.config['data']['sequence_length']
        feature_count = self.config['data']['feature_count']
        
        dummy_input = torch.randn(batch_size, sequence_length, feature_count)
        return dummy_input
    
    def export_to_onnx(self, output_path: str, batch_size: int = 1, 
                      optimize: bool = True) -> str:
        """Export PyTorch model to ONNX format"""
        
        print(f"Exporting model to ONNX format...")
        
        # Create dummy input
        dummy_input = self.create_dummy_input(batch_size)
        
        # Move to device
        dummy_input = dummy_input.to(self.device)
        
        # Set dynamic axes for batch size
        dynamic_axes = {
            'input': {0: 'batch_size'},
            'predictions': {0: 'batch_size'},
            'confidence': {0: 'batch_size'}
        }
        
        # Export to ONNX
        torch.onnx.export(
            self.model,
            dummy_input,
            output_path,
            export_params=True,
            opset_version=self.config['export']['onnx_opset_version'],
            do_constant_folding=True,
            input_names=['input'],
            output_names=['predictions', 'confidence'],
            dynamic_axes=dynamic_axes,
            verbose=False
        )
        
        print(f"Model exported to: {output_path}")
        self.onnx_model_path = output_path
        
        return output_path
    
    def validate_onnx_model(self, onnx_path: str, test_input: torch.Tensor = None) -> bool:
        """Validate ONNX model by comparing outputs with PyTorch model"""
        
        print("Validating ONNX model...")
        
        if test_input is None:
            test_input = self.create_dummy_input()
        
        # Move to device
        test_input = test_input.to(self.device)
        
        # Get PyTorch model output
        with torch.no_grad():
            pytorch_output = self.model(test_input)
            pytorch_predictions = pytorch_output['predictions'].cpu().numpy()
            pytorch_confidence = pytorch_output['confidence'].cpu().numpy()
        
        # Get ONNX model output
        try:
            ort_session = ort.InferenceSession(onnx_path)
            onnx_inputs = {ort_session.get_inputs()[0].name: test_input.cpu().numpy()}
            onnx_outputs = ort_session.run(None, onnx_inputs)
            onnx_predictions = onnx_outputs[0]
            onnx_confidence = onnx_outputs[1]
        except Exception as e:
            print(f"Error running ONNX model: {e}")
            return False
        
        # Compare outputs
        predictions_diff = np.abs(pytorch_predictions - onnx_predictions)
        confidence_diff = np.abs(pytorch_confidence - onnx_confidence)
        
        max_predictions_diff = np.max(predictions_diff)
        max_confidence_diff = np.max(confidence_diff)
        mean_predictions_diff = np.mean(predictions_diff)
        mean_confidence_diff = np.mean(confidence_diff)
        
        print(f"Validation Results:")
        print(f"  Predictions - Max diff: {max_predictions_diff:.6f}, Mean diff: {mean_predictions_diff:.6f}")
        print(f"  Confidence - Max diff: {max_confidence_diff:.6f}, Mean diff: {mean_confidence_diff:.6f}")
        
        # Check if differences are within acceptable tolerance
        tolerance = 1e-4
        predictions_valid = max_predictions_diff < tolerance
        confidence_valid = max_confidence_diff < tolerance
        
        if predictions_valid and confidence_valid:
            print("✅ ONNX model validation passed!")
            return True
        else:
            print("❌ ONNX model validation failed!")
            return False
    
    def optimize_onnx_model(self, input_path: str, output_path: str) -> str:
        """Optimize ONNX model for inference"""
        
        print("Optimizing ONNX model...")
        
        try:
            # Load model
            model = onnx.load(input_path)
            
            # Optimize model
            from onnx import optimizer
            optimized_model = optimizer.optimize(model)
            
            # Save optimized model
            onnx.save(optimized_model, output_path)
            
            print(f"Optimized model saved to: {output_path}")
            return output_path
            
        except Exception as e:
            print(f"Error optimizing model: {e}")
            return input_path
    
    def test_onnx_inference(self, onnx_path: str, n_tests: int = 10) -> Dict:
        """Test ONNX model inference performance"""
        
        print(f"Testing ONNX inference with {n_tests} samples...")
        
        # Create ONNX runtime session
        ort_session = ort.InferenceSession(onnx_path)
        
        # Test with different batch sizes
        batch_sizes = [1, 4, 8, 16]
        results = {}
        
        for batch_size in batch_sizes:
            print(f"Testing batch size: {batch_size}")
            
            # Create test input
            test_input = self.create_dummy_input(batch_size)
            onnx_inputs = {ort_session.get_inputs()[0].name: test_input.numpy()}
            
            # Measure inference time
            import time
            
            # Warmup
            for _ in range(5):
                _ = ort_session.run(None, onnx_inputs)
            
            # Actual timing
            times = []
            for _ in range(n_tests):
                start_time = time.time()
                outputs = ort_session.run(None, onnx_inputs)
                end_time = time.time()
                times.append((end_time - start_time) * 1000)  # Convert to milliseconds
            
            avg_time = np.mean(times)
            std_time = np.std(times)
            min_time = np.min(times)
            max_time = np.max(times)
            
            results[batch_size] = {
                'avg_time_ms': avg_time,
                'std_time_ms': std_time,
                'min_time_ms': min_time,
                'max_time_ms': max_time,
                'throughput_per_sec': batch_size / (avg_time / 1000)
            }
            
            print(f"  Average time: {avg_time:.2f}ms")
            print(f"  Throughput: {batch_size / (avg_time / 1000):.1f} samples/sec")
        
        return results
    
    def get_model_info(self, onnx_path: str) -> Dict:
        """Get information about the ONNX model"""
        
        print("Analyzing ONNX model...")
        
        # Load model
        model = onnx.load(onnx_path)
        
        # Get model size
        model_size_mb = os.path.getsize(onnx_path) / (1024 * 1024)
        
        # Get input/output info
        input_info = {
            'name': model.graph.input[0].name,
            'shape': [dim.dim_value if dim.dim_value > 0 else 'dynamic' 
                     for dim in model.graph.input[0].type.tensor_type.shape.dim],
            'type': model.graph.input[0].type.tensor_type.elem_type
        }
        
        output_info = []
        for output in model.graph.output:
            output_info.append({
                'name': output.name,
                'shape': [dim.dim_value if dim.dim_value > 0 else 'dynamic' 
                         for dim in output.type.tensor_type.shape.dim],
                'type': output.type.tensor_type.elem_type
            })
        
        # Count nodes
        node_count = len(model.graph.node)
        
        info = {
            'model_size_mb': model_size_mb,
            'input_info': input_info,
            'output_info': output_info,
            'node_count': node_count,
            'opset_version': model.opset_import[0].version
        }
        
        print(f"Model Info:")
        print(f"  Size: {model_size_mb:.2f} MB")
        print(f"  Input: {input_info['name']} {input_info['shape']}")
        print(f"  Outputs: {len(output_info)}")
        for output in output_info:
            print(f"    {output['name']}: {output['shape']}")
        print(f"  Nodes: {node_count}")
        print(f"  Opset Version: {info['opset_version']}")
        
        return info
    
    def export_complete(self, model_path: str, output_dir: str = None) -> Dict:
        """Complete export process with validation and optimization"""
        
        if output_dir is None:
            output_dir = self.config['paths']['output_dir']
        
        # Create output directory
        os.makedirs(output_dir, exist_ok=True)
        
        # Load trained model
        self.load_trained_model(model_path)
        
        # Export to ONNX
        onnx_path = os.path.join(output_dir, self.config['paths']['onnx_file'])
        self.export_to_onnx(onnx_path)
        
        # Validate model
        validation_passed = self.validate_onnx_model(onnx_path)
        
        # Optimize model if requested
        if self.config['export']['optimize_for_inference']:
            optimized_path = onnx_path.replace('.onnx', '_optimized.onnx')
            optimized_path = self.optimize_onnx_model(onnx_path, optimized_path)
            
            # Validate optimized model
            if self.validate_onnx_model(optimized_path):
                onnx_path = optimized_path
                print("Using optimized model")
            else:
                print("Optimized model validation failed, using original")
        
        # Test inference performance
        performance_results = self.test_onnx_inference(onnx_path)
        
        # Get model info
        model_info = self.get_model_info(onnx_path)
        
        # Summary
        export_summary = {
            'onnx_path': onnx_path,
            'validation_passed': validation_passed,
            'model_info': model_info,
            'performance_results': performance_results,
            'config': self.config
        }
        
        print("\n" + "="*50)
        print("EXPORT SUMMARY")
        print("="*50)
        print(f"ONNX Model: {onnx_path}")
        print(f"Validation: {'✅ PASSED' if validation_passed else '❌ FAILED'}")
        print(f"Model Size: {model_info['model_size_mb']:.2f} MB")
        print(f"Batch Size 1 Latency: {performance_results[1]['avg_time_ms']:.2f}ms")
        print(f"Batch Size 16 Latency: {performance_results[16]['avg_time_ms']:.2f}ms")
        
        # Check if performance targets are met
        targets = self.config['targets']
        if 'latency_p95' in targets:
            target_latency = targets['latency_p95']
            actual_latency = performance_results[1]['avg_time_ms']
            if actual_latency <= target_latency:
                print(f"✅ Latency target met: {actual_latency:.2f}ms <= {target_latency}ms")
            else:
                print(f"❌ Latency target not met: {actual_latency:.2f}ms > {target_latency}ms")
        
        return export_summary


def main():
    """Main export function"""
    
    # Initialize exporter
    exporter = ONNXExporter()
    
    # Paths
    model_path = os.path.join(exporter.config['paths']['output_dir'], 
                            exporter.config['paths']['model_file'])
    
    if not os.path.exists(model_path):
        print(f"Trained model not found: {model_path}")
        print("Please train the model first.")
        return
    
    # Export model
    export_summary = exporter.export_complete(model_path)
    
    # Save export summary
    summary_path = os.path.join(exporter.config['paths']['output_dir'], 'export_summary.yaml')
    with open(summary_path, 'w') as f:
        yaml.dump(export_summary, f, default_flow_style=False)
    
    print(f"\nExport summary saved to: {summary_path}")
    print("\nExport completed successfully!")


if __name__ == "__main__":
    main()