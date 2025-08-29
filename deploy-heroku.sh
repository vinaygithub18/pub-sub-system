#!/bin/bash

echo "🚀 Deploying Pub/Sub System to Heroku"
echo "======================================"

# Check if Heroku CLI is installed
if ! command -v heroku &> /dev/null; then
    echo "❌ Heroku CLI not found. Please install it first:"
    echo "   macOS: brew tap heroku/brew && brew install heroku"
    echo "   Linux: curl https://cli-assets.heroku.com/install.sh | sh"
    echo "   Windows: Download from https://devcenter.heroku.com/articles/heroku-cli"
    exit 1
fi

# Check if user is logged in
if ! heroku auth:whoami &> /dev/null; then
    echo "🔐 Please login to Heroku first:"
    heroku login
fi

# Check if git repository is initialized
if [ ! -d ".git" ]; then
    echo "📁 Initializing git repository..."
    git init
    git add .
    git commit -m "Initial commit for Heroku deployment"
fi

# Check if Heroku app exists
if ! heroku apps:info &> /dev/null; then
    echo "🏗️  Creating new Heroku app..."
    heroku create
else
    echo "✅ Heroku app already exists"
fi

# Get app name
APP_NAME=$(heroku apps:info --json | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
echo "📱 App name: $APP_NAME"

# Set environment variables
echo "⚙️  Setting environment variables..."
heroku config:set MAX_TOPICS=100
heroku config:set MAX_SUBSCRIBERS=100
heroku config:set MAX_MESSAGES=100

# Deploy to Heroku
echo "🚀 Deploying to Heroku..."
git add .
git commit -m "Deploy to Heroku $(date)"
git push heroku main

# Check deployment status
echo "📊 Checking deployment status..."
heroku ps

# Open the app
echo "🌐 Opening your app..."
heroku open

echo ""
echo "✅ Deployment complete!"
echo "🔗 Your app is available at: https://$APP_NAME.herokuapp.com"
echo "🔌 WebSocket endpoint: wss://$APP_NAME.herokuapp.com/ws"
echo ""
echo "🧪 Test your deployment:"
echo "   Health check: curl https://$APP_NAME.herokuapp.com/health"
echo "   Create topic: curl -X POST https://$APP_NAME.herokuapp.com/topics -H 'Content-Type: application/json' -d '{\"name\":\"test\"}'"
echo ""
echo "📖 View logs: heroku logs --tail"
echo "🔄 Restart app: heroku restart"
