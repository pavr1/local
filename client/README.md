# Ice Cream Store - Client Application

A modern, service-oriented web client for the Ice Cream Store management system.

## 🏗️ Architecture

The client is organized into service-specific directories for better maintainability and scalability:

```
client/
├── index.html                 # Main entry point & service selector
├── auth/                      # Authentication service UI
│   ├── login.html            # Login page
│   ├── dashboard.html        # Auth dashboard
│   ├── auth.js               # Auth service logic
│   └── main.js               # Login page logic
├── orders/                    # Orders service UI
│   ├── index.html            # Orders management page
│   └── orders.js             # Orders service logic
├── shared/                    # Shared components & utilities
│   ├── css/
│   │   └── style.css         # Shared styles
│   ├── js/
│   │   └── navigation.js     # Navigation utilities
│   └── components/           # Reusable UI components
├── core/                      # Core application logic
│   └── js/                   # Core JavaScript modules
└── assets/                    # Static assets
```

## 🚀 Features

### Main Dashboard
- **Service Overview**: Visual cards for each microservice
- **Real-time Status**: Live health checks for all services
- **Quick Actions**: Direct access to common tasks
- **System Information**: Architecture overview and endpoints

### Authentication Service (`/auth/`)
- **Secure Login**: JWT-based authentication
- **Dashboard**: User profile and API testing
- **Token Management**: Automatic token refresh and validation
- **System Integration**: Seamless integration with all services

### Orders Service (`/orders/`)
- **Order Management**: Create, view, and update orders
- **Real-time Filtering**: Filter by status, date, and search terms
- **Statistics Dashboard**: Order analytics and insights
- **API Testing**: Built-in endpoint testing tools

### Shared Components (`/shared/`)
- **Responsive Design**: Mobile-first Bootstrap 5 implementation
- **Consistent Styling**: Shared CSS variables and components
- **Navigation Utils**: Service routing and breadcrumb management
- **Common Utilities**: Reusable JavaScript functions

## 🎨 Design System

### Color Palette
- **Primary**: #667eea (Blue gradient start)
- **Secondary**: #764ba2 (Purple gradient end)
- **Accent**: #45b7d1 (Light blue)
- **Accent Secondary**: #ff6b9d (Pink)
- **Success**: #28a745
- **Warning**: #ffc107
- **Danger**: #dc3545

### Typography
- **Font Family**: Poppins (Google Fonts)
- **Weights**: 300 (Light), 400 (Regular), 500 (Medium), 600 (Semi-bold), 700 (Bold)

### Components
- **Cards**: Rounded corners with subtle shadows
- **Buttons**: Gradient backgrounds with hover effects
- **Status Indicators**: Color-coded with animations
- **Glass Effects**: Backdrop blur for overlays

## 🔧 Getting Started

### Prerequisites
- Modern web browser (Chrome, Firefox, Safari, Edge)
- Ice Cream Store backend services running:
  - Data Service (Port 5432)
  - Auth Service (Port 8081)
  - Orders Service (Port 8083)

### Quick Start
1. **Start Backend Services**:
   ```bash
   # From the root directory
   make fresh
   ```

2. **Open Client**:
   - Navigate to `client/index.html` in your browser
   - Or serve via HTTP server for development

3. **Login**:
   - Click "Authentication" service
   - Use credentials: `admin` / `admin123`

### Development Server
For development, you can use any HTTP server:

```bash
# Python
python -m http.server 8000

# Node.js (http-server)
npx http-server

# PHP
php -S localhost:8000
```

Then visit: `http://localhost:8000`

## 📱 Service Navigation

### Main Entry Point
- **URL**: `/index.html`
- **Purpose**: Service selector and system overview
- **Features**: Real-time service status, quick actions

### Authentication Service
- **Login**: `/auth/login.html`
- **Dashboard**: `/auth/dashboard.html`
- **Features**: JWT authentication, API testing, user management

### Orders Service
- **Main**: `/orders/index.html`
- **Features**: Order CRUD, filtering, statistics, API testing

## 🔒 Security Features

### Authentication Flow
1. **Login**: JWT token obtained from auth service
2. **Storage**: Token stored in session/local storage
3. **Validation**: Automatic token validation on route changes
4. **Refresh**: Silent token refresh when needed
5. **Logout**: Secure token invalidation

### Route Protection
- **Public Routes**: Main dashboard, login page
- **Protected Routes**: All service-specific pages
- **Auto-redirect**: Unauthorized users redirected to login

## 🎯 API Integration

### Service Endpoints
Each service UI integrates with its respective backend:

- **Auth Service**: `http://localhost:8081/api/v1/auth/`
- **Orders Service**: `http://localhost:8083/api/v1/orders/`

### Error Handling
- **Network Errors**: Automatic retry with user feedback
- **Auth Errors**: Redirect to login on token expiration
- **Validation**: Client-side form validation before API calls

## 📊 Features by Service

### Orders Management
- ✅ **List Orders**: Paginated order listing with filters
- ✅ **Create Orders**: Multi-item order creation form
- ✅ **View Details**: Detailed order information
- ✅ **Update Status**: Order workflow management
- ✅ **Statistics**: Order analytics and insights
- ✅ **Real-time Filters**: Status, date, and text search

### Authentication
- ✅ **Secure Login**: JWT-based authentication
- ✅ **Profile Management**: User profile viewing
- ✅ **API Testing**: Built-in endpoint testing
- ✅ **Token Management**: Automatic token handling

## 🛠️ Development

### Adding New Services
1. **Create Service Directory**: `/client/[service-name]/`
2. **Add Main Page**: `index.html` with navigation
3. **Create Service Logic**: `[service-name].js`
4. **Update Navigation**: Add routes to `shared/js/navigation.js`
5. **Update Main Dashboard**: Add service card to `/index.html`

### Code Structure
- **HTML**: Semantic, accessible markup
- **CSS**: BEM methodology with CSS custom properties
- **JavaScript**: ES6+ with class-based architecture
- **Bootstrap**: Utility-first approach with custom components

### Best Practices
- **Responsive First**: Mobile-first design approach
- **Accessibility**: ARIA labels and semantic HTML
- **Performance**: Lazy loading and optimized assets
- **Security**: XSS protection and secure token handling

## 🔍 Troubleshooting

### Common Issues

**Service Not Loading**
- Check backend service is running
- Verify correct port numbers
- Check browser console for errors

**Authentication Failed**
- Ensure auth service is running on port 8081
- Check credentials: `admin` / `admin123`
- Clear browser storage and retry

**API Errors**
- Verify backend services are healthy
- Check network connectivity
- Review browser console for detailed errors

### System Status
Use the main dashboard to check:
- ✅ Service availability
- ✅ Database connectivity
- ✅ API endpoint health

## 📈 Future Enhancements

### Planned Features
- 🔄 **Gateway Integration**: Direct API gateway routing
- 📊 **Advanced Analytics**: Charts and detailed reporting
- 🔔 **Real-time Notifications**: WebSocket integration
- 👥 **Multi-user Support**: Role-based access control
- 🎨 **Theme Support**: Dark/light mode toggle
- 📱 **PWA Support**: Offline functionality

### Service Expansion
- **Inventory Service**: Stock management UI
- **Customer Service**: Customer relationship management
- **Reporting Service**: Advanced analytics dashboard
- **Payment Service**: Payment processing interface

## 🤝 Contributing

1. **Follow Structure**: Use service-oriented organization
2. **Shared Components**: Utilize existing shared resources
3. **Consistent Styling**: Follow established design system
4. **Security First**: Implement proper authentication checks
5. **Documentation**: Update README for new features

## 📄 License

This project is part of the Ice Cream Store management system. See the main project license for details.

---

**🍦 Ice Cream Store Client - Sweet Management Made Simple! 🍦** 