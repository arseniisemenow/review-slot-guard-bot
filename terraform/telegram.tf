# Telegram webhook should be set manually after deployment
# Run the following command after terraform apply:
# curl -X POST "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook" -d "url=$(terraform output -raw api_gateway_url)"
