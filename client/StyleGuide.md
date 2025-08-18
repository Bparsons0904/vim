# Style Guide: SolidJS Project

## Table of Contents

1. [Core Principles](#core-principles)
2. [File Organization](#file-organization)
3. [CSS Architecture](#css-architecture)
4. [Design Tokens](#design-tokens)
5. [Responsive Design](#responsive-design)
6. [Component Styling](#component-styling)
7. [Best Practices](#best-practices)

---

## Core Principles

Our styling approach is built on these core principles:

- **Mobile-first development**: Start with mobile styles and add complexity for larger screens
- **Modular architecture**: Use CSS modules for component encapsulation
- **Consistent design language**: Leverage variables and mixins for consistency
- **Maintainable, scalable code**: Follow consistent patterns for long-term maintainability

---

## File Organization

```
/src
  /styles
    _reset.scss       # Global CSS reset
    _variables.scss   # Design tokens and variables
    _colors.scss      # Color system
    _mixins.scss      # Reusable SCSS mixins
    global.scss       # Global styles (imports the above)
  /components
    /ComponentName
      ComponentName.tsx
      ComponentName.module.scss
```

---

## CSS Architecture

### CSS Modules

- Use `.module.scss` extension for component-specific styles
- Reference styles with the `styles` import:

  ```tsx
  import styles from "./ComponentName.module.scss";

  return <div className={styles.container}>...</div>;
  ```

### Importing Shared Resources

Import shared resources at the top of each SCSS module:

```scss
@use "../../styles/variables";
@use "../../styles/mixins";
@use "../../styles/colors";
```

---

## Design Tokens

### Using Variables

Always use variables for consistent values:

```scss
// WRONG
.element {
  font-size: 1rem;
  padding: 16px;
  color: #333;
}

// RIGHT
.element {
  font-size: variables.$font-size-md;
  padding: variables.$spacing-md;
  color: colors.$text-default;
}
```

### Color System

- Base colors: Use the raw color variables (`$color-primary-500`)
- Semantic colors: Use the functional color variables (`$bg-primary`, `$text-link`)

### Spacing

Use the spacing scale for all margin, padding, and layout-related measurements:

```scss
padding: variables.$spacing-md variables.$spacing-lg;
```

### Typography

Use predefined typography variables for consistency:

```scss
font-size: variables.$font-size-lg;
font-weight: variables.$font-weight-bold;
line-height: variables.$line-height-tight;
```

---

## Responsive Design

### Mobile-First Approach

1. Default styles target mobile devices
2. Use breakpoint mixins to add styles for larger screens
3. Nest breakpoints within component selectors

### Implementing Breakpoints

```scss
.component {
  // Mobile styles (default)
  display: flex;
  flex-direction: column;

  // Tablet and up
  @include mixins.breakpoint(md) {
    flex-direction: row;
  }

  // Desktop and up
  @include mixins.breakpoint(lg) {
    max-width: 1200px;
  }
}
```

### Responsive Values

For responsive property values, use the breakpoint mixin:

```scss
.heading {
  font-size: variables.$font-size-xl;

  @include mixins.breakpoint(md) {
    font-size: variables.$font-size-2xl;
  }

  @include mixins.breakpoint(lg) {
    font-size: variables.$font-size-3xl;
  }
}
```

---

## Component Styling

### Class Naming

- Use camelCase for class names in CSS modules
- Be descriptive but concise
- For related elements, prefix with the parent component name:
  ```scss
  .navbar {
  }
  .navbarMenu {
  }
  .navbarItem {
  }
  ```

### Component Structure

Structure your component styles in a logical order:

1. Component container
2. Layout/positioning styles
3. Child components
4. States and variants
5. Responsive adjustments

### Example Component Style

```scss
// NavBar.module.scss
@use "../../styles/variables";
@use "../../styles/mixins";
@use "../../styles/colors";

.navbar {
  // Base properties
  background-color: colors.$color-primary-900;
  color: colors.$text-inverse;
  padding: variables.$spacing-md 0;

  // Responsive adjustments
  @include mixins.breakpoint(md) {
    padding: variables.$spacing-lg 0;
  }
}

.navbarContainer {
  // Mobile layout (default)
  display: flex;
  flex-direction: column;

  // Larger screen layout
  @include mixins.breakpoint(md) {
    flex-direction: row;
    justify-content: space-between;
  }
}

// Additional component elements...
```

---

## Best Practices

### Using Mixins

Use mixins for repeated patterns:

```scss
.card {
  @include mixins.card;
}

.container {
  @include mixins.container;
}

.heading {
  @include mixins.heading-2;
}
```

### Adopting Common Layouts

Use the flex and grid mixins for common layouts:

```scss
.layout {
  @include mixins.flex(row, space-between, center);

  @include mixins.breakpoint(md) {
    @include mixins.grid(3, variables.$spacing-lg);
  }
}
```

### Media Queries

- Always use the breakpoint mixin for media queries
- Never use `max-width` media queries (desktop-first)
- Test responsive layouts at various viewport sizes

### Accessibility

- Use the appropriate mixins for accessibility:

  ```scss
  .visuallyHidden {
    @include mixins.visually-hidden;
  }

  .focusable {
    @include mixins.focus-ring;
  }
  ```
