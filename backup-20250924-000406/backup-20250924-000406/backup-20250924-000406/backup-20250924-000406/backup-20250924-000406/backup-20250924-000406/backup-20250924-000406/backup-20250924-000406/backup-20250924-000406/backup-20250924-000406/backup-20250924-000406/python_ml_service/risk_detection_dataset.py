#!/usr/bin/env python3
"""
Risk Detection Training Dataset Generator

This module creates a comprehensive training dataset for risk detection models
with minimum 5,000 samples covering all risk categories and severity levels.

Target: 90%+ accuracy for risk detection
"""

import os
import json
import random
import logging
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime
from pathlib import Path
import pandas as pd
import numpy as np

# Configure logging
logger = logging.getLogger(__name__)

class RiskDetectionDatasetGenerator:
    """Generator for risk detection training dataset"""
    
    def __init__(self):
        self.risk_categories = [
            "illegal", "prohibited", "high_risk", "tbml", "sanctions", "fraud"
        ]
        self.risk_severity_levels = ["low", "medium", "high", "critical"]
        
        # Risk category templates
        self.risk_templates = {
            "illegal": {
                "drug_trafficking": [
                    "Cocaine distribution network",
                    "Heroin smuggling operation",
                    "Marijuana cultivation facility",
                    "Methamphetamine production lab",
                    "Drug trafficking organization",
                    "Illegal drug distribution",
                    "Contraband smuggling service",
                    "Underground drug market"
                ],
                "weapons_trafficking": [
                    "Illegal weapons dealer",
                    "Firearms smuggling operation",
                    "Black market weapons",
                    "Unlicensed gun sales",
                    "Weapon trafficking network",
                    "Illegal arms dealer",
                    "Contraband weapons supplier"
                ],
                "human_trafficking": [
                    "Human trafficking ring",
                    "Forced labor operation",
                    "Sex trafficking network",
                    "Human smuggling service",
                    "Trafficking victims",
                    "Modern slavery operation"
                ]
            },
            "prohibited": {
                "gambling": [
                    "Online casino operation",
                    "Sports betting service",
                    "Poker room business",
                    "Gambling establishment",
                    "Betting platform",
                    "Casino gaming service",
                    "Lottery ticket sales"
                ],
                "adult_entertainment": [
                    "Adult entertainment venue",
                    "Strip club business",
                    "Adult content production",
                    "Escort service",
                    "Adult toy store",
                    "Pornography distribution"
                ],
                "cryptocurrency": [
                    "Cryptocurrency exchange",
                    "Bitcoin trading platform",
                    "Digital currency service",
                    "Crypto mining operation",
                    "Blockchain investment",
                    "Digital asset trading"
                ]
            },
            "high_risk": {
                "money_services": [
                    "Money transfer service",
                    "Currency exchange business",
                    "Check cashing service",
                    "Wire transfer company",
                    "Money order service",
                    "International remittance"
                ],
                "prepaid_cards": [
                    "Prepaid card issuer",
                    "Gift card distributor",
                    "Stored value cards",
                    "Prepaid debit cards",
                    "Virtual card service"
                ],
                "high_risk_merchants": [
                    "Travel booking service",
                    "Dating website",
                    "Online auction site",
                    "Peer-to-peer lending",
                    "Crowdfunding platform"
                ]
            },
            "tbml": {
                "shell_companies": [
                    "Shell company formation",
                    "Front company services",
                    "Nominee company setup",
                    "Offshore company registration",
                    "Straw man corporation",
                    "Paper company creation"
                ],
                "trade_finance": [
                    "Trade finance company",
                    "Import export business",
                    "Commodity trading firm",
                    "International trade service",
                    "Cross-border commerce",
                    "Trade documentation service"
                ],
                "complex_structures": [
                    "Complex corporate structure",
                    "Multi-layered organization",
                    "Offshore trust setup",
                    "Tax optimization service",
                    "Asset protection scheme"
                ]
            },
            "sanctions": {
                "sanctions_evasion": [
                    "Sanctions evasion service",
                    "Embargo circumvention",
                    "Blocked entity assistance",
                    "OFAC violation support",
                    "Prohibited transaction facilitation"
                ],
                "restricted_countries": [
                    "Iran business operations",
                    "North Korea trade",
                    "Russia sanctions evasion",
                    "Venezuela commerce",
                    "Cuba business dealings"
                ]
            },
            "fraud": {
                "identity_theft": [
                    "Identity theft service",
                    "Fake ID creation",
                    "Stolen identity sales",
                    "Identity fraud operation",
                    "Personal information theft"
                ],
                "financial_fraud": [
                    "Credit card fraud",
                    "Bank fraud operation",
                    "Wire fraud service",
                    "Investment fraud scheme",
                    "Pyramid scheme operation"
                ],
                "counterfeit": [
                    "Counterfeit goods sales",
                    "Fake product manufacturing",
                    "Forgery service",
                    "Counterfeit currency",
                    "Fake document creation"
                ]
            }
        }
        
        # Low-risk business templates
        self.low_risk_templates = [
            "Restaurant and food service",
            "Retail clothing store",
            "Grocery store chain",
            "Automotive repair shop",
            "Dental practice",
            "Law firm",
            "Accounting services",
            "Real estate agency",
            "Insurance company",
            "Banking services",
            "Technology consulting",
            "Marketing agency",
            "Construction company",
            "Manufacturing plant",
            "Healthcare clinic",
            "Educational institution",
            "Non-profit organization",
            "Government agency",
            "Transportation service",
            "Energy company"
        ]
    
    def generate_dataset(self, num_samples: int = 5000) -> pd.DataFrame:
        """Generate comprehensive risk detection dataset"""
        logger.info(f"Generating risk detection dataset with {num_samples} samples...")
        
        # Calculate distribution
        high_risk_samples = int(num_samples * 0.4)  # 40% high risk
        medium_risk_samples = int(num_samples * 0.3)  # 30% medium risk
        low_risk_samples = int(num_samples * 0.3)  # 30% low risk
        
        dataset = []
        
        # Generate high-risk samples
        logger.info("Generating high-risk samples...")
        high_risk_data = self._generate_risk_samples(high_risk_samples, "high")
        dataset.extend(high_risk_data)
        
        # Generate medium-risk samples
        logger.info("Generating medium-risk samples...")
        medium_risk_data = self._generate_risk_samples(medium_risk_samples, "medium")
        dataset.extend(medium_risk_data)
        
        # Generate low-risk samples
        logger.info("Generating low-risk samples...")
        low_risk_data = self._generate_low_risk_samples(low_risk_samples)
        dataset.extend(low_risk_data)
        
        # Shuffle dataset
        random.shuffle(dataset)
        
        # Convert to DataFrame
        df = pd.DataFrame(dataset)
        
        logger.info(f"✅ Generated dataset with {len(df)} samples")
        logger.info(f"📊 Risk level distribution:")
        logger.info(f"   - High risk: {len(df[df['risk_level'] == 'high'])}")
        logger.info(f"   - Medium risk: {len(df[df['risk_level'] == 'medium'])}")
        logger.info(f"   - Low risk: {len(df[df['risk_level'] == 'low'])}")
        
        return df
    
    def _generate_risk_samples(self, num_samples: int, risk_level: str) -> List[Dict]:
        """Generate risk samples for specific risk level"""
        samples = []
        
        for i in range(num_samples):
            # Select random risk category
            risk_category = random.choice(self.risk_categories)
            
            # Select random subcategory
            subcategories = list(self.risk_templates[risk_category].keys())
            subcategory = random.choice(subcategories)
            
            # Select random template
            templates = self.risk_templates[risk_category][subcategory]
            template = random.choice(templates)
            
            # Generate business name
            business_name = self._generate_business_name(template)
            
            # Generate description
            description = self._generate_description(template, risk_category, subcategory)
            
            # Generate website URL
            website_url = self._generate_website_url(business_name)
            
            # Generate website content
            website_content = self._generate_website_content(template, risk_category, subcategory)
            
            # Calculate risk score based on risk level
            if risk_level == "high":
                risk_score = random.uniform(0.7, 1.0)
                severity = random.choice(["high", "critical"])
            else:  # medium
                risk_score = random.uniform(0.4, 0.7)
                severity = random.choice(["medium", "high"])
            
            sample = {
                'id': f"risk_{i:06d}",
                'business_name': business_name,
                'description': description,
                'website_url': website_url,
                'website_content': website_content,
                'risk_category': risk_category,
                'risk_subcategory': subcategory,
                'risk_level': risk_level,
                'risk_severity': severity,
                'risk_score': risk_score,
                'is_risk': True,
                'created_at': datetime.now().isoformat()
            }
            
            samples.append(sample)
        
        return samples
    
    def _generate_low_risk_samples(self, num_samples: int) -> List[Dict]:
        """Generate low-risk samples"""
        samples = []
        
        for i in range(num_samples):
            # Select random low-risk template
            template = random.choice(self.low_risk_templates)
            
            # Generate business name
            business_name = self._generate_business_name(template)
            
            # Generate description
            description = self._generate_low_risk_description(template)
            
            # Generate website URL
            website_url = self._generate_website_url(business_name)
            
            # Generate website content
            website_content = self._generate_low_risk_website_content(template)
            
            # Low risk characteristics
            risk_score = random.uniform(0.0, 0.3)
            severity = random.choice(["low", "medium"])
            
            sample = {
                'id': f"low_risk_{i:06d}",
                'business_name': business_name,
                'description': description,
                'website_url': website_url,
                'website_content': website_content,
                'risk_category': 'low_risk',
                'risk_subcategory': 'legitimate_business',
                'risk_level': 'low',
                'risk_severity': severity,
                'risk_score': risk_score,
                'is_risk': False,
                'created_at': datetime.now().isoformat()
            }
            
            samples.append(sample)
        
        return samples
    
    def _generate_business_name(self, template: str) -> str:
        """Generate business name from template"""
        # Add company suffixes
        suffixes = ["Inc", "LLC", "Corp", "Ltd", "Co", "Group", "Enterprises", "Services", "Solutions"]
        suffix = random.choice(suffixes)
        
        # Add location
        locations = ["Global", "International", "National", "Regional", "Local", "Metro", "City", "State"]
        location = random.choice(locations)
        
        # Generate name variations
        name_variations = [
            f"{template} {suffix}",
            f"{location} {template}",
            f"{template} {location} {suffix}",
            f"{template} & Associates",
            f"{template} Partners",
            f"{template} Holdings"
        ]
        
        return random.choice(name_variations)
    
    def _generate_description(self, template: str, risk_category: str, subcategory: str) -> str:
        """Generate business description for risk samples"""
        descriptions = {
            "illegal": {
                "drug_trafficking": [
                    f"{template} specializing in pharmaceutical distribution and medical supplies",
                    f"{template} providing healthcare and wellness products",
                    f"{template} offering alternative medicine and natural remedies",
                    f"{template} focused on medical research and development"
                ],
                "weapons_trafficking": [
                    f"{template} providing security equipment and protective gear",
                    f"{template} specializing in outdoor recreation and hunting supplies",
                    f"{template} offering military surplus and collectibles",
                    f"{template} focused on security consulting and training"
                ],
                "human_trafficking": [
                    f"{template} providing employment and staffing services",
                    f"{template} specializing in international recruitment and placement",
                    f"{template} offering travel and tourism services",
                    f"{template} focused on cultural exchange programs"
                ]
            },
            "prohibited": {
                "gambling": [
                    f"{template} providing entertainment and gaming services",
                    f"{template} specializing in recreational activities and leisure",
                    f"{template} offering online entertainment and digital services",
                    f"{template} focused on hospitality and tourism"
                ],
                "adult_entertainment": [
                    f"{template} providing entertainment and nightlife services",
                    f"{template} specializing in hospitality and customer service",
                    f"{template} offering event planning and entertainment",
                    f"{template} focused on leisure and recreation"
                ],
                "cryptocurrency": [
                    f"{template} providing financial technology and digital services",
                    f"{template} specializing in blockchain and distributed systems",
                    f"{template} offering digital asset management and trading",
                    f"{template} focused on fintech innovation and development"
                ]
            },
            "high_risk": {
                "money_services": [
                    f"{template} providing financial services and money management",
                    f"{template} specializing in international remittance and transfers",
                    f"{template} offering currency exchange and financial consulting",
                    f"{template} focused on cross-border financial services"
                ],
                "prepaid_cards": [
                    f"{template} providing payment solutions and financial services",
                    f"{template} specializing in prepaid and stored value products",
                    f"{template} offering digital payment and card services",
                    f"{template} focused on financial technology and innovation"
                ],
                "high_risk_merchants": [
                    f"{template} providing online services and digital platforms",
                    f"{template} specializing in e-commerce and online marketplaces",
                    f"{template} offering technology solutions and digital services",
                    f"{template} focused on internet-based business services"
                ]
            },
            "tbml": {
                "shell_companies": [
                    f"{template} providing corporate services and business formation",
                    f"{template} specializing in international business consulting",
                    f"{template} offering offshore services and corporate solutions",
                    f"{template} focused on business development and consulting"
                ],
                "trade_finance": [
                    f"{template} providing international trade and commerce services",
                    f"{template} specializing in import/export and logistics",
                    f"{template} offering trade finance and commercial services",
                    f"{template} focused on global commerce and international business"
                ],
                "complex_structures": [
                    f"{template} providing corporate structuring and business consulting",
                    f"{template} specializing in tax planning and business optimization",
                    f"{template} offering asset management and financial consulting",
                    f"{template} focused on business development and strategic planning"
                ]
            },
            "sanctions": {
                "sanctions_evasion": [
                    f"{template} providing international business and trade services",
                    f"{template} specializing in global commerce and cross-border trade",
                    f"{template} offering international consulting and business services",
                    f"{template} focused on worldwide business development"
                ],
                "restricted_countries": [
                    f"{template} providing international trade and commerce services",
                    f"{template} specializing in global business and international relations",
                    f"{template} offering cross-border commerce and trade services",
                    f"{template} focused on international business development"
                ]
            },
            "fraud": {
                "identity_theft": [
                    f"{template} providing identity verification and security services",
                    f"{template} specializing in personal information management",
                    f"{template} offering identity protection and security solutions",
                    f"{template} focused on personal security and privacy services"
                ],
                "financial_fraud": [
                    f"{template} providing financial services and investment consulting",
                    f"{template} specializing in wealth management and financial planning",
                    f"{template} offering investment services and financial consulting",
                    f"{template} focused on financial services and wealth management"
                ],
                "counterfeit": [
                    f"{template} providing product manufacturing and distribution",
                    f"{template} specializing in consumer goods and retail products",
                    f"{template} offering manufacturing and product development",
                    f"{template} focused on product design and manufacturing"
                ]
            }
        }
        
        category_descriptions = descriptions.get(risk_category, {}).get(subcategory, [template])
        return random.choice(category_descriptions)
    
    def _generate_low_risk_description(self, template: str) -> str:
        """Generate business description for low-risk samples"""
        descriptions = [
            f"{template} providing quality services and customer satisfaction",
            f"{template} specializing in professional services and expertise",
            f"{template} offering reliable solutions and trusted service",
            f"{template} focused on excellence and customer service",
            f"{template} providing innovative solutions and professional service",
            f"{template} specializing in quality products and reliable service",
            f"{template} offering comprehensive services and expert solutions",
            f"{template} focused on customer satisfaction and quality service"
        ]
        
        return random.choice(descriptions)
    
    def _generate_website_url(self, business_name: str) -> str:
        """Generate website URL from business name"""
        # Clean business name for URL
        url_name = business_name.lower()
        url_name = url_name.replace(" ", "").replace("&", "and").replace(",", "").replace(".", "")
        url_name = url_name.replace("inc", "").replace("llc", "").replace("corp", "").replace("ltd", "")
        
        # Add domain extensions
        extensions = [".com", ".net", ".org", ".biz", ".info"]
        extension = random.choice(extensions)
        
        return f"https://www.{url_name}{extension}"
    
    def _generate_website_content(self, template: str, risk_category: str, subcategory: str) -> str:
        """Generate website content for risk samples"""
        content_templates = {
            "illegal": "Welcome to our professional services. We provide specialized solutions for your needs. Contact us for confidential consultations and discrete services.",
            "prohibited": "Experience our premium services and exclusive offerings. We provide specialized solutions for discerning clients. Join our exclusive community.",
            "high_risk": "Discover our innovative financial solutions and cutting-edge services. We offer specialized products for sophisticated clients and investors.",
            "tbml": "Explore our international business services and global solutions. We provide comprehensive services for international commerce and trade.",
            "sanctions": "Welcome to our global business platform. We offer international services and cross-border solutions for worldwide clients.",
            "fraud": "Experience our secure services and trusted solutions. We provide confidential services and personalized solutions for our clients."
        }
        
        base_content = content_templates.get(risk_category, "Welcome to our professional services.")
        
        # Add additional content
        additional_content = [
            "Our team of experts is ready to assist you with your specific needs.",
            "We maintain the highest standards of confidentiality and professionalism.",
            "Contact us today for a personalized consultation and customized solutions.",
            "We pride ourselves on delivering exceptional results and client satisfaction.",
            "Our services are designed to meet the unique requirements of each client."
        ]
        
        return f"{base_content} {random.choice(additional_content)}"
    
    def _generate_low_risk_website_content(self, template: str) -> str:
        """Generate website content for low-risk samples"""
        content_templates = [
            f"Welcome to {template}. We are committed to providing excellent service and quality products to our customers.",
            f"At {template}, we pride ourselves on our professional approach and customer satisfaction.",
            f"{template} offers reliable services and trusted solutions for all your needs.",
            f"Experience the difference with {template} - your trusted partner for quality service.",
            f"{template} provides comprehensive solutions and expert service to meet your requirements."
        ]
        
        base_content = random.choice(content_templates)
        
        # Add additional content
        additional_content = [
            "We look forward to serving you and building a long-term relationship.",
            "Contact us today to learn more about our services and how we can help you.",
            "Our experienced team is ready to assist you with your specific needs.",
            "We maintain the highest standards of quality and customer service.",
            "Thank you for choosing us for your business needs."
        ]
        
        return f"{base_content} {random.choice(additional_content)}"
    
    def save_dataset(self, df: pd.DataFrame, filepath: str):
        """Save dataset to file"""
        # Save as CSV
        csv_path = filepath.replace('.json', '.csv')
        df.to_csv(csv_path, index=False)
        logger.info(f"Dataset saved to {csv_path}")
        
        # Save as JSON
        df.to_json(filepath, orient='records', indent=2)
        logger.info(f"Dataset saved to {filepath}")
        
        # Save dataset statistics
        stats_path = filepath.replace('.json', '_stats.json')
        stats = {
            'total_samples': len(df),
            'risk_level_distribution': df['risk_level'].value_counts().to_dict(),
            'risk_category_distribution': df['risk_category'].value_counts().to_dict(),
            'risk_severity_distribution': df['risk_severity'].value_counts().to_dict(),
            'risk_score_stats': {
                'mean': float(df['risk_score'].mean()),
                'std': float(df['risk_score'].std()),
                'min': float(df['risk_score'].min()),
                'max': float(df['risk_score'].max())
            },
            'created_at': datetime.now().isoformat()
        }
        
        with open(stats_path, 'w') as f:
            json.dump(stats, f, indent=2)
        
        logger.info(f"Dataset statistics saved to {stats_path}")

def main():
    """Generate risk detection dataset"""
    logger.info("🚀 Starting risk detection dataset generation...")
    
    # Create dataset generator
    generator = RiskDetectionDatasetGenerator()
    
    # Generate dataset
    dataset = generator.generate_dataset(num_samples=5000)
    
    # Save dataset
    output_dir = Path("data/risk_detection")
    output_dir.mkdir(parents=True, exist_ok=True)
    
    dataset_path = output_dir / "risk_detection_dataset.json"
    generator.save_dataset(dataset, str(dataset_path))
    
    logger.info("✅ Risk detection dataset generation completed!")

if __name__ == "__main__":
    main()
