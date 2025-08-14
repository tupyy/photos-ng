# Photos NG - Frontend

A modern React-based photo management application with TypeScript, Redux, and Tailwind CSS.

## 🚀 Tech Stack

- **React 18** with TypeScript
- **Redux Toolkit** for state management
- **Tailwind CSS** for styling
- **React Router DOM** for navigation
- **OpenAPI Generator** for API client generation
- **Webpack** for custom build configuration
- **Flowbite** for UI components
- **Axios** for HTTP requests

## 📁 Project Structure

```
ui/
├── src/
│   ├── generated/          # Auto-generated API clients and types
│   ├── pages/              # Page components organized by feature
│   │   ├── albums/         # Albums page and components
│   │   └── timeline/       # Timeline page and components
│   ├── shared/             # Shared application code
│   │   ├── components/     # Reusable UI components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── layout/         # Layout components (Navbar, Footer, etc.)
│   │   ├── reducers/       # Redux slices and state management
│   │   └── api/            # API configuration
│   ├── App.tsx             # Main application component
│   └── index.tsx           # Application entry point
├── webpack.*.js            # Custom Webpack configurations
├── tailwind.config.js      # Tailwind CSS configuration
└── tsconfig.json           # TypeScript configuration
```

## 🛠️ Development Setup

### Prerequisites

- Node.js (v16 or later)
- npm

### Installation

```bash
# Install dependencies
npm install

# Generate API client from OpenAPI spec
npm run generate:api
```

### Available Scripts

| Command | Description |
|---------|-------------|
| `npm run start:dev` | Start development server with hot reload (local proxy) |
| `npm run start:local` | Start development server with local backend proxy |
| `npm run start:remote` | Start development server with remote backend proxy |
| `npm run build` | Build for production |
| `npm run clean` | Clean build artifacts |
| `npm run generate:api` | Generate TypeScript API client from OpenAPI spec |
| `npm run css:build` | Build Tailwind CSS |
| `npm run css:watch` | Watch and rebuild Tailwind CSS |

### Development Server

The development server supports two proxy modes:

#### Local Development (Default)
```bash
npm run start:dev
# or explicitly
npm run start:local
```
Proxies API requests to `http://localhost:8080` (local backend)

#### Remote Development
```bash
npm run start:remote
```
Proxies API requests to `https://photos.tls.tupangiu.ro` (remote backend)

The application will be available at [http://localhost:3000](http://localhost:3000).

#### Proxy Configuration

You can also control the proxy mode using environment variables:

```bash
# Local backend (default)
PROXY_MODE=local npm run start:dev

# Remote backend
PROXY_MODE=remote npm run start:dev
```

## 🎨 Features

### UI/UX
- **Dark/Light Theme** - Toggle between themes with persistent preference
- **Responsive Design** - Mobile-first approach with Tailwind CSS
- **Modern Layout** - Clean navbar with action menu and progress indicators

### State Management
- **Redux Toolkit** - Centralized state management
- **Async Operations** - Proper handling of API calls with loading states
- **Type Safety** - Full TypeScript integration

### Navigation
- **React Router** - Client-side routing
- **Timeline** - Main photo timeline view (root route)
- **Albums** - Album management and viewing

### API Integration
- **OpenAPI Generated Client** - Type-safe API calls
- **Automatic Sync** - Background synchronization with progress tracking
- **Error Handling** - Proper error states and user feedback

## 🔧 Configuration

### API Configuration

The API client is configured in `src/shared/api/apiConfig.ts`:

```typescript
// Development: Uses proxy (local or remote based on PROXY_MODE)
// Production: Uses REACT_APP_API_URL environment variable
```

#### Proxy Modes

- **Local Mode** (`PROXY_MODE=local`): Proxies to `http://localhost:8080`
  - Used for local backend development
  - Default mode for development
  
- **Remote Mode** (`PROXY_MODE=remote`): Proxies to `https://photos.tls.tupangiu.ro`
  - Used for testing against production backend
  - Enables SSL verification
  - Useful for frontend-only development

### Theme Configuration

Dark mode is implemented using:
- Tailwind CSS `dark:` prefixes
- React Context for theme state
- Local storage for persistence

### Webpack Configuration

Custom Webpack setup with:
- **Development**: Hot reload, source maps, proxy configuration
- **Production**: Minification, optimization, chunking

## 📱 Responsive Design

The application is fully responsive with breakpoints:
- **Mobile**: < 768px (footer hidden for better UX)
- **Tablet**: 768px - 1024px
- **Desktop**: > 1024px

## 🔄 State Management

### Redux Slices

- **Albums**: Album data and operations
- **Media**: Media/photo data and operations  
- **Timeline**: Timeline view state
- **Sync**: Background synchronization state

### Custom Hooks

- `useSync()` - Sync operations and state
- `useApi()` - API interaction helpers

## 🌐 API Integration

The frontend communicates with the backend API through:

1. **Generated Client**: TypeScript client generated from OpenAPI spec
2. **Base Path**: `/api/v1` (handled by development proxy)
3. **Authentication**: Ready for future auth implementation
4. **Error Handling**: Consistent error management across all API calls

## 🚦 Development Workflow

### Local Development
1. **Start Backend**: Ensure the Go backend is running on port 8080
2. **Start Frontend**: Run `npm run start:local` 
3. **API Changes**: Run `npm run generate:api` after OpenAPI spec updates
4. **Styling**: Tailwind classes are processed automatically

### Remote Development (Frontend Only)
1. **Start Frontend**: Run `npm run start:remote`
2. **No Backend Required**: Uses remote production backend
3. **API Changes**: Run `npm run generate:api` after OpenAPI spec updates
4. **Styling**: Tailwind classes are processed automatically

## 📦 Build Process

The build process includes:

1. **TypeScript Compilation**: Full type checking
2. **CSS Processing**: Tailwind CSS compilation and purging
3. **Bundle Optimization**: Code splitting and minification
4. **Asset Optimization**: Image and resource optimization

## 🔍 Troubleshooting

### Common Issues

**API Calls Failing**
- **Local Mode**: Ensure backend is running on port 8080
- **Remote Mode**: Check internet connection and remote server status
- Check proxy configuration in Webpack dev config
- Verify the correct proxy mode is being used (`npm run start:local` vs `npm run start:remote`)

**Styles Not Loading**
- Run `npm run css:build` to regenerate Tailwind CSS
- Check Tailwind config for proper content paths

**Type Errors**
- Regenerate API client: `npm run generate:api`
- Check TypeScript configuration and path aliases

### Debug Mode

Enable detailed logging by setting:
```bash
DEBUG=true npm run start:dev
```

---

For more detailed information about specific features, see:
- [Redux Integration](./README-REDUX.md)
- [Generated API Documentation](./src/generated/README.md)