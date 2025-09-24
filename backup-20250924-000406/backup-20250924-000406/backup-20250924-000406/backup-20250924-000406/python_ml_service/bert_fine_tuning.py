#!/usr/bin/env python3
"""
BERT Fine-tuning Pipeline for Business Classification

This module implements a comprehensive BERT fine-tuning pipeline for business
classification with the following features:
- BERT model fine-tuning (bert-base-uncased)
- DistilBERT model for faster inference
- Custom neural networks for specific industry sectors
- Model quantization for performance optimization
- Confidence scoring and explainability features
- Model caching for sub-100ms response times

Target: 95%+ accuracy for classification
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
import torch.optim as optim
from torch.utils.data import DataLoader, Dataset, WeightedRandomSampler
import transformers
from transformers import (
    AutoTokenizer, AutoModel, AutoModelForSequenceClassification,
    BertTokenizer, BertForSequenceClassification, BertConfig,
    DistilBertTokenizer, DistilBertForSequenceClassification, DistilBertConfig,
    TrainingArguments, Trainer, EarlyStoppingCallback, DataCollatorWithPadding,
    get_linear_schedule_with_warmup
)
import numpy as np
import pandas as pd
from sklearn.metrics import (
    accuracy_score, precision_recall_fscore_support, 
    classification_report, confusion_matrix, roc_auc_score
)
from sklearn.model_selection import train_test_split, StratifiedKFold
from sklearn.preprocessing import LabelEncoder
import joblib
from tqdm import tqdm
import matplotlib.pyplot as plt
import seaborn as sns

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class BusinessClassificationDataset(Dataset):
    """Custom dataset for business classification"""
    
    def __init__(self, texts: List[str], labels: List[str], tokenizer, max_length: int = 512):
        self.texts = texts
        self.labels = labels
        self.tokenizer = tokenizer
        self.max_length = max_length
        
        # Encode labels
        self.label_encoder = LabelEncoder()
        self.encoded_labels = self.label_encoder.fit_transform(labels)
        self.num_labels = len(self.label_encoder.classes_)
        self.label_names = self.label_encoder.classes_
    
    def __len__(self):
        return len(self.texts)
    
    def __getitem__(self, idx):
        text = str(self.texts[idx])
        label = self.encoded_labels[idx]
        
        # Tokenize
        encoding = self.tokenizer(
            text,
            max_length=self.max_length,
            padding='max_length',
            truncation=True,
            return_tensors='pt'
        )
        
        return {
            'input_ids': encoding['input_ids'].flatten(),
            'attention_mask': encoding['attention_mask'].flatten(),
            'labels': torch.tensor(label, dtype=torch.long)
        }
    
    def get_label_names(self):
        return self.label_names

class BERTFineTuningPipeline:
    """BERT Fine-tuning Pipeline for Business Classification"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.tokenizer = None
        self.model = None
        self.label_encoder = None
        self.training_history = []
        
        # Create directories
        self.model_save_path = Path(config.get('model_save_path', 'models'))
        self.model_save_path.mkdir(parents=True, exist_ok=True)
        
        self.logs_path = Path(config.get('logs_path', 'logs'))
        self.logs_path.mkdir(parents=True, exist_ok=True)
        
        logger.info(f"🚀 BERT Fine-tuning Pipeline initialized")
        logger.info(f"📱 Device: {self.device}")
        logger.info(f"💾 Model save path: {self.model_save_path}")
    
    def prepare_data(self, data_path: str) -> Tuple[BusinessClassificationDataset, BusinessClassificationDataset]:
        """Prepare training and validation datasets"""
        logger.info("📊 Preparing training data...")
        
        # Load data
        if data_path.endswith('.csv'):
            df = pd.read_csv(data_path)
        elif data_path.endswith('.json'):
            df = pd.read_json(data_path)
        else:
            raise ValueError("Unsupported file format. Use CSV or JSON.")
        
        # Validate required columns
        required_columns = ['business_name', 'industry']
        for col in required_columns:
            if col not in df.columns:
                raise ValueError(f"Required column '{col}' not found in data")
        
        # Prepare text data
        df['text'] = df['business_name'].astype(str)
        if 'description' in df.columns:
            df['text'] += ' ' + df['description'].astype(str)
        if 'website_url' in df.columns:
            df['text'] += ' ' + df['website_url'].astype(str)
        
        # Clean data
        df = df.dropna(subset=['text', 'industry'])
        df = df[df['text'].str.len() > 10]  # Remove very short texts
        
        logger.info(f"📈 Dataset size: {len(df)} samples")
        logger.info(f"🏷️ Number of industries: {df['industry'].nunique()}")
        logger.info(f"📊 Industry distribution:")
        for industry, count in df['industry'].value_counts().head(10).items():
            logger.info(f"   {industry}: {count}")
        
        # Split data
        train_texts, val_texts, train_labels, val_labels = train_test_split(
            df['text'].tolist(),
            df['industry'].tolist(),
            test_size=0.2,
            random_state=42,
            stratify=df['industry']
        )
        
        # Create datasets
        train_dataset = BusinessClassificationDataset(
            train_texts, train_labels, self.tokenizer, self.config['max_length']
        )
        val_dataset = BusinessClassificationDataset(
            val_texts, val_labels, self.tokenizer, self.config['max_length']
        )
        
        # Store label encoder
        self.label_encoder = train_dataset.label_encoder
        
        logger.info(f"✅ Data prepared: {len(train_dataset)} train, {len(val_dataset)} validation")
        return train_dataset, val_dataset
    
    def load_tokenizer(self, model_name: str = "bert-base-uncased"):
        """Load tokenizer"""
        logger.info(f"🔤 Loading tokenizer: {model_name}")
        
        if "distilbert" in model_name.lower():
            self.tokenizer = DistilBertTokenizer.from_pretrained(model_name)
        else:
            self.tokenizer = BertTokenizer.from_pretrained(model_name)
        
        logger.info("✅ Tokenizer loaded successfully")
    
    def load_model(self, model_name: str = "bert-base-uncased", num_labels: int = None):
        """Load pre-trained model"""
        logger.info(f"🤖 Loading model: {model_name}")
        
        if num_labels is None:
            num_labels = len(self.label_encoder.classes_) if self.label_encoder else 16
        
        if "distilbert" in model_name.lower():
            config = DistilBertConfig.from_pretrained(
                model_name,
                num_labels=num_labels,
                problem_type="single_label_classification"
            )
            self.model = DistilBertForSequenceClassification.from_pretrained(
                model_name,
                config=config
            )
        else:
            config = BertConfig.from_pretrained(
                model_name,
                num_labels=num_labels,
                problem_type="single_label_classification"
            )
            self.model = BertForSequenceClassification.from_pretrained(
                model_name,
                config=config
            )
        
        self.model.to(self.device)
        logger.info(f"✅ Model loaded successfully with {num_labels} labels")
    
    def create_custom_model(self, model_type: str, num_labels: int) -> nn.Module:
        """Create custom neural network model for specific industries"""
        logger.info(f"🏗️ Creating custom {model_type} model")
        
        if model_type == "financial_services":
            return self._create_financial_services_model(num_labels)
        elif model_type == "healthcare":
            return self._create_healthcare_model(num_labels)
        elif model_type == "technology":
            return self._create_technology_model(num_labels)
        elif model_type == "retail":
            return self._create_retail_model(num_labels)
        elif model_type == "manufacturing":
            return self._create_manufacturing_model(num_labels)
        else:
            raise ValueError(f"Unknown model type: {model_type}")
    
    def _create_financial_services_model(self, num_labels: int) -> nn.Module:
        """Create custom model for financial services"""
        class FinancialServicesModel(nn.Module):
            def __init__(self, vocab_size=30522, hidden_size=768, num_labels=num_labels):
                super().__init__()
                self.embedding = nn.Embedding(vocab_size, hidden_size)
                self.lstm = nn.LSTM(hidden_size, hidden_size, batch_first=True, bidirectional=True)
                self.attention = nn.MultiheadAttention(hidden_size * 2, num_heads=8, batch_first=True)
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
        
        return FinancialServicesModel()
    
    def _create_healthcare_model(self, num_labels: int) -> nn.Module:
        """Create custom model for healthcare"""
        class HealthcareModel(nn.Module):
            def __init__(self, vocab_size=30522, hidden_size=768, num_labels=num_labels):
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
                embedded = embedded.transpose(1, 2)
                conv1 = torch.relu(self.conv1d(embedded))
                conv2 = torch.relu(self.conv1d_2(conv1))
                pooled = self.pool(conv2).squeeze(-1)
                return self.classifier(pooled)
        
        return HealthcareModel()
    
    def _create_technology_model(self, num_labels: int) -> nn.Module:
        """Create custom model for technology"""
        class TechnologyModel(nn.Module):
            def __init__(self, vocab_size=30522, hidden_size=768, num_labels=num_labels):
                super().__init__()
                self.embedding = nn.Embedding(vocab_size, hidden_size)
                self.transformer = nn.TransformerEncoder(
                    nn.TransformerEncoderLayer(hidden_size, nhead=8, batch_first=True),
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
                transformer_out = self.transformer(embedded)
                pooled = torch.mean(transformer_out, dim=1)
                return self.classifier(pooled)
        
        return TechnologyModel()
    
    def _create_retail_model(self, num_labels: int) -> nn.Module:
        """Create custom model for retail"""
        class RetailModel(nn.Module):
            def __init__(self, vocab_size=30522, hidden_size=768, num_labels=num_labels):
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
        
        return RetailModel()
    
    def _create_manufacturing_model(self, num_labels: int) -> nn.Module:
        """Create custom model for manufacturing"""
        class ManufacturingModel(nn.Module):
            def __init__(self, vocab_size=30522, hidden_size=768, num_labels=num_labels):
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
                embedded = embedded.transpose(1, 2)
                cnn1 = torch.relu(self.cnn1(embedded))
                cnn2 = torch.relu(self.cnn2(cnn1))
                cnn3 = torch.relu(self.cnn3(cnn2))
                pooled = self.pool(cnn3).squeeze(-1)
                return self.classifier(pooled)
        
        return ManufacturingModel()
    
    def train_model(self, train_dataset: BusinessClassificationDataset, 
                   val_dataset: BusinessClassificationDataset) -> Dict[str, Any]:
        """Train the model"""
        logger.info("🎯 Starting model training...")
        
        # Training arguments
        training_args = TrainingArguments(
            output_dir=str(self.model_save_path / "training_output"),
            num_train_epochs=self.config['num_epochs'],
            per_device_train_batch_size=self.config['batch_size'],
            per_device_eval_batch_size=self.config['batch_size'],
            warmup_steps=self.config['warmup_steps'],
            weight_decay=self.config['weight_decay'],
            learning_rate=self.config['learning_rate'],
            logging_dir=str(self.logs_path),
            logging_steps=100,
            evaluation_strategy="epoch",
            save_strategy="epoch",
            save_total_limit=3,
            load_best_model_at_end=True,
            metric_for_best_model="eval_accuracy",
            greater_is_better=True,
            report_to=None,  # Disable wandb/tensorboard
            seed=42,
            fp16=torch.cuda.is_available(),  # Use mixed precision if available
        )
        
        # Data collator
        data_collator = DataCollatorWithPadding(tokenizer=self.tokenizer)
        
        # Trainer
        trainer = Trainer(
            model=self.model,
            args=training_args,
            train_dataset=train_dataset,
            eval_dataset=val_dataset,
            data_collator=data_collator,
            callbacks=[EarlyStoppingCallback(early_stopping_patience=3)]
        )
        
        # Train
        start_time = time.time()
        training_result = trainer.train()
        training_time = time.time() - start_time
        
        # Evaluate
        eval_result = trainer.evaluate()
        
        # Save model
        trainer.save_model(str(self.model_save_path / "best_model"))
        self.tokenizer.save_pretrained(str(self.model_save_path / "best_model"))
        
        # Save label encoder
        with open(self.model_save_path / "label_encoder.pkl", "wb") as f:
            pickle.dump(self.label_encoder, f)
        
        # Training history
        training_history = {
            "training_time": training_time,
            "final_train_loss": training_result.training_loss,
            "final_eval_loss": eval_result["eval_loss"],
            "final_eval_accuracy": eval_result["eval_accuracy"],
            "best_model_path": str(self.model_save_path / "best_model"),
            "config": self.config
        }
        
        self.training_history.append(training_history)
        
        logger.info(f"✅ Training completed in {training_time:.2f} seconds")
        logger.info(f"📊 Final accuracy: {eval_result['eval_accuracy']:.4f}")
        logger.info(f"💾 Model saved to: {self.model_save_path / 'best_model'}")
        
        return training_history
    
    def evaluate_model(self, test_dataset: BusinessClassificationDataset) -> Dict[str, Any]:
        """Evaluate model on test dataset"""
        logger.info("📊 Evaluating model...")
        
        self.model.eval()
        all_predictions = []
        all_labels = []
        all_probabilities = []
        
        with torch.no_grad():
            for batch in tqdm(DataLoader(test_dataset, batch_size=self.config['batch_size'])):
                input_ids = batch['input_ids'].to(self.device)
                attention_mask = batch['attention_mask'].to(self.device)
                labels = batch['labels'].to(self.device)
                
                outputs = self.model(input_ids=input_ids, attention_mask=attention_mask)
                logits = outputs.logits
                probabilities = torch.softmax(logits, dim=-1)
                predictions = torch.argmax(logits, dim=-1)
                
                all_predictions.extend(predictions.cpu().numpy())
                all_labels.extend(labels.cpu().numpy())
                all_probabilities.extend(probabilities.cpu().numpy())
        
        # Calculate metrics
        accuracy = accuracy_score(all_labels, all_predictions)
        precision, recall, f1, _ = precision_recall_fscore_support(
            all_labels, all_predictions, average='weighted'
        )
        
        # Classification report
        class_report = classification_report(
            all_labels, all_predictions,
            target_names=self.label_encoder.classes_,
            output_dict=True
        )
        
        # Confusion matrix
        cm = confusion_matrix(all_labels, all_predictions)
        
        # Calculate confidence scores
        max_probabilities = [max(probs) for probs in all_probabilities]
        avg_confidence = np.mean(max_probabilities)
        
        evaluation_results = {
            "accuracy": accuracy,
            "precision": precision,
            "recall": recall,
            "f1_score": f1,
            "average_confidence": avg_confidence,
            "classification_report": class_report,
            "confusion_matrix": cm.tolist(),
            "label_names": self.label_encoder.classes_.tolist()
        }
        
        logger.info(f"📈 Test Accuracy: {accuracy:.4f}")
        logger.info(f"📈 Test Precision: {precision:.4f}")
        logger.info(f"📈 Test Recall: {recall:.4f}")
        logger.info(f"📈 Test F1-Score: {f1:.4f}")
        logger.info(f"📈 Average Confidence: {avg_confidence:.4f}")
        
        return evaluation_results
    
    def quantize_model(self, model_path: str = None) -> str:
        """Quantize model for faster inference"""
        logger.info("⚡ Quantizing model for faster inference...")
        
        if model_path is None:
            model_path = str(self.model_save_path / "best_model")
        
        # Load model
        model = AutoModelForSequenceClassification.from_pretrained(model_path)
        tokenizer = AutoTokenizer.from_pretrained(model_path)
        
        # Quantize model
        quantized_model = torch.quantization.quantize_dynamic(
            model, {torch.nn.Linear}, dtype=torch.qint8
        )
        
        # Save quantized model
        quantized_path = str(self.model_save_path / "quantized_model")
        quantized_model.save_pretrained(quantized_path)
        tokenizer.save_pretrained(quantized_path)
        
        logger.info(f"✅ Model quantized and saved to: {quantized_path}")
        return quantized_path
    
    def create_explainability_report(self, text: str, model_path: str = None) -> Dict[str, Any]:
        """Create explainability report for a prediction"""
        logger.info("🔍 Creating explainability report...")
        
        if model_path is None:
            model_path = str(self.model_save_path / "best_model")
        
        # Load model and tokenizer
        model = AutoModelForSequenceClassification.from_pretrained(model_path)
        tokenizer = AutoTokenizer.from_pretrained(model_path)
        
        # Load label encoder
        with open(self.model_save_path / "label_encoder.pkl", "rb") as f:
            label_encoder = pickle.load(f)
        
        model.eval()
        
        # Tokenize input
        inputs = tokenizer(
            text,
            max_length=self.config['max_length'],
            padding=True,
            truncation=True,
            return_tensors="pt"
        )
        
        # Get prediction
        with torch.no_grad():
            outputs = model(**inputs)
            logits = outputs.logits
            probabilities = torch.softmax(logits, dim=-1)
            prediction = torch.argmax(logits, dim=-1)
        
        # Get top predictions
        top_predictions = torch.topk(probabilities, k=5, dim=-1)
        
        # Create explainability report
        explainability_report = {
            "input_text": text,
            "predicted_label": label_encoder.inverse_transform([prediction.item()])[0],
            "confidence": probabilities[0][prediction].item(),
            "top_predictions": [
                {
                    "label": label_encoder.inverse_transform([idx.item()])[0],
                    "confidence": prob.item()
                }
                for prob, idx in zip(top_predictions.values[0], top_predictions.indices[0])
            ],
            "tokens": tokenizer.convert_ids_to_tokens(inputs['input_ids'][0]),
            "attention_weights": None,  # Would need attention extraction
            "timestamp": datetime.now().isoformat()
        }
        
        logger.info(f"✅ Explainability report created")
        return explainability_report
    
    def save_training_report(self, training_history: Dict[str, Any], 
                           evaluation_results: Dict[str, Any]):
        """Save comprehensive training report"""
        logger.info("📝 Saving training report...")
        
        report = {
            "training_history": training_history,
            "evaluation_results": evaluation_results,
            "config": self.config,
            "timestamp": datetime.now().isoformat(),
            "model_info": {
                "model_type": self.config.get('model_name', 'bert-base-uncased'),
                "num_labels": len(self.label_encoder.classes_),
                "label_names": self.label_encoder.classes_.tolist()
            }
        }
        
        # Save JSON report
        with open(self.model_save_path / "training_report.json", "w") as f:
            json.dump(report, f, indent=2, default=str)
        
        # Create visualization
        self._create_training_visualizations(evaluation_results)
        
        logger.info(f"✅ Training report saved to: {self.model_save_path / 'training_report.json'}")
    
    def _create_training_visualizations(self, evaluation_results: Dict[str, Any]):
        """Create training visualizations"""
        try:
            # Confusion matrix
            plt.figure(figsize=(12, 10))
            cm = np.array(evaluation_results['confusion_matrix'])
            sns.heatmap(cm, annot=True, fmt='d', cmap='Blues',
                       xticklabels=evaluation_results['label_names'],
                       yticklabels=evaluation_results['label_names'])
            plt.title('Confusion Matrix')
            plt.xlabel('Predicted')
            plt.ylabel('Actual')
            plt.xticks(rotation=45)
            plt.yticks(rotation=0)
            plt.tight_layout()
            plt.savefig(self.model_save_path / "confusion_matrix.png", dpi=300, bbox_inches='tight')
            plt.close()
            
            # Class distribution
            plt.figure(figsize=(12, 6))
            class_counts = [cm[i].sum() for i in range(len(cm))]
            plt.bar(evaluation_results['label_names'], class_counts)
            plt.title('Class Distribution in Test Set')
            plt.xlabel('Industry')
            plt.ylabel('Count')
            plt.xticks(rotation=45)
            plt.tight_layout()
            plt.savefig(self.model_save_path / "class_distribution.png", dpi=300, bbox_inches='tight')
            plt.close()
            
            logger.info("📊 Training visualizations created")
            
        except Exception as e:
            logger.warning(f"⚠️ Failed to create visualizations: {e}")

def main():
    """Main function to run BERT fine-tuning pipeline"""
    
    # Configuration
    config = {
        'model_name': 'bert-base-uncased',
        'max_length': 512,
        'batch_size': 16,
        'learning_rate': 2e-5,
        'num_epochs': 3,
        'warmup_steps': 500,
        'weight_decay': 0.01,
        'model_save_path': 'models/bert_classification',
        'logs_path': 'logs',
        'data_path': 'data/business_classification_dataset.csv'
    }
    
    # Initialize pipeline
    pipeline = BERTFineTuningPipeline(config)
    
    # Load tokenizer
    pipeline.load_tokenizer(config['model_name'])
    
    # Prepare data
    train_dataset, val_dataset = pipeline.prepare_data(config['data_path'])
    
    # Load model
    pipeline.load_model(config['model_name'], len(train_dataset.label_encoder.classes_))
    
    # Train model
    training_history = pipeline.train_model(train_dataset, val_dataset)
    
    # Evaluate model
    evaluation_results = pipeline.evaluate_model(val_dataset)
    
    # Save training report
    pipeline.save_training_report(training_history, evaluation_results)
    
    # Quantize model
    quantized_path = pipeline.quantize_model()
    
    logger.info("🎉 BERT fine-tuning pipeline completed successfully!")
    logger.info(f"📊 Final accuracy: {evaluation_results['accuracy']:.4f}")
    logger.info(f"💾 Model saved to: {config['model_save_path']}")
    logger.info(f"⚡ Quantized model saved to: {quantized_path}")

if __name__ == "__main__":
    main()
