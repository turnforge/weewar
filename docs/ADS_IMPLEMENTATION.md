# Ads Implementation Plan

**Created**: January 12, 2026
**Status**: Ready for Implementation
**Reference**: [MONETIZATION.md](../MONETIZATION.md)

---

## Overview

This document provides the detailed technical implementation plan for Phase 1 ads with feature flags.

## 1. Feature Flags

### 1.1 New Flags in WeewarApp

Add to `web/server/webapp.go`:

```go
type WeewarApp struct {
    // ... existing fields ...

    // App config
    HideGames  bool
    HideWorlds bool

    // Ad config (NEW)
    AdsEnabled        bool   // Master switch: WEEWAR_ADS_ENABLED (default: true)
    AdsFooterEnabled  bool   // Footer banner: WEEWAR_ADS_FOOTER (default: true)
    AdsHomeEnabled    bool   // Homepage mid-section: WEEWAR_ADS_HOME (default: true)
    AdsListingEnabled bool   // Listing pages: WEEWAR_ADS_LISTING (default: true)
    AdNetworkId       string // Google AdSense publisher ID
}
```

### 1.2 Initialization

```go
// In NewWeewarApp()
weewarApp = &WeewarApp{
    // ... existing ...

    // Ads default to enabled, can be disabled per-placement
    AdsEnabled:        os.Getenv("WEEWAR_ADS_ENABLED") != "false",
    AdsFooterEnabled:  os.Getenv("WEEWAR_ADS_FOOTER") != "false",
    AdsHomeEnabled:    os.Getenv("WEEWAR_ADS_HOME") != "false",
    AdsListingEnabled: os.Getenv("WEEWAR_ADS_LISTING") != "false",
    AdNetworkId:       os.Getenv("WEEWAR_AD_NETWORK_ID"),
}
```

### 1.3 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEEWAR_ADS_ENABLED` | `true` | Master switch for all ads |
| `WEEWAR_ADS_FOOTER` | `true` | Footer banner ads |
| `WEEWAR_ADS_HOME` | `true` | Homepage mid-section ads |
| `WEEWAR_ADS_LISTING` | `true` | Game/World listing page ads |
| `WEEWAR_AD_NETWORK_ID` | (empty) | Google AdSense publisher ID (ca-pub-XXXXX) |

---

## 2. Template Components

### 2.1 AdSlot Component Template

Create `web/templates/components/AdSlot.html`:

```html
{{/*
  AdSlot - Reusable ad container component

  Parameters (via .):
    - SlotId: string - Unique identifier for this ad slot
    - Size: string - "leaderboard" | "mrec" | "skyscraper" | "mobile-banner"
    - Position: string - "footer" | "home-mid" | "listing" | "sidebar"

  Usage:
    {{ template "AdSlot" (dict "SlotId" "footer-1" "Size" "leaderboard" "Position" "footer") }}
*/}}

{{ define "AdSlot" }}
{{ if and (Ctx).AdsEnabled (Ctx).AdNetworkId }}
<div id="ad-{{ .SlotId }}"
     class="ad-container ad-{{ .Size }}"
     data-ad-position="{{ .Position }}"
     data-ad-slot="{{ .SlotId }}">
    <!-- Ad will be injected here by ad network script -->
    <ins class="adsbygoogle"
         style="display:block"
         data-ad-client="{{ (Ctx).AdNetworkId }}"
         data-ad-slot="{{ .SlotId }}"
         data-ad-format="auto"
         data-full-width-responsive="true"></ins>
</div>
{{ else if not (Ctx).AdNetworkId }}
<!-- Ad slot placeholder (no network ID configured) -->
<div id="ad-{{ .SlotId }}"
     class="ad-container ad-{{ .Size }} ad-placeholder"
     data-ad-position="{{ .Position }}">
    <span class="text-xs text-gray-400">Ad Space</span>
</div>
{{ end }}
{{ end }}
```

### 2.2 AdScript Component

Create `web/templates/components/AdScript.html`:

```html
{{/*
  AdScript - Google AdSense script loader
  Include once in the <head> of pages that show ads
*/}}

{{ define "AdScript" }}
{{ if and (Ctx).AdsEnabled (Ctx).AdNetworkId }}
<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client={{ (Ctx).AdNetworkId }}"
        crossorigin="anonymous"></script>
{{ end }}
{{ end }}
```

### 2.3 Footer Banner Ad

Update `web/templates/Footer.html`:

```html
{{ define "Footer" }}
<footer class="border-t border-gray-200 dark:border-gray-700 mt-auto">
    <!-- Footer Ad Slot -->
    {{ if and (Ctx).AdsEnabled (Ctx).AdsFooterEnabled }}
    <div class="ad-footer-wrapper bg-gray-50 dark:bg-gray-900 py-4">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-center">
            <!-- Desktop: Leaderboard -->
            <div class="hidden md:block">
                {{ template "AdSlot" (dict "SlotId" "footer-leaderboard" "Size" "leaderboard" "Position" "footer") }}
            </div>
            <!-- Mobile: Banner -->
            <div class="md:hidden">
                {{ template "AdSlot" (dict "SlotId" "footer-mobile" "Size" "mobile-banner" "Position" "footer") }}
            </div>
        </div>
    </div>
    {{ end }}

    <!-- Existing footer content -->
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="flex flex-col sm:flex-row justify-between items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <div>
                {{ .Header.AppName }}
            </div>
            <div class="flex flex-wrap justify-center gap-x-6 gap-y-2">
                <a href="/about/" class="hover:text-gray-700 dark:hover:text-gray-300">About</a>
                <a href="/contact/" class="hover:text-gray-700 dark:hover:text-gray-300">Contact</a>
                <a href="/privacy/" class="hover:text-gray-700 dark:hover:text-gray-300">Privacy Policy</a>
                <a href="/terms/" class="hover:text-gray-700 dark:hover:text-gray-300">Terms of Service</a>
            </div>
        </div>
    </div>
</footer>
{{ end }}
```

### 2.4 HomePage Mid-Section Ad

Update `web/templates/HomePage.html` - Add between Quick Actions and Recent Activity:

```html
<!-- After Quick Action Buttons (line ~68), before Recent Activity Section (line ~72) -->

<!-- Homepage Mid-Section Ad -->
{{ if and (Ctx).AdsEnabled (Ctx).AdsHomeEnabled }}
<div class="ad-home-mid-wrapper my-8">
    <div class="flex justify-center">
        {{ template "AdSlot" (dict "SlotId" "home-mid" "Size" "mrec" "Position" "home-mid") }}
    </div>
</div>
{{ end }}
```

---

## 3. CSS Styles

### 3.1 Ad Container Styles

Create `web/src/styles/ads.css`:

```css
/* ==========================================================================
   Ad Container Styles
   ========================================================================== */

/* Base ad container */
.ad-container {
    display: flex;
    justify-content: center;
    align-items: center;
    overflow: hidden;
}

/* Size variants */
.ad-leaderboard {
    width: 728px;
    max-width: 100%;
    height: 90px;
}

.ad-mrec {
    width: 300px;
    height: 250px;
}

.ad-skyscraper {
    width: 160px;
    height: 600px;
}

.ad-mobile-banner {
    width: 320px;
    max-width: 100%;
    height: 50px;
}

/* Placeholder styling (when no ad network configured) */
.ad-placeholder {
    background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
    border: 1px dashed #d1d5db;
    border-radius: 4px;
}

.dark .ad-placeholder {
    background: linear-gradient(135deg, #1f2937 0%, #111827 100%);
    border-color: #374151;
}

/* Wrapper styling */
.ad-footer-wrapper {
    border-top: 1px solid #e5e7eb;
}

.dark .ad-footer-wrapper {
    border-top-color: #374151;
}

/* Responsive hiding */
@media (max-width: 767px) {
    .ad-leaderboard {
        display: none;
    }
}

@media (min-width: 768px) {
    .ad-mobile-banner {
        display: none;
    }
}

/* Listing page native ads */
.ad-listing-card {
    /* Match game/world card styling */
    background: white;
    border-radius: 0.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
}

.dark .ad-listing-card {
    background: #1f2937;
}
```

### 3.2 Import in Main CSS

Add to `web/src/styles/main.css` or equivalent:

```css
@import './ads.css';
```

---

## 4. CSP Updates

### 4.1 Security Headers Update

Update `web/server/securityheaders.go`:

```go
func (m *SecurityHeadersMiddleware) buildCSP() string {
    // Base CSP directives
    defaultSrc := "'self'"

    // Script sources
    scriptSrc := "'self' https://unpkg.com"
    if m.IsDevelopment {
        scriptSrc += " 'unsafe-inline' 'unsafe-eval'"
    }
    // Add Google AdSense
    scriptSrc += " https://pagead2.googlesyndication.com https://www.googletagservices.com https://adservice.google.com"

    // Style sources
    styleSrc := "'self' 'unsafe-inline'"  // Tailwind requires unsafe-inline

    // Image sources
    imgSrc := "'self' data: blob: https://pagead2.googlesyndication.com https://www.google.com"

    // Font sources
    fontSrc := "'self'"

    // Connect sources (for AJAX/fetch)
    connectSrc := "'self'"
    if m.IsDevelopment {
        connectSrc += " ws: wss:"
    } else {
        connectSrc += " wss:"
    }
    connectSrc += " https://pagead2.googlesyndication.com"

    // Frame sources (for ad iframes)
    frameSrc := "https://googleads.g.doubleclick.net https://tpc.googlesyndication.com https://www.google.com"

    return fmt.Sprintf(
        "default-src %s; script-src %s; style-src %s; img-src %s; font-src %s; connect-src %s; frame-src %s; frame-ancestors 'none'",
        defaultSrc, scriptSrc, styleSrc, imgSrc, fontSrc, connectSrc, frameSrc,
    )
}
```

---

## 5. Implementation Steps

### Step 1: Add Feature Flags (webapp.go)

```go
// Add to WeewarApp struct
AdsEnabled        bool
AdsFooterEnabled  bool
AdsHomeEnabled    bool
AdsListingEnabled bool
AdNetworkId       string

// Add to NewWeewarApp initialization
AdsEnabled:        os.Getenv("WEEWAR_ADS_ENABLED") != "false",
AdsFooterEnabled:  os.Getenv("WEEWAR_ADS_FOOTER") != "false",
AdsHomeEnabled:    os.Getenv("WEEWAR_ADS_HOME") != "false",
AdsListingEnabled: os.Getenv("WEEWAR_ADS_LISTING") != "false",
AdNetworkId:       os.Getenv("WEEWAR_AD_NETWORK_ID"),
```

### Step 2: Create Ad Component Templates

1. Create `web/templates/components/AdSlot.html`
2. Create `web/templates/components/AdScript.html`

### Step 3: Update Page Templates

1. Update `Footer.html` with ad slot
2. Update `HomePage.html` with mid-section ad
3. Update `BasePage.html` to include AdScript in head (when ads enabled)

### Step 4: Add CSS Styles

1. Create `web/src/styles/ads.css`
2. Import in main stylesheet

### Step 5: Update CSP Headers

1. Modify `securityheaders.go` to allow ad network domains

### Step 6: Test

1. Test with `WEEWAR_ADS_ENABLED=false` (ads hidden)
2. Test with no `WEEWAR_AD_NETWORK_ID` (placeholder shown)
3. Test with valid network ID (ads load)
4. Test dark mode compatibility
5. Test mobile responsiveness

---

## 6. Testing Checklist

### Feature Flag Tests

- [ ] `WEEWAR_ADS_ENABLED=false` hides all ads
- [ ] `WEEWAR_ADS_FOOTER=false` hides only footer ads
- [ ] `WEEWAR_ADS_HOME=false` hides only homepage ads
- [ ] `WEEWAR_ADS_LISTING=false` hides only listing page ads
- [ ] Default (no env vars) shows all ads

### Visual Tests

- [ ] Footer ad displays correctly on desktop
- [ ] Footer ad displays correctly on mobile
- [ ] Homepage mid-section ad centered properly
- [ ] Dark mode: ad placeholders match theme
- [ ] No layout shift when ads load

### CSP Tests

- [ ] AdSense script loads without CSP errors
- [ ] Ad iframes load without CSP errors
- [ ] No console errors related to blocked resources

### Browser Tests

- [ ] Chrome
- [ ] Firefox
- [ ] Safari
- [ ] Mobile Safari
- [ ] Mobile Chrome

---

## 7. Future Enhancements (Phase 2+)

### Listing Page Native Ads

For GameListingPage/WorldListingPage, ads will be injected into the grid every N cards. This requires:

1. Modifying the EntityListing template to support ad injection
2. Creating a native ad card template that matches game/world card styling
3. Server-side logic to determine ad placement frequency

### Game End Screen Ads

Requires:
1. Victory/defeat modal component
2. Interstitial ad integration
3. Skip button after N seconds

### Rewarded Video Ads

Requires:
1. Rewarded video SDK integration (IronSource/Unity Ads)
2. Reward fulfillment logic (coins, undo, etc.)
3. UI for reward redemption

---

## 8. Revenue Tracking

### Metrics to Monitor

1. **Fill Rate**: % of ad requests that return an ad
2. **eCPM**: Effective cost per 1000 impressions
3. **Viewability**: % of ads that were actually seen
4. **Page RPM**: Revenue per 1000 page views

### Google AdSense Dashboard

- Set up Google AdSense account
- Add site verification
- Create ad units for each slot
- Monitor performance in AdSense dashboard

---

## Appendix: Environment Configuration Examples

### Development (.env.dev)

```bash
# Ads disabled in development by default
WEEWAR_ADS_ENABLED=false
WEEWAR_AD_NETWORK_ID=
```

### Staging (.env.staging)

```bash
# Ads enabled with test network ID
WEEWAR_ADS_ENABLED=true
WEEWAR_AD_NETWORK_ID=ca-pub-0000000000000000  # Test ID
```

### Production (.env.prod)

```bash
# Full ad configuration
WEEWAR_ADS_ENABLED=true
WEEWAR_ADS_FOOTER=true
WEEWAR_ADS_HOME=true
WEEWAR_ADS_LISTING=true
WEEWAR_AD_NETWORK_ID=ca-pub-XXXXXXXXXXXXXXXX  # Real AdSense ID
```
