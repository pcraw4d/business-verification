#!/usr/bin/env python3
"""
Lightweight ML Model for Fast-Path Classification (Task 3.1)

This module implements a lightweight classifier optimized for:
- Fast inference (<100ms target)
- Short content (<256 tokens)
- High-confidence keyword matches
- Fast-path requests

Uses a smaller, faster model than the full DistilBART classifier.
"""

import os
import time
import logging
from typing import Dict, List, Optional, Any
from pathlib import Path

import torch
from transformers import pipeline

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class LightweightBusinessClassifier:
    """Lightweight Business Classifier for Fast-Path Requests"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        
        # Model paths
        self.model_save_path = Path(config.get('model_save_path', 'models/lightweight'))
        self.model_save_path.mkdir(parents=True, exist_ok=True)
        
        # Industry labels (same as full classifier for consistency)
        self.industry_labels = config.get('industry_labels', [
            "Technology", "Healthcare", "Financial Services",
            "Retail", "Food & Beverage", "Manufacturing",
            "Construction", "Real Estate", "Transportation",
            "Education", "Professional Services", "Agriculture",
            "Mining & Energy", "Utilities", "Wholesale Trade",
            "Arts & Entertainment", "Accommodation & Hospitality",
            "Administrative Services", "Other Services"
        ])
        
        # Initialize model
        self.classifier = None
        
        logger.info(f"🚀 Lightweight Business Classifier initializing on {self.device}")
        self._load_model()
    
    def _load_model(self):
        """Load lightweight classification model"""
        try:
            # Set Hugging Face cache directory
            cache_dir = os.getenv('TRANSFORMERS_CACHE', os.getenv('HF_HOME', '/app/.cache/huggingface'))
            os.makedirs(cache_dir, exist_ok=True)
            logger.info(f"📁 Using Hugging Face cache directory: {cache_dir}")
            
            # Use DistilBERT-MNLI (already lightweight, but we'll use it with shorter max_length)
            logger.info("📥 Loading lightweight DistilBERT-MNLI for fast classification...")
            self.classifier = pipeline(
                "zero-shot-classification",
                model="typeform/distilbert-base-uncased-mnli",
                device=0 if torch.cuda.is_available() else -1,
                cache_dir=cache_dir
            )
            logger.info("✅ Lightweight classification model loaded")
            
        except Exception as e:
            logger.error(f"❌ Failed to load lightweight model: {e}")
            raise
    
    def classify_fast(
        self,
        content: str,
        business_name: str = "",
        max_length: int = 256  # Shorter for fast inference
    ) -> Dict[str, Any]:
        """
        Fast classification for lightweight model
        
        Args:
            content: Text content to classify (truncated to max_length tokens)
            business_name: Business name (for context)
            max_length: Maximum content length (default 256 for fast inference)
        
        Returns:
            Classification result with confidence scores
        """
        start_time = time.time()
        
        try:
            # Truncate content for fast inference
            # Simple truncation by character count (rough approximation)
            if len(content) > max_length * 4:  # ~4 chars per token
                content = content[:max_length * 4]
                logger.debug(f"Truncated content to {len(content)} chars for fast inference")
            
            # Combine business name and content
            if business_name:
                combined = f"{business_name}. {content}"
            else:
                combined = content
            
            # Perform zero-shot classification
            result = self.classifier(
                combined,
                self.industry_labels,
                multi_label=False
            )
            
            # Extract top predictions
            labels = result.get('labels', [])
            scores = result.get('scores', [])
            
            # Build all_scores dictionary
            all_scores = {}
            for label, score in zip(labels, scores):
                all_scores[label] = float(score)
            
            # Get top prediction
            top_label = labels[0] if labels else "Other Services"
            top_confidence = float(scores[0]) if scores else 0.0
            
            processing_time = time.time() - start_time
            
            logger.info(f"✅ Fast classification completed in {processing_time:.3f}s: {top_label} ({top_confidence:.2%})")
            
            return {
                'industry': top_label,
                'confidence': top_confidence,
                'all_scores': all_scores,
                'processing_time': processing_time,
                'model': 'lightweight-distilbert',
                'method': 'fast_classification'
            }
            
        except Exception as e:
            logger.error(f"❌ Fast classification error: {e}")
            processing_time = time.time() - start_time
            return {
                'industry': 'Other Services',
                'confidence': 0.0,
                'all_scores': {},
                'processing_time': processing_time,
                'model': 'lightweight-distilbert',
                'method': 'fast_classification',
                'error': str(e)
            }
    
    def should_use_lightweight(
        self,
        content: str,
        use_fast_path: bool = False,
        keyword_confidence: float = 0.0
    ) -> bool:
        """
        Determine if lightweight model should be used
        
        Args:
            content: Text content length
            use_fast_path: Whether fast-path mode is enabled
            keyword_confidence: Confidence from keyword-based classification
        
        Returns:
            True if lightweight model should be used
        """
        # Use lightweight if:
        # 1. Fast-path mode is enabled
        # 2. Content is short (<256 tokens ~ 1024 chars)
        # 3. Keyword confidence is high (>0.85)
        
        content_short = len(content) < 1024
        high_keyword_confidence = keyword_confidence > 0.85
        
        should_use = use_fast_path or content_short or high_keyword_confidence
        
        if should_use:
            logger.debug(f"Using lightweight model: fast_path={use_fast_path}, short_content={content_short}, high_keyword={high_keyword_confidence}")
        
        return should_use

