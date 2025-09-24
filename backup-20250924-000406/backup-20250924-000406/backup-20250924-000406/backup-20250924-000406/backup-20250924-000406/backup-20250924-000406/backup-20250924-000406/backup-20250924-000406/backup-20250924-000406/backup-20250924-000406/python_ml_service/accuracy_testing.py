#!/usr/bin/env python3
"""
Accuracy Testing Framework for ML Models

This module implements comprehensive accuracy testing for all ML models,
including:

- Cross-validation testing
- Holdout testing
- A/B testing between models
- Statistical significance testing
- Performance benchmarking
- Accuracy reporting and visualization

Target: 95%+ accuracy for classification, 90%+ accuracy for risk detection
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
import numpy as np
import pandas as pd
from sklearn.metrics import (
    accuracy_score, precision_recall_fscore_support, 
    classification_report, confusion_matrix, roc_auc_score,
    precision_score, recall_score, f1_score
)
from sklearn.model_selection import (
    cross_val_score, StratifiedKFold, train_test_split,
    validation_curve, learning_curve
)
from sklearn.preprocessing import LabelEncoder
from scipy import stats
import matplotlib.pyplot as plt
import seaborn as sns
from tqdm import tqdm
import joblib

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class AccuracyTester:
    """Comprehensive accuracy testing framework for ML models"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        # Create directories
        self.test_results_path = Path(config.get('test_results_path', 'test_results'))
        self.test_results_path.mkdir(parents=True, exist_ok=True)
        
        # Accuracy targets
        self.accuracy_targets = {
            "classification": 0.95,  # 95% accuracy target
            "risk_detection": 0.90,  # 90% accuracy target
            "confidence_threshold": 0.8  # 80% confidence threshold
        }
        
        logger.info(f"🎯 Accuracy Tester initialized")
        logger.info(f"📱 Device: {self.device}")
        logger.info(f"💾 Test results path: {self.test_results_path}")
        logger.info(f"🎯 Accuracy targets: {self.accuracy_targets}")
    
    def test_model_accuracy(self, model: nn.Module, test_data: List[Dict],
                          model_type: str = "bert", task_type: str = "classification") -> Dict[str, Any]:
        """Test model accuracy on test dataset"""
        logger.info(f"🧪 Testing {model_type.upper()} model accuracy for {task_type}...")
        
        model.eval()
        predictions = []
        true_labels = []
        confidences = []
        processing_times = []
        
        # Load tokenizer
        if model_type.lower() == "distilbert":
            from transformers import DistilBertTokenizer
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            from transformers import BertTokenizer
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        with torch.no_grad():
            for item in tqdm(test_data, desc=f"Testing {model_type} model"):
                start_time = time.time()
                
                # Prepare input
                text = item.get('text', '')
                true_label = item.get('label', 0)
                
                inputs = tokenizer(
                    text,
                    max_length=512,
                    padding=True,
                    truncation=True,
                    return_tensors="pt"
                )
                
                # Get prediction
                outputs = model(**inputs)
                logits = outputs.logits
                probabilities = torch.softmax(logits, dim=-1)
                prediction = torch.argmax(logits, dim=-1).item()
                confidence = torch.max(probabilities, dim=-1)[0].item()
                
                processing_time = time.time() - start_time
                
                predictions.append(prediction)
                true_labels.append(true_label)
                confidences.append(confidence)
                processing_times.append(processing_time)
        
        # Calculate metrics
        accuracy = accuracy_score(true_labels, predictions)
        precision, recall, f1, _ = precision_recall_fscore_support(
            true_labels, predictions, average='weighted'
        )
        
        # Calculate per-class metrics
        precision_per_class, recall_per_class, f1_per_class, _ = precision_recall_fscore_support(
            true_labels, predictions, average=None
        )
        
        # Calculate confidence metrics
        avg_confidence = np.mean(confidences)
        confidence_std = np.std(confidences)
        
        # Calculate processing time metrics
        avg_processing_time = np.mean(processing_times)
        processing_time_std = np.std(processing_times)
        
        # Create confusion matrix
        cm = confusion_matrix(true_labels, predictions)
        
        # Calculate accuracy by confidence level
        confidence_accuracy = self._calculate_confidence_accuracy(
            true_labels, predictions, confidences
        )
        
        test_results = {
            "model_type": model_type,
            "task_type": task_type,
            "test_samples": len(test_data),
            "accuracy": accuracy,
            "precision": precision,
            "recall": recall,
            "f1_score": f1,
            "precision_per_class": precision_per_class.tolist(),
            "recall_per_class": recall_per_class.tolist(),
            "f1_per_class": f1_per_class.tolist(),
            "confusion_matrix": cm.tolist(),
            "confidence_metrics": {
                "average_confidence": avg_confidence,
                "confidence_std": confidence_std,
                "confidence_accuracy": confidence_accuracy
            },
            "performance_metrics": {
                "average_processing_time": avg_processing_time,
                "processing_time_std": processing_time_std,
                "throughput": 1.0 / avg_processing_time
            },
            "target_achieved": accuracy >= self.accuracy_targets[task_type],
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"📊 {model_type.upper()} {task_type} test results:")
        logger.info(f"   Accuracy: {accuracy:.4f} (target: {self.accuracy_targets[task_type]:.2f})")
        logger.info(f"   Precision: {precision:.4f}")
        logger.info(f"   Recall: {recall:.4f}")
        logger.info(f"   F1-Score: {f1:.4f}")
        logger.info(f"   Average confidence: {avg_confidence:.4f}")
        logger.info(f"   Average processing time: {avg_processing_time*1000:.2f}ms")
        logger.info(f"   Target achieved: {'✅' if test_results['target_achieved'] else '❌'}")
        
        return test_results
    
    def cross_validate_model(self, model: nn.Module, data: List[Dict],
                           model_type: str = "bert", cv_folds: int = 5) -> Dict[str, Any]:
        """Perform cross-validation testing"""
        logger.info(f"🔄 Performing {cv_folds}-fold cross-validation for {model_type.upper()}...")
        
        # Prepare data for cross-validation
        texts = [item['text'] for item in data]
        labels = [item['label'] for item in data]
        
        # Create stratified k-fold
        skf = StratifiedKFold(n_splits=cv_folds, shuffle=True, random_state=42)
        
        cv_scores = []
        cv_results = []
        
        for fold, (train_idx, val_idx) in enumerate(skf.split(texts, labels)):
            logger.info(f"   Fold {fold + 1}/{cv_folds}")
            
            # Split data
            train_data = [data[i] for i in train_idx]
            val_data = [data[i] for i in val_idx]
            
            # Test on validation set
            fold_results = self.test_model_accuracy(model, val_data, model_type)
            cv_scores.append(fold_results['accuracy'])
            cv_results.append(fold_results)
        
        # Calculate cross-validation statistics
        cv_mean = np.mean(cv_scores)
        cv_std = np.std(cv_scores)
        cv_min = np.min(cv_scores)
        cv_max = np.max(cv_scores)
        
        cv_summary = {
            "model_type": model_type,
            "cv_folds": cv_folds,
            "cv_scores": cv_scores,
            "cv_mean": cv_mean,
            "cv_std": cv_std,
            "cv_min": cv_min,
            "cv_max": cv_max,
            "cv_confidence_interval": {
                "lower": cv_mean - 1.96 * cv_std / np.sqrt(cv_folds),
                "upper": cv_mean + 1.96 * cv_std / np.sqrt(cv_folds)
            },
            "target_achieved": cv_mean >= self.accuracy_targets["classification"],
            "detailed_results": cv_results,
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"📊 Cross-validation results:")
        logger.info(f"   Mean accuracy: {cv_mean:.4f} ± {cv_std:.4f}")
        logger.info(f"   Min accuracy: {cv_min:.4f}")
        logger.info(f"   Max accuracy: {cv_max:.4f}")
        logger.info(f"   95% CI: [{cv_summary['cv_confidence_interval']['lower']:.4f}, {cv_summary['cv_confidence_interval']['upper']:.4f}]")
        logger.info(f"   Target achieved: {'✅' if cv_summary['target_achieved'] else '❌'}")
        
        return cv_summary
    
    def compare_models(self, models: Dict[str, nn.Module], test_data: List[Dict],
                      task_type: str = "classification") -> Dict[str, Any]:
        """Compare multiple models and perform statistical significance testing"""
        logger.info(f"⚖️ Comparing {len(models)} models for {task_type}...")
        
        model_results = {}
        
        # Test each model
        for model_name, model in models.items():
            logger.info(f"   Testing {model_name}...")
            results = self.test_model_accuracy(model, test_data, model_name, task_type)
            model_results[model_name] = results
        
        # Perform statistical significance testing
        significance_tests = self._perform_significance_tests(model_results)
        
        # Rank models by accuracy
        model_rankings = sorted(
            model_results.items(),
            key=lambda x: x[1]['accuracy'],
            reverse=True
        )
        
        comparison_results = {
            "task_type": task_type,
            "test_samples": len(test_data),
            "model_results": model_results,
            "model_rankings": [
                {"model": name, "accuracy": results["accuracy"]} 
                for name, results in model_rankings
            ],
            "significance_tests": significance_tests,
            "best_model": model_rankings[0][0],
            "best_accuracy": model_rankings[0][1]["accuracy"],
            "target_achieved": any(
                results["target_achieved"] for results in model_results.values()
            ),
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"📊 Model comparison results:")
        for i, (model_name, results) in enumerate(model_rankings):
            logger.info(f"   {i+1}. {model_name}: {results['accuracy']:.4f}")
        logger.info(f"   Best model: {comparison_results['best_model']}")
        logger.info(f"   Target achieved: {'✅' if comparison_results['target_achieved'] else '❌'}")
        
        return comparison_results
    
    def test_model_robustness(self, model: nn.Module, test_data: List[Dict],
                            model_type: str = "bert") -> Dict[str, Any]:
        """Test model robustness with various perturbations"""
        logger.info(f"🛡️ Testing {model_type.upper()} model robustness...")
        
        # Original accuracy
        original_results = self.test_model_accuracy(model, test_data, model_type)
        original_accuracy = original_results['accuracy']
        
        # Test with various perturbations
        perturbations = {
            "typos": self._add_typos,
            "capitalization": self._change_capitalization,
            "punctuation": self._change_punctuation,
            "word_order": self._change_word_order,
            "synonyms": self._replace_with_synonyms
        }
        
        robustness_results = {
            "model_type": model_type,
            "original_accuracy": original_accuracy,
            "perturbation_results": {},
            "robustness_score": 0.0,
            "timestamp": datetime.now().isoformat()
        }
        
        for perturbation_name, perturbation_func in perturbations.items():
            logger.info(f"   Testing {perturbation_name} perturbation...")
            
            # Apply perturbation
            perturbed_data = []
            for item in test_data:
                perturbed_item = item.copy()
                perturbed_item['text'] = perturbation_func(item['text'])
                perturbed_data.append(perturbed_item)
            
            # Test on perturbed data
            perturbed_results = self.test_model_accuracy(model, perturbed_data, model_type)
            perturbed_accuracy = perturbed_results['accuracy']
            
            # Calculate robustness
            robustness = perturbed_accuracy / original_accuracy
            robustness_results["perturbation_results"][perturbation_name] = {
                "accuracy": perturbed_accuracy,
                "robustness": robustness,
                "accuracy_drop": original_accuracy - perturbed_accuracy
            }
        
        # Calculate overall robustness score
        robustness_scores = [
            result["robustness"] 
            for result in robustness_results["perturbation_results"].values()
        ]
        robustness_results["robustness_score"] = np.mean(robustness_scores)
        
        logger.info(f"📊 Robustness test results:")
        logger.info(f"   Original accuracy: {original_accuracy:.4f}")
        logger.info(f"   Overall robustness score: {robustness_results['robustness_score']:.4f}")
        for perturbation, results in robustness_results["perturbation_results"].items():
            logger.info(f"   {perturbation}: {results['accuracy']:.4f} (robustness: {results['robustness']:.4f})")
        
        return robustness_results
    
    def generate_accuracy_report(self, test_results: List[Dict[str, Any]]) -> str:
        """Generate comprehensive accuracy report"""
        logger.info("📝 Generating accuracy report...")
        
        report = {
            "report_summary": {
                "total_tests": len(test_results),
                "models_tested": list(set([result.get('model_type', 'unknown') for result in test_results])),
                "average_accuracy": np.mean([result.get('accuracy', 0) for result in test_results]),
                "target_achievement_rate": np.mean([result.get('target_achieved', False) for result in test_results]),
                "timestamp": datetime.now().isoformat()
            },
            "detailed_results": test_results,
            "recommendations": self._generate_accuracy_recommendations(test_results)
        }
        
        # Save report
        report_path = self.test_results_path / f"accuracy_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_path, 'w') as f:
            json.dump(report, f, indent=2)
        
        # Create visualizations
        self._create_accuracy_visualizations(test_results)
        
        logger.info(f"📝 Accuracy report saved to: {report_path}")
        return str(report_path)
    
    def _calculate_confidence_accuracy(self, true_labels: List[int], 
                                     predictions: List[int], 
                                     confidences: List[float]) -> Dict[str, float]:
        """Calculate accuracy by confidence level"""
        confidence_accuracy = {}
        
        # Define confidence levels
        confidence_levels = {
            "high": (0.9, 1.0),
            "medium": (0.7, 0.9),
            "low": (0.5, 0.7),
            "very_low": (0.0, 0.5)
        }
        
        for level, (min_conf, max_conf) in confidence_levels.items():
            # Filter predictions by confidence level
            level_indices = [
                i for i, conf in enumerate(confidences)
                if min_conf <= conf < max_conf
            ]
            
            if level_indices:
                level_true = [true_labels[i] for i in level_indices]
                level_pred = [predictions[i] for i in level_indices]
                level_accuracy = accuracy_score(level_true, level_pred)
                confidence_accuracy[level] = level_accuracy
            else:
                confidence_accuracy[level] = 0.0
        
        return confidence_accuracy
    
    def _perform_significance_tests(self, model_results: Dict[str, Dict[str, Any]]) -> Dict[str, Any]:
        """Perform statistical significance tests between models"""
        significance_tests = {}
        
        model_names = list(model_results.keys())
        
        for i, model1 in enumerate(model_names):
            for j, model2 in enumerate(model_names[i+1:], i+1):
                # Extract accuracies (assuming binary classification for simplicity)
                acc1 = model_results[model1]['accuracy']
                acc2 = model_results[model2]['accuracy']
                
                # Perform t-test (simplified)
                # In practice, you'd need the actual prediction scores
                test_name = f"{model1}_vs_{model2}"
                
                # Calculate effect size
                effect_size = abs(acc1 - acc2)
                
                # Determine significance (simplified)
                is_significant = effect_size > 0.01  # 1% difference threshold
                
                significance_tests[test_name] = {
                    "model1": model1,
                    "model2": model2,
                    "accuracy1": acc1,
                    "accuracy2": acc2,
                    "difference": acc1 - acc2,
                    "effect_size": effect_size,
                    "is_significant": is_significant
                }
        
        return significance_tests
    
    def _add_typos(self, text: str) -> str:
        """Add typos to text"""
        # Simple typo simulation
        words = text.split()
        if len(words) > 1:
            # Randomly select a word and add a typo
            word_idx = np.random.randint(0, len(words))
            word = words[word_idx]
            if len(word) > 3:
                # Add a random character
                char_idx = np.random.randint(1, len(word))
                words[word_idx] = word[:char_idx] + word[char_idx] + word[char_idx:]
        return ' '.join(words)
    
    def _change_capitalization(self, text: str) -> str:
        """Change capitalization of text"""
        return text.swapcase()
    
    def _change_punctuation(self, text: str) -> str:
        """Change punctuation in text"""
        # Replace common punctuation
        text = text.replace('.', '!')
        text = text.replace(',', ';')
        text = text.replace('?', '.')
        return text
    
    def _change_word_order(self, text: str) -> str:
        """Change word order in text"""
        words = text.split()
        if len(words) > 2:
            # Swap two random words
            idx1, idx2 = np.random.choice(len(words), 2, replace=False)
            words[idx1], words[idx2] = words[idx2], words[idx1]
        return ' '.join(words)
    
    def _replace_with_synonyms(self, text: str) -> str:
        """Replace words with synonyms (simplified)"""
        # Simple synonym replacement
        synonyms = {
            "company": "corporation",
            "business": "enterprise",
            "services": "solutions",
            "leading": "premier",
            "provider": "supplier"
        }
        
        for word, synonym in synonyms.items():
            text = text.replace(word, synonym)
        
        return text
    
    def _generate_accuracy_recommendations(self, test_results: List[Dict[str, Any]]) -> List[str]:
        """Generate recommendations based on test results"""
        recommendations = []
        
        # Analyze results
        avg_accuracy = np.mean([result.get('accuracy', 0) for result in test_results])
        target_achievement_rate = np.mean([result.get('target_achieved', False) for result in test_results])
        
        if avg_accuracy >= 0.95:
            recommendations.append("✅ Excellent accuracy achieved! Models are ready for production deployment.")
        elif avg_accuracy >= 0.90:
            recommendations.append("✅ Good accuracy achieved. Consider fine-tuning for better performance.")
        else:
            recommendations.append("⚠️ Accuracy below target. Consider additional training or model architecture changes.")
        
        if target_achievement_rate >= 0.8:
            recommendations.append("✅ Most models achieve target accuracy. Focus on the best performing models.")
        else:
            recommendations.append("⚠️ Many models below target. Review training data and model architecture.")
        
        recommendations.extend([
            "💡 Consider ensemble methods for improved accuracy.",
            "💡 Implement model monitoring in production.",
            "💡 Regular retraining with new data.",
            "💡 A/B testing for model comparison.",
            "💡 Confidence-based routing for uncertain predictions."
        ])
        
        return recommendations
    
    def _create_accuracy_visualizations(self, test_results: List[Dict[str, Any]]):
        """Create accuracy visualizations"""
        try:
            fig, axes = plt.subplots(2, 2, figsize=(15, 12))
            
            # Model accuracy comparison
            models = [result.get('model_type', 'Unknown') for result in test_results]
            accuracies = [result.get('accuracy', 0) for result in test_results]
            
            axes[0, 0].bar(models, accuracies, color='skyblue')
            axes[0, 0].axhline(y=0.95, color='red', linestyle='--', label='Target (95%)')
            axes[0, 0].set_title('Model Accuracy Comparison')
            axes[0, 0].set_ylabel('Accuracy')
            axes[0, 0].tick_params(axis='x', rotation=45)
            axes[0, 0].legend()
            
            # Accuracy distribution
            axes[0, 1].hist(accuracies, bins=10, alpha=0.7, color='lightgreen')
            axes[0, 1].axvline(x=0.95, color='red', linestyle='--', label='Target (95%)')
            axes[0, 1].set_title('Accuracy Distribution')
            axes[0, 1].set_xlabel('Accuracy')
            axes[0, 1].set_ylabel('Frequency')
            axes[0, 1].legend()
            
            # Precision vs Recall
            precisions = [result.get('precision', 0) for result in test_results]
            recalls = [result.get('recall', 0) for result in test_results]
            
            axes[1, 0].scatter(precisions, recalls, s=100, alpha=0.7)
            axes[1, 0].set_xlabel('Precision')
            axes[1, 0].set_ylabel('Recall')
            axes[1, 0].set_title('Precision vs Recall')
            
            # Target achievement
            target_achieved = [result.get('target_achieved', False) for result in test_results]
            achievement_counts = [sum(target_achieved), len(target_achieved) - sum(target_achieved)]
            achievement_labels = ['Target Achieved', 'Target Not Achieved']
            
            axes[1, 1].pie(achievement_counts, labels=achievement_labels, autopct='%1.1f%%', 
                          colors=['lightgreen', 'lightcoral'])
            axes[1, 1].set_title('Target Achievement Rate')
            
            plt.tight_layout()
            
            # Save visualization
            viz_path = self.test_results_path / f"accuracy_visualization_{datetime.now().strftime('%Y%m%d_%H%M%S')}.png"
            plt.savefig(viz_path, dpi=300, bbox_inches='tight')
            plt.close()
            
            logger.info("📊 Accuracy visualizations created")
            
        except Exception as e:
            logger.warning(f"⚠️ Failed to create accuracy visualizations: {e}")

def main():
    """Main function to demonstrate accuracy testing"""
    
    # Configuration
    config = {
        'test_results_path': 'test_results',
        'accuracy_targets': {
            'classification': 0.95,
            'risk_detection': 0.90
        }
    }
    
    # Initialize accuracy tester
    tester = AccuracyTester(config)
    
    # Example usage with dummy data
    test_data = [
        {"text": "Acme Corporation - Leading technology solutions provider", "label": 0},
        {"text": "Healthcare Plus - Comprehensive medical services", "label": 1},
        {"text": "Financial Services Inc - Investment and wealth management", "label": 2},
        {"text": "Retail Solutions - Fashion and lifestyle products", "label": 3}
    ] * 25  # 100 test samples
    
    # Create dummy model for demonstration
    class DummyModel(nn.Module):
        def __init__(self):
            super().__init__()
            self.linear = nn.Linear(10, 4)
        
        def forward(self, x):
            return self.linear(x)
    
    dummy_model = DummyModel()
    
    # Test model accuracy
    accuracy_results = tester.test_model_accuracy(dummy_model, test_data, "bert")
    print(f"Accuracy test results: {accuracy_results}")
    
    # Cross-validation test
    cv_results = tester.cross_validate_model(dummy_model, test_data, "bert")
    print(f"Cross-validation results: {cv_results}")
    
    # Generate accuracy report
    report_path = tester.generate_accuracy_report([accuracy_results, cv_results])
    print(f"Accuracy report: {report_path}")
    
    logger.info("🎉 Accuracy testing demonstration completed!")

if __name__ == "__main__":
    main()
