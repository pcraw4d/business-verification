#!/bin/bash

# KYB Platform - Infrastructure Deployment Script
# Handles Terraform deployment with proper validation and error handling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="production"
TERRAFORM_DIR="deployments/terraform"
PLAN_FILE="terraform.tfplan"
AUTO_APPROVE=false
DRY_RUN=false
BACKEND_INIT=false

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  init        - Initialize Terraform backend and modules"
    echo "  plan        - Create Terraform execution plan"
    echo "  apply       - Apply Terraform configuration"
    echo "  destroy     - Destroy infrastructure (use with caution)"
    echo "  validate    - Validate Terraform configuration"
    echo "  fmt         - Format Terraform files"
    echo "  output      - Show Terraform outputs"
    echo "  state       - Show Terraform state"
    echo "  refresh     - Refresh Terraform state"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV - Environment (production, staging, development)"
    echo "  -d, --dir DIR         - Terraform directory (default: deployments/terraform)"
    echo "  -a, --auto-approve    - Auto-approve changes (use with caution)"
    echo "  --dry-run             - Show what would be done without applying"
    echo "  --backend-init        - Initialize backend configuration"
    echo "  -h, --help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 init"
    echo "  $0 plan -e production"
    echo "  $0 apply -e production --auto-approve"
    echo "  $0 validate"
    echo "  $0 output"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Terraform is installed
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install Terraform first."
        exit 1
    fi
    
    # Check Terraform version
    local tf_version=$(terraform version -json | jq -r '.terraform_version')
    print_status "Terraform version: $tf_version"
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install AWS CLI first."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials not configured. Please run 'aws configure' first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install jq first."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to validate environment
validate_environment() {
    local env=$1
    
    case $env in
        production|staging|development)
            print_status "Environment: $env"
            ;;
        *)
            print_error "Invalid environment: $env"
            print_error "Valid environments: production, staging, development"
            exit 1
            ;;
    esac
}

# Function to get environment variables file
get_env_file() {
    local env=$1
    local env_file="$TERRAFORM_DIR/environments/${env}.tfvars"
    
    if [ ! -f "$env_file" ]; then
        print_error "Environment file not found: $env_file"
        exit 1
    fi
    
    echo "$env_file"
}

# Function to initialize Terraform
init_terraform() {
    print_status "Initializing Terraform..."
    
    cd "$TERRAFORM_DIR"
    
    # Initialize backend if requested
    if [ "$BACKEND_INIT" = true ]; then
        print_status "Initializing backend..."
        terraform init -backend=true
    else
        terraform init
    fi
    
    # Get modules
    terraform get
    
    print_success "Terraform initialized successfully"
}

# Function to validate Terraform configuration
validate_terraform() {
    print_status "Validating Terraform configuration..."
    
    cd "$TERRAFORM_DIR"
    
    # Format check
    if ! terraform fmt -check -recursive; then
        print_warning "Terraform files need formatting. Run 'terraform fmt' to fix."
    fi
    
    # Validation
    if terraform validate; then
        print_success "Terraform configuration is valid"
    else
        print_error "Terraform configuration validation failed"
        exit 1
    fi
}

# Function to create Terraform plan
plan_terraform() {
    local env_file=$1
    
    print_status "Creating Terraform plan..."
    
    cd "$TERRAFORM_DIR"
    
    # Create plan
    if terraform plan -var-file="$env_file" -out="$PLAN_FILE"; then
        print_success "Terraform plan created successfully"
        
        # Show plan summary
        print_status "Plan summary:"
        terraform show -json "$PLAN_FILE" | jq -r '.resource_changes[] | "\(.change.actions[]) \(.type) \(.name)"' | sort
    else
        print_error "Failed to create Terraform plan"
        exit 1
    fi
}

# Function to apply Terraform configuration
apply_terraform() {
    local env_file=$1
    
    print_status "Applying Terraform configuration..."
    
    cd "$TERRAFORM_DIR"
    
    # Check if plan exists
    if [ ! -f "$PLAN_FILE" ]; then
        print_warning "No plan file found. Creating new plan..."
        plan_terraform "$env_file"
    fi
    
    # Apply configuration
    if [ "$AUTO_APPROVE" = true ]; then
        if terraform apply "$PLAN_FILE"; then
            print_success "Terraform configuration applied successfully"
        else
            print_error "Failed to apply Terraform configuration"
            exit 1
        fi
    else
        print_status "Review the plan above and confirm to apply (y/N):"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            if terraform apply "$PLAN_FILE"; then
                print_success "Terraform configuration applied successfully"
            else
                print_error "Failed to apply Terraform configuration"
                exit 1
            fi
        else
            print_status "Deployment cancelled"
            exit 0
        fi
    fi
}

# Function to destroy infrastructure
destroy_terraform() {
    local env_file=$1
    
    print_warning "DESTROYING INFRASTRUCTURE - This action cannot be undone!"
    print_warning "This will delete all resources in the $ENVIRONMENT environment."
    
    if [ "$AUTO_APPROVE" = true ]; then
        print_warning "Auto-approve enabled. Proceeding with destruction..."
    else
        print_status "Type 'DESTROY' to confirm destruction:"
        read -r response
        if [ "$response" != "DESTROY" ]; then
            print_status "Destruction cancelled"
            exit 0
        fi
    fi
    
    cd "$TERRAFORM_DIR"
    
    if terraform destroy -var-file="$env_file" -auto-approve="$AUTO_APPROVE"; then
        print_success "Infrastructure destroyed successfully"
    else
        print_error "Failed to destroy infrastructure"
        exit 1
    fi
}

# Function to show outputs
show_outputs() {
    print_status "Showing Terraform outputs..."
    
    cd "$TERRAFORM_DIR"
    
    if terraform output -json | jq -r 'to_entries[] | "\(.key): \(.value.value)"'; then
        print_success "Outputs displayed successfully"
    else
        print_error "Failed to show outputs"
        exit 1
    fi
}

# Function to show state
show_state() {
    print_status "Showing Terraform state..."
    
    cd "$TERRAFORM_DIR"
    
    if terraform show -json | jq -r '.values.root_module.resources[] | "\(.type) \(.name): \(.mode)"'; then
        print_success "State displayed successfully"
    else
        print_error "Failed to show state"
        exit 1
    fi
}

# Function to refresh state
refresh_state() {
    print_status "Refreshing Terraform state..."
    
    cd "$TERRAFORM_DIR"
    
    if terraform refresh; then
        print_success "State refreshed successfully"
    else
        print_error "Failed to refresh state"
        exit 1
    fi
}

# Function to format Terraform files
format_terraform() {
    print_status "Formatting Terraform files..."
    
    cd "$TERRAFORM_DIR"
    
    if terraform fmt -recursive; then
        print_success "Terraform files formatted successfully"
    else
        print_error "Failed to format Terraform files"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -d|--dir)
            TERRAFORM_DIR="$2"
            shift 2
            ;;
        -a|--auto-approve)
            AUTO_APPROVE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --backend-init)
            BACKEND_INIT=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        init|plan|apply|destroy|validate|fmt|output|state|refresh)
            COMMAND="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if command is provided
if [ -z "$COMMAND" ]; then
    print_error "No command specified"
    show_usage
    exit 1
fi

# Main execution
print_status "Starting infrastructure deployment"
print_status "Environment: $ENVIRONMENT"
print_status "Terraform directory: $TERRAFORM_DIR"

# Check prerequisites
check_prerequisites

# Validate environment
validate_environment "$ENVIRONMENT"

# Get environment file
ENV_FILE=$(get_env_file "$ENVIRONMENT")

# Execute command
case $COMMAND in
    init)
        init_terraform
        ;;
    validate)
        validate_terraform
        ;;
    fmt)
        format_terraform
        ;;
    plan)
        init_terraform
        validate_terraform
        plan_terraform "$ENV_FILE"
        ;;
    apply)
        init_terraform
        validate_terraform
        if [ "$DRY_RUN" = true ]; then
            plan_terraform "$ENV_FILE"
        else
            apply_terraform "$ENV_FILE"
        fi
        ;;
    destroy)
        init_terraform
        destroy_terraform "$ENV_FILE"
        ;;
    output)
        show_outputs
        ;;
    state)
        show_state
        ;;
    refresh)
        refresh_state
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Infrastructure deployment completed successfully!"
