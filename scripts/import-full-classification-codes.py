#!/usr/bin/env python3
"""
Full Classification Codes Import Script

This script imports all classification codes from CSV files into the Supabase database.
It handles MCC, NAICS, and SIC codes with proper industry mapping.

Usage:
    python import-full-classification-codes.py

Environment Variables Required:
    SUPABASE_URL              Your Supabase project URL
    SUPABASE_SERVICE_ROLE_KEY Your Supabase service role key
"""

import os
import sys
import csv
import json
import logging
import requests
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional
from pathlib import Path

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class ClassificationCodeImporter:
    """Imports classification codes from CSV files to Supabase."""
    
    def __init__(self, supabase_url: str, service_role_key: str):
        """Initialize the importer."""
        self.supabase_url = supabase_url
        self.service_role_key = service_role_key
        self.headers = {
            "apikey": service_role_key,
            "Authorization": f"Bearer {service_role_key}",
            "Content-Type": "application/json",
            "Prefer": "return=representation"
        }
        
        # Industry mapping for classification codes
        self.industry_mapping = {
            # Technology
            "technology": {
                "keywords": ["software", "computer", "technology", "tech", "digital", "IT", "information", "data", "cloud", "AI", "artificial intelligence"],
                "industry_id": 1
            },
            # Financial Services
            "financial": {
                "keywords": ["bank", "financial", "insurance", "credit", "loan", "investment", "finance", "money", "payment", "trading"],
                "industry_id": 2
            },
            # Healthcare
            "healthcare": {
                "keywords": ["medical", "health", "hospital", "doctor", "clinic", "pharmacy", "healthcare", "medicine", "dental", "veterinary"],
                "industry_id": 3
            },
            # Manufacturing
            "manufacturing": {
                "keywords": ["manufacturing", "production", "factory", "industrial", "machinery", "equipment", "assembly", "fabrication"],
                "industry_id": 4
            },
            # Retail
            "retail": {
                "keywords": ["retail", "store", "shop", "merchandise", "sales", "commerce", "trade", "wholesale", "distribution"],
                "industry_id": 5
            },
            # General Business (fallback)
            "general": {
                "keywords": ["business", "service", "consulting", "professional", "management", "administration"],
                "industry_id": 6
            }
        }
    
    def make_api_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
        """Make API request to Supabase."""
        url = f"{self.supabase_url}/rest/v1/{endpoint}"
        
        try:
            if method.upper() == "GET":
                response = requests.get(url, headers=self.headers)
            elif method.upper() == "POST":
                response = requests.post(url, headers=self.headers, json=data)
            elif method.upper() == "DELETE":
                response = requests.delete(url, headers=self.headers)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
            
            response.raise_for_status()
            return response.json() if response.content else {}
            
        except requests.exceptions.RequestException as e:
            logger.error(f"API request failed: {e}")
            raise
    
    def get_industry_id_for_code(self, code: str, description: str, code_type: str) -> int:
        """Determine the appropriate industry ID for a classification code."""
        text_to_analyze = f"{code} {description}".lower()
        
        # Score each industry based on keyword matches
        industry_scores = {}
        for industry_name, industry_data in self.industry_mapping.items():
            score = 0
            for keyword in industry_data["keywords"]:
                if keyword.lower() in text_to_analyze:
                    score += 1
            industry_scores[industry_name] = score
        
        # Find the industry with the highest score
        best_industry = max(industry_scores, key=industry_scores.get)
        
        # If no keywords matched, use general business
        if industry_scores[best_industry] == 0:
            best_industry = "general"
        
        return self.industry_mapping[best_industry]["industry_id"]
    
    def import_mcc_codes(self, csv_file: str) -> int:
        """Import MCC codes from CSV file."""
        logger.info(f"Importing MCC codes from {csv_file}")
        
        codes_to_import = []
        with open(csv_file, 'r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                mcc_code = row.get('MCC', '').strip()
                description = row.get('Description', '').strip()
                
                if mcc_code and description:
                    industry_id = self.get_industry_id_for_code(mcc_code, description, 'MCC')
                    
                    codes_to_import.append({
                        "code": mcc_code,
                        "description": description,
                        "code_type": "MCC",
                        "industry_id": industry_id,
                        "is_active": True,
                        "created_at": datetime.now(timezone.utc).isoformat(),
                        "updated_at": datetime.now(timezone.utc).isoformat()
                    })
        
        # Import in batches
        batch_size = 100
        total_imported = 0
        
        for i in range(0, len(codes_to_import), batch_size):
            batch = codes_to_import[i:i + batch_size]
            try:
                result = self.make_api_request("POST", "classification_codes", batch)
                total_imported += len(batch)
                logger.info(f"Imported MCC batch {i//batch_size + 1}: {len(batch)} codes")
            except Exception as e:
                logger.error(f"Failed to import MCC batch {i//batch_size + 1}: {e}")
        
        logger.info(f"Successfully imported {total_imported} MCC codes")
        return total_imported
    
    def import_naics_codes(self, csv_file: str) -> int:
        """Import NAICS codes from CSV file."""
        logger.info(f"Importing NAICS codes from {csv_file}")
        
        codes_to_import = []
        with open(csv_file, 'r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                naics_code = row.get('2022 NAICS Code', '').strip()
                description = row.get('2022 NAICS Title', '').strip()
                
                if naics_code and description and naics_code != '':
                    industry_id = self.get_industry_id_for_code(naics_code, description, 'NAICS')
                    
                    codes_to_import.append({
                        "code": naics_code,
                        "description": description,
                        "code_type": "NAICS",
                        "industry_id": industry_id,
                        "is_active": True,
                        "created_at": datetime.now(timezone.utc).isoformat(),
                        "updated_at": datetime.now(timezone.utc).isoformat()
                    })
        
        # Import in batches
        batch_size = 100
        total_imported = 0
        
        for i in range(0, len(codes_to_import), batch_size):
            batch = codes_to_import[i:i + batch_size]
            try:
                result = self.make_api_request("POST", "classification_codes", batch)
                total_imported += len(batch)
                logger.info(f"Imported NAICS batch {i//batch_size + 1}: {len(batch)} codes")
            except Exception as e:
                logger.error(f"Failed to import NAICS batch {i//batch_size + 1}: {e}")
        
        logger.info(f"Successfully imported {total_imported} NAICS codes")
        return total_imported
    
    def import_sic_codes(self, csv_file: str) -> int:
        """Import SIC codes from CSV file."""
        logger.info(f"Importing SIC codes from {csv_file}")
        
        codes_to_import = []
        with open(csv_file, 'r', encoding='utf-8') as file:
            reader = csv.DictReader(file)
            for row in reader:
                sic_code = row.get('SIC', '').strip()
                description = row.get('Description', '').strip()
                
                if sic_code and description:
                    industry_id = self.get_industry_id_for_code(sic_code, description, 'SIC')
                    
                    codes_to_import.append({
                        "code": sic_code,
                        "description": description,
                        "code_type": "SIC",
                        "industry_id": industry_id,
                        "is_active": True,
                        "created_at": datetime.now(timezone.utc).isoformat(),
                        "updated_at": datetime.now(timezone.utc).isoformat()
                    })
        
        # Import in batches
        batch_size = 100
        total_imported = 0
        
        for i in range(0, len(codes_to_import), batch_size):
            batch = codes_to_import[i:i + batch_size]
            try:
                result = self.make_api_request("POST", "classification_codes", batch)
                total_imported += len(batch)
                logger.info(f"Imported SIC batch {i//batch_size + 1}: {len(batch)} codes")
            except Exception as e:
                logger.error(f"Failed to import SIC batch {i//batch_size + 1}: {e}")
        
        logger.info(f"Successfully imported {total_imported} SIC codes")
        return total_imported
    
    def clear_existing_codes(self) -> None:
        """Clear existing classification codes (except sample data)."""
        logger.info("Clearing existing classification codes...")
        
        try:
            # Delete all existing codes
            result = self.make_api_request("DELETE", "classification_codes?is_active=eq.true")
            logger.info("Cleared existing classification codes")
        except Exception as e:
            logger.error(f"Failed to clear existing codes: {e}")
            raise
    
    def get_current_code_count(self) -> int:
        """Get current count of classification codes."""
        try:
            result = self.make_api_request("GET", "classification_codes?select=count")
            return result[0]["count"] if result else 0
        except Exception as e:
            logger.error(f"Failed to get code count: {e}")
            return 0
    
    def import_all_codes(self) -> Dict[str, int]:
        """Import all classification codes from CSV files."""
        logger.info("Starting full classification codes import...")
        
        # Get current count
        initial_count = self.get_current_code_count()
        logger.info(f"Initial classification codes count: {initial_count}")
        
        # Clear existing codes
        self.clear_existing_codes()
        
        # Import codes from each CSV file
        results = {}
        
        # Import MCC codes
        mcc_file = "codes/mcc_codes.csv"
        if os.path.exists(mcc_file):
            results["MCC"] = self.import_mcc_codes(mcc_file)
        else:
            logger.error(f"MCC codes file not found: {mcc_file}")
            results["MCC"] = 0
        
        # Import NAICS codes
        naics_file = "codes/NAICS-2022-Codes_industries.csv"
        if os.path.exists(naics_file):
            results["NAICS"] = self.import_naics_codes(naics_file)
        else:
            logger.error(f"NAICS codes file not found: {naics_file}")
            results["NAICS"] = 0
        
        # Import SIC codes
        sic_file = "codes/sic-codes.csv"
        if os.path.exists(sic_file):
            results["SIC"] = self.import_sic_codes(sic_file)
        else:
            logger.error(f"SIC codes file not found: {sic_file}")
            results["SIC"] = 0
        
        # Get final count
        final_count = self.get_current_code_count()
        total_imported = sum(results.values())
        
        logger.info(f"Import completed!")
        logger.info(f"Total codes imported: {total_imported}")
        logger.info(f"Final classification codes count: {final_count}")
        logger.info(f"Breakdown: MCC={results['MCC']}, NAICS={results['NAICS']}, SIC={results['SIC']}")
        
        return results

def main():
    """Main function."""
    # Check environment variables
    supabase_url = os.getenv("SUPABASE_URL")
    service_role_key = os.getenv("SUPABASE_SERVICE_ROLE_KEY")
    
    if not supabase_url:
        logger.error("SUPABASE_URL environment variable is not set")
        sys.exit(1)
    
    if not service_role_key:
        logger.error("SUPABASE_SERVICE_ROLE_KEY environment variable is not set")
        sys.exit(1)
    
    # Create importer and run import
    importer = ClassificationCodeImporter(supabase_url, service_role_key)
    
    try:
        results = importer.import_all_codes()
        
        # Print summary
        print("\n" + "="*60)
        print("CLASSIFICATION CODES IMPORT SUMMARY")
        print("="*60)
        print(f"MCC Codes Imported:  {results['MCC']:,}")
        print(f"NAICS Codes Imported: {results['NAICS']:,}")
        print(f"SIC Codes Imported:   {results['SIC']:,}")
        print(f"Total Codes Imported: {sum(results.values()):,}")
        print("="*60)
        
    except Exception as e:
        logger.error(f"Import failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
