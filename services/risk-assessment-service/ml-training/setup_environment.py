"""
Setup Script for LSTM Training Environment

Initializes the Python ML environment and generates synthetic training data.
"""

import os
import sys
import subprocess
import yaml
from pathlib import Path


def run_command(command: str, description: str) -> bool:
    """Run a command and return success status"""
    print(f"\n{description}...")
    print(f"Running: {command}")
    
    try:
        result = subprocess.run(command, shell=True, check=True, capture_output=True, text=True)
        print("✅ Success")
        if result.stdout:
            print(f"Output: {result.stdout}")
        return True
    except subprocess.CalledProcessError as e:
        print(f"❌ Failed with exit code {e.returncode}")
        if e.stdout:
            print(f"Output: {e.stdout}")
        if e.stderr:
            print(f"Error: {e.stderr}")
        return False


def setup_python_environment():
    """Set up Python virtual environment and install dependencies"""
    
    print("Setting up Python ML environment...")
    
    # Check if Python is available
    if not run_command("python3 --version", "Checking Python version"):
        print("Python3 not found. Please install Python 3.8+ first.")
        return False
    
    # Create virtual environment
    if not run_command("python3 -m venv venv", "Creating virtual environment"):
        return False
    
    # Activate virtual environment and install dependencies
    if os.name == 'nt':  # Windows
        activate_cmd = "venv\\Scripts\\activate"
        pip_cmd = "venv\\Scripts\\pip"
    else:  # Unix/Linux/macOS
        activate_cmd = "source venv/bin/activate"
        pip_cmd = "venv/bin/pip"
    
    # Upgrade pip
    if not run_command(f"{pip_cmd} install --upgrade pip", "Upgrading pip"):
        return False
    
    # Install dependencies
    if not run_command(f"{pip_cmd} install -r requirements.txt", "Installing dependencies"):
        return False
    
    print("✅ Python environment setup completed!")
    return True


def create_directories():
    """Create necessary directories"""
    
    print("\nCreating directories...")
    
    directories = [
        "data",
        "models", 
        "training",
        "export",
        "notebooks",
        "output",
        "checkpoints",
        "logs"
    ]
    
    for directory in directories:
        os.makedirs(directory, exist_ok=True)
        print(f"✅ Created directory: {directory}")
    
    print("✅ All directories created!")


def generate_synthetic_data():
    """Generate synthetic training data"""
    
    print("\nGenerating synthetic training data...")
    
    # Check if virtual environment exists
    if not os.path.exists("venv"):
        print("Virtual environment not found. Please run setup_python_environment() first.")
        return False
    
    # Activate virtual environment and run data generator
    if os.name == 'nt':  # Windows
        python_cmd = "venv\\Scripts\\python"
    else:  # Unix/Linux/macOS
        python_cmd = "venv/bin/python"
    
    # Run synthetic data generator
    if not run_command(f"{python_cmd} data/synthetic_generator.py", "Generating synthetic data"):
        return False
    
    print("✅ Synthetic data generation completed!")
    return True


def validate_setup():
    """Validate the setup by running a quick test"""
    
    print("\nValidating setup...")
    
    # Check if virtual environment exists
    if not os.path.exists("venv"):
        print("❌ Virtual environment not found")
        return False
    
    # Check if data file exists
    if not os.path.exists("data/synthetic_risk_data.parquet"):
        print("❌ Synthetic data file not found")
        return False
    
    # Check if config file exists
    if not os.path.exists("models/model_config.yaml"):
        print("❌ Model config file not found")
        return False
    
    # Test model creation
    if os.name == 'nt':  # Windows
        python_cmd = "venv\\Scripts\\python"
    else:  # Unix/Linux/macOS
        python_cmd = "venv/bin/python"
    
    if not run_command(f"{python_cmd} -c \"import torch; print('PyTorch version:', torch.__version__)\"", "Testing PyTorch"):
        return False
    
    if not run_command(f"{python_cmd} -c \"import onnx; print('ONNX version:', onnx.__version__)\"", "Testing ONNX"):
        return False
    
    print("✅ Setup validation completed!")
    return True


def main():
    """Main setup function"""
    
    print("="*60)
    print("LSTM TRAINING ENVIRONMENT SETUP")
    print("="*60)
    
    # Change to the ml-training directory
    script_dir = Path(__file__).parent
    os.chdir(script_dir)
    
    # Step 1: Set up Python environment
    if not setup_python_environment():
        print("❌ Failed to set up Python environment")
        return False
    
    # Step 2: Create directories
    create_directories()
    
    # Step 3: Generate synthetic data
    if not generate_synthetic_data():
        print("❌ Failed to generate synthetic data")
        return False
    
    # Step 4: Validate setup
    if not validate_setup():
        print("❌ Setup validation failed")
        return False
    
    print("\n" + "="*60)
    print("SETUP COMPLETED SUCCESSFULLY!")
    print("="*60)
    print("\nNext steps:")
    print("1. Activate the virtual environment:")
    if os.name == 'nt':  # Windows
        print("   venv\\Scripts\\activate")
    else:  # Unix/Linux/macOS
        print("   source venv/bin/activate")
    print("2. Train the LSTM model:")
    print("   python training/train_lstm.py")
    print("3. Export to ONNX:")
    print("   python export/export_to_onnx.py")
    print("\nFor more information, see the README.md file.")
    
    return True


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
