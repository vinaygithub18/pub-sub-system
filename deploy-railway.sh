#!/bin/bash

echo "🚂 Deploying Pub/Sub System to Railway"
echo "======================================"

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing now..."
    npm install -g @railway/cli
fi

# Check if user is logged in
if ! railway whoami &> /dev/null; then
    echo "🔐 Please login to Railway first:"
    railway login
fi

# Check if git repository is initialized
if [ ! -d ".git" ]; then
    echo "📁 Initializing git repository..."
    git init
    git add .
    git commit -m "Initial commit for Railway deployment"
fi

# Check if Railway project exists
if [ ! -f ".railway" ]; then
    echo "🏗️  Initializing Railway project..."
    railway init
else
    echo "✅ Railway project already exists"
fi

# Deploy to Railway
echo "🚀 Deploying to Railway..."
railway up

# Check deployment status
echo "📊 Checking deployment status..."
railway status

# Get the app URL
echo "🔗 Getting your app URL..."
APP_URL=$(railway status --json | grep -o '"url":"[^"]*"' | cut -d'"' -f4)

if [ ! -z "$APP_URL" ]; then
    echo ""
    echo "✅ Deployment complete!"
    echo "🔗 Your app is available at: $APP_URL"
    echo "🔌 WebSocket endpoint: ${APP_URL/https:/wss:}/ws"
    echo ""
    echo "🧪 Test your deployment:"
    echo "   Health check: curl $APP_URL/health"
    echo "   Create topic: curl -X POST $APP_URL/topics -H 'Content-Type: application/json' -d '{\"name\":\"test\"}'"
    echo ""
    echo "📖 View logs: railway logs"
    echo "🔄 Restart app: railway restart"
    echo "🌐 Open dashboard: railway open"
else
    echo ""
    echo "⚠️  Deployment may still be in progress..."
    echo "📊 Check status: railway status"
    echo "📖 View logs: railway logs"
fi
