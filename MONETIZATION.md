# WeeWar Monetization Strategy

**Created**: January 12, 2026
**Last Updated**: January 12, 2026
**Status**: Planning Phase
**Focus**: Ad-based monetization (pre-subscription)

---

## Executive Summary

This document outlines a phased ad monetization strategy for WeeWar, designed to maximize revenue while maintaining a quality user experience. The approach prioritizes non-intrusive placements and natural game flow interruption points.

**Key Principles**:
1. **Player experience first** - Ads should not disrupt active gameplay
2. **Natural pause points** - Leverage turn-based mechanics for ad placement
3. **Value exchange** - Rewarded ads offer tangible in-game benefits
4. **Progressive implementation** - Start simple, iterate based on data

---

## Competitive Landscape

### Direct Competitors

| Platform | Model | Ads? | Monthly Users | Notes |
|----------|-------|------|---------------|-------|
| **Board Game Arena** | Subscription ($5/mo) | No | 10M+ | Market leader, Asmodee-owned |
| **Tabletopia** | Freemium tiers | Minimal | 500K+ | Publisher revenue share |
| **Tabletop Simulator** | One-time ($20) | No | Steam-based | Creator marketplace coming |
| **Yucata** | Donations | No | 100K+ | Volunteer-run 20+ years |
| **BoardGameGeek** | Ads + GeekGold | Heavy | 17M visits/mo | Database/community only |

### Key Insight

**Nobody in the board game space is doing ads well.** BGA and Tabletopia went subscription-first. BGG has ads but dated UX and no actual gameplay. This creates an opportunity: WeeWar can be the **polished, ad-supported, free-to-play** alternative.

### Revenue Benchmarks (Browser Games)

| Metric | Low | Average | High |
|--------|-----|---------|------|
| Banner CPM | $0.50 | $2.00 | $5.00 |
| Interstitial CPM | $2.00 | $8.00 | $15.00 |
| Rewarded Video CPM | $5.00 | $15.00 | $30.00 |
| ARPDAU (Ad Revenue) | $0.01 | $0.05 | $0.15 |

---

## Phase 1: Foundation (Launch Ready)

**Goal**: Establish baseline ad infrastructure with minimal development effort.
**Timeline**: Can ship with launch
**Revenue Impact**: Low-Medium

### 1.1 Footer Banner Ads

**Placement**: All pages, below main content
**Format**: 728x90 leaderboard (desktop), 320x50 mobile banner
**Visibility**: Always visible, non-intrusive
**Implementation**: Simple ad container in Footer component

```html
<!-- web/templates/components/Footer.html -->
<footer>
  <div id="ad-footer" class="ad-container ad-leaderboard">
    <!-- Ad network script -->
  </div>
  <div class="footer-links">
    <!-- Existing footer links -->
  </div>
</footer>
```

**Estimated Revenue**: $0.50-2.00 CPM

### 1.2 Homepage Interstitial Zone

**Placement**: Between "Recent Games" and "Recent Worlds" sections
**Format**: 300x250 medium rectangle or native ad
**Visibility**: High traffic area, all authenticated users
**Context**: Non-gameplay, browsing state

```html
<!-- web/templates/HomePage.html -->
<section class="recent-games">...</section>

<div id="ad-homepage-mid" class="ad-container ad-mrec my-8">
  <!-- Ad network script -->
</div>

<section class="recent-worlds">...</section>
```

**Estimated Revenue**: $1.00-3.00 CPM

### 1.3 Listing Page Native Ads

**Placement**: GameListingPage, WorldListingPage - between cards
**Format**: Native ad styled like game/world card
**Frequency**: Every 6-8 cards
**Visibility**: Blends with content, less intrusive

**Estimated Revenue**: $2.00-5.00 CPM (native ads command premium)

### Phase 1 Technical Requirements

1. **Ad container components**: Reusable `AdSlot` component with size variants
2. **Ad network integration**: Google AdSense for initial launch (easiest setup)
3. **Dark mode support**: Ensure ads don't break dark theme
4. **Responsive sizing**: Mobile-specific ad sizes
5. **Ad blocker detection**: Optional, for analytics only (no gate)

---

## Phase 2: Game Integration (Post-Launch)

**Goal**: Integrate ads into natural game flow pause points.
**Timeline**: 2-4 weeks post-launch
**Revenue Impact**: Medium-High

### 2.1 Game End Screen (HIGH VALUE)

**Trigger**: When game concludes (victory/defeat)
**Format**: Full-screen interstitial or rewarded video
**User State**: Peak attention, completed activity
**Skip**: 5-second skip button

**Implementation Required**:
- Create GameEndModal component
- Victory/defeat detection in game state
- Interstitial ad integration

```typescript
// web/pages/game/GameEndScreen.ts
class GameEndScreen {
  show(result: GameResult) {
    // Show interstitial ad
    showInterstitialAd('game-end').then(() => {
      // Show victory/defeat UI
      this.renderResults(result);
    });
  }
}
```

**Estimated Revenue**: $8.00-15.00 CPM (captive audience)

### 2.2 Turn Transition Ads (MEDIUM VALUE)

**Trigger**: Between player turns (2-3 second natural pause)
**Format**: Brief interstitial (skippable after 2s) or banner expansion
**Frequency**: Every 3-5 turns (not every turn)
**Skip**: Auto-advance after 3 seconds

**User Experience Consideration**: Must be very brief. Turn-based games have natural pauses but players expect quick transitions.

```typescript
// Frequency control
const TURNS_BETWEEN_ADS = 4;
if (turnCounter % TURNS_BETWEEN_ADS === 0) {
  showBriefInterstitial('turn-transition');
}
```

**Estimated Revenue**: $5.00-10.00 CPM

### 2.3 Right Panel Ads (Desktop Only)

**Placement**: Below Turn Options panel in GameViewerPage
**Format**: 160x600 skyscraper or 300x250 medium rectangle
**Visibility**: Persistent during gameplay, low priority
**Context**: Doesn't block game scene

```html
<!-- GameViewerPage right panels -->
<div class="terrain-panel">...</div>
<div class="unit-panel">...</div>
<div class="turn-options-panel">...</div>
<div id="ad-game-sidebar" class="ad-container ad-skyscraper">
  <!-- Ad network script -->
</div>
```

**Estimated Revenue**: $1.00-3.00 CPM

---

## Phase 3: Rewarded Ads (Value Exchange)

**Goal**: Offer players tangible benefits for watching ads.
**Timeline**: 4-8 weeks post-launch
**Revenue Impact**: High (best eCPM)

### 3.1 Bonus Coins

**Trigger**: Player clicks "Watch Ad for +100 Coins"
**Format**: 15-30 second rewarded video
**Reward**: In-game currency bonus (configurable)
**Limit**: 3-5 per day per player

**UI Location**:
- Game setup screen (before game starts)
- End of turn (optional boost)
- Low coins warning state

```typescript
// Rewarded ad flow
async function watchAdForCoins() {
  const completed = await showRewardedAd('bonus-coins');
  if (completed) {
    await grantCoins(player, 100);
    showToast('You received 100 bonus coins!');
  }
}
```

**Estimated Revenue**: $15.00-30.00 CPM

### 3.2 Undo Move

**Trigger**: Player wants to undo last move
**Format**: 15-second rewarded video
**Reward**: Undo last action (normally not available)
**Limit**: 1 per game

**Implementation**:
- Add "Undo" button to game UI
- Undo requires watching ad OR premium status
- Server validates undo eligibility

**Estimated Revenue**: $15.00-25.00 CPM

### 3.3 Extra Time (Future - Async Games)

**Trigger**: Player running low on turn timer
**Format**: 15-second rewarded video
**Reward**: +24 hours turn time
**Context**: Only relevant for async multiplayer with timers

---

## Phase 4: Premium Lite (Ad-Free Tier)

**Goal**: Simple paid tier to remove ads.
**Timeline**: 2-3 months post-launch
**Revenue Impact**: Recurring revenue stream

### Pricing

| Tier | Price | Features |
|------|-------|----------|
| **Free** | $0 | Full game access, ads shown |
| **Supporter** | $3/month | Ad-free experience, supporter badge |
| **Premium** | $6/month | Ad-free + cosmetics + priority support |

### Implementation

1. **User flag**: `user.premium_status` in database
2. **Ad gating**: Check premium status before showing ads
3. **Payment**: Stripe integration for subscriptions
4. **Badge**: Visual indicator in game UI

```go
// Check before showing ad
func shouldShowAd(user *User) bool {
    return user.PremiumStatus == PremiumStatus_FREE
}
```

---

## Ad Network Recommendations

### Tier 1: Launch (Easiest Setup)

| Network | Best For | Setup Time |
|---------|----------|------------|
| **Google AdSense** | Banner/display | 1-2 days |
| **Google Ad Manager** | Direct deals + programmatic | 3-5 days |

### Tier 2: Optimization (Higher Revenue)

| Network | Best For | Notes |
|---------|----------|-------|
| **GameDistribution** | HTML5 games, rewarded video | Game-specific network |
| **IronSource** | Rewarded video, mediation | Good eCPM for games |
| **Unity Ads** | Rewarded video | Strong game focus |
| **AppLovin MAX** | Mediation platform | Optimizes across networks |

### Tier 3: Direct Sales (Highest Value)

| Approach | Best For | Notes |
|----------|----------|-------|
| **Direct sponsorships** | Homepage takeover | BoardGameGeek charges $700+ |
| **Publisher partnerships** | Promoted games/worlds | Revenue share model |

---

## Technical Implementation

**Detailed Implementation Plan**: See [docs/ADS_IMPLEMENTATION.md](./docs/ADS_IMPLEMENTATION.md)

### Feature Flags

| Variable | Default | Description |
|----------|---------|-------------|
| `LILBATTLE_ADS_ENABLED` | `true` | Master switch for all ads |
| `LILBATTLE_ADS_FOOTER` | `true` | Footer banner ads |
| `LILBATTLE_ADS_HOME` | `true` | Homepage mid-section ads |
| `LILBATTLE_ADS_LISTING` | `true` | Game/World listing page ads |
| `LILBATTLE_AD_NETWORK_ID` | (empty) | Google AdSense publisher ID (ca-pub-XXXXX) |

### Ad Container Component

```typescript
// web/src/components/AdSlot.ts
interface AdSlotProps {
  id: string;
  size: 'leaderboard' | 'mrec' | 'skyscraper' | 'mobile-banner';
  position: 'footer' | 'sidebar' | 'interstitial';
}

class AdSlot {
  constructor(private props: AdSlotProps) {}

  render(container: HTMLElement) {
    if (isPremiumUser()) return; // No ads for premium

    const slot = document.createElement('div');
    slot.id = `ad-${this.props.id}`;
    slot.className = `ad-container ad-${this.props.size}`;
    slot.dataset.adPosition = this.props.position;
    container.appendChild(slot);

    // Initialize ad network
    this.loadAd(slot);
  }
}
```

### CSS for Ad Containers

```css
/* web/src/styles/ads.css */
.ad-container {
  display: flex;
  justify-content: center;
  align-items: center;
  background: var(--ad-bg, #f5f5f5);
  min-height: 50px;
}

.dark .ad-container {
  --ad-bg: #1a1a1a;
}

.ad-leaderboard { width: 728px; height: 90px; }
.ad-mrec { width: 300px; height: 250px; }
.ad-skyscraper { width: 160px; height: 600px; }
.ad-mobile-banner { width: 320px; height: 50px; }

@media (max-width: 768px) {
  .ad-leaderboard { display: none; }
  .ad-mobile-banner { display: flex; }
}
```

### Content Security Policy Updates

```go
// web/server/securityheaders.go
// Add ad network domains to CSP
scriptSrc := "'self' https://pagead2.googlesyndication.com https://www.googletagservices.com"
frameSrc := "'self' https://googleads.g.doubleclick.net https://tpc.googlesyndication.com"
```

---

## Revenue Projections

### Conservative Estimates (1,000 DAU)

| Source | Impressions/Day | CPM | Daily Revenue |
|--------|-----------------|-----|---------------|
| Footer banner | 3,000 | $1.00 | $3.00 |
| Homepage mid | 1,000 | $2.00 | $2.00 |
| Game end interstitial | 500 | $10.00 | $5.00 |
| Rewarded video | 200 | $20.00 | $4.00 |
| **Total** | | | **$14.00/day** |

**Monthly**: ~$420
**Annual**: ~$5,000

### Growth Estimates (10,000 DAU)

| Source | Impressions/Day | CPM | Daily Revenue |
|--------|-----------------|-----|---------------|
| Footer banner | 30,000 | $1.50 | $45.00 |
| Homepage mid | 10,000 | $2.50 | $25.00 |
| Game end interstitial | 5,000 | $12.00 | $60.00 |
| Rewarded video | 2,000 | $22.00 | $44.00 |
| Turn transition | 8,000 | $6.00 | $48.00 |
| **Total** | | | **$222/day** |

**Monthly**: ~$6,660
**Annual**: ~$80,000

### Premium Conversion (10,000 DAU)

Assuming 2% premium conversion at $5/month average:
- 200 subscribers x $5 = $1,000/month
- Annual: $12,000

**Combined Annual (10K DAU)**: ~$92,000

---

## Metrics to Track

### Ad Performance

| Metric | Definition | Target |
|--------|------------|--------|
| **Fill Rate** | % of ad requests filled | >90% |
| **eCPM** | Effective cost per 1000 impressions | >$5 blended |
| **Viewability** | % of ads actually seen | >70% |
| **CTR** | Click-through rate | 0.5-2% (varies by format) |

### User Impact

| Metric | Definition | Warning Sign |
|--------|------------|--------------|
| **Session Duration** | Time per session | Drops >20% after ads |
| **Games per Session** | Games started per visit | Drops after ads |
| **Day-1 Retention** | Users returning next day | Drops below 30% |
| **Premium Conversion** | % upgrading to ad-free | Below 1% |

### A/B Testing Plan

1. **Ad frequency**: Test 1 vs 2 vs 3 game-end ads per session
2. **Placement**: Test sidebar vs below-content ads
3. **Rewarded amounts**: Test 50 vs 100 vs 200 coin rewards
4. **Skip timing**: Test 3s vs 5s skip buttons

---

## Implementation Checklist

### Phase 1 (Launch)

- [ ] Create `AdSlot` component with size variants
- [ ] Add footer ad container to Footer component
- [ ] Add homepage mid-section ad container
- [ ] Set up Google AdSense account
- [ ] Update CSP headers for ad network domains
- [ ] Add premium user check to skip ads
- [ ] Test dark mode compatibility
- [ ] Test mobile responsiveness

### Phase 2 (Post-Launch)

- [ ] Implement GameEndScreen component
- [ ] Add game end ad trigger
- [ ] Implement turn transition ads with frequency control
- [ ] Add right sidebar ad slot (desktop)
- [ ] Set up analytics for ad performance

### Phase 3 (Rewarded)

- [ ] Integrate rewarded video SDK
- [ ] Implement bonus coins flow
- [ ] Implement undo move flow
- [ ] Add daily reward limits
- [ ] Create reward UI components

### Phase 4 (Premium)

- [ ] Design premium tier pricing
- [ ] Integrate Stripe for payments
- [ ] Add premium status to user model
- [ ] Create supporter badge assets
- [ ] Implement ad-free experience

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Ad blockers** | 20-30% users block ads | Focus on premium conversion, rewarded ads |
| **User churn** | Ads drive users away | A/B test aggressively, monitor retention |
| **Low fill rate** | Revenue below projections | Use mediation, multiple networks |
| **Policy violations** | Account suspension | Review ad placement guidelines carefully |
| **Performance impact** | Slower page loads | Lazy load ads, async scripts |

---

## Conclusion

This phased approach allows WeeWar to:

1. **Launch with basic ads** - Footer and homepage ads require minimal development
2. **Iterate based on data** - Measure impact before adding more placements
3. **Maximize high-value moments** - Game end and rewarded ads have best eCPM
4. **Offer premium escape** - Users who dislike ads can pay to remove them

The turn-based nature of WeeWar is actually an advantage for ad monetization - natural pause points exist where ads feel less intrusive than in real-time games.

**Recommended First Step**: Set up Google AdSense with footer banner ads only. Measure impact on user behavior for 2-4 weeks before expanding to other placements.
