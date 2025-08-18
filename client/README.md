# Billy Wu - Client

A modern SolidJS frontend for Billy Wu, a collaborative storytelling app where friends and classmates take turns writing stories together.

## Features

- **Modern SolidJS**: Fast, reactive UI with SolidJS
- **TypeScript**: Full type safety throughout
- **SCSS Modules**: Organized styling with CSS modules and SCSS
- **Design System**: Comprehensive color palette and component system
- **Authentication**: JWT-based auth with context management
- **Real-time**: WebSocket integration for live collaboration
- **Responsive**: Mobile-first design with clean breakpoints

## Getting Started

### Prerequisites

- Node.js 22+ (see `.nvmrc`)
- npm, pnpm, or yarn

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   npm install
   # or
   pnpm install
   ```
3. Copy environment variables:
   ```bash
   cp .env.example .env
   ```
4. Update `.env` with your backend URL:
   ```
   VITE_API_URL=http://localhost:8280
   VITE_WS_URL=ws://localhost:8280
   VITE_ENV=local
   ```

### Development Commands

Start development server:

```bash
npm run dev
# or
npm start
```

Build for production:

```bash
npm run build
```

Preview production build:

```bash
npm run serve
```

### Project Structure

```
client/
├── src/
│   ├── components/
│   │   ├── common/        # Reusable UI components
│   │   └── layout/        # Layout components (Navbar, etc.)
│   ├── context/           # React-style contexts (Auth, WebSocket)
│   ├── pages/             # Page components
│   ├── services/          # API services and utilities
│   ├── styles/            # Global styles and design system
│   │   ├── _colors.scss   # Color palette
│   │   ├── _variables.scss # Design tokens
│   │   ├── _mixins.scss   # SCSS mixins
│   │   └── global.scss    # Global styles
│   └── types/             # TypeScript type definitions
├── index.html             # HTML entry point
├── vite.config.ts         # Vite configuration
└── tsconfig.json          # TypeScript configuration
```

### Design System

The app uses a comprehensive design system with:

- **Colors**: Purple/cyan theme with semantic color tokens
- **Typography**: Consistent font scales and weights
- **Spacing**: 4px base unit with consistent spacing scale
- **Components**: Reusable Button, TextInput, and layout components
- **Responsive**: Mobile-first with 6 breakpoint system

#### Using the Design System

```scss
// Import design tokens
@use "@styles/variables" as *;
@use "@styles/mixins" as *;
@use "@styles/colors" as *;

.my-component {
  padding: $spacing-md;
  background: $bg-primary;
  border-radius: $border-radius-md;

  @include breakpoint(md) {
    padding: $spacing-lg;
  }
}
```

### Path Aliases

The project uses TypeScript path aliases:

- `@styles/*` → `src/styles/*`
- `@components/*` → `src/components/*`
- `@layout/*` → `src/components/layout/*`
- `@pages/*` → `src/pages/*`
- `@services/*` → `src/services/*`
- `@context/*` → `src/context/*`

### State Management

- **SolidJS Stores**: For local component state
- **Context API**: For auth state and WebSocket connections
- **TanStack Query**: For server state management

### Authentication Flow

1. User logs in via `/login` page
2. JWT token stored in HTTP-only cookie
3. `AuthContext` manages auth state globally
4. Protected routes check auth status
5. WebSocket connection established when authenticated

### Styling Guidelines

- Use CSS Modules (`.module.scss`) for component styles
- Follow mobile-first responsive design
- Use design tokens from `_variables.scss` and `_colors.scss`
- Leverage mixins for common patterns
- Import shared resources at the top of each SCSS file

### API Integration

The app connects to the Go backend via:

- Axios HTTP client with interceptors
- TanStack Query for caching and mutations
- WebSocket connection for real-time features

### Development

- Hot reload enabled via Vite
- TypeScript strict mode
- ESLint configuration for code quality
- SCSS preprocessing with automatic imports
