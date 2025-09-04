#!/usr/bin/env python3
"""
Supabase Bulk Operations Script

This script provides advanced bulk import/export functionality for keyword management
using Supabase's REST API with Python for more complex operations.

Usage:
    python supabase_bulk_operations.py [command] [options]

Commands:
    export-all          Export all data to structured format
    import-all          Import all data from structured format
    sync-data           Sync data between environments
    validate-data       Validate data integrity
    generate-sample     Generate sample data for testing
    migrate-data        Migrate data between schemas
"""

import os
import sys
import json
import csv
import argparse
import logging
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional
from pathlib import Path

import requests
import pandas as pd
from supabase import create_client, Client

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class SupabaseBulkOperations:
    """Handles bulk operations with Supabase for keyword management."""
    
    def __init__(self, supabase_url: str, service_role_key: str):
        """Initialize the Supabase client."""
        self.supabase_url = supabase_url
        self.service_role_key = service_role_key
        self.client: Client = create_client(supabase_url, service_role_key)
        
        # Headers for direct API calls
        self.headers = {
            'apikey': service_role_key,
            'Authorization': f'Bearer {service_role_key}',
            'Content-Type': 'application/json',
            'Prefer': 'return=representation'
        }
    
    def _make_request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict:
        """Make a request to the Supabase API."""
        url = f"{self.supabase_url}/rest/v1/{endpoint}"
        
        try:
            if method.upper() == 'GET':
                response = requests.get(url, headers=self.headers)
            elif method.upper() == 'POST':
                response = requests.post(url, headers=self.headers, json=data)
            elif method.upper() == 'PUT':
                response = requests.put(url, headers=self.headers, json=data)
            elif method.upper() == 'DELETE':
                response = requests.delete(url, headers=self.headers)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
            
            response.raise_for_status()
            return response.json() if response.content else {}
            
        except requests.exceptions.RequestException as e:
            logger.error(f"API request failed: {e}")
            raise
    
    def export_all_data(self, output_dir: str) -> Dict[str, str]:
        """Export all data to structured format."""
        logger.info("Starting full data export...")
        
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        export_info = {
            'timestamp': timestamp,
            'export_date': datetime.now(timezone.utc).isoformat(),
            'files': {}
        }
        
        # Export industries
        logger.info("Exporting industries...")
        industries = self._make_request('GET', 'industries?select=*')
        industries_file = output_path / f"industries_{timestamp}.json"
        with open(industries_file, 'w') as f:
            json.dump(industries, f, indent=2, default=str)
        export_info['files']['industries'] = str(industries_file)
        
        # Export keywords
        logger.info("Exporting keywords...")
        keywords = self._make_request('GET', 'industry_keywords?select=*,industries(name)')
        keywords_file = output_path / f"keywords_{timestamp}.json"
        with open(keywords_file, 'w') as f:
            json.dump(keywords, f, indent=2, default=str)
        export_info['files']['keywords'] = str(keywords_file)
        
        # Export classification codes
        logger.info("Exporting classification codes...")
        codes = self._make_request('GET', 'classification_codes?select=*,industries(name)')
        codes_file = output_path / f"codes_{timestamp}.json"
        with open(codes_file, 'w') as f:
            json.dump(codes, f, indent=2, default=str)
        export_info['files']['codes'] = str(codes_file)
        
        # Export code keywords
        logger.info("Exporting code keywords...")
        code_keywords = self._make_request('GET', 'code_keywords?select=*,classification_codes(code,code_type)')
        code_keywords_file = output_path / f"code_keywords_{timestamp}.json"
        with open(code_keywords_file, 'w') as f:
            json.dump(code_keywords, f, indent=2, default=str)
        export_info['files']['code_keywords'] = str(code_keywords_file)
        
        # Export classification history (last 30 days)
        logger.info("Exporting classification history...")
        history = self._make_request('GET', 'classification_history?select=*&created_at=gte.2024-12-20')
        history_file = output_path / f"history_{timestamp}.json"
        with open(history_file, 'w') as f:
            json.dump(history, f, indent=2, default=str)
        export_info['files']['history'] = str(history_file)
        
        # Create export manifest
        manifest_file = output_path / f"export_manifest_{timestamp}.json"
        with open(manifest_file, 'w') as f:
            json.dump(export_info, f, indent=2)
        
        logger.info(f"Export completed. Manifest: {manifest_file}")
        return export_info
    
    def import_all_data(self, import_dir: str, validate: bool = True) -> Dict[str, int]:
        """Import all data from structured format."""
        logger.info(f"Starting full data import from: {import_dir}")
        
        import_path = Path(import_dir)
        if not import_path.exists():
            raise FileNotFoundError(f"Import directory not found: {import_dir}")
        
        # Find manifest file
        manifest_files = list(import_path.glob("export_manifest_*.json"))
        if not manifest_files:
            raise FileNotFoundError("No export manifest found in import directory")
        
        manifest_file = manifest_files[0]  # Use the first manifest found
        with open(manifest_file, 'r') as f:
            manifest = json.load(f)
        
        logger.info(f"Using manifest: {manifest_file}")
        logger.info(f"Export date: {manifest.get('export_date', 'Unknown')}")
        
        import_stats = {}
        
        # Import industries
        if 'industries' in manifest['files']:
            industries_file = Path(manifest['files']['industries'])
            if industries_file.exists():
                logger.info("Importing industries...")
                with open(industries_file, 'r') as f:
                    industries = json.load(f)
                
                if validate:
                    self._validate_industries(industries)
                
                # Clear existing data (optional - comment out for append mode)
                # self._make_request('DELETE', 'industries')
                
                # Import new data
                result = self._make_request('POST', 'industries', industries)
                import_stats['industries'] = len(result) if isinstance(result, list) else 1
                logger.info(f"Imported {import_stats['industries']} industries")
        
        # Import keywords
        if 'keywords' in manifest['files']:
            keywords_file = Path(manifest['files']['keywords'])
            if keywords_file.exists():
                logger.info("Importing keywords...")
                with open(keywords_file, 'r') as f:
                    keywords = json.load(f)
                
                if validate:
                    self._validate_keywords(keywords)
                
                # Clear existing data (optional)
                # self._make_request('DELETE', 'industry_keywords')
                
                # Import new data
                result = self._make_request('POST', 'industry_keywords', keywords)
                import_stats['keywords'] = len(result) if isinstance(result, list) else 1
                logger.info(f"Imported {import_stats['keywords']} keywords")
        
        # Import classification codes
        if 'codes' in manifest['files']:
            codes_file = Path(manifest['files']['codes'])
            if codes_file.exists():
                logger.info("Importing classification codes...")
                with open(codes_file, 'r') as f:
                    codes = json.load(f)
                
                if validate:
                    self._validate_codes(codes)
                
                # Clear existing data (optional)
                # self._make_request('DELETE', 'classification_codes')
                
                # Import new data
                result = self._make_request('POST', 'classification_codes', codes)
                import_stats['codes'] = len(result) if isinstance(result, list) else 1
                logger.info(f"Imported {import_stats['codes']} classification codes")
        
        # Import code keywords
        if 'code_keywords' in manifest['files']:
            code_keywords_file = Path(manifest['files']['code_keywords'])
            if code_keywords_file.exists():
                logger.info("Importing code keywords...")
                with open(code_keywords_file, 'r') as f:
                    code_keywords = json.load(f)
                
                if validate:
                    self._validate_code_keywords(code_keywords)
                
                # Clear existing data (optional)
                # self._make_request('DELETE', 'code_keywords')
                
                # Import new data
                result = self._make_request('POST', 'code_keywords', code_keywords)
                import_stats['code_keywords'] = len(result) if isinstance(result, list) else 1
                logger.info(f"Imported {import_stats['code_keywords']} code keywords")
        
        logger.info(f"Import completed. Stats: {import_stats}")
        return import_stats
    
    def sync_data(self, source_url: str, source_key: str, tables: List[str] = None) -> Dict[str, int]:
        """Sync data between Supabase environments."""
        logger.info(f"Syncing data from {source_url} to {self.supabase_url}")
        
        if tables is None:
            tables = ['industries', 'industry_keywords', 'classification_codes', 'code_keywords']
        
        # Create source client
        source_client = create_client(source_url, source_key)
        source_headers = {
            'apikey': source_key,
            'Authorization': f'Bearer {source_key}',
            'Content-Type': 'application/json'
        }
        
        sync_stats = {}
        
        for table in tables:
            logger.info(f"Syncing table: {table}")
            
            # Get data from source
            source_url_endpoint = f"{source_url}/rest/v1/{table}?select=*"
            response = requests.get(source_url_endpoint, headers=source_headers)
            response.raise_for_status()
            source_data = response.json()
            
            # Clear target data
            self._make_request('DELETE', table)
            
            # Import to target
            if source_data:
                result = self._make_request('POST', table, source_data)
                sync_stats[table] = len(result) if isinstance(result, list) else 1
                logger.info(f"Synced {sync_stats[table]} records for {table}")
            else:
                sync_stats[table] = 0
                logger.info(f"No data to sync for {table}")
        
        logger.info(f"Sync completed. Stats: {sync_stats}")
        return sync_stats
    
    def validate_data(self) -> Dict[str, List[str]]:
        """Validate data integrity across all tables."""
        logger.info("Starting data validation...")
        
        validation_results = {
            'errors': [],
            'warnings': [],
            'info': []
        }
        
        # Validate industries
        industries = self._make_request('GET', 'industries?select=*')
        validation_results['info'].append(f"Found {len(industries)} industries")
        
        # Check for duplicate industry names
        industry_names = [i['name'] for i in industries]
        duplicates = [name for name in set(industry_names) if industry_names.count(name) > 1]
        if duplicates:
            validation_results['errors'].append(f"Duplicate industry names: {duplicates}")
        
        # Validate keywords
        keywords = self._make_request('GET', 'industry_keywords?select=*')
        validation_results['info'].append(f"Found {len(keywords)} keywords")
        
        # Check for orphaned keywords
        industry_ids = {i['id'] for i in industries}
        orphaned_keywords = [k for k in keywords if k['industry_id'] not in industry_ids]
        if orphaned_keywords:
            validation_results['errors'].append(f"Found {len(orphaned_keywords)} orphaned keywords")
        
        # Check for invalid weights
        invalid_weights = [k for k in keywords if k['weight'] < 0 or k['weight'] > 1]
        if invalid_weights:
            validation_results['warnings'].append(f"Found {len(invalid_weights)} keywords with invalid weights")
        
        # Validate classification codes
        codes = self._make_request('GET', 'classification_codes?select=*')
        validation_results['info'].append(f"Found {len(codes)} classification codes")
        
        # Check for orphaned codes
        orphaned_codes = [c for c in codes if c['industry_id'] not in industry_ids]
        if orphaned_codes:
            validation_results['errors'].append(f"Found {len(orphaned_codes)} orphaned classification codes")
        
        # Check for duplicate codes
        code_identifiers = [(c['code'], c['code_type']) for c in codes]
        duplicate_codes = [code for code in set(code_identifiers) if code_identifiers.count(code) > 1]
        if duplicate_codes:
            validation_results['errors'].append(f"Duplicate classification codes: {duplicate_codes}")
        
        # Print results
        for level, messages in validation_results.items():
            if messages:
                logger.info(f"{level.upper()}:")
                for message in messages:
                    logger.info(f"  - {message}")
        
        return validation_results
    
    def generate_sample_data(self, output_dir: str) -> Dict[str, str]:
        """Generate sample data for testing."""
        logger.info("Generating sample data...")
        
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        # Sample industries
        sample_industries = [
            {
                "name": "Technology",
                "description": "Software development and technology services",
                "category": "Technology",
                "is_active": True,
                "created_at": datetime.now(timezone.utc).isoformat(),
                "updated_at": datetime.now(timezone.utc).isoformat()
            },
            {
                "name": "Healthcare",
                "description": "Medical and healthcare services",
                "category": "Healthcare",
                "is_active": True,
                "created_at": datetime.now(timezone.utc).isoformat(),
                "updated_at": datetime.now(timezone.utc).isoformat()
            },
            {
                "name": "Finance",
                "description": "Financial services and banking",
                "category": "Finance",
                "is_active": True,
                "created_at": datetime.now(timezone.utc).isoformat(),
                "updated_at": datetime.now(timezone.utc).isoformat()
            }
        ]
        
        # Sample keywords
        sample_keywords = [
            {"industry_id": 1, "keyword": "software", "weight": 0.9, "keyword_type": "primary", "is_active": True},
            {"industry_id": 1, "keyword": "technology", "weight": 0.8, "keyword_type": "primary", "is_active": True},
            {"industry_id": 1, "keyword": "development", "weight": 0.7, "keyword_type": "secondary", "is_active": True},
            {"industry_id": 2, "keyword": "medical", "weight": 0.9, "keyword_type": "primary", "is_active": True},
            {"industry_id": 2, "keyword": "healthcare", "weight": 0.8, "keyword_type": "primary", "is_active": True},
            {"industry_id": 2, "keyword": "patient", "weight": 0.6, "keyword_type": "secondary", "is_active": True},
            {"industry_id": 3, "keyword": "banking", "weight": 0.9, "keyword_type": "primary", "is_active": True},
            {"industry_id": 3, "keyword": "finance", "weight": 0.8, "keyword_type": "primary", "is_active": True},
            {"industry_id": 3, "keyword": "investment", "weight": 0.7, "keyword_type": "secondary", "is_active": True}
        ]
        
        # Sample classification codes
        sample_codes = [
            {"code": "541511", "description": "Custom Computer Programming Services", "code_type": "NAICS", "industry_id": 1, "is_active": True},
            {"code": "541512", "description": "Computer Systems Design Services", "code_type": "NAICS", "industry_id": 1, "is_active": True},
            {"code": "621111", "description": "Offices of Physicians", "code_type": "NAICS", "industry_id": 2, "is_active": True},
            {"code": "621112", "description": "Offices of Physicians, Mental Health Specialists", "code_type": "NAICS", "industry_id": 2, "is_active": True},
            {"code": "522110", "description": "Commercial Banking", "code_type": "NAICS", "industry_id": 3, "is_active": True},
            {"code": "522120", "description": "Savings Institutions", "code_type": "NAICS", "industry_id": 3, "is_active": True}
        ]
        
        # Save sample data
        files = {}
        
        industries_file = output_path / "sample_industries.json"
        with open(industries_file, 'w') as f:
            json.dump(sample_industries, f, indent=2)
        files['industries'] = str(industries_file)
        
        keywords_file = output_path / "sample_keywords.json"
        with open(keywords_file, 'w') as f:
            json.dump(sample_keywords, f, indent=2)
        files['keywords'] = str(keywords_file)
        
        codes_file = output_path / "sample_codes.json"
        with open(codes_file, 'w') as f:
            json.dump(sample_codes, f, indent=2)
        files['codes'] = str(codes_file)
        
        logger.info(f"Sample data generated in: {output_path}")
        return files
    
    def _validate_industries(self, industries: List[Dict]) -> None:
        """Validate industries data."""
        required_fields = ['name', 'description', 'category', 'is_active']
        for i, industry in enumerate(industries):
            for field in required_fields:
                if field not in industry:
                    raise ValueError(f"Industry {i}: Missing required field '{field}'")
    
    def _validate_keywords(self, keywords: List[Dict]) -> None:
        """Validate keywords data."""
        required_fields = ['industry_id', 'keyword', 'weight', 'keyword_type', 'is_active']
        for i, keyword in enumerate(keywords):
            for field in required_fields:
                if field not in keyword:
                    raise ValueError(f"Keyword {i}: Missing required field '{field}'")
            
            if not (0 <= keyword['weight'] <= 1):
                raise ValueError(f"Keyword {i}: Weight must be between 0 and 1")
    
    def _validate_codes(self, codes: List[Dict]) -> None:
        """Validate classification codes data."""
        required_fields = ['code', 'description', 'code_type', 'industry_id', 'is_active']
        for i, code in enumerate(codes):
            for field in required_fields:
                if field not in code:
                    raise ValueError(f"Code {i}: Missing required field '{field}'")
            
            if code['code_type'] not in ['NAICS', 'MCC', 'SIC']:
                raise ValueError(f"Code {i}: Invalid code_type '{code['code_type']}'")
    
    def _validate_code_keywords(self, code_keywords: List[Dict]) -> None:
        """Validate code keywords data."""
        required_fields = ['code_id', 'keyword', 'weight', 'is_active']
        for i, code_keyword in enumerate(code_keywords):
            for field in required_fields:
                if field not in code_keyword:
                    raise ValueError(f"Code keyword {i}: Missing required field '{field}'")
            
            if not (0 <= code_keyword['weight'] <= 1):
                raise ValueError(f"Code keyword {i}: Weight must be between 0 and 1")


def main():
    """Main function to handle command line arguments."""
    parser = argparse.ArgumentParser(description='Supabase Bulk Operations Script')
    parser.add_argument('command', choices=[
        'export-all', 'import-all', 'sync-data', 'validate-data', 
        'generate-sample', 'migrate-data'
    ], help='Command to execute')
    parser.add_argument('--input-dir', help='Input directory for import operations')
    parser.add_argument('--output-dir', help='Output directory for export operations')
    parser.add_argument('--source-url', help='Source URL for sync operations')
    parser.add_argument('--source-key', help='Source API key for sync operations')
    parser.add_argument('--tables', nargs='+', help='Tables to sync (for sync-data command)')
    parser.add_argument('--no-validate', action='store_true', help='Skip validation during import')
    parser.add_argument('--verbose', '-v', action='store_true', help='Enable verbose logging')
    
    args = parser.parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    # Get environment variables
    supabase_url = os.getenv('SUPABASE_URL')
    service_role_key = os.getenv('SUPABASE_SERVICE_ROLE_KEY')
    
    if not supabase_url or not service_role_key:
        logger.error("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY environment variables are required")
        sys.exit(1)
    
    # Initialize operations
    ops = SupabaseBulkOperations(supabase_url, service_role_key)
    
    try:
        if args.command == 'export-all':
            if not args.output_dir:
                args.output_dir = f"exports/export_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            result = ops.export_all_data(args.output_dir)
            logger.info(f"Export completed: {result}")
            
        elif args.command == 'import-all':
            if not args.input_dir:
                logger.error("--input-dir is required for import-all command")
                sys.exit(1)
            result = ops.import_all_data(args.input_dir, validate=not args.no_validate)
            logger.info(f"Import completed: {result}")
            
        elif args.command == 'sync-data':
            if not args.source_url or not args.source_key:
                logger.error("--source-url and --source-key are required for sync-data command")
                sys.exit(1)
            result = ops.sync_data(args.source_url, args.source_key, args.tables)
            logger.info(f"Sync completed: {result}")
            
        elif args.command == 'validate-data':
            result = ops.validate_data()
            if result['errors']:
                logger.error("Data validation failed")
                sys.exit(1)
            else:
                logger.info("Data validation passed")
                
        elif args.command == 'generate-sample':
            if not args.output_dir:
                args.output_dir = f"samples/sample_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            result = ops.generate_sample_data(args.output_dir)
            logger.info(f"Sample data generated: {result}")
            
        else:
            logger.error(f"Command '{args.command}' not implemented yet")
            sys.exit(1)
            
    except Exception as e:
        logger.error(f"Operation failed: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()
