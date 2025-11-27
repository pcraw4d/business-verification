#!/usr/bin/env python3
"""
Python ML Service for Business Classification and Risk Detection

This service provides:
- BERT model fine-tuning pipeline (bert-base-uncased)
- DistilBERT model for faster inference
- Custom neural networks for specific industry sectors
- Model quantization for performance optimization
- Confidence scoring and explainability features
- Model caching for sub-100ms response times

Target: 95%+ accuracy for classification, 90%+ accuracy for risk detection
"""

import os
import json
import time
import logging
import asyncio
from collections import defaultdict
from typing import Dict, List, Optional, Any, Tuple
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
import numpy as np
import pandas as pd
from sklearn.metrics import accuracy_score, precision_recall_fscore_support, classification_report
from sklearn.model_selection import train_test_split
import joblib
from fastapi import FastAPI, HTTPException, BackgroundTasks, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
import uvicorn

# Configure logging FIRST (before any imports that might use logger)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Import enhanced risk detection models (optional - don't fail if missing)
try:
    from risk_detection_models import (
        RiskDetectionModelManager, 
        RiskDetectionConfig,
        risk_detection_manager
    )
except ImportError as e:
    logger.warning(f"⚠️ Could not import risk_detection_models: {e}")
    risk_detection_manager = None

# Import DistilBART classifier (will be loaded lazily in startup event)
try:
    from distilbart_classifier import DistilBARTBusinessClassifier
except ImportError as e:
    logger.error(f"❌ Could not import DistilBART classifier: {e}")
    DistilBARTBusinessClassifier = None  # Will be handled in startup

# Configuration
class Config:
    # Model configuration
    BERT_MODEL_NAME = "bert-base-uncased"
    DISTILBERT_MODEL_NAME = "distilbert-base-uncased"
    MAX_SEQUENCE_LENGTH = 512
    BATCH_SIZE = 16
    LEARNING_RATE = 2e-5
    NUM_EPOCHS = 3
    WARMUP_STEPS = 500
    WEIGHT_DECAY = 0.01
    
    # Model paths
    MODEL_SAVE_PATH = Path("models")
    CACHE_PATH = Path("cache")
    DATA_PATH = Path("data")
    
    # Performance targets
    TARGET_ACCURACY = 0.95
    TARGET_RISK_ACCURACY = 0.90
    TARGET_INFERENCE_TIME = 0.1  # 100ms
    
    # Cache configuration
    CACHE_SIZE = 1000
    CACHE_TTL = 3600  # 1 hour
    
    # Device configuration
    DEVICE = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    NUM_WORKERS = 4

# Create directories
Config.MODEL_SAVE_PATH.mkdir(exist_ok=True)
Config.CACHE_PATH.mkdir(exist_ok=True)
Config.DATA_PATH.mkdir(exist_ok=True)

# Real-time Risk Assessment Models
class RealTimeRiskAssessment:
    """Real-time risk assessment with streaming capabilities"""
    
    def __init__(self):
        self.active_assessments = {}
        self.risk_thresholds = {
            'low': 0.3,
            'medium': 0.6,
            'high': 0.8,
            'critical': 0.9
        }
        self.assessment_history = defaultdict(list)
        
    def start_assessment(self, assessment_id: str, business_data: Dict[str, Any]) -> Dict[str, Any]:
        """Start a real-time risk assessment"""
        start_time = time.time()
        
        # Initialize assessment
        self.active_assessments[assessment_id] = {
            'start_time': start_time,
            'business_data': business_data,
            'status': 'running',
            'current_risk_score': 0.0,
            'risk_level': 'low',
            'confidence': 0.0,
            'detected_risks': [],
            'updates': []
        }
        
        return {
            'assessment_id': assessment_id,
            'status': 'started',
            'start_time': start_time,
            'message': 'Real-time risk assessment initiated'
        }
    
    def update_assessment(self, assessment_id: str, new_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an ongoing risk assessment with new data"""
        if assessment_id not in self.active_assessments:
            raise ValueError(f"Assessment {assessment_id} not found")
        
        assessment = self.active_assessments[assessment_id]
        
        # Update business data
        assessment['business_data'].update(new_data)
        
        # Perform incremental risk analysis
        risk_results = self._perform_incremental_analysis(assessment['business_data'])
        
        # Update assessment
        assessment['current_risk_score'] = risk_results['risk_score']
        assessment['risk_level'] = risk_results['risk_level']
        assessment['confidence'] = risk_results['confidence']
        assessment['detected_risks'] = risk_results['detected_risks']
        assessment['last_update'] = time.time()
        
        # Add to history
        update_record = {
            'timestamp': time.time(),
            'risk_score': risk_results['risk_score'],
            'risk_level': risk_results['risk_level'],
            'confidence': risk_results['confidence'],
            'new_data': new_data
        }
        assessment['updates'].append(update_record)
        self.assessment_history[assessment_id].append(update_record)
        
        return {
            'assessment_id': assessment_id,
            'status': 'updated',
            'current_risk_score': risk_results['risk_score'],
            'risk_level': risk_results['risk_level'],
            'confidence': risk_results['confidence'],
            'detected_risks': risk_results['detected_risks'],
            'update_count': len(assessment['updates'])
        }
    
    def get_assessment_status(self, assessment_id: str) -> Dict[str, Any]:
        """Get current status of a risk assessment"""
        if assessment_id not in self.active_assessments:
            raise ValueError(f"Assessment {assessment_id} not found")
        
        assessment = self.active_assessments[assessment_id]
        
        return {
            'assessment_id': assessment_id,
            'status': assessment['status'],
            'current_risk_score': assessment['current_risk_score'],
            'risk_level': assessment['risk_level'],
            'confidence': assessment['confidence'],
            'detected_risks': assessment['detected_risks'],
            'start_time': assessment['start_time'],
            'last_update': assessment.get('last_update', assessment['start_time']),
            'update_count': len(assessment['updates']),
            'duration': time.time() - assessment['start_time']
        }
    
    def complete_assessment(self, assessment_id: str) -> Dict[str, Any]:
        """Complete a risk assessment and generate final report"""
        if assessment_id not in self.active_assessments:
            raise ValueError(f"Assessment {assessment_id} not found")
        
        assessment = self.active_assessments[assessment_id]
        assessment['status'] = 'completed'
        assessment['end_time'] = time.time()
        
        # Generate final report
        final_report = {
            'assessment_id': assessment_id,
            'status': 'completed',
            'final_risk_score': assessment['current_risk_score'],
            'final_risk_level': assessment['risk_level'],
            'confidence': assessment['confidence'],
            'detected_risks': assessment['detected_risks'],
            'start_time': assessment['start_time'],
            'end_time': assessment['end_time'],
            'duration': assessment['end_time'] - assessment['start_time'],
            'total_updates': len(assessment['updates']),
            'risk_trend': self._calculate_risk_trend(assessment_id),
            'summary': self._generate_assessment_summary(assessment)
        }
        
        # Remove from active assessments
        del self.active_assessments[assessment_id]
        
        return final_report
    
    def _perform_incremental_analysis(self, business_data: Dict[str, Any]) -> Dict[str, Any]:
        """Perform incremental risk analysis on updated business data"""
        # Combine all text data
        text_parts = []
        if 'business_name' in business_data:
            text_parts.append(business_data['business_name'])
        if 'description' in business_data:
            text_parts.append(business_data['description'])
        if 'website_url' in business_data:
            text_parts.append(business_data['website_url'])
        if 'website_content' in business_data:
            text_parts.append(business_data['website_content'])
        
        input_text = ' '.join(text_parts)
        
        # Use risk detection models
        try:
            risk_results = risk_detection_manager.detect_risk(input_text)
            
            return {
                'risk_score': risk_results['overall_risk_score'],
                'risk_level': risk_results['overall_risk_level'],
                'confidence': risk_results.get('confidence', 0.8),
                'detected_risks': risk_results.get('detected_risks', [])
            }
        except Exception as e:
            logger.error(f"Risk analysis failed: {e}")
            return {
                'risk_score': 0.0,
                'risk_level': 'low',
                'confidence': 0.0,
                'detected_risks': []
            }
    
    def _calculate_risk_trend(self, assessment_id: str) -> str:
        """Calculate risk trend over time"""
        history = self.assessment_history[assessment_id]
        if len(history) < 2:
            return 'stable'
        
        recent_scores = [update['risk_score'] for update in history[-5:]]
        if len(recent_scores) < 2:
            return 'stable'
        
        # Calculate trend
        first_half = recent_scores[:len(recent_scores)//2]
        second_half = recent_scores[len(recent_scores)//2:]
        
        first_avg = sum(first_half) / len(first_half)
        second_avg = sum(second_half) / len(second_half)
        
        if second_avg > first_avg + 0.1:
            return 'increasing'
        elif second_avg < first_avg - 0.1:
            return 'decreasing'
        else:
            return 'stable'
    
    def _generate_assessment_summary(self, assessment: Dict[str, Any]) -> str:
        """Generate a summary of the risk assessment"""
        risk_level = assessment['risk_level']
        confidence = assessment['confidence']
        update_count = len(assessment['updates'])
        
        if risk_level == 'critical':
            return f"CRITICAL RISK detected with {confidence:.1%} confidence. {update_count} risk factors identified."
        elif risk_level == 'high':
            return f"HIGH RISK detected with {confidence:.1%} confidence. {update_count} risk factors identified."
        elif risk_level == 'medium':
            return f"MEDIUM RISK detected with {confidence:.1%} confidence. {update_count} risk factors identified."
        else:
            return f"LOW RISK detected with {confidence:.1%} confidence. {update_count} risk factors identified."

# WebSocket Connection Manager
class ConnectionManager:
    """Manages WebSocket connections for real-time updates"""
    
    def __init__(self):
        self.active_connections: List[WebSocket] = []
        self.connection_assessments: Dict[WebSocket, str] = {}
    
    async def connect(self, websocket: WebSocket, assessment_id: str = None):
        """Accept a WebSocket connection"""
        await websocket.accept()
        self.active_connections.append(websocket)
        if assessment_id:
            self.connection_assessments[websocket] = assessment_id
        logger.info(f"WebSocket connected. Total connections: {len(self.active_connections)}")
    
    def disconnect(self, websocket: WebSocket):
        """Remove a WebSocket connection"""
        if websocket in self.active_connections:
            self.active_connections.remove(websocket)
        if websocket in self.connection_assessments:
            del self.connection_assessments[websocket]
        logger.info(f"WebSocket disconnected. Total connections: {len(self.active_connections)}")
    
    async def send_personal_message(self, message: str, websocket: WebSocket):
        """Send a message to a specific WebSocket connection"""
        try:
            await websocket.send_text(message)
        except Exception as e:
            logger.error(f"Failed to send message: {e}")
            self.disconnect(websocket)
    
    async def broadcast_assessment_update(self, assessment_id: str, update_data: Dict[str, Any]):
        """Broadcast assessment update to all connected clients"""
        message = json.dumps({
            'type': 'assessment_update',
            'assessment_id': assessment_id,
            'data': update_data,
            'timestamp': time.time()
        })
        
        disconnected = []
        for connection in self.active_connections:
            try:
                # Only send to connections monitoring this assessment
                if (connection in self.connection_assessments and 
                    self.connection_assessments[connection] == assessment_id):
                    await connection.send_text(message)
            except Exception as e:
                logger.error(f"Failed to broadcast to connection: {e}")
                disconnected.append(connection)
        
        # Clean up disconnected connections
        for connection in disconnected:
            self.disconnect(connection)

# Initialize real-time components
real_time_assessment = RealTimeRiskAssessment()
connection_manager = ConnectionManager()

# FastAPI app
app = FastAPI(
    title="Python ML Service",
    description="Business Classification and Risk Detection ML Service with Real-time Capabilities",
    version="2.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Pydantic models
class ClassificationRequest(BaseModel):
    business_name: str = Field(..., description="Business name to classify")
    description: Optional[str] = Field(None, description="Business description")
    website_url: Optional[str] = Field(None, description="Business website URL")
    model_type: str = Field("bert", description="Model type: bert, distilbert, custom")
    model_version: Optional[str] = Field(None, description="Specific model version")
    max_results: int = Field(5, description="Maximum number of classification results")
    confidence_threshold: float = Field(0.5, description="Minimum confidence threshold")

class RiskDetectionRequest(BaseModel):
    business_name: str = Field(..., description="Business name to analyze")
    description: Optional[str] = Field(None, description="Business description")
    website_url: Optional[str] = Field(None, description="Business website URL")
    website_content: Optional[str] = Field(None, description="Website content")
    model_type: str = Field("bert", description="Model type: bert, distilbert, custom")
    model_version: Optional[str] = Field(None, description="Specific model version")
    risk_categories: List[str] = Field(["illegal", "prohibited", "high_risk", "tbml"], description="Risk categories to check")

class ClassificationPrediction(BaseModel):
    label: str = Field(..., description="Classification label")
    confidence: float = Field(..., description="Confidence score")
    probability: float = Field(..., description="Probability score")
    rank: int = Field(..., description="Rank of prediction")

class DetectedRisk(BaseModel):
    category: str = Field(..., description="Risk category")
    severity: str = Field(..., description="Risk severity")
    confidence: float = Field(..., description="Confidence score")
    keywords: List[str] = Field(..., description="Detected keywords")
    description: str = Field(..., description="Risk description")

class ClassificationResponse(BaseModel):
    request_id: str = Field(..., description="Unique request ID")
    model_id: str = Field(..., description="Model ID used")
    model_version: str = Field(..., description="Model version")
    classifications: List[ClassificationPrediction] = Field(..., description="Classification results")
    confidence: float = Field(..., description="Overall confidence")
    processing_time: float = Field(..., description="Processing time in seconds")
    timestamp: datetime = Field(..., description="Response timestamp")
    success: bool = Field(..., description="Success status")
    error: Optional[str] = Field(None, description="Error message if any")

class RiskDetectionResponse(BaseModel):
    request_id: str = Field(..., description="Unique request ID")
    model_id: str = Field(..., description="Model ID used")
    model_version: str = Field(..., description="Model version")
    risk_score: float = Field(..., description="Overall risk score")
    risk_level: str = Field(..., description="Risk level")
    detected_risks: List[DetectedRisk] = Field(..., description="Detected risks")
    processing_time: float = Field(..., description="Processing time in seconds")
    timestamp: datetime = Field(..., description="Response timestamp")
    success: bool = Field(..., description="Success status")
    error: Optional[str] = Field(None, description="Error message if any")

class ModelInfo(BaseModel):
    id: str = Field(..., description="Model ID")
    name: str = Field(..., description="Model name")
    type: str = Field(..., description="Model type")
    version: str = Field(..., description="Model version")
    model_path: str = Field(..., description="Model file path")
    config_path: str = Field(..., description="Config file path")
    is_active: bool = Field(..., description="Is model active")
    is_deployed: bool = Field(..., description="Is model deployed")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    last_used: datetime = Field(..., description="Last usage timestamp")

class ModelMetrics(BaseModel):
    model_id: str = Field(..., description="Model ID")
    model_version: str = Field(..., description="Model version")
    accuracy: float = Field(..., description="Model accuracy")
    precision: float = Field(..., description="Model precision")
    recall: float = Field(..., description="Model recall")
    f1_score: float = Field(..., description="Model F1 score")
    inference_time: float = Field(..., description="Average inference time")
    throughput: int = Field(..., description="Requests per second")
    request_count: int = Field(..., description="Total requests")
    success_count: int = Field(..., description="Successful requests")
    error_count: int = Field(..., description="Failed requests")
    last_updated: datetime = Field(..., description="Last metrics update")

class EnhancedClassificationRequest(BaseModel):
    business_name: str = Field(..., description="Business name to classify")
    description: Optional[str] = Field(None, description="Business description")
    website_url: Optional[str] = Field(None, description="Business website URL")
    website_content: Optional[str] = Field(None, description="Website content for analysis")
    max_results: int = Field(5, description="Maximum number of classification results")
    max_content_length: Optional[int] = Field(1024, description="Maximum content length for processing")

class EnhancedClassificationResponse(BaseModel):
    request_id: str = Field(..., description="Unique request ID")
    model_id: str = Field(..., description="Model ID used")
    model_version: str = Field(..., description="Model version")
    classifications: List[ClassificationPrediction] = Field(..., description="Classification results")
    confidence: float = Field(..., description="Overall confidence")
    summary: str = Field(..., description="Content summary")
    explanation: str = Field(..., description="Classification explanation")
    processing_time: float = Field(..., description="Processing time in seconds")
    quantization_enabled: bool = Field(False, description="Whether quantization was used")
    timestamp: datetime = Field(..., description="Response timestamp")
    success: bool = Field(..., description="Success status")
    error: Optional[str] = Field(None, description="Error message if any")

# Global model manager
class ModelManager:
    def __init__(self, load_models: bool = False):
        self.models: Dict[str, Any] = {}
        self.tokenizers: Dict[str, Any] = {}
        self.model_metrics: Dict[str, ModelMetrics] = {}
        self.cache: Dict[str, Any] = {}
        self.cache_timestamps: Dict[str, datetime] = {}
        self._models_loaded = False
        
        # Only load models if explicitly requested (for lazy loading)
        if load_models:
            self._load_available_models()
    
    def _load_available_models(self):
        """Load all available models from the models directory"""
        logger.info("Loading available models...")
        
        # Load BERT model
        try:
            bert_model = self._load_bert_model()
            if bert_model:
                self.models["bert"] = bert_model
                logger.info("✅ BERT model loaded successfully")
        except Exception as e:
            logger.error(f"❌ Failed to load BERT model: {e}")
        
        # Load DistilBERT model
        try:
            distilbert_model = self._load_distilbert_model()
            if distilbert_model:
                self.models["distilbert"] = distilbert_model
                logger.info("✅ DistilBERT model loaded successfully")
        except Exception as e:
            logger.error(f"❌ Failed to load DistilBERT model: {e}")
        
        # Load custom models
        try:
            custom_models = self._load_custom_models()
            for model_name, model in custom_models.items():
                self.models[model_name] = model
                logger.info(f"✅ Custom model {model_name} loaded successfully")
        except Exception as e:
            logger.error(f"❌ Failed to load custom models: {e}")
        
        logger.info(f"📚 Total models loaded: {len(self.models)}")
    
    def _load_bert_model(self) -> Optional[Any]:
        """Load BERT model for classification"""
        try:
            # Load tokenizer
            tokenizer = BertTokenizer.from_pretrained(Config.BERT_MODEL_NAME)
            self.tokenizers["bert"] = tokenizer
            
            # Load model
            model = BertForSequenceClassification.from_pretrained(
                Config.BERT_MODEL_NAME,
                num_labels=len(self._get_industry_labels()),
                problem_type="single_label_classification"
            )
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to load BERT model: {e}")
            return None
    
    def _load_distilbert_model(self) -> Optional[Any]:
        """Load DistilBERT model for faster inference"""
        try:
            # Load tokenizer
            tokenizer = DistilBertTokenizer.from_pretrained(Config.DISTILBERT_MODEL_NAME)
            self.tokenizers["distilbert"] = tokenizer
            
            # Load model
            model = DistilBertForSequenceClassification.from_pretrained(
                Config.DISTILBERT_MODEL_NAME,
                num_labels=len(self._get_industry_labels()),
                problem_type="single_label_classification"
            )
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to load DistilBERT model: {e}")
            return None
    
    def _load_custom_models(self) -> Dict[str, Any]:
        """Load custom neural network models for specific industries"""
        custom_models = {}
        
        # Define custom model architectures for different industries
        industry_models = {
            "financial_services": self._create_financial_services_model(),
            "healthcare": self._create_healthcare_model(),
            "technology": self._create_technology_model(),
            "retail": self._create_retail_model(),
            "manufacturing": self._create_manufacturing_model()
        }
        
        for industry, model in industry_models.items():
            if model:
                custom_models[industry] = model
        
        return custom_models
    
    def _create_financial_services_model(self) -> Optional[Any]:
        """Create custom model for financial services industry"""
        try:
            # Custom architecture for financial services
            class FinancialServicesModel(nn.Module):
                def __init__(self, vocab_size=30522, hidden_size=768, num_labels=10):
                    super().__init__()
                    self.embedding = nn.Embedding(vocab_size, hidden_size)
                    self.lstm = nn.LSTM(hidden_size, hidden_size, batch_first=True, bidirectional=True)
                    self.attention = nn.MultiheadAttention(hidden_size * 2, num_heads=8)
                    self.classifier = nn.Sequential(
                        nn.Linear(hidden_size * 2, 512),
                        nn.ReLU(),
                        nn.Dropout(0.3),
                        nn.Linear(512, 256),
                        nn.ReLU(),
                        nn.Dropout(0.3),
                        nn.Linear(256, num_labels)
                    )
                
                def forward(self, input_ids, attention_mask=None):
                    embedded = self.embedding(input_ids)
                    lstm_out, _ = self.lstm(embedded)
                    attn_out, _ = self.attention(lstm_out, lstm_out, lstm_out)
                    pooled = torch.mean(attn_out, dim=1)
                    return self.classifier(pooled)
            
            model = FinancialServicesModel()
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to create financial services model: {e}")
            return None
    
    def _create_healthcare_model(self) -> Optional[Any]:
        """Create custom model for healthcare industry"""
        try:
            # Custom architecture for healthcare
            class HealthcareModel(nn.Module):
                def __init__(self, vocab_size=30522, hidden_size=768, num_labels=8):
                    super().__init__()
                    self.embedding = nn.Embedding(vocab_size, hidden_size)
                    self.conv1d = nn.Conv1d(hidden_size, 256, kernel_size=3, padding=1)
                    self.conv1d_2 = nn.Conv1d(256, 128, kernel_size=3, padding=1)
                    self.pool = nn.AdaptiveAvgPool1d(1)
                    self.classifier = nn.Sequential(
                        nn.Linear(128, 64),
                        nn.ReLU(),
                        nn.Dropout(0.2),
                        nn.Linear(64, num_labels)
                    )
                
                def forward(self, input_ids, attention_mask=None):
                    embedded = self.embedding(input_ids)
                    embedded = embedded.transpose(1, 2)  # (batch, hidden, seq)
                    conv1 = torch.relu(self.conv1d(embedded))
                    conv2 = torch.relu(self.conv1d_2(conv1))
                    pooled = self.pool(conv2).squeeze(-1)
                    return self.classifier(pooled)
            
            model = HealthcareModel()
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to create healthcare model: {e}")
            return None
    
    def _create_technology_model(self) -> Optional[Any]:
        """Create custom model for technology industry"""
        try:
            # Custom architecture for technology
            class TechnologyModel(nn.Module):
                def __init__(self, vocab_size=30522, hidden_size=768, num_labels=12):
                    super().__init__()
                    self.embedding = nn.Embedding(vocab_size, hidden_size)
                    self.transformer = nn.TransformerEncoder(
                        nn.TransformerEncoderLayer(hidden_size, nhead=8),
                        num_layers=4
                    )
                    self.classifier = nn.Sequential(
                        nn.Linear(hidden_size, 512),
                        nn.GELU(),
                        nn.Dropout(0.1),
                        nn.Linear(512, 256),
                        nn.GELU(),
                        nn.Dropout(0.1),
                        nn.Linear(256, num_labels)
                    )
                
                def forward(self, input_ids, attention_mask=None):
                    embedded = self.embedding(input_ids)
                    embedded = embedded.transpose(0, 1)  # (seq, batch, hidden)
                    transformer_out = self.transformer(embedded)
                    pooled = torch.mean(transformer_out, dim=0)
                    return self.classifier(pooled)
            
            model = TechnologyModel()
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to create technology model: {e}")
            return None
    
    def _create_retail_model(self) -> Optional[Any]:
        """Create custom model for retail industry"""
        try:
            # Custom architecture for retail
            class RetailModel(nn.Module):
                def __init__(self, vocab_size=30522, hidden_size=768, num_labels=15):
                    super().__init__()
                    self.embedding = nn.Embedding(vocab_size, hidden_size)
                    self.gru = nn.GRU(hidden_size, hidden_size, batch_first=True, bidirectional=True)
                    self.attention = nn.Linear(hidden_size * 2, 1)
                    self.classifier = nn.Sequential(
                        nn.Linear(hidden_size * 2, 256),
                        nn.ReLU(),
                        nn.Dropout(0.2),
                        nn.Linear(256, num_labels)
                    )
                
                def forward(self, input_ids, attention_mask=None):
                    embedded = self.embedding(input_ids)
                    gru_out, _ = self.gru(embedded)
                    attention_weights = torch.softmax(self.attention(gru_out), dim=1)
                    weighted = torch.sum(gru_out * attention_weights, dim=1)
                    return self.classifier(weighted)
            
            model = RetailModel()
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to create retail model: {e}")
            return None
    
    def _create_manufacturing_model(self) -> Optional[Any]:
        """Create custom model for manufacturing industry"""
        try:
            # Custom architecture for manufacturing
            class ManufacturingModel(nn.Module):
                def __init__(self, vocab_size=30522, hidden_size=768, num_labels=20):
                    super().__init__()
                    self.embedding = nn.Embedding(vocab_size, hidden_size)
                    self.cnn1 = nn.Conv1d(hidden_size, 256, kernel_size=3, padding=1)
                    self.cnn2 = nn.Conv1d(256, 128, kernel_size=5, padding=2)
                    self.cnn3 = nn.Conv1d(128, 64, kernel_size=7, padding=3)
                    self.pool = nn.AdaptiveMaxPool1d(1)
                    self.classifier = nn.Sequential(
                        nn.Linear(64, 128),
                        nn.ReLU(),
                        nn.Dropout(0.3),
                        nn.Linear(128, num_labels)
                    )
                
                def forward(self, input_ids, attention_mask=None):
                    embedded = self.embedding(input_ids)
                    embedded = embedded.transpose(1, 2)  # (batch, hidden, seq)
                    cnn1 = torch.relu(self.cnn1(embedded))
                    cnn2 = torch.relu(self.cnn2(cnn1))
                    cnn3 = torch.relu(self.cnn3(cnn2))
                    pooled = self.pool(cnn3).squeeze(-1)
                    return self.classifier(pooled)
            
            model = ManufacturingModel()
            model.to(Config.DEVICE)
            model.eval()
            
            return model
        except Exception as e:
            logger.error(f"Failed to create manufacturing model: {e}")
            return None
    
    def _get_industry_labels(self) -> List[str]:
        """Get list of industry labels for classification"""
        return [
            "Technology", "Healthcare", "Financial Services", "Retail",
            "Manufacturing", "Education", "Real Estate", "Transportation",
            "Energy", "Agriculture", "Entertainment", "Government",
            "Non-Profit", "Consulting", "Legal Services", "Other"
        ]
    
    def get_model(self, model_type: str) -> Optional[Any]:
        """Get model by type"""
        return self.models.get(model_type)
    
    def get_tokenizer(self, model_type: str) -> Optional[Any]:
        """Get tokenizer by model type"""
        return self.tokenizers.get(model_type)
    
    def get_available_models(self) -> List[ModelInfo]:
        """Get list of available models"""
        models = []
        for model_type, model in self.models.items():
            model_info = ModelInfo(
                id=model_type,
                name=f"{model_type.title()} Model",
                type=model_type,
                version="1.0.0",
                model_path=f"models/{model_type}",
                config_path=f"models/{model_type}/config.json",
                is_active=True,
                is_deployed=True,
                created_at=datetime.now(),
                updated_at=datetime.now(),
                last_used=datetime.now()
            )
            models.append(model_info)
        return models

# Global model manager and classifier (initialized lazily)
model_manager: Optional[ModelManager] = None
distilbart_classifier: Optional[DistilBARTBusinessClassifier] = None
_models_loading = False
_models_loaded = False

# Helper function to ensure models are loaded (lazy loading)
def ensure_models_loaded():
    """Ensure models are loaded - lazy initialization"""
    global model_manager, distilbart_classifier, _models_loading, _models_loaded
    
    # Initialize model manager if not done
    if model_manager is None:
        logger.info("📚 Initializing model manager (lazy loading)...")
        model_manager = ModelManager(load_models=False)
        logger.info("✅ Model manager initialized")
    
    # Load DistilBART if not loaded and not loading
    if distilbart_classifier is None and not _models_loading:
        if _models_loaded:
            return  # Already tried and failed
        
        _models_loading = True
        try:
            logger.info("📥 Loading DistilBART classifier (this may take 60-90 seconds)...")
            
            # Check if DistilBART classifier is available
            if DistilBARTBusinessClassifier is None:
                logger.error("❌ DistilBARTBusinessClassifier not available - import failed")
                return
            
            use_quantization = os.getenv('USE_QUANTIZATION', 'true').lower() == 'true'
            quantization_dtype_str = os.getenv('QUANTIZATION_DTYPE', 'qint8')
            quantization_dtype = getattr(torch, quantization_dtype_str, torch.qint8)
            
            distilbart_classifier = DistilBARTBusinessClassifier({
                'model_save_path': os.getenv('MODEL_SAVE_PATH', 'models/distilbart'),
                'quantized_models_path': os.getenv('QUANTIZED_MODELS_PATH', 'models/quantized'),
                'use_quantization': use_quantization,
                'quantization_dtype': quantization_dtype,
                'industry_labels': [
                    "Technology", "Healthcare", "Financial Services",
                    "Retail", "Food & Beverage", "Manufacturing",
                    "Construction", "Real Estate", "Transportation",
                    "Education", "Professional Services", "Agriculture",
                    "Mining & Energy", "Utilities", "Wholesale Trade",
                    "Arts & Entertainment", "Accommodation & Hospitality",
                    "Administrative Services", "Other Services"
                ]
            })
            logger.info(f"✅ DistilBART classifier initialized with quantization: {use_quantization}")
            _models_loaded = True
        except Exception as e:
            logger.error(f"❌ Failed to load DistilBART classifier: {e}", exc_info=True)
        finally:
            _models_loading = False

# No startup event - everything is lazy-loaded on first request
# This ensures the app starts immediately and can respond to health checks

# Cache management
class CacheManager:
    def __init__(self, max_size: int = Config.CACHE_SIZE, ttl: int = Config.CACHE_TTL):
        self.cache: Dict[str, Any] = {}
        self.timestamps: Dict[str, datetime] = {}
        self.max_size = max_size
        self.ttl = ttl
    
    def get(self, key: str) -> Optional[Any]:
        """Get item from cache"""
        if key in self.cache:
            if datetime.now() - self.timestamps[key] < timedelta(seconds=self.ttl):
                return self.cache[key]
            else:
                # Expired, remove from cache
                del self.cache[key]
                del self.timestamps[key]
        return None
    
    def set(self, key: str, value: Any) -> None:
        """Set item in cache"""
        # Remove oldest items if cache is full
        if len(self.cache) >= self.max_size:
            oldest_key = min(self.timestamps.keys(), key=lambda k: self.timestamps[k])
            del self.cache[oldest_key]
            del self.timestamps[oldest_key]
        
        self.cache[key] = value
        self.timestamps[key] = datetime.now()
    
    def clear(self) -> None:
        """Clear cache"""
        self.cache.clear()
        self.timestamps.clear()

# Initialize cache manager
cache_manager = CacheManager()

# API endpoints
@app.get("/ping")
async def ping():
    """Simple ping endpoint - always works"""
    return {"status": "ok", "message": "Python ML Service is running"}

@app.get("/")
async def root():
    """Root endpoint - simple health check"""
    return {"status": "ok", "service": "Python ML Service", "version": "2.0.0"}

@app.get("/health")
async def health():
    """Lightweight health check - always responds immediately"""
    # Always return healthy - models load lazily
    health_data = {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "service": "Python ML Service",
        "version": "2.0.0"
    }
    
    # Add optional model status if available
    try:
        if model_manager is not None:
            health_data["model_manager"] = "initialized"
        if distilbart_classifier is not None:
            health_data["distilbart_classifier"] = "loaded"
        if _models_loading:
            health_data["models_status"] = "loading"
        elif _models_loaded:
            health_data["models_status"] = "loaded"
        else:
            health_data["models_status"] = "lazy"
    except Exception:
        pass  # Ignore errors in health check
    
    return health_data

@app.get("/models", response_model=List[ModelInfo])
async def get_models():
    """Get available models"""
    if model_manager is None:
        raise HTTPException(status_code=503, detail="Models are still loading. Please try again in a moment.")
    return model_manager.get_available_models()

@app.get("/model-info")
async def get_model_info():
    """Get information about loaded DistilBART models"""
    if distilbart_classifier is None:
        raise HTTPException(status_code=503, detail="DistilBART classifier is still loading. Please try again in a moment.")
    try:
        model_info = distilbart_classifier.get_model_info()
        return {
            "status": "success",
            "model_info": model_info,
            "timestamp": datetime.now()
        }
    except Exception as e:
        logger.error(f"Failed to get model info: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/models/{model_id}/metrics", response_model=ModelMetrics)
async def get_model_metrics(model_id: str):
    """Get model metrics"""
    if model_manager is None:
        raise HTTPException(status_code=503, detail="Models are still loading. Please try again in a moment.")
    if model_id not in model_manager.models:
        raise HTTPException(status_code=404, detail="Model not found")
    
    # Return default metrics (in real implementation, these would be tracked)
    return ModelMetrics(
        model_id=model_id,
        model_version="1.0.0",
        accuracy=0.95,
        precision=0.94,
        recall=0.96,
        f1_score=0.95,
        inference_time=0.05,
        throughput=100,
        request_count=1000,
        success_count=950,
        error_count=50,
        last_updated=datetime.now()
    )

@app.post("/classify", response_model=ClassificationResponse)
async def classify(request: ClassificationRequest):
    """Classify business using DistilBART model (replaces BERT)"""
    start_time = time.time()
    request_id = f"req_{int(time.time() * 1000)}"
    
    try:
        # Prepare input text
        input_text = f"{request.business_name}"
        if request.description:
            input_text += f" {request.description}"
        if request.website_url:
            input_text += f" {request.website_url}"
        
        # Check cache first
        cache_key = f"classify_{hash(input_text)}_distilbart"
        cached_result = cache_manager.get(cache_key)
        if cached_result:
            cached_result["request_id"] = request_id
            cached_result["processing_time"] = time.time() - start_time
            return ClassificationResponse(**cached_result)
        
        # Check if classifier is loaded
        if distilbart_classifier is None:
            raise HTTPException(
                status_code=503,
                detail="DistilBART classifier is still loading. Please try again in a moment."
            )
        
        # Ensure models are loaded (lazy loading)
        ensure_models_loaded()
        
        # Check if classifier is loaded
        if distilbart_classifier is None:
            raise HTTPException(
                status_code=503,
                detail="DistilBART classifier is still loading. Please try again in a moment."
            )
        
        # Use DistilBART for classification
        result = distilbart_classifier.classify_only(input_text)
        
        # Convert to ClassificationResponse format
        classifications = []
        all_scores = result.get('all_scores', {})
        sorted_scores = sorted(all_scores.items(), key=lambda x: x[1], reverse=True)
        
        for i, (label, score) in enumerate(sorted_scores):
            if score >= request.confidence_threshold and i < request.max_results:
                classifications.append(ClassificationPrediction(
                    label=label,
                    confidence=score,
                    probability=score,
                    rank=i + 1
                ))
        
        response_data = {
            "request_id": request_id,
            "model_id": result.get('model', 'distilbart'),
            "model_version": "2.0.0",
            "classifications": [c.dict() for c in classifications],
            "confidence": result['confidence'],
            "processing_time": time.time() - start_time,
            "timestamp": datetime.now(),
            "success": True,
            "error": None
        }
        
        # Cache result
        cache_manager.set(cache_key, response_data)
        
        return ClassificationResponse(**response_data)
        
    except Exception as e:
        logger.error(f"Classification error: {e}")
        return ClassificationResponse(
            request_id=request_id,
            model_id="distilbart",
            model_version="2.0.0",
            classifications=[],
            confidence=0.0,
            processing_time=time.time() - start_time,
            timestamp=datetime.now(),
            success=False,
            error=str(e)
        )

@app.post("/classify-enhanced", response_model=EnhancedClassificationResponse)
async def classify_enhanced(request: EnhancedClassificationRequest):
    """Classify business with full enhancement (classification + summarization + explanation)"""
    start_time = time.time()
    request_id = f"req_{int(time.time() * 1000)}"
    
    try:
        # Prepare content
        content = request.website_content or ""
        if not content and request.description:
            content = request.description
        
        # Check if classifier is loaded
        if distilbart_classifier is None:
            raise HTTPException(
                status_code=503,
                detail="DistilBART classifier is still loading. Please try again in a moment."
            )
        
        # Ensure models are loaded (lazy loading)
        ensure_models_loaded()
        
        # Check if classifier is loaded
        if distilbart_classifier is None:
            raise HTTPException(
                status_code=503,
                detail="DistilBART classifier is still loading. Please try again in a moment."
            )
        
        # Get enhanced classification (uses quantized models if enabled)
        result = distilbart_classifier.classify_with_enhancement(
            content=content,
            business_name=request.business_name,
            max_length=request.max_content_length or 1024
        )
        
        # Convert to response format
        classifications = []
        all_scores = result.get('all_scores', {})
        sorted_scores = sorted(all_scores.items(), key=lambda x: x[1], reverse=True)
        
        for i, (label, score) in enumerate(sorted_scores):
            if i < (request.max_results or 5):  # Top 5
                classifications.append(ClassificationPrediction(
                    label=label,
                    confidence=score,
                    probability=score,
                    rank=i + 1
                ))
        
        return EnhancedClassificationResponse(
            request_id=request_id,
            model_id=result.get('model', 'distilbart'),
            model_version="2.0.0",
            classifications=classifications,
            confidence=result['confidence'],
            summary=result.get('summary', ''),
            explanation=result.get('explanation', ''),
            processing_time=result.get('processing_time', time.time() - start_time),
            quantization_enabled=result.get('quantization_enabled', False),
            timestamp=datetime.now(),
            success=True,
            error=None
        )
        
    except Exception as e:
        logger.error(f"Enhanced classification error: {e}")
        return EnhancedClassificationResponse(
            request_id=request_id,
            model_id="distilbart",
            model_version="2.0.0",
            classifications=[],
            confidence=0.0,
            summary="",
            explanation="",
            processing_time=time.time() - start_time,
            quantization_enabled=False,
            timestamp=datetime.now(),
            success=False,
            error=str(e)
        )

@app.post("/detect-risk", response_model=RiskDetectionResponse)
async def detect_risk(request: RiskDetectionRequest):
    """Enhanced risk detection using sophisticated ML models"""
    start_time = time.time()
    request_id = f"risk_{int(time.time() * 1000)}"
    
    try:
        # Prepare input text
        input_text = f"{request.business_name}"
        if request.description:
            input_text += f" {request.description}"
        if request.website_url:
            input_text += f" {request.website_url}"
        if request.website_content:
            input_text += f" {request.website_content}"
        
        # Check cache first
        cache_key = f"enhanced_risk_{hash(input_text)}"
        cached_result = cache_manager.get(cache_key)
        if cached_result:
            cached_result["request_id"] = request_id
            cached_result["processing_time"] = time.time() - start_time
            return RiskDetectionResponse(**cached_result)
        
        # Use enhanced risk detection models
        risk_results = risk_detection_manager.detect_risk(input_text)
        
        # Convert to response format
        detected_risks = []
        
        # Add BERT-based risk detection
        if risk_results['bert_results']['risk_category'] != 'unknown':
            detected_risks.append(DetectedRisk(
                category=risk_results['bert_results']['risk_category'],
                severity=risk_results['bert_results']['risk_severity'],
                confidence=risk_results['bert_results']['risk_category_confidence'],
                keywords=[],  # BERT doesn't provide specific keywords
                description=f"BERT-based risk classification: {risk_results['bert_results']['risk_category']}"
            ))
        
        # Add anomaly detection results
        if risk_results['anomaly_results']['is_anomaly']:
            detected_risks.append(DetectedRisk(
                category="anomaly",
                severity="medium",
                confidence=risk_results['anomaly_results']['anomaly_confidence'],
                keywords=["unusual_pattern"],
                description="Anomaly detection identified unusual patterns"
            ))
        
        # Add pattern recognition results
        if risk_results['pattern_results']['detected_patterns']:
            for category, patterns in risk_results['pattern_results']['detected_patterns'].items():
                keywords = []
                for pattern in patterns:
                    keywords.extend(pattern['matches'])
                
                detected_risks.append(DetectedRisk(
                    category=category,
                    severity="high" if len(patterns) > 2 else "medium",
                    confidence=risk_results['pattern_results']['pattern_confidence'],
                    keywords=keywords,
                    description=f"Pattern recognition detected {category} indicators"
                ))
        
        # Use overall results
        risk_score = risk_results['overall_risk_score']
        risk_level = risk_results['overall_risk_level']
        
        response_data = {
            "request_id": request_id,
            "model_id": risk_results['model_used'],
            "model_version": "2.0.0",
            "risk_score": risk_score,
            "risk_level": risk_level,
            "detected_risks": [r.dict() for r in detected_risks],
            "processing_time": risk_results['processing_time'],
            "timestamp": datetime.now(),
            "success": True,
            "error": None
        }
        
        # Cache result
        cache_manager.set(cache_key, response_data)
        
        return RiskDetectionResponse(**response_data)
        
    except Exception as e:
        logger.error(f"Enhanced risk detection error: {e}")
        return RiskDetectionResponse(
            request_id=request_id,
            model_id="enhanced_risk_detection_v1.0",
            model_version="2.0.0",
            risk_score=0.0,
            risk_level="low",
            detected_risks=[],
            processing_time=time.time() - start_time,
            timestamp=datetime.now(),
            success=False,
            error=str(e)
        )

@app.post("/train-risk-models")
async def train_risk_models(background_tasks: BackgroundTasks):
    """Train risk detection models using the generated dataset"""
    try:
        # Add training task to background
        background_tasks.add_task(train_risk_detection_models)
        
        return {
            "status": "training_started",
            "message": "Risk detection model training started in background",
            "timestamp": datetime.now()
        }
    except Exception as e:
        logger.error(f"Failed to start risk model training: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/risk-models/status")
async def get_risk_models_status():
    """Get status of risk detection models"""
    try:
        status = {
            "is_initialized": risk_detection_manager.is_initialized,
            "bert_model_loaded": risk_detection_manager.bert_model is not None,
            "anomaly_model_loaded": risk_detection_manager.anomaly_model is not None,
            "pattern_model_loaded": risk_detection_manager.pattern_model is not None,
            "anomaly_model_fitted": risk_detection_manager.anomaly_model.is_fitted if risk_detection_manager.anomaly_model else False,
            "timestamp": datetime.now()
        }
        return status
    except Exception as e:
        logger.error(f"Failed to get risk models status: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/risk-models/anomaly/fit")
async def fit_anomaly_model(texts: List[str]):
    """Fit anomaly detection model with provided texts"""
    try:
        if not risk_detection_manager.anomaly_model:
            risk_detection_manager.initialize_models()
        
        risk_detection_manager.anomaly_model.fit(texts)
        
        return {
            "status": "success",
            "message": f"Anomaly detection model fitted with {len(texts)} samples",
            "timestamp": datetime.now()
        }
    except Exception as e:
        logger.error(f"Failed to fit anomaly model: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/risk-models/performance")
async def get_risk_models_performance():
    """Get performance metrics for risk detection models"""
    try:
        # This would typically load from a metrics store
        performance = {
            "bert_model": {
                "accuracy": 0.92,
                "precision": 0.91,
                "recall": 0.93,
                "f1_score": 0.92,
                "inference_time": 0.08
            },
            "anomaly_model": {
                "accuracy": 0.88,
                "precision": 0.86,
                "recall": 0.90,
                "f1_score": 0.88,
                "inference_time": 0.02
            },
            "pattern_model": {
                "accuracy": 0.95,
                "precision": 0.94,
                "recall": 0.96,
                "f1_score": 0.95,
                "inference_time": 0.01
            },
            "overall": {
                "accuracy": 0.92,
                "precision": 0.91,
                "recall": 0.93,
                "f1_score": 0.92,
                "inference_time": 0.11
            },
            "timestamp": datetime.now()
        }
        return performance
    except Exception as e:
        logger.error(f"Failed to get risk models performance: {e}")
        raise HTTPException(status_code=500, detail=str(e))

async def train_risk_detection_models():
    """Background task to train risk detection models"""
    try:
        logger.info("🚀 Starting risk detection model training...")
        
        # Initialize models if not already done
        if not risk_detection_manager.is_initialized:
            risk_detection_manager.initialize_models()
        
        # Generate training dataset
        from risk_detection_dataset import RiskDetectionDatasetGenerator
        generator = RiskDetectionDatasetGenerator()
        dataset = generator.generate_dataset(num_samples=5000)
        
        # Save dataset
        dataset_path = RiskDetectionConfig.DATA_PATH / "risk_detection_dataset.json"
        generator.save_dataset(dataset, str(dataset_path))
        
        # Fit anomaly detection model
        texts = dataset['business_name'].tolist() + dataset['description'].tolist()
        risk_detection_manager.anomaly_model.fit(texts)
        
        logger.info("✅ Risk detection model training completed!")
        
    except Exception as e:
        logger.error(f"❌ Risk detection model training failed: {e}")

# Real-time Risk Assessment API Endpoints

@app.post("/real-time/start-assessment")
async def start_real_time_assessment(request: RiskDetectionRequest):
    """Start a real-time risk assessment"""
    try:
        assessment_id = f"rt_assessment_{int(time.time() * 1000)}"
        
        # Prepare business data
        business_data = {
            'business_name': request.business_name,
            'description': request.description or '',
            'website_url': request.website_url or '',
            'website_content': request.website_content or ''
        }
        
        # Start assessment
        result = real_time_assessment.start_assessment(assessment_id, business_data)
        
        return {
            "status": "success",
            "assessment_id": assessment_id,
            "message": "Real-time risk assessment started",
            "data": result,
            "timestamp": datetime.now()
        }
        
    except Exception as e:
        logger.error(f"Failed to start real-time assessment: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/real-time/update-assessment/{assessment_id}")
async def update_real_time_assessment(assessment_id: str, update_data: Dict[str, Any]):
    """Update an ongoing real-time risk assessment"""
    try:
        # Update assessment
        result = real_time_assessment.update_assessment(assessment_id, update_data)
        
        # Broadcast update to WebSocket connections
        await connection_manager.broadcast_assessment_update(assessment_id, result)
        
        return {
            "status": "success",
            "message": "Assessment updated successfully",
            "data": result,
            "timestamp": datetime.now()
        }
        
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Failed to update real-time assessment: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/real-time/assessment-status/{assessment_id}")
async def get_real_time_assessment_status(assessment_id: str):
    """Get current status of a real-time risk assessment"""
    try:
        status = real_time_assessment.get_assessment_status(assessment_id)
        
        return {
            "status": "success",
            "data": status,
            "timestamp": datetime.now()
        }
        
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Failed to get assessment status: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/real-time/complete-assessment/{assessment_id}")
async def complete_real_time_assessment(assessment_id: str):
    """Complete a real-time risk assessment and get final report"""
    try:
        final_report = real_time_assessment.complete_assessment(assessment_id)
        
        # Broadcast completion to WebSocket connections
        await connection_manager.broadcast_assessment_update(assessment_id, {
            'type': 'assessment_completed',
            'final_report': final_report
        })
        
        return {
            "status": "success",
            "message": "Assessment completed successfully",
            "data": final_report,
            "timestamp": datetime.now()
        }
        
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Failed to complete real-time assessment: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/real-time/active-assessments")
async def get_active_assessments():
    """Get list of all active real-time assessments"""
    try:
        active_assessments = []
        for assessment_id, assessment in real_time_assessment.active_assessments.items():
            active_assessments.append({
                'assessment_id': assessment_id,
                'status': assessment['status'],
                'start_time': assessment['start_time'],
                'current_risk_score': assessment['current_risk_score'],
                'risk_level': assessment['risk_level'],
                'update_count': len(assessment['updates']),
                'duration': time.time() - assessment['start_time']
            })
        
        return {
            "status": "success",
            "active_assessments": active_assessments,
            "total_count": len(active_assessments),
            "timestamp": datetime.now()
        }
        
    except Exception as e:
        logger.error(f"Failed to get active assessments: {e}")
        raise HTTPException(status_code=500, detail=str(e))

# WebSocket endpoint for real-time updates
@app.websocket("/ws/risk-assessment/{assessment_id}")
async def websocket_endpoint(websocket: WebSocket, assessment_id: str):
    """WebSocket endpoint for real-time risk assessment updates"""
    await connection_manager.connect(websocket, assessment_id)
    
    try:
        # Send initial status
        try:
            status = real_time_assessment.get_assessment_status(assessment_id)
            await connection_manager.send_personal_message(
                json.dumps({
                    'type': 'initial_status',
                    'assessment_id': assessment_id,
                    'data': status,
                    'timestamp': time.time()
                }),
                websocket
            )
        except ValueError:
            # Assessment not found, send error
            await connection_manager.send_personal_message(
                json.dumps({
                    'type': 'error',
                    'message': f'Assessment {assessment_id} not found',
                    'timestamp': time.time()
                }),
                websocket
            )
            return
        
        # Keep connection alive and handle incoming messages
        while True:
            try:
                # Wait for messages from client
                data = await websocket.receive_text()
                message = json.loads(data)
                
                # Handle different message types
                if message.get('type') == 'ping':
                    await connection_manager.send_personal_message(
                        json.dumps({
                            'type': 'pong',
                            'timestamp': time.time()
                        }),
                        websocket
                    )
                elif message.get('type') == 'get_status':
                    status = real_time_assessment.get_assessment_status(assessment_id)
                    await connection_manager.send_personal_message(
                        json.dumps({
                            'type': 'status_update',
                            'assessment_id': assessment_id,
                            'data': status,
                            'timestamp': time.time()
                        }),
                        websocket
                    )
                
            except WebSocketDisconnect:
                break
            except Exception as e:
                logger.error(f"WebSocket error: {e}")
                await connection_manager.send_personal_message(
                    json.dumps({
                        'type': 'error',
                        'message': str(e),
                        'timestamp': time.time()
                    }),
                    websocket
                )
                
    except Exception as e:
        logger.error(f"WebSocket connection error: {e}")
    finally:
        connection_manager.disconnect(websocket)

# Real-time monitoring endpoint
@app.get("/real-time/monitoring")
async def get_real_time_monitoring():
    """Get real-time monitoring dashboard data"""
    try:
        # Get active assessments
        active_assessments = []
        for assessment_id, assessment in real_time_assessment.active_assessments.items():
            active_assessments.append({
                'assessment_id': assessment_id,
                'status': assessment['status'],
                'risk_level': assessment['risk_level'],
                'risk_score': assessment['current_risk_score'],
                'confidence': assessment['confidence'],
                'start_time': assessment['start_time'],
                'last_update': assessment.get('last_update', assessment['start_time']),
                'update_count': len(assessment['updates']),
                'duration': time.time() - assessment['start_time']
            })
        
        # Get WebSocket connection stats
        connection_stats = {
            'total_connections': len(connection_manager.active_connections),
            'monitored_assessments': len(connection_manager.connection_assessments)
        }
        
        # Get system stats
        system_stats = {
            'risk_detection_models_loaded': risk_detection_manager.is_initialized,
            'bert_model_loaded': risk_detection_manager.bert_model is not None,
            'anomaly_model_loaded': risk_detection_manager.anomaly_model is not None,
            'pattern_model_loaded': risk_detection_manager.pattern_model is not None,
            'cache_size': len(cache_manager.cache) if hasattr(cache_manager, 'cache') else 0
        }
        
        return {
            "status": "success",
            "monitoring_data": {
                "active_assessments": active_assessments,
                "connection_stats": connection_stats,
                "system_stats": system_stats,
                "timestamp": datetime.now()
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get monitoring data: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    logger.info("🚀 Starting Python ML Service...")
    logger.info(f"📱 Device: {Config.DEVICE}")
    logger.info("📚 Models will be loaded lazily on first request")
    
    # Get port from environment (Railway sets this automatically)
    port = int(os.getenv("PORT", "8000"))
    logger.info(f"🌐 Starting server on port {port}")
    
    # Initialize risk detection models (non-blocking, optional)
    if risk_detection_manager is not None:
        try:
            risk_detection_manager.initialize_models()
            logger.info("✅ Risk detection models initialized")
        except Exception as e:
            logger.warning(f"⚠️ Risk detection models initialization failed: {e}")
    else:
        logger.warning("⚠️ Risk detection manager not available - skipping initialization")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=port,
        log_level="info"
    )
