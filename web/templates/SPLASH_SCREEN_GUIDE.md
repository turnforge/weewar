# Splash Screen System

The splash screen system provides an instant-loading overlay that displays before JavaScript loads, with progress tracking support.

## Features

- **Instant loading**: Renders with pure HTML/CSS before any JavaScript
- **Translucent backdrop**: 95% opacity with blur effect
- **Progress bar**: Updateable percentage display
- **Customizable**: Per-page custom splash screens via templates
- **Theme-aware**: Supports light/dark mode

## Basic Usage

### 1. Using the Default Splash Screen

The default splash screen is automatically included in all pages that inherit from `BasePage.html`. No additional setup required.

**Server-side** (Go template data):
```go
data := map[string]interface{}{
    "SplashTitle":   "Loading Dashboard",
    "SplashMessage": "Fetching your data...",
}
```

**Client-side** (TypeScript):
```typescript
import { SplashScreen } from './lib/SplashScreen';

// Update progress
SplashScreen.updateProgress(50);

// Update message
SplashScreen.updateMessage("Loading complete!", "Rendering interface...");

// Update both at once
SplashScreen.update({
    title: "Almost there",
    message: "Finalizing...",
    progress: 90
});

// Dismiss when ready
SplashScreen.dismiss();
```

### 2. Creating a Custom Splash Screen

Create a new template file (e.g., `MyPageSplash.html`):

```html
<!-- templates/MyPageSplash.html -->
{{ define "MyPageSplash" }}
<div class="flex flex-col items-center space-y-8">
  <!-- Your custom loading animation -->
  <div class="relative w-24 h-24">
    <!-- Custom spinner or icon -->
  </div>

  <!-- Message area (use data attributes for JS updates) -->
  <div class="text-center space-y-3">
    <h2 class="text-2xl font-bold" data-splash-title>Custom Loading</h2>
    <p class="text-base" data-splash-message>Please wait...</p>
  </div>

  <!-- Progress bar (required attributes) -->
  <div class="w-96 max-w-full">
    <div class="h-3 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
      <div
        data-splash-progress-bar
        class="h-full bg-blue-600 transition-all duration-300"
        style="width: 0%"
      ></div>
    </div>
    <p class="text-sm text-center mt-2" data-splash-progress-text>0%</p>
  </div>
</div>
{{ end }}
```

### 3. Using a Custom Splash Screen in Your Page

**Option A: Inline custom splash** (simpler, recommended):

```html
<!-- templates/MyPage.html -->
{{# include "BasePage.html" #}}

{{ define "SplashScreenContent" }}
<div class="flex flex-col items-center space-y-6">
  <!-- Your custom splash content with required data attributes -->
  <h2 data-splash-title>My Custom Loading</h2>
  <p data-splash-message>Loading...</p>
  <div data-splash-progress-bar style="width: 0%"></div>
  <p data-splash-progress-text>0%</p>
</div>
{{ end }}

{{ define "BodySection" }}
  <!-- Your page content -->
{{ end }}

{{ define "MyPage" }}
{{ template "BasePage" . }}
{{ end }}
```

**Option B: Separate template file** (for reusable splash screens):

Create `MyPageSplash.html`:
```html
{{ define "MyPageSplash" }}
<div class="flex flex-col items-center space-y-6">
  <!-- Your custom splash content -->
</div>
{{ end }}
```

Then in your page:
```html
<!-- templates/MyPage.html -->
{{# include "BasePage.html" #}}
{{# include "MyPageSplash.html" #}}

{{ define "SplashScreenContent" }}
  {{ template "MyPageSplash" . }}
{{ end }}

{{ define "MyPage" }}
{{ template "BasePage" . }}
{{ end }}
```

**Note**: If you encounter "multiple definition" errors, you may need to adjust the template include order or inline the splash content directly.

### 4. Disabling the Splash Screen

If a page doesn't need a splash screen:

**Server-side**:
```go
data := map[string]interface{}{
    "DisableSplashScreen": true,
}
```

## Required Data Attributes

For JavaScript updates to work, your custom splash screens must include these attributes:

- `data-splash-title`: Element to update with title text
- `data-splash-message`: Element to update with message text
- `data-splash-progress-bar`: Element to animate width (0-100%)
- `data-splash-progress-text`: Element to show percentage text

## TypeScript API

### `SplashScreen.dismiss()`
Fade out and remove the splash screen. Safe to call multiple times.

### `SplashScreen.updateProgress(percent: number)`
Update the progress bar (0-100).

### `SplashScreen.updateMessage(title?: string, message?: string)`
Update the title and/or message text.

### `SplashScreen.update(options)`
Update multiple properties at once:
```typescript
SplashScreen.update({
    title: "New title",
    message: "New message",
    progress: 75
});
```

### `SplashScreen.isVisible(): boolean`
Check if splash screen is still showing.

## Example: WorldEditorPage

See `WorldEditorSplash.html` for a full custom implementation with:
- Custom loading animation (map icon)
- Gradient progress bar
- Pulsing loading dots
- Larger text and spacing

## Best Practices

1. **Start with low progress**: Initialize at 0% and increment as loading progresses
2. **Update incrementally**: Show progress at key milestones (25%, 50%, 75%, 100%)
3. **Dismiss on ready**: Call `dismiss()` in your page's `onReady()` or similar lifecycle method
4. **Use meaningful messages**: Tell users what's actually happening ("Loading map data...", "Initializing engine...")
5. **Test both themes**: Verify your custom splash works in light and dark mode
