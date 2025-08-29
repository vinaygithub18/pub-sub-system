# 🚀 Heroku Deployment Guide

## Prerequisites

1. **Install Heroku CLI**
   ```bash
   # macOS
   brew tap heroku/brew && brew install heroku
   
   # Windows
   # Download from: https://devcenter.heroku.com/articles/heroku-cli
   
   # Linux
   curl https://cli-assets.heroku.com/install.sh | sh
   ```

2. **Login to Heroku**
   ```bash
   heroku login
   ```

## Deployment Steps

### 1. **Initialize Git Repository** (if not already done)
   ```bash
   git init
   git add .
   git commit -m "Initial commit for Heroku deployment"
   ```

### 2. **Create Heroku App**
   ```bash
   heroku create your-app-name
   # or let Heroku generate a name:
   heroku create
   ```

### 3. **Set Environment Variables** (Optional)
   ```bash
   heroku config:set MAX_TOPICS=100
   heroku config:set MAX_SUBSCRIBERS=100
   heroku config:set MAX_MESSAGES=100
   ```

### 4. **Deploy to Heroku**
   ```bash
   git push heroku main
   # or if your branch is called master:
   git push heroku master
   ```

### 5. **Open Your App**
   ```bash
   heroku open
   ```

## File Structure for Heroku

Your project should have these files:
- ✅ `main.go` - Main Go application
- ✅ `go.mod` - Go module file
- ✅ `go.sum` - Go dependencies
- ✅ `Procfile` - Heroku process definition
- ✅ `app.json` - Heroku app configuration

## Testing Your Deployment

### 1. **Health Check**
   ```bash
   curl https://your-app-name.herokuapp.com/health
   ```

### 2. **Create a Topic**
   ```bash
   curl -X POST https://your-app-name.herokuapp.com/topics \
     -H "Content-Type: application/json" \
     -d '{"name": "test-topic"}'
   ```

### 3. **List Topics**
   ```bash
   curl https://your-app-name.herokuapp.com/topics
   ```

### 4. **WebSocket Connection**
   ```bash
   # Update websocket_test.html with your Heroku URL
   # Change ws://localhost:8080/ws to:
   # wss://your-app-name.herokuapp.com/ws
   ```

## Important Notes

### ⚠️ **WebSocket Support**
- Heroku **fully supports** WebSocket connections
- Use `wss://` (secure WebSocket) for production
- Your existing WebSocket code will work without changes

### 🔧 **Port Configuration**
- Heroku automatically sets the `PORT` environment variable
- Your app now uses `os.Getenv("PORT")` instead of hardcoded port
- No additional configuration needed

### 📊 **Scaling**
- Start with the free tier for testing
- Upgrade to paid dynos for production use
- Use `heroku ps:scale web=1` to scale up

## Troubleshooting

### **Build Failures**
   ```bash
   # Check build logs
   heroku logs --tail
   
   # Check buildpack
   heroku buildpacks
   ```

### **Runtime Errors**
   ```bash
   # View application logs
   heroku logs --tail
   
   # Check app status
   heroku ps
   ```

### **WebSocket Issues**
   - Ensure you're using `wss://` for secure connections
   - Check that your client supports secure WebSockets
   - Verify CORS settings if needed

## Production Considerations

### 1. **Custom Domain** (Optional)
   ```bash
   heroku domains:add www.yourdomain.com
   ```

### 2. **SSL Certificate**
   - Heroku provides automatic SSL for `*.herokuapp.com` domains
   - Custom domains require SSL certificate setup

### 3. **Monitoring**
   ```bash
   # Add monitoring add-ons
   heroku addons:create papertrail:choklad
   heroku addons:create newrelic:wayne
   ```

### 4. **Database** (Future Enhancement)
   ```bash
   # Add PostgreSQL for persistence
   heroku addons:create heroku-postgresql:mini
   ```

## Quick Commands Reference

```bash
# Deploy updates
git add .
git commit -m "Update message"
git push heroku main

# View logs
heroku logs --tail

# Restart app
heroku restart

# Check app status
heroku ps

# Open app
heroku open

# Run locally with Heroku config
heroku local web
```

## Next Steps After Deployment

1. **Test all endpoints** with your Heroku URL
2. **Update WebSocket client** to use `wss://your-app.herokuapp.com/ws`
3. **Monitor logs** for any issues
4. **Set up monitoring** for production use
5. **Consider adding persistence** with Heroku Postgres

## Support

- **Heroku Documentation**: https://devcenter.heroku.com/
- **Go on Heroku**: https://devcenter.heroku.com/articles/go-support
- **WebSocket Support**: https://devcenter.heroku.com/articles/websockets

Your Pub/Sub system is now ready for production deployment on Heroku with full WebSocket support! 🎉
