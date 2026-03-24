---
name: requirements
description: >
  Use this skill whenever the user wants to define, refine, document, or discuss requirements
  for the TurnScore tabletop tournament table rating application. Triggers include: defining
  new features, writing user stories, discussing the rating system, QR code workflow,
  authentication model, tournament management, theme/layout system, GDPR compliance, i18n,
  or any time the user says "I want", "new idea", "feature", "requirement" or describes how
  something should work. Always use this skill before writing any code for a new feature —
  requirements first, implementation second.
---

# Requirements Skill — TurnScore

## Project Context

**Application name:** TurnScore
**Domain:** Tabletop miniature game tournaments (system-agnostic)
**Purpose:** Organizers present game tables at a tournament online; participants rate them
based on defined criteria via QR code and smartphone.

---

## Domain Knowledge

### Tabletop Tournament Context (system-agnostic)
- The app is **not tied to any specific game system** — no branding for Infinity, Warhammer, etc.
- Supported genres: **Fantasy** and **Sci-Fi** (covers ~99% of all tabletop systems)
- The **app layout and theming adapts to the tournament type** (Fantasy vs. Sci-Fi)
- Tournaments consist of multiple **game tables** (typically 6–20 tables)
- Each table has **Deployment Zones A and B** — the two deployment areas for players
- Table quality significantly impacts the playing experience

### Tournament Types & Theming
| Type | Visual Theme | Example Systems |
|---|---|---|
| `fantasy` | Medieval, organic, stone/wood aesthetic | Warhammer Fantasy, Age of Sigmar, Kings of War |
| `scifi` | Futuristic, technical, metal/neon aesthetic | Warhammer 40K, Infinity, Star Wars Legion |

The tournament type is set when creating a tournament and determines:
- Color palette and icons in the frontend
- Background images/textures
- Terminology (e.g. "terrain" vs. "cover", etc.)

---

## Rating Criteria

### Rating Scale: **1–6 (School Grade System)**
- **1** = Excellent
- **2** = Good
- **3** = Satisfactory
- **4** = Sufficient
- **5** = Poor
- **6** = Failing

⚠️ Important for results: **Lower values = better** (like school grades). Rankings must be sorted ascending (ASC).

### Core Criteria (always active)
| Criterion | Description | Type |
|---|---|---|
| **Balance** | Equal cover/line-of-sight for both sides | Scale 1–6 |
| **Aesthetics** | Visual impression, painting, coherence | Scale 1–6 |
| **Terrain Density** | Sufficient terrain for tactical play | Scale 1–6 |
| **Labeling** | Clear markers for terrain rules | Scale 1–6 |
| **Overall** | General overall impression | Scale 1–6 |
| **Deployment Zone A** | Rating of deployment zone A | Scale 1–6 |
| **Deployment Zone B** | Rating of deployment zone B | Scale 1–6 |
| **Table Played** | Did I play on this table? | Choice: No / Zone A / Zone B |
| **Free Comment** | Open text comment about the table | Text (max. 1000 chars) |

### Optional Criteria (configurable per tournament)
- Catering / Food
- Venue / Location
- Organization / Event Flow

### Deployment Zone Photos
Each table has **three photo categories**:
- `general` — Overall table photos
- `zone_a` — Photos specific to Deployment Zone A
- `zone_b` — Photos specific to Deployment Zone B

When rating a zone, the rater automatically sees the corresponding zone photos.

---

## Roles & Permissions

### Role: Organizer
- Creates and manages tournaments (creator = organizer automatically)
- **Full permissions** for all tournament-related actions
- Only role that can delete/archive tournaments
- Only role that can configure and view results
- Can invite and remove helpers

### Role: Helper
- Added to a tournament by the organizer
- **Almost identical permissions to organizer**, including:
  - Create, edit, describe tables
  - Upload, categorize, and delete photos
  - **Register raters and assign codes**
  - Generate QR codes and export as PDF
- ❌ No access to: results, tournament deletion/archiving, helper management

### Role: Rater — NO classic login
- Authentication via: **nickname + 4–6 digit numeric code**
- Code assigned by organizer **or helper**, or auto-generated
- Prevents duplicate ratings (one rating per rater per table)
- No password, no email required
- **All submitted ratings are fully anonymous externally** — no nickname, no identity visible to anyone

---

## Authentication (Organizer & Helper)

### Login Strategy Phase 1: Google OAuth Only
**Recommendation: implement only "Login with Google"**

Reasons:
- No custom password reset flow needed
- No email sending setup (SMTP, SendGrid, etc.)
- No two-factor authentication concept needed
- Faster, more secure implementation
- Virtually every organizer has a Google account

**Phase 2 (backlog):** Email + password as additional option — then with 2FA and secure password management.

### Rater Login (separate flow)
- Select tournament (or pre-filled via QR code) → enter nickname → enter code
- JWT with limited lifetime (tournament end + 24h buffer)
- No Google login for raters

---

## Core Entities

```
Tournament
  ├── id, slug, name
  ├── type: fantasy | scifi                    ← determines UI theme
  ├── description (free text, Markdown support)
  ├── links: [{url, label}]                    ← always shown with disclaimer
  ├── location, event_date
  ├── organizer_id → User
  ├── table_count (set at creation, immutable after activation)
  ├── status: draft | active | voting | archived
  ├── voting_start, voting_end
  ├── active_criteria: []CriteriaKey
  ├── result_config: {
  │     show_comments: bool,
  │     visible_comment_criteria: []CriteriaKey
  │   }
  └── tables []Table

Table
  ├── id, tournament_id, number (1..table_count), name, description
  ├── photos: [{id, url, thumbnail_url, category: general|zone_a|zone_b}]
  ├── qr_code_url
  └── ratings []Rating

Rater
  ├── id, tournament_id
  ├── nickname (unique per tournament)
  └── code (4–6 digits, unique per tournament)

Rating
  ├── id, table_id, rater_id      ← rater_id NEVER exposed externally
  ├── criteria_scores: map[CriteriaKey]int (1–6)
  ├── played_zone: none | zone_a | zone_b
  ├── comment (optional, max. 1000 chars)
  └── created_at

User (system user)
  ├── id, google_sub (OAuth ID), email, name, avatar_url
  ├── role: organizer | helper
  └── tournament_memberships []TournamentMembership
```

---

## User Stories

### Epic 1: Tournament Management
- **US-01:** As an organizer I want to create a tournament with name, type (Fantasy/Sci-Fi), date, location, description, and number of tables.
- **US-02:** As an organizer I want to add any number of links to the tournament (with label), always displayed with a liability disclaimer.
- **US-03:** As an organizer I want to choose which optional rating criteria are active.
- **US-04:** As an organizer I want to set a voting period (start/end).
- **US-05:** As an organizer I want to invite helpers via Google account / email.
- **US-06:** As an organizer I want to configure before showing results: whether comments are visible, and if so, which criteria comments to show.

### Epic 2: Table Management
- **US-10:** As an organizer/helper I want to create tables (count taken from tournament config).
- **US-11:** As an organizer/helper I want to upload photos and assign them to a category (General / Zone A / Zone B).
- **US-12:** As an organizer/helper I want to generate a QR code per table.
- **US-13:** As an organizer/helper I want to export all QR codes as a print-ready PDF.
- **US-14:** As an organizer/helper I want to register participants (raters) and assign or auto-generate codes.

### Epic 3: Rating
- **US-20:** As a participant I want to log in with nickname + code.
- **US-21:** As a participant I want to rate tables on the active criteria using a 1–6 scale.
- **US-22:** As a participant I want to see the corresponding zone photos when rating Zone A/B.
- **US-23:** As a participant I want to indicate whether I played on this table (No / Zone A / Zone B).
- **US-24:** As a participant I want to leave an optional free-text comment (max. 1000 chars).
- **US-25:** The system prevents me from rating the same table twice.
- **US-26:** As a participant I want to see all tables of a tournament in an overview.

### Epic 4: Results (organizer/helper only)
- **US-30:** As an organizer/helper I want to see a ranking of tables by overall score (ascending, 1 = best grade).
- **US-31:** As an organizer/helper I want to see average scores per criterion per table.
- **US-32:** As an organizer/helper I want to see anonymized comments (if enabled in config) — without rater attribution.
- **US-33:** As an organizer/helper I want to see how many players played Zone A vs Zone B per table.

### Epic 5: Home Page & Overview
- **US-40:** As a visitor I see on the home page up to 5 upcoming and 5 past tournaments.
- **US-41:** As a visitor I can identify the tournament type via a badge (Fantasy 🏰 / Sci-Fi 🚀).
- **US-42:** As a visitor I see on a tournament card: name, date, location, type badge, table count.

---

## QR Code Workflow

1. Organizer/helper generates QR codes (individually or all at once)
2. QR code URL: `https://app.example.com/rate/{tournament_slug}/{table_number}`
3. Export: PNG (single) or PDF sheet (all tables, print-ready)
4. Print placed on the table (table card ~9×5 cm)
5. Participant scans → table detail page → login with nickname+code → submit rating

**QR code card contains:** table number, table name, QR code, (optional) tournament name

---

## Home Page Layout

```
[Header: TurnScore | 🇩🇪 / 🇬🇧 language switcher | Login]

─── Upcoming Tournaments (max. 5) ───────────────────
  🚀 Sci-Fi  | Tournament Name     | 12 Jul 2025 | Essen
  🏰 Fantasy | Tournament Name     | 19 Jul 2025 | Berlin
  [Show all tournaments →]

─── Past Tournaments (max. 5) ───────────────────────
  🚀 Sci-Fi  | Tournament Name     | 03 Mar 2025 | Hamburg
  [Show all past tournaments →]

[Footer: Imprint | Privacy Policy | Cookie Settings]
```

---

## Privacy & Data Protection

### Anonymity Principle
- Ratings are **internally** linked to the rater (rater_id) for duplicate prevention
- **Externally — including to organizers** — ratings are fully anonymous
- In results views: no nicknames, no rater IDs visible
- Comments appear without author attribution

### GDPR Requirements
- **Cookie banner** on first visit (opt-in for non-essential cookies)
- **Privacy policy** as a standalone page (linked in footer)
- **Legal notice / Imprint** as a standalone page
- Minimal data storage: raters have no email, no full name
- Google login: OAuth scopes limited to minimum (email + profile only)
- Right to deletion: organizer account and all related data deletable
- External links disclaimer: *"We are not responsible for the content of external links."*

### Required Pages
- `/imprint` — legal notice (required under German law)
- `/privacy` — privacy policy (GDPR)
- Both linked in footer, accessible without login

---

## Internationalization (i18n)

- **Languages:** English (default) and German
- **Language switcher:** In header or sidebar, with flag icons (🇬🇧 / 🇩🇪)
- All UI texts via i18n keys (`react-i18next`), no text hardcoding
- Tournament content (name, description, links) = user input, not translated
- Locale persistence: stored in localStorage or cookie

---

## Non-Functional Requirements

- **Responsive:** Mobile-first (raters primarily use smartphones)
- **Offline-tolerant:** Ratings work with poor tournament Wi-Fi (retry logic)
- **Performance:** Photo upload with automatic compression + thumbnail generation
- **Theming:** Fantasy/Sci-Fi CSS theme loaded from tournament data, no app restart needed
- **Accessibility:** WCAG 2.1 AA targeted

---

## Backlog
- Live results display during the tournament
- Public results page after tournament ends
- CSV/Excel export for organizers
- Comment moderation
- Table comparison view
- Dark mode (per theme)
- Email + password login (Phase 2, with 2FA)
- Additional themes (e.g. Historical, Cyberpunk)
- Notifications (e.g. voting period starts)

---

## Checklist for New Requirements

1. **Classify** — which epic does it belong to? New entity needed?
2. **User story** — "As a [role] I want to [action] so that [benefit]."
3. **Acceptance criteria** — what must be true for the story to be "done"?
4. **Anonymity check** — is rater identity exposed anywhere externally?
5. **GDPR check** — are new personal data being collected?
6. **Theme check** — does the Fantasy/Sci-Fi theming need to be updated?
7. **Dependencies** — DB schema changes, new API endpoints, i18n keys needed?
