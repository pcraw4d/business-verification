#!/bin/bash

# KYB Platform - CI/CD Integration Script
# This script provides integration with various CI/CD platforms

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# GitHub Actions integration
setup_github_actions() {
    log_info "Setting up GitHub Actions integration..."
    
    mkdir -p .github/workflows
    
    # Create main CI/CD workflow
    cat > .github/workflows/ci-cd.yml << 'EOF'
name: KYB Platform CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  release:
    types: [ published ]

env:
  GO_VERSION: '1.22'
  DOCKER_REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Install dependencies
      run: go mod download

    - name: Run tests
      run: go test ./... -v -race -coverprofile=coverage.out

    - name: Run security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.DOCKER_REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile.production
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Deploy to staging
      run: |
        chmod +x scripts/deploy-automation.sh
        ./scripts/deploy-automation.sh \
          -e staging \
          -s github \
          -b ${{ github.run_number }} \
          -c ${{ github.sha }} \
          -r ${{ github.ref_name }} \
          -n slack
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

  deploy-production:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Deploy to production
      run: |
        chmod +x scripts/deploy-automation.sh
        ./scripts/deploy-automation.sh \
          -e production \
          -s github \
          -b ${{ github.run_number }} \
          -c ${{ github.sha }} \
          -r ${{ github.ref_name }} \
          -n all
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
        TEAMS_WEBHOOK_URL: ${{ secrets.TEAMS_WEBHOOK_URL }}
        EMAIL_RECIPIENTS: ${{ secrets.EMAIL_RECIPIENTS }}

  security-scan:
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results to GitHub Security tab
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'
EOF

    # Create release workflow
    cat > .github/workflows/release.yml << 'EOF'
name: Release

on:
  release:
    types: [ published ]

env:
  DOCKER_REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.DOCKER_REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Build and push release image
      uses: docker/build-push-action@v5
      with:
        context: .
        file: ./Dockerfile.production
        push: true
        tags: |
          ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.event.release.tag_name }}
          ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}:latest
        labels: |
          org.opencontainers.image.title=KYB Platform
          org.opencontainers.image.description=Know Your Business verification platform
          org.opencontainers.image.version=${{ github.event.release.tag_name }}
          org.opencontainers.image.created=${{ github.event.release.published_at }}
          org.opencontainers.image.revision=${{ github.sha }}
          org.opencontainers.image.licenses=MIT

    - name: Deploy release to production
      run: |
        chmod +x scripts/deploy-automation.sh
        ./scripts/deploy-automation.sh \
          -e production \
          -t manual \
          -s github \
          -b ${{ github.run_number }} \
          -c ${{ github.sha }} \
          -r ${{ github.ref_name }} \
          -n all
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
        TEAMS_WEBHOOK_URL: ${{ secrets.TEAMS_WEBHOOK_URL }}
        EMAIL_RECIPIENTS: ${{ secrets.EMAIL_RECIPIENTS }}
EOF

    log_success "GitHub Actions integration setup completed"
}

# GitLab CI integration
setup_gitlab_ci() {
    log_info "Setting up GitLab CI integration..."
    
    cat > .gitlab-ci.yml << 'EOF'
stages:
  - test
  - build
  - security
  - deploy

variables:
  GO_VERSION: "1.22"
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

services:
  - docker:24-dind

before_script:
  - apk add --no-cache git
  - go version

test:
  stage: test
  image: golang:1.22-alpine
  script:
    - go mod download
    - go test ./... -v -race -coverprofile=coverage.out
    - go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    - gosec ./...
  coverage: '/coverage: \d+\.\d+%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.out
    expire_in: 1 week

build:
  stage: build
  image: docker:24
  script:
    - docker build -f Dockerfile.production -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker build -f Dockerfile.production -t $CI_REGISTRY_IMAGE:latest .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
  only:
    - main
    - develop

security:
  stage: security
  image: docker:24
  script:
    - docker run --rm -v $(pwd):/workspace aquasec/trivy fs /workspace
  allow_failure: true

deploy_staging:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache bash curl
  script:
    - chmod +x scripts/deploy-automation.sh
    - ./scripts/deploy-automation.sh -e staging -s gitlab -b $CI_PIPELINE_ID -c $CI_COMMIT_SHA -r $CI_COMMIT_REF_NAME -n slack
  environment:
    name: staging
    url: https://staging.kybplatform.com
  only:
    - develop

deploy_production:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache bash curl
  script:
    - chmod +x scripts/deploy-automation.sh
    - ./scripts/deploy-automation.sh -e production -s gitlab -b $CI_PIPELINE_ID -c $CI_COMMIT_SHA -r $CI_COMMIT_REF_NAME -n all
  environment:
    name: production
    url: https://api.kybplatform.com
  only:
    - main
  when: manual
EOF

    log_success "GitLab CI integration setup completed"
}

# Jenkins integration
setup_jenkins() {
    log_info "Setting up Jenkins integration..."
    
    mkdir -p jenkins
    
    # Create Jenkinsfile
    cat > Jenkinsfile << 'EOF'
pipeline {
    agent any
    
    environment {
        GO_VERSION = '1.22'
        DOCKER_REGISTRY = 'your-registry.com'
        IMAGE_NAME = 'kyb-platform'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Test') {
            steps {
                sh '''
                    go version
                    go mod download
                    go test ./... -v -race -coverprofile=coverage.out
                '''
            }
            post {
                always {
                    publishCoverage adapters: [
                        coberturaAdapter('coverage.xml')
                    ], sourceFileResolver: sourceFiles('STORE_LAST_BUILD')
                }
            }
        }
        
        stage('Security Scan') {
            steps {
                sh '''
                    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
                    gosec ./...
                '''
            }
        }
        
        stage('Build') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            steps {
                script {
                    def image = docker.build("${DOCKER_REGISTRY}/${IMAGE_NAME}:${BUILD_NUMBER}")
                    docker.withRegistry("https://${DOCKER_REGISTRY}", 'docker-registry-credentials') {
                        image.push()
                        image.push('latest')
                    }
                }
            }
        }
        
        stage('Deploy Staging') {
            when {
                branch 'develop'
            }
            steps {
                sh '''
                    chmod +x scripts/deploy-automation.sh
                    ./scripts/deploy-automation.sh \
                        -e staging \
                        -s jenkins \
                        -b $BUILD_NUMBER \
                        -c $GIT_COMMIT \
                        -r $GIT_BRANCH \
                        -n slack
                '''
            }
        }
        
        stage('Deploy Production') {
            when {
                branch 'main'
            }
            steps {
                input message: 'Deploy to production?', ok: 'Deploy'
                sh '''
                    chmod +x scripts/deploy-automation.sh
                    ./scripts/deploy-automation.sh \
                        -e production \
                        -s jenkins \
                        -b $BUILD_NUMBER \
                        -c $GIT_COMMIT \
                        -r $GIT_BRANCH \
                        -n all
                '''
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
        success {
            slackSend channel: '#deployments',
                      color: 'good',
                      message: "✅ Build ${env.BUILD_NUMBER} succeeded for ${env.JOB_NAME}"
        }
        failure {
            slackSend channel: '#deployments',
                      color: 'danger',
                      message: "❌ Build ${env.BUILD_NUMBER} failed for ${env.JOB_NAME}"
        }
    }
}
EOF

    # Create Jenkins job configuration
    cat > jenkins/job-config.xml << 'EOF'
<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job@2.41">
  <description>KYB Platform CI/CD Pipeline</description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <jenkins.model.BuildDiscarderProperty>
      <strategy class="hudson.tasks.LogRotator">
        <daysToKeep>30</daysToKeep>
        <numToKeep>50</numToKeep>
        <artifactDaysToKeep>-1</artifactDaysToKeep>
        <artifactNumToKeep>-1</artifactNumToKeep>
      </strategy>
    </jenkins.model.BuildDiscarderProperty>
  </properties>
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition" plugin="workflow-cps@2.90">
    <scm class="hudson.plugins.git.GitSCM" plugin="git@4.8.3">
      <configVersion>2</configVersion>
      <userRemoteConfigs>
        <hudson.plugins.git.UserRemoteConfig>
          <url>https://github.com/your-org/kyb-platform.git</url>
        </hudson.plugins.git.UserRemoteConfig>
      </userRemoteConfigs>
      <branches>
        <hudson.plugins.git.BranchSpec>
          <name>*/main</name>
        </hudson.plugins.git.BranchSpec>
        <hudson.plugins.git.BranchSpec>
          <name>*/develop</name>
        </hudson.plugins.git.BranchSpec>
      </branches>
      <doGenerateSubmoduleConfigurations>false</doGenerateSubmoduleConfigurations>
      <submoduleCfg class="list"/>
      <extensions/>
    </scm>
    <scriptPath>Jenkinsfile</scriptPath>
    <lightweight>false</lightweight>
  </definition>
  <triggers>
    <hudson.triggers.SCMTrigger>
      <spec>H/5 * * * *</spec>
    </hudson.triggers.SCMTrigger>
  </triggers>
  <disabled>false</disabled>
</flow-definition>
EOF

    log_success "Jenkins integration setup completed"
}

# Create deployment hooks
create_deployment_hooks() {
    log_info "Creating deployment hooks..."
    
    mkdir -p hooks
    
    # Pre-deployment hook
    cat > hooks/pre-deploy.sh << 'EOF'
#!/bin/bash
# Pre-deployment hook
# This script runs before deployment

echo "Running pre-deployment checks..."

# Check if database migrations are needed
if [ -f "scripts/check-migrations.sh" ]; then
    ./scripts/check-migrations.sh
fi

# Check if configuration is valid
if [ -f "scripts/validate-config.sh" ]; then
    ./scripts/validate-config.sh
fi

# Check if dependencies are available
if [ -f "scripts/check-dependencies.sh" ]; then
    ./scripts/check-dependencies.sh
fi

echo "Pre-deployment checks completed"
EOF

    # Post-deployment hook
    cat > hooks/post-deploy.sh << 'EOF'
#!/bin/bash
# Post-deployment hook
# This script runs after deployment

echo "Running post-deployment tasks..."

# Run database migrations
if [ -f "scripts/run-migrations.sh" ]; then
    ./scripts/run-migrations.sh
fi

# Clear caches
if [ -f "scripts/clear-caches.sh" ]; then
    ./scripts/clear-caches.sh
fi

# Warm up application
if [ -f "scripts/warm-up.sh" ]; then
    ./scripts/warm-up.sh
fi

# Send deployment notification
if [ -f "scripts/send-notification.sh" ]; then
    ./scripts/send-notification.sh "deployment_success"
fi

echo "Post-deployment tasks completed"
EOF

    # Make hooks executable
    chmod +x hooks/*.sh
    
    log_success "Deployment hooks created"
}

# Create monitoring integration
create_monitoring_integration() {
    log_info "Creating monitoring integration..."
    
    mkdir -p monitoring/integrations
    
    # Prometheus configuration for CI/CD metrics
    cat > monitoring/integrations/prometheus-cicd.yml << 'EOF'
# Prometheus configuration for CI/CD metrics
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "cicd-alerts.yml"

scrape_configs:
  - job_name: 'cicd-metrics'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: /metrics
    scrape_interval: 30s

  - job_name: 'deployment-metrics'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics/deployment
    scrape_interval: 60s
EOF

    # CI/CD alert rules
    cat > monitoring/integrations/cicd-alerts.yml << 'EOF'
groups:
  - name: cicd
    rules:
      - alert: DeploymentFailed
        expr: deployment_status{status="failed"} == 1
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Deployment failed"
          description: "Deployment {{ $labels.deployment_id }} failed in environment {{ $labels.environment }}"

      - alert: DeploymentDurationHigh
        expr: deployment_duration_seconds > 1800
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Deployment taking too long"
          description: "Deployment {{ $labels.deployment_id }} is taking longer than expected"

      - alert: TestFailureRateHigh
        expr: test_failure_rate > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High test failure rate"
          description: "Test failure rate is {{ $value | humanizePercentage }}"
EOF

    log_success "Monitoring integration created"
}

# Main setup function
main() {
    echo "=========================================="
    echo "      KYB Platform CI/CD Integration"
    echo "=========================================="
    echo
    
    # Setup GitHub Actions
    setup_github_actions
    
    # Setup GitLab CI
    setup_gitlab_ci
    
    # Setup Jenkins
    setup_jenkins
    
    # Create deployment hooks
    create_deployment_hooks
    
    # Create monitoring integration
    create_monitoring_integration
    
    echo
    echo "=========================================="
    echo "         INTEGRATION SUMMARY"
    echo "=========================================="
    echo "✅ GitHub Actions workflows created"
    echo "✅ GitLab CI configuration created"
    echo "✅ Jenkins pipeline created"
    echo "✅ Deployment hooks created"
    echo "✅ Monitoring integration created"
    echo "=========================================="
    echo
    
    log_success "CI/CD integration setup completed successfully!"
    
    echo
    echo "Next steps:"
    echo "1. Configure secrets in your CI/CD platform"
    echo "2. Set up webhook URLs for notifications"
    echo "3. Configure deployment environments"
    echo "4. Test the integration with a sample deployment"
}

# Run main function
main "$@"
