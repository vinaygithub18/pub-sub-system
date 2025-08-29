# 🚂 Railway Deployment Guide

## Why Railway?

✅ **Free Tier Available** - No credit card required  
✅ **Full WebSocket Support** - Perfect for your Pub/Sub system  
✅ **Go Native Support** - Automatic Go detection and build  
✅ **Global CDN** - Fast worldwide access  
✅ **Automatic HTTPS** - SSL certificates included  
✅ **Easy Scaling** - From free to enterprise  

## Prerequisites

1. **GitHub Account** (for code repository)
2. **Railway Account** (free at railway.app)
3. **Railway CLI** (already installed)

## 🚀 **Deployment Steps**

### 1. **Login to Railway**
```bash
railway login
```
- Opens browser for authentication
- Connect with GitHub account

### 2. **Initialize Railway Project**
```bash
railway init
```
- Creates new Railway project
- Links to your GitHub repository

### 3. **Deploy Your App**
```bash
railway up
```
- Builds and deploys your Go application
- Automatically detects Go and builds

### 4. **Get Your App URL**
```bash
railway status
```
- Shows your deployed app URL
- Usually: `https://your-app-name.railway.app`

## 🔧 **Configuration**

### **Environment Variables** (Optional)
```bash
railway variables set MAX_TOPICS=100
railway variables set MAX_SUBSCRIBERS=100
railway variables set MAX_MESSAGES=100
```

### **Custom Domain** (Optional)
```bash
railway domain yourdomain.com
```

## 🧪 **Testing Your Deployment**

### **Health Check**
```bash
curl https://your-app-name.railway.app/health
```

### **Create Topic**
```bash
curl -X POST https://your-app-name.railway.app/topics \
  -H "Content-Type: application/json" \
  -d '{"name": "test-topic"}'
```

### **List Topics**
```bash
curl https://your-app-name.railway.app/topics
```

### **WebSocket Connection**
```bash
# Update websocket_test.html with your Railway URL
# Change ws://localhost:8080/ws to:
# wss://your-app-name.railway.app/ws
```

## 📊 **Monitoring & Management**

### **View Logs**
```bash
railway logs
```

### **Check Status**
```bash
railway status
```

### **Restart App**
```bash
railway restart
```

### **Open Dashboard**
```bash
railway open
```

## 🚨 **Important Notes**

### **WebSocket Support**
- Railway **fully supports** WebSocket connections
- Use `wss://` (secure WebSocket) for production
- No additional configuration needed

### **Port Configuration**
- Railway automatically sets the `PORT` environment variable
- Your app uses `os.Getenv("PORT")` (already configured)
- No changes needed to your code

### **Free Tier Limits**
- **Build Time**: 500 minutes/month
- **Deployments**: Unlimited
- **Bandwidth**: 100GB/month
- **Perfect for development and small production apps**

## 🔄 **Continuous Deployment**

### **Automatic Deploys**
- Connect your GitHub repository
- Railway automatically deploys on every push
- No manual deployment needed

### **Manual Deploy**
```bash
railway up
```

## 🛠️ **Troubleshooting**

### **Build Failures**
```bash
# Check build logs
railway logs

# Verify Go version
go version
```

### **Runtime Errors**
```bash
# View application logs
railway logs

# Check app status
railway status
```

### **WebSocket Issues**
- Ensure you're using `wss://` for secure connections
- Check that your client supports secure WebSockets
- Verify CORS settings if needed

## 📈 **Scaling Options**

### **Free Tier**
- Perfect for development and testing
- 100GB bandwidth per month

### **Pro Plan** ($5/month)
- 1TB bandwidth per month
- Custom domains
- Priority support

### **Enterprise**
- Unlimited bandwidth
- Advanced monitoring
- Dedicated support

## 🎯 **Next Steps After Deployment**

1. **Test all endpoints** with your Railway URL
2. **Update WebSocket client** to use `wss://your-app.railway.app/ws`
3. **Monitor logs** for any issues
4. **Set up monitoring** for production use
5. **Consider adding persistence** with Railway Postgres

## 🆘 **Support**

- **Railway Documentation**: https://docs.railway.app/
- **Go on Railway**: https://docs.railway.app/deploy/deployments
- **Community Discord**: https://railway.app/discord

## 🎉 **Deployment Commands Summary**

```bash
# Login and setup
railway login
railway init

# Deploy
railway up

# Monitor
railway status
railway logs

# Manage
railway restart
railway open
```

Your Pub/Sub system is now ready for production deployment on Railway with full WebSocket support! 🚂✨
