#!/usr/bin/env python3
"""
Business Classification Training Dataset Generator

This module generates a comprehensive training dataset for business classification
with a minimum of 10,000 samples. The dataset includes:

- Business names from various industries
- Descriptions and website URLs
- Industry labels with proper distribution
- Synthetic data generation for underrepresented industries
- Data augmentation techniques
- Quality validation and filtering

Target: 10,000+ high-quality training samples with balanced industry distribution
"""

import os
import json
import random
import logging
import requests
from typing import Dict, List, Optional, Any, Tuple
from datetime import datetime
from pathlib import Path
import pandas as pd
import numpy as np
from faker import Faker
import re
from urllib.parse import urlparse
import time

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class BusinessDatasetGenerator:
    """Generator for business classification training dataset"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.fake = Faker()
        
        # Create directories
        self.data_path = Path(config.get('data_path', 'data'))
        self.data_path.mkdir(parents=True, exist_ok=True)
        
        # Industry definitions with keywords and patterns
        self.industries = self._define_industries()
        
        # Business name patterns by industry
        self.business_patterns = self._define_business_patterns()
        
        # Website domain patterns
        self.domain_patterns = self._define_domain_patterns()
        
        logger.info(f"🏭 Business Dataset Generator initialized")
        logger.info(f"📊 Target samples: {config.get('target_samples', 10000)}")
        logger.info(f"🏷️ Industries: {len(self.industries)}")
    
    def _define_industries(self) -> Dict[str, Dict[str, Any]]:
        """Define industry categories with keywords and characteristics"""
        return {
            "Technology": {
                "keywords": [
                    "software", "technology", "tech", "digital", "cyber", "data", "cloud",
                    "ai", "artificial intelligence", "machine learning", "blockchain",
                    "mobile", "app", "platform", "system", "solution", "innovation",
                    "computing", "network", "security", "analytics", "automation"
                ],
                "business_types": [
                    "Software Company", "Tech Solutions", "Digital Agency", "Data Analytics",
                    "Cloud Services", "AI Solutions", "Mobile Apps", "Cybersecurity",
                    "IT Consulting", "Tech Innovation", "Software Development", "Tech Startup"
                ],
                "weight": 0.15  # 15% of dataset
            },
            "Healthcare": {
                "keywords": [
                    "medical", "health", "healthcare", "clinic", "hospital", "doctor",
                    "pharmacy", "dental", "wellness", "therapy", "treatment", "care",
                    "medicine", "nursing", "mental health", "rehabilitation", "diagnostic",
                    "surgical", "pediatric", "geriatric", "emergency", "urgent care"
                ],
                "business_types": [
                    "Medical Center", "Healthcare Clinic", "Dental Practice", "Pharmacy",
                    "Wellness Center", "Therapy Services", "Medical Group", "Health System",
                    "Urgent Care", "Specialty Clinic", "Mental Health Services", "Rehabilitation"
                ],
                "weight": 0.12  # 12% of dataset
            },
            "Financial Services": {
                "keywords": [
                    "bank", "financial", "finance", "investment", "credit", "loan",
                    "insurance", "wealth", "capital", "trading", "advisory", "consulting",
                    "mortgage", "retirement", "pension", "fund", "asset", "portfolio",
                    "securities", "brokerage", "accounting", "tax", "audit"
                ],
                "business_types": [
                    "Bank", "Credit Union", "Investment Firm", "Insurance Agency",
                    "Financial Advisory", "Wealth Management", "Mortgage Company",
                    "Accounting Firm", "Tax Services", "Financial Planning", "Brokerage",
                    "Asset Management", "Financial Consulting"
                ],
                "weight": 0.10  # 10% of dataset
            },
            "Retail": {
                "keywords": [
                    "store", "shop", "retail", "market", "boutique", "outlet", "mall",
                    "fashion", "clothing", "apparel", "shoes", "accessories", "jewelry",
                    "electronics", "furniture", "home", "garden", "sports", "toys",
                    "books", "gifts", "department", "supermarket", "grocery"
                ],
                "business_types": [
                    "Retail Store", "Fashion Boutique", "Department Store", "Electronics Store",
                    "Furniture Store", "Home Improvement", "Sports Store", "Toy Store",
                    "Bookstore", "Gift Shop", "Supermarket", "Grocery Store", "Outlet Store"
                ],
                "weight": 0.12  # 12% of dataset
            },
            "Manufacturing": {
                "keywords": [
                    "manufacturing", "factory", "production", "industrial", "machinery",
                    "equipment", "automotive", "aerospace", "steel", "metal", "plastic",
                    "chemical", "pharmaceutical", "food processing", "textile", "fabrication",
                    "assembly", "quality control", "supply chain", "logistics"
                ],
                "business_types": [
                    "Manufacturing Company", "Industrial Equipment", "Automotive Parts",
                    "Aerospace Components", "Steel Production", "Chemical Manufacturing",
                    "Food Processing", "Textile Manufacturing", "Machinery Company",
                    "Industrial Solutions", "Production Facility", "Manufacturing Plant"
                ],
                "weight": 0.08  # 8% of dataset
            },
            "Education": {
                "keywords": [
                    "school", "university", "college", "education", "learning", "training",
                    "academy", "institute", "tutoring", "coaching", "curriculum", "student",
                    "teacher", "professor", "research", "scholarship", "degree", "certificate",
                    "online learning", "e-learning", "educational", "academic"
                ],
                "business_types": [
                    "University", "College", "School", "Academy", "Training Institute",
                    "Tutoring Center", "Educational Services", "Online Learning", "Research Institute",
                    "Educational Consulting", "Language School", "Vocational Training"
                ],
                "weight": 0.08  # 8% of dataset
            },
            "Real Estate": {
                "keywords": [
                    "real estate", "property", "housing", "apartment", "condo", "house",
                    "commercial", "residential", "development", "construction", "building",
                    "realtor", "broker", "agent", "mortgage", "rental", "leasing",
                    "property management", "land", "estate", "investment property"
                ],
                "business_types": [
                    "Real Estate Agency", "Property Management", "Real Estate Development",
                    "Commercial Real Estate", "Residential Sales", "Property Investment",
                    "Real Estate Brokerage", "Housing Development", "Property Services",
                    "Real Estate Consulting", "Land Development", "Estate Planning"
                ],
                "weight": 0.07  # 7% of dataset
            },
            "Transportation": {
                "keywords": [
                    "transportation", "shipping", "logistics", "delivery", "freight",
                    "trucking", "airline", "airport", "taxi", "uber", "lyft", "bus",
                    "railway", "train", "subway", "metro", "transit", "courier",
                    "express", "cargo", "warehouse", "distribution"
                ],
                "business_types": [
                    "Transportation Company", "Logistics Services", "Shipping Company",
                    "Trucking Company", "Delivery Service", "Courier Service", "Freight Company",
                    "Transportation Solutions", "Logistics Management", "Supply Chain",
                    "Transportation Consulting", "Fleet Management"
                ],
                "weight": 0.06  # 6% of dataset
            },
            "Energy": {
                "keywords": [
                    "energy", "power", "electric", "gas", "oil", "renewable", "solar",
                    "wind", "nuclear", "coal", "petroleum", "utilities", "grid",
                    "generation", "transmission", "distribution", "conservation", "efficiency",
                    "sustainability", "green energy", "clean energy", "alternative energy"
                ],
                "business_types": [
                    "Energy Company", "Power Generation", "Electric Utility", "Gas Company",
                    "Renewable Energy", "Solar Company", "Wind Energy", "Energy Consulting",
                    "Power Solutions", "Energy Management", "Utility Services", "Energy Trading"
                ],
                "weight": 0.05  # 5% of dataset
            },
            "Agriculture": {
                "keywords": [
                    "agriculture", "farming", "crop", "livestock", "dairy", "poultry",
                    "organic", "sustainable", "greenhouse", "irrigation", "fertilizer",
                    "seed", "grain", "produce", "food", "farm", "ranch", "vineyard",
                    "orchard", "agricultural", "agribusiness", "farming equipment"
                ],
                "business_types": [
                    "Farm", "Agricultural Company", "Dairy Farm", "Crop Production",
                    "Livestock Farm", "Organic Farm", "Greenhouse", "Agricultural Services",
                    "Farming Equipment", "Agricultural Consulting", "Food Production",
                    "Agricultural Supply"
                ],
                "weight": 0.04  # 4% of dataset
            },
            "Entertainment": {
                "keywords": [
                    "entertainment", "media", "film", "movie", "television", "tv",
                    "music", "theater", "theatre", "concert", "event", "production",
                    "studio", "broadcasting", "streaming", "gaming", "sports", "recreation",
                    "leisure", "amusement", "casino", "gambling"
                ],
                "business_types": [
                    "Entertainment Company", "Media Production", "Film Studio", "TV Production",
                    "Music Label", "Theater Company", "Event Planning", "Broadcasting",
                    "Streaming Service", "Gaming Company", "Sports Entertainment", "Casino"
                ],
                "weight": 0.05  # 5% of dataset
            },
            "Government": {
                "keywords": [
                    "government", "public", "municipal", "city", "county", "state",
                    "federal", "agency", "department", "bureau", "office", "administration",
                    "services", "public works", "utilities", "infrastructure", "policy",
                    "regulation", "compliance", "public safety", "emergency services"
                ],
                "business_types": [
                    "Government Agency", "Municipal Services", "Public Works", "City Services",
                    "County Government", "State Agency", "Federal Office", "Public Utilities",
                    "Government Consulting", "Public Administration", "Regulatory Agency",
                    "Public Safety"
                ],
                "weight": 0.03  # 3% of dataset
            },
            "Non-Profit": {
                "keywords": [
                    "foundation", "charity", "non-profit", "nonprofit", "organization",
                    "association", "society", "institute", "center", "alliance", "coalition",
                    "network", "community", "volunteer", "donation", "fundraising",
                    "advocacy", "social", "humanitarian", "environmental", "conservation"
                ],
                "business_types": [
                    "Foundation", "Charity Organization", "Non-Profit", "Community Organization",
                    "Advocacy Group", "Social Services", "Environmental Organization",
                    "Humanitarian Aid", "Volunteer Organization", "Community Center",
                    "Social Impact", "Non-Profit Consulting"
                ],
                "weight": 0.03  # 3% of dataset
            },
            "Consulting": {
                "keywords": [
                    "consulting", "advisory", "consultant", "strategy", "management",
                    "business", "professional", "services", "expertise", "specialist",
                    "analyst", "advisor", "coach", "mentor", "trainer", "facilitator",
                    "project management", "change management", "organizational", "leadership"
                ],
                "business_types": [
                    "Consulting Firm", "Management Consulting", "Business Advisory",
                    "Strategy Consulting", "Professional Services", "Business Consulting",
                    "Organizational Consulting", "Project Management", "Change Management",
                    "Leadership Consulting", "Specialized Consulting", "Industry Consulting"
                ],
                "weight": 0.06  # 6% of dataset
            },
            "Legal Services": {
                "keywords": [
                    "law", "legal", "attorney", "lawyer", "law firm", "counsel",
                    "litigation", "corporate", "criminal", "civil", "family", "immigration",
                    "patent", "trademark", "copyright", "intellectual property", "contract",
                    "compliance", "regulatory", "dispute", "mediation", "arbitration"
                ],
                "business_types": [
                    "Law Firm", "Legal Services", "Attorney Office", "Legal Consulting",
                    "Corporate Law", "Criminal Defense", "Family Law", "Immigration Law",
                    "Patent Law", "Intellectual Property", "Legal Advisory", "Litigation"
                ],
                "weight": 0.04  # 4% of dataset
            },
            "Other": {
                "keywords": [
                    "services", "solutions", "company", "corporation", "enterprise",
                    "business", "group", "holdings", "ventures", "partners", "associates",
                    "international", "global", "worldwide", "diversified", "multinational"
                ],
                "business_types": [
                    "Business Services", "General Services", "Diversified Company",
                    "Holding Company", "Investment Group", "Business Solutions",
                    "Professional Services", "General Business", "Enterprise Services",
                    "Business Group", "Service Company", "Solutions Provider"
                ],
                "weight": 0.02  # 2% of dataset
            }
        }
    
    def _define_business_patterns(self) -> Dict[str, List[str]]:
        """Define business name patterns by industry"""
        return {
            "Technology": [
                "{name} Technologies", "{name} Solutions", "{name} Systems", "{name} Labs",
                "{name} Innovations", "{name} Digital", "{name} Tech", "{name} Software",
                "{name} Data", "{name} Cloud", "{name} AI", "{name} Cyber"
            ],
            "Healthcare": [
                "{name} Medical", "{name} Health", "{name} Healthcare", "{name} Clinic",
                "{name} Medical Center", "{name} Health Services", "{name} Wellness",
                "{name} Care", "{name} Medical Group", "{name} Health System"
            ],
            "Financial Services": [
                "{name} Bank", "{name} Financial", "{name} Capital", "{name} Investment",
                "{name} Wealth", "{name} Credit", "{name} Insurance", "{name} Finance",
                "{name} Financial Services", "{name} Asset Management"
            ],
            "Retail": [
                "{name} Store", "{name} Shop", "{name} Retail", "{name} Market",
                "{name} Boutique", "{name} Outlet", "{name} Department Store",
                "{name} Supermarket", "{name} Grocery", "{name} Fashion"
            ],
            "Manufacturing": [
                "{name} Manufacturing", "{name} Industries", "{name} Production",
                "{name} Industrial", "{name} Equipment", "{name} Machinery",
                "{name} Steel", "{name} Metal", "{name} Chemical", "{name} Processing"
            ],
            "Education": [
                "{name} University", "{name} College", "{name} School", "{name} Academy",
                "{name} Institute", "{name} Learning", "{name} Education", "{name} Training",
                "{name} Educational Services", "{name} Research"
            ],
            "Real Estate": [
                "{name} Real Estate", "{name} Properties", "{name} Development",
                "{name} Realty", "{name} Property", "{name} Housing", "{name} Commercial",
                "{name} Residential", "{name} Land", "{name} Estate"
            ],
            "Transportation": [
                "{name} Transportation", "{name} Logistics", "{name} Shipping",
                "{name} Delivery", "{name} Freight", "{name} Trucking", "{name} Express",
                "{name} Courier", "{name} Transport", "{name} Logistics"
            ],
            "Energy": [
                "{name} Energy", "{name} Power", "{name} Electric", "{name} Gas",
                "{name} Oil", "{name} Renewable", "{name} Solar", "{name} Wind",
                "{name} Utilities", "{name} Power Generation"
            ],
            "Agriculture": [
                "{name} Farm", "{name} Agriculture", "{name} Farming", "{name} Crop",
                "{name} Livestock", "{name} Dairy", "{name} Organic", "{name} Produce",
                "{name} Agricultural", "{name} Agribusiness"
            ],
            "Entertainment": [
                "{name} Entertainment", "{name} Media", "{name} Production", "{name} Studio",
                "{name} Broadcasting", "{name} Streaming", "{name} Gaming", "{name} Events",
                "{name} Theater", "{name} Music"
            ],
            "Government": [
                "{name} Government", "{name} Municipal", "{name} City", "{name} County",
                "{name} State", "{name} Federal", "{name} Agency", "{name} Department",
                "{name} Public Services", "{name} Administration"
            ],
            "Non-Profit": [
                "{name} Foundation", "{name} Charity", "{name} Organization", "{name} Society",
                "{name} Institute", "{name} Center", "{name} Alliance", "{name} Network",
                "{name} Community", "{name} Services"
            ],
            "Consulting": [
                "{name} Consulting", "{name} Advisory", "{name} Services", "{name} Solutions",
                "{name} Management", "{name} Strategy", "{name} Professional", "{name} Business",
                "{name} Group", "{name} Partners"
            ],
            "Legal Services": [
                "{name} Law", "{name} Legal", "{name} Attorney", "{name} Law Firm",
                "{name} Counsel", "{name} Legal Services", "{name} Legal Group",
                "{name} Law Office", "{name} Legal Advisory", "{name} Law Partners"
            ],
            "Other": [
                "{name} Company", "{name} Corporation", "{name} Enterprise", "{name} Group",
                "{name} Holdings", "{name} Ventures", "{name} Partners", "{name} Associates",
                "{name} International", "{name} Global"
            ]
        }
    
    def _define_domain_patterns(self) -> Dict[str, List[str]]:
        """Define website domain patterns by industry"""
        return {
            "Technology": [".com", ".io", ".tech", ".ai", ".app", ".dev", ".cloud"],
            "Healthcare": [".com", ".org", ".health", ".medical", ".care", ".clinic"],
            "Financial Services": [".com", ".bank", ".finance", ".invest", ".capital"],
            "Retail": [".com", ".store", ".shop", ".market", ".boutique", ".outlet"],
            "Manufacturing": [".com", ".industrial", ".manufacturing", ".equipment"],
            "Education": [".edu", ".org", ".academy", ".school", ".university", ".institute"],
            "Real Estate": [".com", ".realty", ".properties", ".estate", ".homes"],
            "Transportation": [".com", ".logistics", ".shipping", ".transport", ".delivery"],
            "Energy": [".com", ".energy", ".power", ".electric", ".utilities"],
            "Agriculture": [".com", ".farm", ".agriculture", ".organic", ".produce"],
            "Entertainment": [".com", ".entertainment", ".media", ".studio", ".events"],
            "Government": [".gov", ".org", ".municipal", ".city", ".county"],
            "Non-Profit": [".org", ".foundation", ".charity", ".nonprofit"],
            "Consulting": [".com", ".consulting", ".advisory", ".services", ".solutions"],
            "Legal Services": [".com", ".law", ".legal", ".attorney", ".counsel"],
            "Other": [".com", ".org", ".net", ".biz", ".co"]
        }
    
    def generate_business_name(self, industry: str) -> str:
        """Generate a realistic business name for the given industry"""
        # Get industry patterns
        patterns = self.business_patterns.get(industry, self.business_patterns["Other"])
        
        # Generate base name
        base_names = [
            self.fake.company(),
            self.fake.last_name() + " " + self.fake.word().title(),
            self.fake.word().title() + " " + self.fake.word().title(),
            self.fake.city() + " " + self.fake.word().title(),
            self.fake.first_name() + " " + self.fake.word().title()
        ]
        
        base_name = random.choice(base_names)
        
        # Apply pattern
        pattern = random.choice(patterns)
        business_name = pattern.format(name=base_name)
        
        return business_name
    
    def generate_business_description(self, industry: str, business_name: str) -> str:
        """Generate a realistic business description"""
        industry_info = self.industries[industry]
        keywords = industry_info["keywords"]
        
        # Base description templates
        templates = [
            f"{business_name} is a leading provider of {random.choice(keywords)} services.",
            f"At {business_name}, we specialize in {random.choice(keywords)} and {random.choice(keywords)}.",
            f"{business_name} offers comprehensive {random.choice(keywords)} solutions for businesses and individuals.",
            f"With years of experience, {business_name} delivers exceptional {random.choice(keywords)} services.",
            f"{business_name} is committed to providing high-quality {random.choice(keywords)} and {random.choice(keywords)}.",
            f"Our team at {business_name} brings expertise in {random.choice(keywords)} and {random.choice(keywords)}.",
            f"{business_name} serves clients with innovative {random.choice(keywords)} solutions.",
            f"Trust {business_name} for professional {random.choice(keywords)} services and support."
        ]
        
        description = random.choice(templates)
        
        # Add additional details
        if random.random() < 0.7:
            additional_details = [
                f" We serve clients nationwide with cutting-edge technology and personalized service.",
                f" Our experienced team ensures the highest standards of quality and customer satisfaction.",
                f" Contact us today to learn more about our comprehensive {random.choice(keywords)} solutions.",
                f" We pride ourselves on delivering reliable, efficient, and cost-effective services.",
                f" With a focus on innovation and excellence, we help our clients achieve their goals.",
                f" Our commitment to quality and customer service sets us apart in the industry."
            ]
            description += random.choice(additional_details)
        
        return description
    
    def generate_website_url(self, business_name: str, industry: str) -> str:
        """Generate a realistic website URL"""
        # Clean business name for URL
        clean_name = re.sub(r'[^a-zA-Z0-9\s]', '', business_name.lower())
        clean_name = re.sub(r'\s+', '', clean_name)
        
        # Get domain patterns for industry
        domain_patterns = self.domain_patterns.get(industry, self.domain_patterns["Other"])
        domain = random.choice(domain_patterns)
        
        # Generate URL
        if domain.startswith('.'):
            url = f"https://www.{clean_name}{domain}"
        else:
            url = f"https://www.{clean_name}.{domain}"
        
        return url
    
    def generate_single_business(self, industry: str) -> Dict[str, str]:
        """Generate a single business record"""
        business_name = self.generate_business_name(industry)
        description = self.generate_business_description(industry, business_name)
        website_url = self.generate_website_url(business_name, industry)
        
        return {
            "business_name": business_name,
            "description": description,
            "website_url": website_url,
            "industry": industry
        }
    
    def generate_dataset(self, target_samples: int = 10000) -> pd.DataFrame:
        """Generate the complete training dataset"""
        logger.info(f"🏭 Generating {target_samples} business samples...")
        
        dataset = []
        
        # Calculate samples per industry based on weights
        industry_samples = {}
        for industry, info in self.industries.items():
            samples = int(target_samples * info["weight"])
            industry_samples[industry] = samples
            logger.info(f"   {industry}: {samples} samples")
        
        # Generate samples for each industry
        for industry, num_samples in industry_samples.items():
            logger.info(f"📊 Generating {num_samples} samples for {industry}...")
            
            for _ in range(num_samples):
                business = self.generate_single_business(industry)
                dataset.append(business)
        
        # Convert to DataFrame
        df = pd.DataFrame(dataset)
        
        # Add some data augmentation
        df = self._augment_dataset(df)
        
        # Validate and clean data
        df = self._validate_dataset(df)
        
        logger.info(f"✅ Dataset generated successfully!")
        logger.info(f"📊 Total samples: {len(df)}")
        logger.info(f"🏷️ Industries: {df['industry'].nunique()}")
        logger.info(f"📈 Industry distribution:")
        for industry, count in df['industry'].value_counts().items():
            logger.info(f"   {industry}: {count} ({count/len(df)*100:.1f}%)")
        
        return df
    
    def _augment_dataset(self, df: pd.DataFrame) -> pd.DataFrame:
        """Apply data augmentation techniques"""
        logger.info("🔄 Applying data augmentation...")
        
        augmented_data = []
        
        for _, row in df.iterrows():
            # Original data
            augmented_data.append(row.to_dict())
            
            # Augmentation 1: Add variations to business name
            if random.random() < 0.3:
                augmented_row = row.copy()
                business_name = augmented_row['business_name']
                
                # Add common variations
                variations = [
                    business_name.replace(" Inc", ""),
                    business_name.replace(" LLC", ""),
                    business_name.replace(" Corp", ""),
                    business_name.replace(" Ltd", ""),
                    business_name + " Inc",
                    business_name + " LLC",
                    business_name + " Corp"
                ]
                
                if variations:
                    augmented_row['business_name'] = random.choice(variations)
                    augmented_data.append(augmented_row.to_dict())
            
            # Augmentation 2: Add variations to description
            if random.random() < 0.2:
                augmented_row = row.copy()
                description = augmented_row['description']
                
                # Add common variations
                variations = [
                    description.replace("leading", "premier"),
                    description.replace("comprehensive", "complete"),
                    description.replace("high-quality", "top-quality"),
                    description.replace("experienced", "professional"),
                    description.replace("innovative", "cutting-edge")
                ]
                
                if variations:
                    augmented_row['description'] = random.choice(variations)
                    augmented_data.append(augmented_row.to_dict())
        
        # Convert back to DataFrame
        augmented_df = pd.DataFrame(augmented_data)
        
        # Remove duplicates
        augmented_df = augmented_df.drop_duplicates(subset=['business_name', 'industry'])
        
        logger.info(f"📈 Dataset augmented: {len(df)} -> {len(augmented_df)} samples")
        return augmented_df
    
    def _validate_dataset(self, df: pd.DataFrame) -> pd.DataFrame:
        """Validate and clean the dataset"""
        logger.info("🔍 Validating and cleaning dataset...")
        
        initial_count = len(df)
        
        # Remove rows with missing required fields
        df = df.dropna(subset=['business_name', 'industry'])
        
        # Remove very short business names
        df = df[df['business_name'].str.len() >= 3]
        
        # Remove very short descriptions
        df = df[df['description'].str.len() >= 20]
        
        # Validate website URLs
        df = df[df['website_url'].str.contains(r'^https?://', regex=True)]
        
        # Remove duplicate business names within same industry
        df = df.drop_duplicates(subset=['business_name', 'industry'])
        
        # Ensure minimum samples per industry
        min_samples_per_industry = 50
        industry_counts = df['industry'].value_counts()
        valid_industries = industry_counts[industry_counts >= min_samples_per_industry].index
        df = df[df['industry'].isin(valid_industries)]
        
        final_count = len(df)
        logger.info(f"🧹 Dataset cleaned: {initial_count} -> {final_count} samples")
        
        return df
    
    def save_dataset(self, df: pd.DataFrame, filename: str = None) -> str:
        """Save dataset to file"""
        if filename is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            filename = f"business_classification_dataset_{timestamp}.csv"
        
        filepath = self.data_path / filename
        
        # Save as CSV
        df.to_csv(filepath, index=False)
        
        # Save metadata
        metadata = {
            "filename": filename,
            "total_samples": len(df),
            "industries": df['industry'].nunique(),
            "industry_distribution": df['industry'].value_counts().to_dict(),
            "generated_at": datetime.now().isoformat(),
            "config": self.config
        }
        
        metadata_path = self.data_path / f"{filename.replace('.csv', '_metadata.json')}"
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2)
        
        logger.info(f"💾 Dataset saved to: {filepath}")
        logger.info(f"📝 Metadata saved to: {metadata_path}")
        
        return str(filepath)
    
    def create_sample_dataset(self, num_samples: int = 1000) -> pd.DataFrame:
        """Create a smaller sample dataset for testing"""
        logger.info(f"🧪 Creating sample dataset with {num_samples} samples...")
        
        # Temporarily adjust weights for smaller dataset
        original_weights = {industry: info["weight"] for industry, info in self.industries.items()}
        
        # Generate sample dataset
        sample_df = self.generate_dataset(num_samples)
        
        # Restore original weights
        for industry, weight in original_weights.items():
            self.industries[industry]["weight"] = weight
        
        return sample_df

def main():
    """Main function to generate business classification dataset"""
    
    # Configuration
    config = {
        'data_path': 'data',
        'target_samples': 10000,
        'min_samples_per_industry': 50,
        'augmentation_enabled': True,
        'validation_enabled': True
    }
    
    # Initialize generator
    generator = BusinessDatasetGenerator(config)
    
    # Generate dataset
    df = generator.generate_dataset(config['target_samples'])
    
    # Save dataset
    filepath = generator.save_dataset(df)
    
    # Create sample dataset for testing
    sample_df = generator.create_sample_dataset(1000)
    sample_filepath = generator.save_dataset(sample_df, "business_classification_sample.csv")
    
    logger.info("🎉 Business classification dataset generation completed!")
    logger.info(f"📊 Main dataset: {filepath}")
    logger.info(f"🧪 Sample dataset: {sample_filepath}")
    logger.info(f"📈 Total samples: {len(df)}")
    logger.info(f"🏷️ Industries: {df['industry'].nunique()}")

if __name__ == "__main__":
    main()
