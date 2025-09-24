#!/usr/bin/env python3
"""
Model Quantization for Performance Optimization

This module implements model quantization techniques to optimize ML models for
faster inference while maintaining high accuracy. It includes:

- Dynamic quantization for BERT and DistilBERT models
- Static quantization for custom neural networks
- Quantization-aware training (QAT)
- Model compression and optimization
- Performance benchmarking and comparison
- Export to optimized formats (ONNX, TensorRT)

Target: 2-4x faster inference with minimal accuracy loss
"""

import os
import json
import time
import logging
import pickle
from typing import Dict, List, Optional, Any, Tuple, Union
from datetime import datetime
from pathlib import Path
import warnings
warnings.filterwarnings("ignore")

import torch
import torch.nn as nn
import torch.quantization as quantization
from torch.quantization import quantize_dynamic, quantize_static, prepare_qat
import numpy as np
import pandas as pd
from transformers import (
    AutoTokenizer, AutoModel, AutoModelForSequenceClassification,
    BertTokenizer, BertForSequenceClassification,
    DistilBertTokenizer, DistilBertForSequenceClassification
)
import onnx
import onnxruntime as ort
from tqdm import tqdm
import matplotlib.pyplot as plt
import seaborn as sns

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class ModelQuantizer:
    """Model Quantization for Performance Optimization"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        # Create directories
        self.quantized_models_path = Path(config.get('quantized_models_path', 'models/quantized'))
        self.quantized_models_path.mkdir(parents=True, exist_ok=True)
        
        self.benchmark_results_path = Path(config.get('benchmark_results_path', 'benchmarks'))
        self.benchmark_results_path.mkdir(parents=True, exist_ok=True)
        
        logger.info(f"⚡ Model Quantizer initialized")
        logger.info(f"📱 Device: {self.device}")
        logger.info(f"💾 Quantized models path: {self.quantized_models_path}")
    
    def quantize_bert_model(self, model_path: str, model_type: str = "bert") -> Dict[str, Any]:
        """Quantize BERT model using dynamic quantization"""
        logger.info(f"🤖 Quantizing {model_type.upper()} model from: {model_path}")
        
        # Load model and tokenizer
        if model_type.lower() == "distilbert":
            model = DistilBertForSequenceClassification.from_pretrained(model_path)
            tokenizer = DistilBertTokenizer.from_pretrained(model_path)
        else:
            model = BertForSequenceClassification.from_pretrained(model_path)
            tokenizer = BertTokenizer.from_pretrained(model_path)
        
        # Set model to evaluation mode
        model.eval()
        
        # Dynamic quantization
        quantized_model = quantize_dynamic(
            model, 
            {nn.Linear, nn.LSTM, nn.GRU}, 
            dtype=torch.qint8
        )
        
        # Save quantized model
        quantized_path = self.quantized_models_path / f"{model_type}_quantized"
        quantized_path.mkdir(parents=True, exist_ok=True)
        
        quantized_model.save_pretrained(str(quantized_path))
        tokenizer.save_pretrained(str(quantized_path))
        
        # Copy additional files
        self._copy_model_files(model_path, str(quantized_path))
        
        # Calculate model size reduction
        original_size = self._get_model_size(model_path)
        quantized_size = self._get_model_size(str(quantized_path))
        size_reduction = (original_size - quantized_size) / original_size * 100
        
        result = {
            "model_type": model_type,
            "original_path": model_path,
            "quantized_path": str(quantized_path),
            "original_size_mb": original_size,
            "quantized_size_mb": quantized_size,
            "size_reduction_percent": size_reduction,
            "quantization_method": "dynamic",
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"✅ {model_type.upper()} model quantized successfully")
        logger.info(f"📊 Size reduction: {size_reduction:.1f}%")
        logger.info(f"💾 Quantized model saved to: {quantized_path}")
        
        return result
    
    def quantize_custom_model(self, model_path: str, model_name: str) -> Dict[str, Any]:
        """Quantize custom neural network model using static quantization"""
        logger.info(f"🏗️ Quantizing custom model: {model_name}")
        
        # Load model
        model = torch.load(model_path, map_location=self.device)
        model.eval()
        
        # Prepare model for quantization
        model.qconfig = quantization.get_default_qconfig('fbgemm')
        prepared_model = quantization.prepare(model)
        
        # Calibration (using dummy data)
        calibration_data = self._generate_calibration_data(model_name)
        self._calibrate_model(prepared_model, calibration_data)
        
        # Convert to quantized model
        quantized_model = quantization.convert(prepared_model)
        
        # Save quantized model
        quantized_path = self.quantized_models_path / f"{model_name}_quantized"
        quantized_path.mkdir(parents=True, exist_ok=True)
        
        torch.save(quantized_model, quantized_path / "model.pt")
        
        # Calculate model size reduction
        original_size = os.path.getsize(model_path) / (1024 * 1024)  # MB
        quantized_size = os.path.getsize(quantized_path / "model.pt") / (1024 * 1024)  # MB
        size_reduction = (original_size - quantized_size) / original_size * 100
        
        result = {
            "model_name": model_name,
            "original_path": model_path,
            "quantized_path": str(quantized_path),
            "original_size_mb": original_size,
            "quantized_size_mb": quantized_size,
            "size_reduction_percent": size_reduction,
            "quantization_method": "static",
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"✅ Custom model {model_name} quantized successfully")
        logger.info(f"📊 Size reduction: {size_reduction:.1f}%")
        logger.info(f"💾 Quantized model saved to: {quantized_path}")
        
        return result
    
    def quantize_aware_training(self, model: nn.Module, train_loader, val_loader, 
                              num_epochs: int = 5) -> nn.Module:
        """Quantization-aware training for better accuracy"""
        logger.info("🎯 Starting quantization-aware training...")
        
        # Set quantization configuration
        model.qconfig = quantization.get_default_qconfig('fbgemm')
        
        # Prepare model for QAT
        model.train()
        prepared_model = prepare_qat(model)
        
        # Training setup
        optimizer = torch.optim.Adam(prepared_model.parameters(), lr=1e-4)
        criterion = nn.CrossEntropyLoss()
        
        # Training loop
        for epoch in range(num_epochs):
            prepared_model.train()
            total_loss = 0
            
            for batch_idx, (data, target) in enumerate(train_loader):
                optimizer.zero_grad()
                output = prepared_model(data)
                loss = criterion(output, target)
                loss.backward()
                optimizer.step()
                
                total_loss += loss.item()
                
                if batch_idx % 100 == 0:
                    logger.info(f"Epoch {epoch+1}/{num_epochs}, Batch {batch_idx}, Loss: {loss.item():.4f}")
            
            # Validation
            if val_loader:
                val_accuracy = self._evaluate_model(prepared_model, val_loader)
                logger.info(f"Epoch {epoch+1}/{num_epochs}, Validation Accuracy: {val_accuracy:.4f}")
        
        # Convert to quantized model
        prepared_model.eval()
        quantized_model = quantization.convert(prepared_model)
        
        logger.info("✅ Quantization-aware training completed")
        return quantized_model
    
    def export_to_onnx(self, model_path: str, model_type: str = "bert") -> str:
        """Export model to ONNX format for optimized inference"""
        logger.info(f"📤 Exporting {model_type.upper()} model to ONNX...")
        
        # Load model and tokenizer
        if model_type.lower() == "distilbert":
            model = DistilBertForSequenceClassification.from_pretrained(model_path)
            tokenizer = DistilBertTokenizer.from_pretrained(model_path)
        else:
            model = BertForSequenceClassification.from_pretrained(model_path)
            tokenizer = BertTokenizer.from_pretrained(model_path)
        
        model.eval()
        
        # Create dummy input
        dummy_input = tokenizer(
            "This is a test sentence for ONNX export",
            max_length=512,
            padding=True,
            truncation=True,
            return_tensors="pt"
        )
        
        # Export to ONNX
        onnx_path = self.quantized_models_path / f"{model_type}_model.onnx"
        
        torch.onnx.export(
            model,
            (dummy_input['input_ids'], dummy_input['attention_mask']),
            str(onnx_path),
            export_params=True,
            opset_version=11,
            do_constant_folding=True,
            input_names=['input_ids', 'attention_mask'],
            output_names=['logits'],
            dynamic_axes={
                'input_ids': {0: 'batch_size', 1: 'sequence'},
                'attention_mask': {0: 'batch_size', 1: 'sequence'},
                'logits': {0: 'batch_size'}
            }
        )
        
        logger.info(f"✅ ONNX model exported to: {onnx_path}")
        return str(onnx_path)
    
    def benchmark_models(self, original_model_path: str, quantized_model_path: str,
                        test_data: List[str], model_type: str = "bert") -> Dict[str, Any]:
        """Benchmark original vs quantized models"""
        logger.info(f"⏱️ Benchmarking {model_type.upper()} models...")
        
        # Load models
        if model_type.lower() == "distilbert":
            original_model = DistilBertForSequenceClassification.from_pretrained(original_model_path)
            quantized_model = DistilBertForSequenceClassification.from_pretrained(quantized_model_path)
            tokenizer = DistilBertTokenizer.from_pretrained(original_model_path)
        else:
            original_model = BertForSequenceClassification.from_pretrained(original_model_path)
            quantized_model = BertForSequenceClassification.from_pretrained(quantized_model_path)
            tokenizer = BertTokenizer.from_pretrained(original_model_path)
        
        original_model.eval()
        quantized_model.eval()
        
        # Benchmark original model
        original_times = []
        original_predictions = []
        
        for text in tqdm(test_data, desc="Benchmarking original model"):
            start_time = time.time()
            
            inputs = tokenizer(
                text,
                max_length=512,
                padding=True,
                truncation=True,
                return_tensors="pt"
            )
            
            with torch.no_grad():
                outputs = original_model(**inputs)
                prediction = torch.argmax(outputs.logits, dim=-1)
                original_predictions.append(prediction.item())
            
            original_times.append(time.time() - start_time)
        
        # Benchmark quantized model
        quantized_times = []
        quantized_predictions = []
        
        for text in tqdm(test_data, desc="Benchmarking quantized model"):
            start_time = time.time()
            
            inputs = tokenizer(
                text,
                max_length=512,
                padding=True,
                truncation=True,
                return_tensors="pt"
            )
            
            with torch.no_grad():
                outputs = quantized_model(**inputs)
                prediction = torch.argmax(outputs.logits, dim=-1)
                quantized_predictions.append(prediction.item())
            
            quantized_times.append(time.time() - start_time)
        
        # Calculate metrics
        original_avg_time = np.mean(original_times)
        quantized_avg_time = np.mean(quantized_times)
        speedup = original_avg_time / quantized_avg_time
        
        # Calculate accuracy difference
        accuracy_diff = np.mean(np.array(original_predictions) == np.array(quantized_predictions))
        
        benchmark_results = {
            "model_type": model_type,
            "original_model_path": original_model_path,
            "quantized_model_path": quantized_model_path,
            "test_samples": len(test_data),
            "original_avg_time_ms": original_avg_time * 1000,
            "quantized_avg_time_ms": quantized_avg_time * 1000,
            "speedup_factor": speedup,
            "accuracy_agreement": accuracy_diff,
            "original_throughput": 1.0 / original_avg_time,
            "quantized_throughput": 1.0 / quantized_avg_time,
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"📊 Benchmark Results:")
        logger.info(f"   Original avg time: {original_avg_time*1000:.2f}ms")
        logger.info(f"   Quantized avg time: {quantized_avg_time*1000:.2f}ms")
        logger.info(f"   Speedup: {speedup:.2f}x")
        logger.info(f"   Accuracy agreement: {accuracy_diff:.4f}")
        
        return benchmark_results
    
    def benchmark_onnx_model(self, onnx_path: str, test_data: List[str]) -> Dict[str, Any]:
        """Benchmark ONNX model performance"""
        logger.info(f"📤 Benchmarking ONNX model: {onnx_path}")
        
        # Load ONNX model
        ort_session = ort.InferenceSession(onnx_path)
        
        # Load tokenizer
        tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        # Benchmark ONNX model
        onnx_times = []
        
        for text in tqdm(test_data, desc="Benchmarking ONNX model"):
            start_time = time.time()
            
            inputs = tokenizer(
                text,
                max_length=512,
                padding=True,
                truncation=True,
                return_tensors="np"
            )
            
            # Run inference
            ort_inputs = {
                'input_ids': inputs['input_ids'].astype(np.int64),
                'attention_mask': inputs['attention_mask'].astype(np.int64)
            }
            
            ort_session.run(None, ort_inputs)
            onnx_times.append(time.time() - start_time)
        
        onnx_avg_time = np.mean(onnx_times)
        
        benchmark_results = {
            "model_format": "onnx",
            "onnx_path": onnx_path,
            "test_samples": len(test_data),
            "avg_time_ms": onnx_avg_time * 1000,
            "throughput": 1.0 / onnx_avg_time,
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"📊 ONNX Benchmark Results:")
        logger.info(f"   Average time: {onnx_avg_time*1000:.2f}ms")
        logger.info(f"   Throughput: {1.0/onnx_avg_time:.1f} predictions/second")
        
        return benchmark_results
    
    def create_optimization_report(self, benchmark_results: List[Dict[str, Any]]) -> str:
        """Create comprehensive optimization report"""
        logger.info("📝 Creating optimization report...")
        
        report = {
            "optimization_summary": {
                "total_models_optimized": len(benchmark_results),
                "average_speedup": np.mean([r.get('speedup_factor', 1.0) for r in benchmark_results]),
                "average_accuracy_agreement": np.mean([r.get('accuracy_agreement', 1.0) for r in benchmark_results]),
                "timestamp": datetime.now().isoformat()
            },
            "detailed_results": benchmark_results,
            "recommendations": self._generate_optimization_recommendations(benchmark_results)
        }
        
        # Save report
        report_path = self.benchmark_results_path / f"optimization_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_path, 'w') as f:
            json.dump(report, f, indent=2)
        
        # Create visualization
        self._create_optimization_visualizations(benchmark_results)
        
        logger.info(f"📝 Optimization report saved to: {report_path}")
        return str(report_path)
    
    def _copy_model_files(self, source_path: str, dest_path: str):
        """Copy additional model files"""
        import shutil
        
        files_to_copy = ['config.json', 'label_encoder.pkl', 'training_report.json']
        
        for file in files_to_copy:
            source_file = Path(source_path) / file
            dest_file = Path(dest_path) / file
            
            if source_file.exists():
                shutil.copy2(source_file, dest_file)
    
    def _get_model_size(self, model_path: str) -> float:
        """Get model size in MB"""
        total_size = 0
        
        for dirpath, dirnames, filenames in os.walk(model_path):
            for filename in filenames:
                filepath = os.path.join(dirpath, filename)
                total_size += os.path.getsize(filepath)
        
        return total_size / (1024 * 1024)  # Convert to MB
    
    def _generate_calibration_data(self, model_name: str) -> List[torch.Tensor]:
        """Generate calibration data for static quantization"""
        # Generate dummy calibration data
        calibration_data = []
        
        for _ in range(100):  # 100 calibration samples
            # Create dummy input based on model type
            if "financial" in model_name.lower():
                dummy_input = torch.randint(0, 1000, (1, 128))  # Financial services model
            elif "healthcare" in model_name.lower():
                dummy_input = torch.randint(0, 1000, (1, 256))  # Healthcare model
            else:
                dummy_input = torch.randint(0, 1000, (1, 512))  # Default
            
            calibration_data.append(dummy_input)
        
        return calibration_data
    
    def _calibrate_model(self, model: nn.Module, calibration_data: List[torch.Tensor]):
        """Calibrate model for static quantization"""
        model.eval()
        
        with torch.no_grad():
            for data in calibration_data:
                model(data)
    
    def _evaluate_model(self, model: nn.Module, data_loader) -> float:
        """Evaluate model accuracy"""
        model.eval()
        correct = 0
        total = 0
        
        with torch.no_grad():
            for data, target in data_loader:
                output = model(data)
                pred = output.argmax(dim=1, keepdim=True)
                correct += pred.eq(target.view_as(pred)).sum().item()
                total += target.size(0)
        
        return correct / total
    
    def _generate_optimization_recommendations(self, benchmark_results: List[Dict[str, Any]]) -> List[str]:
        """Generate optimization recommendations"""
        recommendations = []
        
        # Analyze results
        avg_speedup = np.mean([r.get('speedup_factor', 1.0) for r in benchmark_results])
        avg_accuracy = np.mean([r.get('accuracy_agreement', 1.0) for r in benchmark_results])
        
        if avg_speedup > 2.0:
            recommendations.append("✅ Excellent speedup achieved! Consider deploying quantized models in production.")
        elif avg_speedup > 1.5:
            recommendations.append("✅ Good speedup achieved. Quantized models provide significant performance benefits.")
        else:
            recommendations.append("⚠️ Limited speedup achieved. Consider alternative optimization techniques.")
        
        if avg_accuracy > 0.95:
            recommendations.append("✅ High accuracy maintained. Quantization has minimal impact on model performance.")
        elif avg_accuracy > 0.90:
            recommendations.append("✅ Good accuracy maintained. Monitor model performance in production.")
        else:
            recommendations.append("⚠️ Accuracy degradation detected. Consider quantization-aware training.")
        
        recommendations.extend([
            "💡 Consider ONNX export for cross-platform deployment.",
            "💡 Implement model caching for frequently used predictions.",
            "💡 Use batch processing for multiple predictions.",
            "💡 Monitor model performance and accuracy in production.",
            "💡 Consider model distillation for further optimization."
        ])
        
        return recommendations
    
    def _create_optimization_visualizations(self, benchmark_results: List[Dict[str, Any]]):
        """Create optimization visualizations"""
        try:
            # Performance comparison
            plt.figure(figsize=(12, 8))
            
            models = [r.get('model_type', 'Unknown') for r in benchmark_results]
            speedups = [r.get('speedup_factor', 1.0) for r in benchmark_results]
            accuracies = [r.get('accuracy_agreement', 1.0) for r in benchmark_results]
            
            # Speedup comparison
            plt.subplot(2, 2, 1)
            plt.bar(models, speedups, color='skyblue')
            plt.title('Model Speedup Comparison')
            plt.ylabel('Speedup Factor')
            plt.xticks(rotation=45)
            
            # Accuracy comparison
            plt.subplot(2, 2, 2)
            plt.bar(models, accuracies, color='lightgreen')
            plt.title('Accuracy Agreement')
            plt.ylabel('Accuracy Agreement')
            plt.xticks(rotation=45)
            
            # Speedup vs Accuracy scatter
            plt.subplot(2, 2, 3)
            plt.scatter(speedups, accuracies, s=100, alpha=0.7)
            plt.xlabel('Speedup Factor')
            plt.ylabel('Accuracy Agreement')
            plt.title('Speedup vs Accuracy Trade-off')
            
            # Performance improvement
            plt.subplot(2, 2, 4)
            improvements = [(s - 1) * 100 for s in speedups]
            plt.bar(models, improvements, color='orange')
            plt.title('Performance Improvement (%)')
            plt.ylabel('Improvement (%)')
            plt.xticks(rotation=45)
            
            plt.tight_layout()
            plt.savefig(self.benchmark_results_path / "optimization_visualization.png", 
                       dpi=300, bbox_inches='tight')
            plt.close()
            
            logger.info("📊 Optimization visualizations created")
            
        except Exception as e:
            logger.warning(f"⚠️ Failed to create visualizations: {e}")

def main():
    """Main function to demonstrate model quantization"""
    
    # Configuration
    config = {
        'quantized_models_path': 'models/quantized',
        'benchmark_results_path': 'benchmarks',
        'test_samples': 100
    }
    
    # Initialize quantizer
    quantizer = ModelQuantizer(config)
    
    # Example usage
    model_path = "models/bert_classification/best_model"
    
    # Quantize BERT model
    if os.path.exists(model_path):
        bert_result = quantizer.quantize_bert_model(model_path, "bert")
        print(f"BERT quantization result: {bert_result}")
        
        # Export to ONNX
        onnx_path = quantizer.export_to_onnx(model_path, "bert")
        print(f"ONNX model exported to: {onnx_path}")
        
        # Benchmark models
        test_data = [
            "Acme Corporation - Leading technology solutions provider",
            "Healthcare Plus - Comprehensive medical services",
            "Financial Services Inc - Investment and wealth management",
            "Retail Solutions - Fashion and lifestyle products"
        ] * 25  # 100 test samples
        
        benchmark_results = quantizer.benchmark_models(
            model_path, 
            bert_result['quantized_path'], 
            test_data, 
            "bert"
        )
        
        print(f"Benchmark results: {benchmark_results}")
        
        # Create optimization report
        report_path = quantizer.create_optimization_report([benchmark_results])
        print(f"Optimization report: {report_path}")
    
    logger.info("🎉 Model quantization demonstration completed!")

if __name__ == "__main__":
    main()
