#!/bin/bash
set -e

echo "========================================="
echo "Review Slot Guard Bot - Deployment Script"
echo "========================================="
echo ""

# Check if terraform.tfvars exists
if [ ! -f "terraform/terraform.tfvars" ]; then
    echo "❌ terraform.tfvars not found!"
    echo "Creating from example..."
    cp terraform/terraform.tfvars.example terraform/terraform.tfvars
    echo ""
    echo "⚠️  Please edit terraform/terraform.tfvars with your values:"
    echo "   - folder_id"
    echo "   - cloud_id"
    echo "   - telegram_bot_token"
    echo ""
    read -p "Press Enter after editing terraform.tfvars..."
fi

# Build functions
echo "📦 Building functions..."
cd functions/telegram_handler
GOOS=linux GOARCH=amd64 go build -o main .
cd ../periodic_job
GOOS=linux GOARCH=amd64 go build -o main .
cd ../..
echo "✅ Functions built"
echo ""

# Deploy with Terraform
echo "🚀 Deploying to Yandex Cloud..."
cd terraform
terraform init
terraform plan
echo ""
echo "Applying Terraform configuration..."
terraform apply
echo ""

# Show outputs
echo "========================================="
echo "Deployment Complete!"
echo "========================================="
echo ""
terraform output
echo ""
echo "📱 Send /start to your bot to begin!"
