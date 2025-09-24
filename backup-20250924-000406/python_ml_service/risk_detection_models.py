#!/usr/bin/env python3
"""
Enhanced Risk Detection Models for Business Verification

This module provides sophisticated ML models for risk detection including:
- BERT-based risk classification models
- Anomaly detection models for unusual patterns
- Pattern recognition models for complex risk scenarios
- Risk scoring and confidence metrics
- Real-time risk assessment capabilities

Target: 90%+ accuracy for risk detection
"""

import os
import json
import time
import logging
import re
import numpy as np
import pandas as pd
from typing import Dict, List, Optional, Any, Tuple, Union
from datetime import datetime, timedelta
from pathlib import Path

import torch
import torch.nn as nn
import torch.optim as optim
from torch.utils.data import DataLoader, Dataset
import transformers
from transformers import (
    AutoTokenizer, AutoModel, AutoModelForSequenceClassification,
    BertTokenizer, BertForSequenceClassification,
    DistilBertTokenizer, DistilBertForSequenceClassification,
    TrainingArguments, Trainer, EarlyStoppingCallback
)
from sklearn.metrics import accuracy_score, precision_recall_fscore_support, classification_report
from sklearn.model_selection import train_test_split
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
from sklearn.decomposition import PCA
import joblib

# Configure logging
logger = logging.getLogger(__name__)

class RiskDetectionConfig:
    """Configuration for risk detection models"""
    
    # Model configuration
    BERT_MODEL_NAME = "bert-base-uncased"
    DISTILBERT_MODEL_NAME = "distilbert-base-uncased"
    MAX_SEQUENCE_LENGTH = 512
    BATCH_SIZE = 16
    LEARNING_RATE = 2e-5
    NUM_EPOCHS = 5
    WARMUP_STEPS = 500
    WEIGHT_DECAY = 0.01
    
    # Risk categories
    RISK_CATEGORIES = [
        "illegal", "prohibited", "high_risk", "tbml", "sanctions", "fraud"
    ]
    
    # Risk severity levels
    RISK_SEVERITY_LEVELS = ["low", "medium", "high", "critical"]
    
    # Performance targets
    TARGET_RISK_ACCURACY = 0.90
    TARGET_INFERENCE_TIME = 0.1  # 100ms
    
    # Model paths
    MODEL_SAVE_PATH = Path("models/risk_detection")
    CACHE_PATH = Path("cache/risk_detection")
    DATA_PATH = Path("data/risk_detection")
    
    # Device configuration
    DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")

# Create directories
RiskDetectionConfig.MODEL_SAVE_PATH.mkdir(parents=True, exist_ok=True)
RiskDetectionConfig.CACHE_PATH.mkdir(parents=True, exist_ok=True)
RiskDetectionConfig.DATA_PATH.mkdir(parents=True, exist_ok=True)

class RiskDetectionDataset(Dataset):
    """Dataset for risk detection training"""
    
    def __init__(self, texts: List[str], labels: List[int], tokenizer, max_length: int = 512):
        self.texts = texts
        self.labels = labels
        self.tokenizer = tokenizer
        self.max_length = max_length
    
    def __len__(self):
        return len(self.texts)
    
    def __getitem__(self, idx):
        text = str(self.texts[idx])
        label = self.labels[idx]
        
        encoding = self.tokenizer(
            text,
            truncation=True,
            padding='max_length',
            max_length=self.max_length,
            return_tensors='pt'
        )
        
        return {
            'input_ids': encoding['input_ids'].flatten(),
            'attention_mask': encoding['attention_mask'].flatten(),
            'labels': torch.tensor(label, dtype=torch.long)
        }

class BERTRiskClassificationModel(nn.Module):
    """Enhanced BERT model for risk classification"""
    
    def __init__(self, model_name: str, num_risk_categories: int, num_severity_levels: int):
        super().__init__()
        self.bert = AutoModel.from_pretrained(model_name)
        self.dropout = nn.Dropout(0.3)
        
        # Risk category classification head
        self.risk_category_classifier = nn.Linear(self.bert.config.hidden_size, num_risk_categories)
        
        # Risk severity classification head
        self.risk_severity_classifier = nn.Linear(self.bert.config.hidden_size, num_severity_levels)
        
        # Risk score regression head
        self.risk_score_regressor = nn.Sequential(
            nn.Linear(self.bert.config.hidden_size, 256),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(256, 128),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(128, 1),
            nn.Sigmoid()
        )
        
        # Attention mechanism for explainability
        self.attention = nn.MultiheadAttention(
            embed_dim=self.bert.config.hidden_size,
            num_heads=8,
            dropout=0.1
        )
    
    def forward(self, input_ids, attention_mask=None):
        # Get BERT outputs
        outputs = self.bert(input_ids=input_ids, attention_mask=attention_mask)
        sequence_output = outputs.last_hidden_state
        pooled_output = outputs.pooler_output
        
        # Apply attention mechanism
        attn_output, attn_weights = self.attention(
            sequence_output.transpose(0, 1),
            sequence_output.transpose(0, 1),
            sequence_output.transpose(0, 1)
        )
        attn_output = attn_output.transpose(0, 1)
        
        # Combine pooled output with attention
        enhanced_output = pooled_output + attn_output.mean(dim=1)
        enhanced_output = self.dropout(enhanced_output)
        
        # Get predictions
        risk_category_logits = self.risk_category_classifier(enhanced_output)
        risk_severity_logits = self.risk_severity_classifier(enhanced_output)
        risk_score = self.risk_score_regressor(enhanced_output)
        
        return {
            'risk_category_logits': risk_category_logits,
            'risk_severity_logits': risk_severity_logits,
            'risk_score': risk_score,
            'attention_weights': attn_weights
        }

class AnomalyDetectionModel:
    """Anomaly detection model for unusual patterns"""
    
    def __init__(self, contamination: float = 0.1):
        self.isolation_forest = IsolationForest(
            contamination=contamination,
            random_state=42,
            n_estimators=100
        )
        self.scaler = StandardScaler()
        self.pca = PCA(n_components=0.95)
        self.is_fitted = False
    
    def extract_features(self, texts: List[str]) -> np.ndarray:
        """Extract features from text for anomaly detection"""
        features = []
        
        for text in texts:
            # Basic text features
            text_features = [
                len(text),  # Length
                len(text.split()),  # Word count
                text.count(' '),  # Space count
                text.count('\n'),  # Line count
                text.count('!'),  # Exclamation count
                text.count('?'),  # Question count
                text.count('$'),  # Dollar sign count
                text.count('%'),  # Percentage count
                text.count('@'),  # At sign count
                text.count('#'),  # Hash count
                text.count('&'),  # Ampersand count
                text.count('*'),  # Asterisk count
                text.count('+'),  # Plus count
                text.count('-'),  # Minus count
                text.count('='),  # Equals count
                text.count('_'),  # Underscore count
                text.count('|'),  # Pipe count
                text.count('\\'),  # Backslash count
                text.count('/'),  # Forward slash count
                text.count('('),  # Left parenthesis count
                text.count(')'),  # Right parenthesis count
                text.count('['),  # Left bracket count
                text.count(']'),  # Right bracket count
                text.count('{'),  # Left brace count
                text.count('}'),  # Right brace count
                text.count('<'),  # Less than count
                text.count('>'),  # Greater than count
                text.count('"'),  # Double quote count
                text.count("'"),  # Single quote count
                text.count('`'),  # Backtick count
                text.count('~'),  # Tilde count
                text.count('^'),  # Caret count
                text.count(';'),  # Semicolon count
                text.count(':'),  # Colon count
                text.count(','),  # Comma count
                text.count('.'),  # Period count
            ]
            
            # Character frequency features
            char_counts = {}
            for char in text.lower():
                char_counts[char] = char_counts.get(char, 0) + 1
            
            # Add character frequency features (top 26 most common)
            common_chars = 'abcdefghijklmnopqrstuvwxyz'
            for char in common_chars:
                features.append(char_counts.get(char, 0))
            
            # Add basic text features
            features.extend(text_features)
        
        return np.array(features).reshape(len(texts), -1)
    
    def fit(self, texts: List[str]):
        """Fit the anomaly detection model"""
        logger.info("Training anomaly detection model...")
        
        # Extract features
        features = self.extract_features(texts)
        
        # Scale features
        features_scaled = self.scaler.fit_transform(features)
        
        # Apply PCA
        features_pca = self.pca.fit_transform(features_scaled)
        
        # Fit isolation forest
        self.isolation_forest.fit(features_pca)
        self.is_fitted = True
        
        logger.info(f"Anomaly detection model trained on {len(texts)} samples")
        logger.info(f"Feature dimensions: {features.shape[1]} -> {features_pca.shape[1]} (after PCA)")
    
    def predict(self, texts: List[str]) -> Tuple[np.ndarray, np.ndarray]:
        """Predict anomalies in texts"""
        if not self.is_fitted:
            raise ValueError("Model must be fitted before prediction")
        
        # Extract features
        features = self.extract_features(texts)
        
        # Scale features
        features_scaled = self.scaler.transform(features)
        
        # Apply PCA
        features_pca = self.pca.transform(features_scaled)
        
        # Predict anomalies
        anomaly_scores = self.isolation_forest.decision_function(features_pca)
        is_anomaly = self.isolation_forest.predict(features_pca)
        
        return is_anomaly, anomaly_scores
    
    def save(self, path: str):
        """Save the model"""
        model_data = {
            'isolation_forest': self.isolation_forest,
            'scaler': self.scaler,
            'pca': self.pca,
            'is_fitted': self.is_fitted
        }
        joblib.dump(model_data, path)
        logger.info(f"Anomaly detection model saved to {path}")
    
    def load(self, path: str):
        """Load the model"""
        model_data = joblib.load(path)
        self.isolation_forest = model_data['isolation_forest']
        self.scaler = model_data['scaler']
        self.pca = model_data['pca']
        self.is_fitted = model_data['is_fitted']
        logger.info(f"Anomaly detection model loaded from {path}")

class PatternRecognitionModel:
    """Pattern recognition model for complex risk scenarios"""
    
    def __init__(self):
        self.patterns = {
            'money_laundering': [
                r'\b(shell company|front company|straw man|nominee)\b',
                r'\b(offshore|tax haven|cayman|bermuda|swiss)\b',
                r'\b(round trip|circular|layering|integration)\b',
                r'\b(cash intensive|high volume|bulk cash)\b',
                r'\b(structured|smurfing|placement|lifting)\b'
            ],
            'terrorist_financing': [
                r'\b(terrorist|extremist|radical|jihad|islamic state)\b',
                r'\b(funding|financing|support|donation|charity)\b',
                r'\b(weapon|explosive|bomb|attack|violence)\b',
                r'\b(recruitment|training|camp|cell|network)\b'
            ],
            'drug_trafficking': [
                r'\b(drug|cocaine|heroin|marijuana|cannabis|methamphetamine)\b',
                r'\b(trafficking|smuggling|distribution|dealer|cartel)\b',
                r'\b(illegal|contraband|underground|black market)\b',
                r'\b(production|manufacturing|cultivation|growing)\b'
            ],
            'fraud': [
                r'\b(fraud|scam|fake|counterfeit|forgery)\b',
                r'\b(identity theft|phishing|social engineering)\b',
                r'\b(pyramid scheme|ponzi|investment fraud)\b',
                r'\b(credit card fraud|bank fraud|wire fraud)\b'
            ],
            'sanctions_evasion': [
                r'\b(sanctions|embargo|blocked|prohibited)\b',
                r'\b(ofac|treasury|compliance|violation)\b',
                r'\b(iran|north korea|russia|venezuela|cuba)\b',
                r'\b(evasion|circumvention|bypass|workaround)\b'
            ]
        }
        
        self.compiled_patterns = {}
        for category, patterns in self.patterns.items():
            self.compiled_patterns[category] = [
                re.compile(pattern, re.IGNORECASE) for pattern in patterns
            ]
    
    def detect_patterns(self, text: str) -> Dict[str, List[Dict]]:
        """Detect risk patterns in text"""
        detected_patterns = {}
        
        for category, compiled_patterns in self.compiled_patterns.items():
            category_matches = []
            
            for pattern in compiled_patterns:
                matches = pattern.findall(text)
                if matches:
                    category_matches.append({
                        'pattern': pattern.pattern,
                        'matches': matches,
                        'count': len(matches)
                    })
            
            if category_matches:
                detected_patterns[category] = category_matches
        
        return detected_patterns
    
    def calculate_pattern_risk_score(self, detected_patterns: Dict[str, List[Dict]]) -> float:
        """Calculate risk score based on detected patterns"""
        if not detected_patterns:
            return 0.0
        
        # Risk weights for different categories
        category_weights = {
            'money_laundering': 0.9,
            'terrorist_financing': 1.0,
            'drug_trafficking': 0.95,
            'fraud': 0.8,
            'sanctions_evasion': 0.85
        }
        
        total_score = 0.0
        total_weight = 0.0
        
        for category, matches in detected_patterns.items():
            weight = category_weights.get(category, 0.5)
            category_score = min(1.0, sum(match['count'] for match in matches) * 0.2)
            
            total_score += category_score * weight
            total_weight += weight
        
        return min(1.0, total_score / total_weight if total_weight > 0 else 0.0)

class RiskDetectionModelManager:
    """Manager for all risk detection models"""
    
    def __init__(self):
        self.bert_model = None
        self.bert_tokenizer = None
        self.anomaly_model = None
        self.pattern_model = None
        self.is_initialized = False
    
    def initialize_models(self):
        """Initialize all risk detection models"""
        logger.info("Initializing risk detection models...")
        
        try:
            # Initialize BERT model
            self.bert_tokenizer = BertTokenizer.from_pretrained(RiskDetectionConfig.BERT_MODEL_NAME)
            self.bert_model = BERTRiskClassificationModel(
                model_name=RiskDetectionConfig.BERT_MODEL_NAME,
                num_risk_categories=len(RiskDetectionConfig.RISK_CATEGORIES),
                num_severity_levels=len(RiskDetectionConfig.RISK_SEVERITY_LEVELS)
            )
            self.bert_model.to(RiskDetectionConfig.DEVICE)
            self.bert_model.eval()
            
            # Initialize anomaly detection model
            self.anomaly_model = AnomalyDetectionModel()
            
            # Initialize pattern recognition model
            self.pattern_model = PatternRecognitionModel()
            
            self.is_initialized = True
            logger.info("✅ All risk detection models initialized successfully")
            
        except Exception as e:
            logger.error(f"❌ Failed to initialize risk detection models: {e}")
            raise
    
    def detect_risk(self, text: str) -> Dict[str, Any]:
        """Comprehensive risk detection using all models"""
        if not self.is_initialized:
            self.initialize_models()
        
        start_time = time.time()
        
        # BERT-based risk classification
        bert_results = self._bert_risk_classification(text)
        
        # Anomaly detection
        anomaly_results = self._anomaly_detection(text)
        
        # Pattern recognition
        pattern_results = self._pattern_recognition(text)
        
        # Combine results
        combined_results = self._combine_results(bert_results, anomaly_results, pattern_results)
        
        processing_time = time.time() - start_time
        combined_results['processing_time'] = processing_time
        
        return combined_results
    
    def _bert_risk_classification(self, text: str) -> Dict[str, Any]:
        """BERT-based risk classification"""
        try:
            # Tokenize input
            inputs = self.bert_tokenizer(
                text,
                max_length=RiskDetectionConfig.MAX_SEQUENCE_LENGTH,
                padding=True,
                truncation=True,
                return_tensors="pt"
            )
            
            # Move to device
            inputs = {k: v.to(RiskDetectionConfig.DEVICE) for k, v in inputs.items()}
            
            # Make prediction
            with torch.no_grad():
                outputs = self.bert_model(**inputs)
                
                # Get risk category predictions
                risk_category_probs = torch.softmax(outputs['risk_category_logits'], dim=-1)
                risk_category_pred = torch.argmax(risk_category_probs, dim=-1)
                
                # Get risk severity predictions
                risk_severity_probs = torch.softmax(outputs['risk_severity_logits'], dim=-1)
                risk_severity_pred = torch.argmax(risk_severity_probs, dim=-1)
                
                # Get risk score
                risk_score = outputs['risk_score'].item()
                
                return {
                    'risk_category': RiskDetectionConfig.RISK_CATEGORIES[risk_category_pred.item()],
                    'risk_category_confidence': risk_category_probs[0][risk_category_pred].item(),
                    'risk_severity': RiskDetectionConfig.RISK_SEVERITY_LEVELS[risk_severity_pred.item()],
                    'risk_severity_confidence': risk_severity_probs[0][risk_severity_pred].item(),
                    'risk_score': risk_score,
                    'attention_weights': outputs['attention_weights']
                }
                
        except Exception as e:
            logger.error(f"BERT risk classification failed: {e}")
            return {
                'risk_category': 'unknown',
                'risk_category_confidence': 0.0,
                'risk_severity': 'low',
                'risk_severity_confidence': 0.0,
                'risk_score': 0.0,
                'attention_weights': None
            }
    
    def _anomaly_detection(self, text: str) -> Dict[str, Any]:
        """Anomaly detection for unusual patterns"""
        try:
            if not self.anomaly_model.is_fitted:
                return {
                    'is_anomaly': False,
                    'anomaly_score': 0.0,
                    'anomaly_confidence': 0.0
                }
            
            is_anomaly, anomaly_scores = self.anomaly_model.predict([text])
            
            return {
                'is_anomaly': bool(is_anomaly[0] == -1),
                'anomaly_score': float(anomaly_scores[0]),
                'anomaly_confidence': abs(float(anomaly_scores[0]))
            }
            
        except Exception as e:
            logger.error(f"Anomaly detection failed: {e}")
            return {
                'is_anomaly': False,
                'anomaly_score': 0.0,
                'anomaly_confidence': 0.0
            }
    
    def _pattern_recognition(self, text: str) -> Dict[str, Any]:
        """Pattern recognition for complex risk scenarios"""
        try:
            detected_patterns = self.pattern_model.detect_patterns(text)
            pattern_risk_score = self.pattern_model.calculate_pattern_risk_score(detected_patterns)
            
            return {
                'detected_patterns': detected_patterns,
                'pattern_risk_score': pattern_risk_score,
                'pattern_confidence': min(1.0, len(detected_patterns) * 0.3)
            }
            
        except Exception as e:
            logger.error(f"Pattern recognition failed: {e}")
            return {
                'detected_patterns': {},
                'pattern_risk_score': 0.0,
                'pattern_confidence': 0.0
            }
    
    def _combine_results(self, bert_results: Dict, anomaly_results: Dict, pattern_results: Dict) -> Dict[str, Any]:
        """Combine results from all models"""
        # Calculate overall risk score
        bert_score = bert_results['risk_score']
        anomaly_score = 0.8 if anomaly_results['is_anomaly'] else 0.2
        pattern_score = pattern_results['pattern_risk_score']
        
        # Weighted combination
        overall_risk_score = (
            bert_score * 0.5 +
            anomaly_score * 0.3 +
            pattern_score * 0.2
        )
        
        # Determine overall risk level
        if overall_risk_score >= 0.8:
            overall_risk_level = "critical"
        elif overall_risk_score >= 0.6:
            overall_risk_level = "high"
        elif overall_risk_score >= 0.4:
            overall_risk_level = "medium"
        else:
            overall_risk_level = "low"
        
        # Calculate overall confidence
        overall_confidence = (
            bert_results['risk_category_confidence'] * 0.4 +
            anomaly_results['anomaly_confidence'] * 0.3 +
            pattern_results['pattern_confidence'] * 0.3
        )
        
        return {
            'overall_risk_score': overall_risk_score,
            'overall_risk_level': overall_risk_level,
            'overall_confidence': overall_confidence,
            'bert_results': bert_results,
            'anomaly_results': anomaly_results,
            'pattern_results': pattern_results,
            'model_used': 'enhanced_risk_detection_v1.0'
        }

# Global risk detection model manager
risk_detection_manager = RiskDetectionModelManager()
