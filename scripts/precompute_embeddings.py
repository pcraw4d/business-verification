#!/usr/bin/env python3
"""
Pre-compute embeddings for all classification codes.
Run once after enabling pgvector.
"""

import os
import sys
from typing import List, Dict
from sentence_transformers import SentenceTransformer
from supabase import create_client, Client
from tqdm import tqdm
import time

# Configuration
MODEL_NAME = 'sentence-transformers/all-MiniLM-L6-v2'
BATCH_SIZE = 50
SUPABASE_URL = os.getenv('SUPABASE_URL')
SUPABASE_KEY = os.getenv('SUPABASE_SERVICE_KEY')  # Use service key for admin access

def main():
    print("=" * 60)
    print("CODE EMBEDDINGS PRE-COMPUTATION")
    print("=" * 60)
    
    # Validate environment
    if not SUPABASE_URL or not SUPABASE_KEY:
        print("❌ Error: SUPABASE_URL and SUPABASE_SERVICE_KEY must be set")
        sys.exit(1)
    
    # Initialize Supabase client
    print("\n1. Connecting to Supabase...")
    supabase: Client = create_client(SUPABASE_URL, SUPABASE_KEY)
    print("✅ Connected")
    
    # Initialize embedding model
    print(f"\n2. Loading embedding model: {MODEL_NAME}...")
    model = SentenceTransformer(MODEL_NAME)
    print(f"✅ Model loaded (embedding dimension: {model.get_sentence_embedding_dimension()})")
    
    # Fetch all codes
    print("\n3. Fetching classification codes from database...")
    codes = fetch_all_codes(supabase)
    print(f"✅ Fetched {len(codes)} codes")
    print(f"   - MCC: {sum(1 for c in codes if c['code_type'] == 'MCC')}")
    print(f"   - SIC: {sum(1 for c in codes if c['code_type'] == 'SIC')}")
    print(f"   - NAICS: {sum(1 for c in codes if c['code_type'] == 'NAICS')}")
    
    # Enrich descriptions
    print("\n4. Enriching code descriptions with context...")
    enriched_codes = enrich_codes_with_context(codes, supabase)
    print("✅ Descriptions enriched")
    
    # Generate embeddings
    print(f"\n5. Generating embeddings (batch size: {BATCH_SIZE})...")
    embeddings_data = generate_embeddings_batch(enriched_codes, model, BATCH_SIZE)
    print(f"✅ Generated {len(embeddings_data)} embeddings")
    
    # Insert into database
    print("\n6. Inserting embeddings into database...")
    insert_embeddings(embeddings_data, supabase, BATCH_SIZE)
    print("✅ All embeddings inserted")
    
    # Verify
    print("\n7. Verifying insertion...")
    verify_embeddings(supabase)
    
    print("\n" + "=" * 60)
    print("✅ EMBEDDINGS PRE-COMPUTATION COMPLETE!")
    print("=" * 60)

def fetch_all_codes(supabase: Client) -> List[Dict]:
    """Fetch all classification codes from database."""
    response = supabase.table('classification_codes').select('*').execute()
    return response.data

def enrich_codes_with_context(codes: List[Dict], supabase: Client) -> List[Dict]:
    """Enrich code descriptions with keywords and additional context."""
    enriched = []
    
    for code in tqdm(codes, desc="Enriching"):
        code_type = code['code_type']
        code_value = code['code']
        code_id = code['id']
        description = code['description']
        
        # Build extended description for better embeddings
        extended_parts = [description]
        
        # Add industry context if available
        if code.get('industry_id'):
            # Try to fetch industry name
            try:
                industry_response = supabase.table('industries').select('name').eq('id', code['industry_id']).single().execute()
                if industry_response.data:
                    extended_parts.append(f"Industry: {industry_response.data['name']}")
            except:
                pass
        
        # Fetch related keywords using code_id (code_keywords references classification_codes.id)
        try:
            keywords_response = supabase.table('code_keywords').select('keyword').eq(
                'code_id', code_id
            ).limit(15).execute()
            
            if keywords_response.data:
                keywords = [kw['keyword'] for kw in keywords_response.data]
                extended_parts.append(f"Related terms: {', '.join(keywords)}")
        except Exception as e:
            # If query fails, continue without keywords
            pass
        
        # Fetch examples if available (check if code_metadata table exists)
        try:
            metadata_response = supabase.table('code_metadata').select('examples').eq(
                'code_type', code_type
            ).eq('code', code_value).single().execute()
            
            if metadata_response.data and metadata_response.data.get('examples'):
                extended_parts.append(f"Examples: {metadata_response.data['examples']}")
        except:
            # Table might not exist or no examples, continue
            pass
        
        # Add business type hints
        business_type_hints = get_business_type_hints(code_type, code_value)
        if business_type_hints:
            extended_parts.append(business_type_hints)
        
        enriched.append({
            'code_type': code_type,
            'code': code_value,
            'description': description,
            'extended_description': '. '.join(extended_parts),
            'industry_context': code.get('industry_id', ''),
        })
    
    return enriched

def get_business_type_hints(code_type: str, code: str) -> str:
    """Add contextual hints for better embeddings."""
    hints_map = {
        # MCC hints
        ('MCC', '5812'): "Restaurants, cafes, dining establishments, food service",
        ('MCC', '5814'): "Fast food restaurants, quick service, QSR",
        ('MCC', '5411'): "Grocery stores, supermarkets, food retail",
        ('MCC', '7372'): "Software development, programming services, IT consulting",
        ('MCC', '8011'): "Medical doctors, physicians, healthcare providers",
        ('MCC', '8021'): "Dental offices, dentists, orthodontists",
        # Add more hints for common codes
    }
    
    return hints_map.get((code_type, code), "")

def generate_embeddings_batch(
    codes: List[Dict],
    model: SentenceTransformer,
    batch_size: int
) -> List[Dict]:
    """Generate embeddings in batches for efficiency."""
    embeddings_data = []
    
    # Process in batches
    for i in tqdm(range(0, len(codes), batch_size), desc="Generating embeddings"):
        batch = codes[i:i+batch_size]
        
        # Extract texts for embedding
        texts = [code['extended_description'] for code in batch]
        
        # Generate embeddings for batch
        embeddings = model.encode(texts, show_progress_bar=False)
        
        # Combine with metadata
        for code, embedding in zip(batch, embeddings):
            embeddings_data.append({
                'code_type': code['code_type'],
                'code': code['code'],
                'description': code['description'],
                'extended_description': code['extended_description'],
                'industry_context': str(code['industry_context']),
                'embedding': embedding.tolist(),
            })
    
    return embeddings_data

def insert_embeddings(
    embeddings_data: List[Dict],
    supabase: Client,
    batch_size: int
):
    """Insert embeddings into database in batches."""
    total = len(embeddings_data)
    
    for i in tqdm(range(0, total, batch_size), desc="Inserting"):
        batch = embeddings_data[i:i+batch_size]
        
        try:
            supabase.table('code_embeddings').insert(batch).execute()
            time.sleep(0.1)  # Brief pause to avoid rate limits
        except Exception as e:
            print(f"\n❌ Error inserting batch {i//batch_size + 1}: {e}")
            # Try inserting one by one for this batch
            for item in batch:
                try:
                    supabase.table('code_embeddings').insert(item).execute()
                except Exception as e2:
                    print(f"   ❌ Failed to insert {item['code_type']} {item['code']}: {e2}")

def verify_embeddings(supabase: Client):
    """Verify embeddings were inserted correctly."""
    # Count total
    response = supabase.table('code_embeddings').select('*', count='exact').execute()
    total_count = response.count
    
    print(f"   Total embeddings in database: {total_count}")
    
    # Count by type
    for code_type in ['MCC', 'SIC', 'NAICS']:
        response = supabase.table('code_embeddings').select(
            '*', count='exact'
        ).eq('code_type', code_type).execute()
        print(f"   - {code_type}: {response.count}")
    
    # Test a similarity search
    print("\n   Testing similarity search...")
    test_embedding = [0.0] * 384  # Dummy embedding
    test_embedding[0] = 1.0
    
    try:
        result = supabase.rpc(
            'match_code_embeddings',
            {
                'query_embedding': test_embedding,
                'code_type_filter': 'MCC',
                'match_threshold': 0.0,
                'match_count': 3
            }
        ).execute()
        
        print(f"   ✅ Similarity search working (returned {len(result.data)} results)")
    except Exception as e:
        print(f"   ❌ Similarity search test failed: {e}")

if __name__ == '__main__':
    main()

