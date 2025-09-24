#!/usr/bin/env python3
"""
Confidence Scoring and Explainability Features

This module implements advanced confidence scoring and explainability features for
ML models, including:

- Multi-level confidence scoring (prediction, model, ensemble)
- Attention-based explainability for BERT models
- Feature importance analysis
- Uncertainty quantification
- Model interpretability techniques
- Explainability report generation

Target: Comprehensive confidence scoring and explainability for all ML models
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
import torch.nn.functional as F
import numpy as np
import pandas as pd
from transformers import (
    AutoTokenizer, AutoModel, AutoModelForSequenceClassification,
    BertTokenizer, BertForSequenceClassification,
    DistilBertTokenizer, DistilBertForSequenceClassification
)
from sklearn.metrics import accuracy_score, precision_recall_fscore_support
from sklearn.ensemble import RandomForestClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.preprocessing import StandardScaler
import shap
import lime
import lime.lime_tabular
from lime.lime_text import LimeTextExplainer
import matplotlib.pyplot as plt
import seaborn as sns
from tqdm import tqdm

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class ConfidenceScorer:
    """Advanced confidence scoring for ML models"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        # Create directories
        self.explainability_path = Path(config.get('explainability_path', 'explainability'))
        self.explainability_path.mkdir(parents=True, exist_ok=True)
        
        # Confidence thresholds
        self.confidence_thresholds = {
            "high": 0.9,
            "medium": 0.7,
            "low": 0.5,
            "very_low": 0.3
        }
        
        logger.info(f"🎯 Confidence Scorer initialized")
        logger.info(f"📱 Device: {self.device}")
        logger.info(f"💾 Explainability path: {self.explainability_path}")
    
    def calculate_prediction_confidence(self, logits: torch.Tensor, 
                                      method: str = "softmax") -> Dict[str, float]:
        """Calculate prediction confidence using various methods"""
        
        if method == "softmax":
            probabilities = F.softmax(logits, dim=-1)
            max_prob = torch.max(probabilities, dim=-1)[0]
            confidence = max_prob.item()
            
        elif method == "entropy":
            probabilities = F.softmax(logits, dim=-1)
            entropy = -torch.sum(probabilities * torch.log(probabilities + 1e-8), dim=-1)
            max_entropy = torch.log(torch.tensor(logits.size(-1), dtype=torch.float))
            confidence = 1 - (entropy / max_entropy).item()
            
        elif method == "margin":
            probabilities = F.softmax(logits, dim=-1)
            sorted_probs, _ = torch.sort(probabilities, descending=True)
            margin = sorted_probs[0] - sorted_probs[1]
            confidence = margin.item()
            
        elif method == "temperature_scaling":
            # Temperature scaling for better calibration
            temperature = 2.0  # Learned parameter
            scaled_logits = logits / temperature
            probabilities = F.softmax(scaled_logits, dim=-1)
            max_prob = torch.max(probabilities, dim=-1)[0]
            confidence = max_prob.item()
            
        else:
            raise ValueError(f"Unknown confidence method: {method}")
        
        return {
            "confidence": confidence,
            "method": method,
            "confidence_level": self._get_confidence_level(confidence)
        }
    
    def calculate_model_confidence(self, model: nn.Module, test_data: List[Dict],
                                 model_type: str = "bert") -> Dict[str, Any]:
        """Calculate overall model confidence"""
        logger.info(f"📊 Calculating model confidence for {model_type.upper()}...")
        
        model.eval()
        confidences = []
        predictions = []
        true_labels = []
        
        # Load tokenizer
        if model_type.lower() == "distilbert":
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        with torch.no_grad():
            for item in tqdm(test_data, desc="Calculating model confidence"):
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
                
                # Calculate confidence
                confidence_result = self.calculate_prediction_confidence(logits)
                confidences.append(confidence_result['confidence'])
                
                # Get prediction
                prediction = torch.argmax(logits, dim=-1).item()
                predictions.append(prediction)
                true_labels.append(true_label)
        
        # Calculate model-level metrics
        accuracy = accuracy_score(true_labels, predictions)
        avg_confidence = np.mean(confidences)
        confidence_std = np.std(confidences)
        
        # Calculate confidence-accuracy correlation
        confidence_accuracy_corr = np.corrcoef(confidences, 
                                             [1 if p == t else 0 for p, t in zip(predictions, true_labels)])[0, 1]
        
        model_confidence = {
            "model_type": model_type,
            "accuracy": accuracy,
            "average_confidence": avg_confidence,
            "confidence_std": confidence_std,
            "confidence_accuracy_correlation": confidence_accuracy_corr,
            "confidence_distribution": {
                "high": sum(1 for c in confidences if c >= self.confidence_thresholds["high"]),
                "medium": sum(1 for c in confidences if self.confidence_thresholds["medium"] <= c < self.confidence_thresholds["high"]),
                "low": sum(1 for c in confidences if self.confidence_thresholds["low"] <= c < self.confidence_thresholds["medium"]),
                "very_low": sum(1 for c in confidences if c < self.confidence_thresholds["low"])
            },
            "total_samples": len(test_data)
        }
        
        logger.info(f"📊 Model confidence calculated:")
        logger.info(f"   Accuracy: {accuracy:.4f}")
        logger.info(f"   Average confidence: {avg_confidence:.4f}")
        logger.info(f"   Confidence-accuracy correlation: {confidence_accuracy_corr:.4f}")
        
        return model_confidence
    
    def calculate_ensemble_confidence(self, models: List[nn.Module], 
                                    test_data: List[Dict]) -> Dict[str, Any]:
        """Calculate ensemble confidence from multiple models"""
        logger.info("🎯 Calculating ensemble confidence...")
        
        ensemble_predictions = []
        ensemble_confidences = []
        
        for model in models:
            model.eval()
            model_predictions = []
            model_confidences = []
            
            # Load appropriate tokenizer
            if hasattr(model, 'config') and 'distilbert' in str(model.config).lower():
                tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
            else:
                tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
            
            with torch.no_grad():
                for item in tqdm(test_data, desc=f"Processing model {len(ensemble_predictions)+1}"):
                    text = item.get('text', '')
                    
                    inputs = tokenizer(
                        text,
                        max_length=512,
                        padding=True,
                        truncation=True,
                        return_tensors="pt"
                    )
                    
                    outputs = model(**inputs)
                    logits = outputs.logits
                    
                    # Calculate confidence
                    confidence_result = self.calculate_prediction_confidence(logits)
                    model_confidences.append(confidence_result['confidence'])
                    
                    # Get prediction
                    prediction = torch.argmax(logits, dim=-1).item()
                    model_predictions.append(prediction)
            
            ensemble_predictions.append(model_predictions)
            ensemble_confidences.append(model_confidences)
        
        # Calculate ensemble metrics
        ensemble_predictions = np.array(ensemble_predictions)
        ensemble_confidences = np.array(ensemble_confidences)
        
        # Average predictions and confidences
        avg_predictions = np.mean(ensemble_predictions, axis=0)
        avg_confidences = np.mean(ensemble_confidences, axis=0)
        
        # Calculate agreement between models
        model_agreement = []
        for i in range(len(test_data)):
            predictions_for_sample = ensemble_predictions[:, i]
            agreement = len(set(predictions_for_sample)) == 1  # All models agree
            model_agreement.append(agreement)
        
        agreement_rate = np.mean(model_agreement)
        
        ensemble_confidence = {
            "num_models": len(models),
            "average_confidence": np.mean(avg_confidences),
            "confidence_std": np.std(avg_confidences),
            "model_agreement_rate": agreement_rate,
            "confidence_distribution": {
                "high": sum(1 for c in avg_confidences if c >= self.confidence_thresholds["high"]),
                "medium": sum(1 for c in avg_confidences if self.confidence_thresholds["medium"] <= c < self.confidence_thresholds["high"]),
                "low": sum(1 for c in avg_confidences if self.confidence_thresholds["low"] <= c < self.confidence_thresholds["medium"]),
                "very_low": sum(1 for c in avg_confidences if c < self.confidence_thresholds["low"])
            },
            "total_samples": len(test_data)
        }
        
        logger.info(f"🎯 Ensemble confidence calculated:")
        logger.info(f"   Number of models: {len(models)}")
        logger.info(f"   Average confidence: {np.mean(avg_confidences):.4f}")
        logger.info(f"   Model agreement rate: {agreement_rate:.4f}")
        
        return ensemble_confidence
    
    def extract_attention_weights(self, model: nn.Module, text: str, 
                                model_type: str = "bert") -> Dict[str, Any]:
        """Extract attention weights for explainability"""
        logger.info(f"🔍 Extracting attention weights for {model_type.upper()}...")
        
        # Load tokenizer
        if model_type.lower() == "distilbert":
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        # Tokenize input
        inputs = tokenizer(
            text,
            max_length=512,
            padding=True,
            truncation=True,
            return_tensors="pt"
        )
        
        tokens = tokenizer.convert_ids_to_tokens(inputs['input_ids'][0])
        
        # Get attention weights
        model.eval()
        with torch.no_grad():
            outputs = model(**inputs, output_attentions=True)
            attentions = outputs.attentions  # List of attention tensors
        
        # Process attention weights
        attention_analysis = {
            "tokens": tokens,
            "num_layers": len(attentions),
            "num_heads": attentions[0].size(1),
            "attention_weights": [],
            "layer_importance": [],
            "head_importance": []
        }
        
        # Analyze each layer
        for layer_idx, attention in enumerate(attentions):
            # Average across heads
            avg_attention = torch.mean(attention, dim=1)  # [batch, seq, seq]
            
            # Get attention to [CLS] token (first token)
            cls_attention = avg_attention[0, 0, :].cpu().numpy()
            
            attention_analysis["attention_weights"].append(cls_attention.tolist())
            
            # Calculate layer importance
            layer_importance = torch.mean(avg_attention).item()
            attention_analysis["layer_importance"].append(layer_importance)
        
        # Calculate head importance
        for head_idx in range(attentions[0].size(1)):
            head_attention = torch.mean(attentions[0][0, head_idx, :, :]).item()
            attention_analysis["head_importance"].append(head_attention)
        
        logger.info(f"🔍 Attention weights extracted:")
        logger.info(f"   Number of layers: {len(attentions)}")
        logger.info(f"   Number of heads: {attentions[0].size(1)}")
        logger.info(f"   Number of tokens: {len(tokens)}")
        
        return attention_analysis
    
    def create_lime_explanation(self, model: nn.Module, text: str, 
                              model_type: str = "bert") -> Dict[str, Any]:
        """Create LIME explanation for model prediction"""
        logger.info(f"🍋 Creating LIME explanation for {model_type.upper()}...")
        
        # Load tokenizer
        if model_type.lower() == "distilbert":
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        # Create LIME explainer
        explainer = LimeTextExplainer(class_names=[f"Class_{i}" for i in range(16)])
        
        # Define prediction function
        def predict_proba(texts):
            model.eval()
            probabilities = []
            
            for text in texts:
                inputs = tokenizer(
                    text,
                    max_length=512,
                    padding=True,
                    truncation=True,
                    return_tensors="pt"
                )
                
                with torch.no_grad():
                    outputs = model(**inputs)
                    logits = outputs.logits
                    prob = F.softmax(logits, dim=-1)
                    probabilities.append(prob.cpu().numpy()[0])
            
            return np.array(probabilities)
        
        # Generate explanation
        explanation = explainer.explain_instance(text, predict_proba, num_features=10)
        
        # Extract explanation data
        lime_explanation = {
            "text": text,
            "explanation": explanation.as_list(),
            "explanation_html": explanation.as_html(),
            "prediction": explanation.predict_proba.tolist(),
            "num_features": len(explanation.as_list())
        }
        
        logger.info(f"🍋 LIME explanation created with {len(explanation.as_list())} features")
        return lime_explanation
    
    def create_shap_explanation(self, model: nn.Module, text: str, 
                              model_type: str = "bert") -> Dict[str, Any]:
        """Create SHAP explanation for model prediction"""
        logger.info(f"🔮 Creating SHAP explanation for {model_type.upper()}...")
        
        # Load tokenizer
        if model_type.lower() == "distilbert":
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        # Tokenize input
        inputs = tokenizer(
            text,
            max_length=512,
            padding=True,
            truncation=True,
            return_tensors="pt"
        )
        
        tokens = tokenizer.convert_ids_to_tokens(inputs['input_ids'][0])
        
        # Create SHAP explainer
        def model_predict(texts):
            model.eval()
            predictions = []
            
            for text in texts:
                inputs = tokenizer(
                    text,
                    max_length=512,
                    padding=True,
                    truncation=True,
                    return_tensors="pt"
                )
                
                with torch.no_grad():
                    outputs = model(**inputs)
                    logits = outputs.logits
                    prob = F.softmax(logits, dim=-1)
                    predictions.append(prob.cpu().numpy()[0])
            
            return np.array(predictions)
        
        # Create SHAP explainer
        explainer = shap.Explainer(model_predict, tokenizer)
        
        # Generate explanation
        shap_values = explainer([text])
        
        # Extract explanation data
        shap_explanation = {
            "text": text,
            "tokens": tokens,
            "shap_values": shap_values.values[0].tolist(),
            "base_values": shap_values.base_values[0].tolist(),
            "data": shap_values.data[0].tolist()
        }
        
        logger.info(f"🔮 SHAP explanation created")
        return shap_explanation
    
    def create_comprehensive_explainability_report(self, model: nn.Module, text: str,
                                                 model_type: str = "bert") -> Dict[str, Any]:
        """Create comprehensive explainability report"""
        logger.info(f"📊 Creating comprehensive explainability report for {model_type.upper()}...")
        
        # Get prediction
        if model_type.lower() == "distilbert":
            tokenizer = DistilBertTokenizer.from_pretrained("distilbert-base-uncased")
        else:
            tokenizer = BertTokenizer.from_pretrained("bert-base-uncased")
        
        inputs = tokenizer(
            text,
            max_length=512,
            padding=True,
            truncation=True,
            return_tensors="pt"
        )
        
        model.eval()
        with torch.no_grad():
            outputs = model(**inputs)
            logits = outputs.logits
            probabilities = F.softmax(logits, dim=-1)
            prediction = torch.argmax(logits, dim=-1).item()
            confidence = torch.max(probabilities, dim=-1)[0].item()
        
        # Create comprehensive report
        explainability_report = {
            "input_text": text,
            "model_type": model_type,
            "prediction": prediction,
            "confidence": confidence,
            "confidence_level": self._get_confidence_level(confidence),
            "probabilities": probabilities[0].tolist(),
            "timestamp": datetime.now().isoformat(),
            "explanations": {}
        }
        
        # Add attention weights
        try:
            attention_analysis = self.extract_attention_weights(model, text, model_type)
            explainability_report["explanations"]["attention"] = attention_analysis
        except Exception as e:
            logger.warning(f"⚠️ Failed to extract attention weights: {e}")
            explainability_report["explanations"]["attention"] = {"error": str(e)}
        
        # Add LIME explanation
        try:
            lime_explanation = self.create_lime_explanation(model, text, model_type)
            explainability_report["explanations"]["lime"] = lime_explanation
        except Exception as e:
            logger.warning(f"⚠️ Failed to create LIME explanation: {e}")
            explainability_report["explanations"]["lime"] = {"error": str(e)}
        
        # Add SHAP explanation
        try:
            shap_explanation = self.create_shap_explanation(model, text, model_type)
            explainability_report["explanations"]["shap"] = shap_explanation
        except Exception as e:
            logger.warning(f"⚠️ Failed to create SHAP explanation: {e}")
            explainability_report["explanations"]["shap"] = {"error": str(e)}
        
        # Save report
        report_path = self.explainability_path / f"explainability_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(report_path, 'w') as f:
            json.dump(explainability_report, f, indent=2)
        
        logger.info(f"📊 Comprehensive explainability report created")
        logger.info(f"💾 Report saved to: {report_path}")
        
        return explainability_report
    
    def _get_confidence_level(self, confidence: float) -> str:
        """Get confidence level based on threshold"""
        if confidence >= self.confidence_thresholds["high"]:
            return "high"
        elif confidence >= self.confidence_thresholds["medium"]:
            return "medium"
        elif confidence >= self.confidence_thresholds["low"]:
            return "low"
        else:
            return "very_low"
    
    def create_confidence_visualization(self, confidence_data: Dict[str, Any]) -> str:
        """Create confidence visualization"""
        logger.info("📊 Creating confidence visualization...")
        
        try:
            fig, axes = plt.subplots(2, 2, figsize=(15, 12))
            
            # Confidence distribution
            if "confidence_distribution" in confidence_data:
                dist = confidence_data["confidence_distribution"]
                labels = list(dist.keys())
                values = list(dist.values())
                
                axes[0, 0].bar(labels, values, color=['green', 'yellow', 'orange', 'red'])
                axes[0, 0].set_title('Confidence Level Distribution')
                axes[0, 0].set_ylabel('Number of Samples')
            
            # Confidence vs Accuracy scatter
            if "confidence_accuracy_correlation" in confidence_data:
                corr = confidence_data["confidence_accuracy_correlation"]
                axes[0, 1].text(0.5, 0.5, f'Correlation: {corr:.4f}', 
                               ha='center', va='center', fontsize=16)
                axes[0, 1].set_title('Confidence-Accuracy Correlation')
            
            # Model performance metrics
            metrics = ["accuracy", "average_confidence", "confidence_std"]
            values = [confidence_data.get(metric, 0) for metric in metrics]
            
            axes[1, 0].bar(metrics, values, color=['blue', 'green', 'orange'])
            axes[1, 0].set_title('Model Performance Metrics')
            axes[1, 0].set_ylabel('Value')
            axes[1, 0].tick_params(axis='x', rotation=45)
            
            # Confidence histogram
            if "confidences" in confidence_data:
                confidences = confidence_data["confidences"]
                axes[1, 1].hist(confidences, bins=20, alpha=0.7, color='skyblue')
                axes[1, 1].set_title('Confidence Score Distribution')
                axes[1, 1].set_xlabel('Confidence Score')
                axes[1, 1].set_ylabel('Frequency')
            
            plt.tight_layout()
            
            # Save visualization
            viz_path = self.explainability_path / f"confidence_visualization_{datetime.now().strftime('%Y%m%d_%H%M%S')}.png"
            plt.savefig(viz_path, dpi=300, bbox_inches='tight')
            plt.close()
            
            logger.info(f"📊 Confidence visualization saved to: {viz_path}")
            return str(viz_path)
            
        except Exception as e:
            logger.warning(f"⚠️ Failed to create confidence visualization: {e}")
            return ""

def main():
    """Main function to demonstrate confidence scoring and explainability"""
    
    # Configuration
    config = {
        'explainability_path': 'explainability',
        'confidence_thresholds': {
            'high': 0.9,
            'medium': 0.7,
            'low': 0.5,
            'very_low': 0.3
        }
    }
    
    # Initialize confidence scorer
    scorer = ConfidenceScorer(config)
    
    # Example usage
    test_text = "Acme Corporation - Leading provider of financial services and investment solutions"
    
    # Create dummy model for demonstration
    class DummyModel(nn.Module):
        def __init__(self):
            super().__init__()
            self.linear = nn.Linear(10, 16)
        
        def forward(self, x):
            return self.linear(x)
    
    dummy_model = DummyModel()
    
    # Calculate prediction confidence
    dummy_logits = torch.randn(1, 16)
    confidence_result = scorer.calculate_prediction_confidence(dummy_logits)
    print(f"Prediction confidence: {confidence_result}")
    
    # Create comprehensive explainability report
    explainability_report = scorer.create_comprehensive_explainability_report(
        dummy_model, test_text, "bert"
    )
    print(f"Explainability report created with {len(explainability_report['explanations'])} explanations")
    
    logger.info("🎉 Confidence scoring and explainability demonstration completed!")

if __name__ == "__main__":
    main()
