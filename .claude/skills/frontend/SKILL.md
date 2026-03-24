---
name: frontend
description: >
  Use this skill for all React + TypeScript frontend development tasks in TurnScore.
  Triggers include: building UI components, implementing pages, routing, TanStack Query
  API hooks, openapi-typescript generated types, theming (Fantasy/Sci-Fi CSS variables),
  i18n (react-i18next EN/DE), accessibility (WCAG 2.1 AA, jest-axe), rating interface
  with 1-6 school grade scale, photo upload with camera access, Google OAuth login,
  Playwright E2E tests, Vitest unit tests, or any time React/TypeScript code needs to be
  written or reviewed. Always consult this skill when writing frontend code.
---

# Frontend Skill — React + TypeScript Development

## Tech Stack

| Tool | Version | Purpose |
|---|---|---|
| React | 18 | UI framework |
| TypeScript | 5 | Type safety |
| Vite | 5 | Build tool / dev server |
| React Router | 6 | Client-side routing |
| TanStack Query | 5 | Server-state management |
| Zustand | 4 | Client-state (auth) |
| shadcn/ui + Radix | latest | Components (a11y built-in) |
| Tailwind CSS | 3 | Styling + CSS variables for themes |
| React Hook Form | 7 | Forms |
| Zod | 3 | Schema validation |
| axios | 1.7 | HTTP client |
| openapi-typescript | 7 | Generate types from OpenAPI spec |
| react-i18next | 15 | EN/DE internationalization |
| Vitest | 2 | Unit tests |
| React Testing Library | 16 | Component tests |
| jest-axe | 9 | Accessibility tests |
| Playwright | 1.47 | E2E tests |

---

## Project Structure

```
src/
├── api/
│   ├── generated/          # ⛔ do not edit — openapi-typescript output
│   │   └── api.ts
│   ├── client.ts           # Axios instance + TanStack Query setup
│   └── hooks/              # All API hooks (TanStack Query)
│       ├── useTournaments.ts
│       ├── useTables.ts
│       └── useRatings.ts
├── components/
│   ├── ui/                 # ⛔ shadcn/ui (do not edit)
│   ├── layout/
│   │   ├── AppLayout.tsx
│   │   ├── Navbar.tsx      # Language switcher + auth status
│   │   └── Footer.tsx      # Imprint / Privacy / Cookie link
│   ├── tournament/
│   │   ├── TournamentCard.tsx   # Badge: Fantasy🏰 / Sci-Fi🚀
│   │   └── TournamentForm.tsx
│   ├── table/
│   │   ├── TableCard.tsx
│   │   ├── TableGrid.tsx
│   │   ├── PhotoGallery.tsx     # Swipeable, category filter
│   │   └── PhotoUpload.tsx      # Drag & drop + camera
│   └── rating/
│       ├── RatingForm.tsx       # Core: 1-6 grade sliders + zone picker
│       ├── CriteriaItem.tsx     # Single criterion with slider
│       ├── PlayedZonePicker.tsx # No / Zone A / Zone B
│       └── RatingResults.tsx    # Results view (organizer)
├── pages/
│   ├── HomePage.tsx             # 5 upcoming + 5 past tournaments
│   ├── TournamentPage.tsx       # Public tournament view + table list
│   ├── RatePage.tsx             # Rating page (QR code landing)
│   ├── LoginPage.tsx            # Google OAuth button
│   ├── RaterLoginPage.tsx       # Nickname + code form
│   ├── OAuthCallbackPage.tsx    # Handles Google redirect
│   ├── ImprintPage.tsx
│   ├── PrivacyPage.tsx
│   └── dashboard/
│       ├── DashboardPage.tsx
│       ├── TournamentEditPage.tsx
│       ├── TableManagementPage.tsx
│       └── ResultsPage.tsx
├── hooks/
│   └── useTheme.ts              # Fantasy/Sci-Fi theme switching
├── stores/
│   └── authStore.ts             # Zustand: JWT + user role
├── themes/
│   ├── fantasy.css              # CSS custom properties — fantasy
│   └── scifi.css                # CSS custom properties — sci-fi
├── i18n/
│   ├── index.ts                 # react-i18next setup
│   ├── en.json                  # All English texts
│   └── de.json                  # All German texts
└── e2e/                         # Playwright E2E tests
    ├── rating-flow.spec.ts
    └── organizer.spec.ts
```

---

## Generated TypeScript Types (openapi-typescript)

Types are **never written manually** — they come from the OpenAPI spec:

```bash
# Makefile target:
npx openapi-typescript ../api/openapi.yaml -o src/api/generated/api.ts
```

Usage:

```typescript
// src/api/hooks/useTournaments.ts
import type { components, paths } from '@/api/generated/api';

type Tournament = components['schemas']['Tournament'];
type CreateTournamentBody =
  paths['/api/v1/tournaments']['post']['requestBody']['content']['application/json'];

export function useTournaments() {
  return useQuery<Tournament[]>({
    queryKey: ['tournaments'],
    queryFn: () => apiClient.get('/tournaments').then(r => r.data),
  });
}

export function useCreateTournament() {
  const qc = useQueryClient();
  return useMutation<Tournament, Error, CreateTournamentBody>({
    mutationFn: (body) => apiClient.post('/tournaments', body).then(r => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tournaments'] }),
  });
}
```

---

## Theming: Fantasy vs. Sci-Fi

Themes are switched via **CSS custom properties** — no JS theme system needed:

```css
/* src/themes/fantasy.css */
:root[data-theme="fantasy"] {
  --color-primary:     #8B5E3C;   /* Warm brown */
  --color-secondary:   #4A7C59;   /* Forest green */
  --color-background:  #F5F0E8;   /* Parchment */
  --color-surface:     #EDE6D6;
  --color-text:        #2C1810;
  --color-accent:      #C4922A;   /* Gold */
  --font-heading:      'Cinzel', serif;
  --font-body:         'Crimson Text', serif;
  --border-radius:     4px;
  --shadow:            2px 2px 8px rgba(44, 24, 16, 0.3);
}

/* src/themes/scifi.css */
:root[data-theme="scifi"] {
  --color-primary:     #00D4FF;   /* Cyan/neon */
  --color-secondary:   #7B2FBE;   /* Purple */
  --color-background:  #0A0E1A;   /* Dark navy */
  --color-surface:     #111827;
  --color-text:        #E2E8F0;
  --color-accent:      #00FF88;   /* Neon green */
  --font-heading:      'Orbitron', sans-serif;
  --font-body:         'Share Tech Mono', monospace;
  --border-radius:     2px;
  --shadow:            0 0 12px rgba(0, 212, 255, 0.3);
}
```

```typescript
// src/hooks/useTheme.ts
export function useTheme() {
  const applyTheme = (type: 'fantasy' | 'scifi') => {
    document.documentElement.setAttribute('data-theme', type);
  };
  return { applyTheme };
}

// On TournamentPage — apply automatically:
const { applyTheme } = useTheme();
useEffect(() => { applyTheme(tournament.type); }, [tournament.type]);
```

---

## i18n: English / German

```typescript
// src/i18n/index.ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import en from './en.json';
import de from './de.json';

i18n.use(initReactI18next).init({
  resources: { en: { translation: en }, de: { translation: de } },
  lng: localStorage.getItem('language') || 'en',
  fallbackLng: 'en',
});

// Language switcher in Navbar
function LanguageSwitcher() {
  const { i18n } = useTranslation();
  return (
    <div className="flex gap-2" role="group" aria-label="Select language">
      <button
        onClick={() => { i18n.changeLanguage('en'); localStorage.setItem('language', 'en'); }}
        aria-pressed={i18n.language === 'en'}
        className={i18n.language === 'en' ? 'opacity-100' : 'opacity-50'}
      >
        🇬🇧
      </button>
      <button
        onClick={() => { i18n.changeLanguage('de'); localStorage.setItem('language', 'de'); }}
        aria-pressed={i18n.language === 'de'}
        className={i18n.language === 'de' ? 'opacity-100' : 'opacity-50'}
      >
        🇩🇪
      </button>
    </div>
  );
}
```

```json
// src/i18n/en.json (excerpt)
{
  "rating": {
    "title": "Rate Table",
    "submit": "Submit Rating",
    "submitting": "Saving...",
    "already_rated": "You have already rated this table.",
    "grade_1": "1 – Excellent",
    "grade_2": "2 – Good",
    "grade_3": "3 – Satisfactory",
    "grade_4": "4 – Sufficient",
    "grade_5": "5 – Poor",
    "grade_6": "6 – Failing",
    "played_zone": "Did I play on this table?",
    "played_none": "No",
    "played_zone_a": "Yes, Zone A",
    "played_zone_b": "Yes, Zone B",
    "comment_placeholder": "Optional comment (max. 1000 characters)..."
  },
  "criteria": {
    "balance": "Balance",
    "aesthetics": "Aesthetics",
    "terrain_density": "Terrain Density",
    "labeling": "Labeling",
    "overall": "Overall Impression",
    "zone_a": "Deployment Zone A",
    "zone_b": "Deployment Zone B",
    "catering": "Catering",
    "venue": "Venue"
  }
}
```

---

## Rating Form (1–6 School Grade System)

```tsx
// src/components/rating/CriteriaItem.tsx
interface CriteriaItemProps {
  criteriaKey: string;
  label: string;
  value: number;        // 1–6
  onChange: (v: number) => void;
  zonePhotos?: Photo[]; // auto-shown for zone_a/zone_b
}

export function CriteriaItem({ criteriaKey, label, value, onChange, zonePhotos }: CriteriaItemProps) {
  const { t } = useTranslation();
  const id = `criteria-${criteriaKey}`;

  return (
    <div className="space-y-2">
      {/* Zone photos shown automatically when rating A/B */}
      {zonePhotos && zonePhotos.length > 0 && (
        <PhotoGallery photos={zonePhotos} compact />
      )}

      <div className="flex justify-between items-center">
        <label htmlFor={id} className="font-medium text-sm">
          {label}
        </label>
        <span className="text-xl font-bold text-primary" aria-live="polite">
          {value} — {t(`rating.grade_${value}`)}
        </span>
      </div>

      {/* Slider: min=1 (excellent) max=6 (failing) */}
      <Slider
        id={id}
        min={1} max={6} step={1}
        value={[value]}
        onValueChange={([v]) => onChange(v)}
        className="w-full"
        aria-valuemin={1}
        aria-valuemax={6}
        aria-valuenow={value}
        aria-valuetext={t(`rating.grade_${value}`)}
      />

      <div className="flex justify-between text-xs text-muted-foreground" aria-hidden>
        <span>1 – {t('rating.grade_1_short')}</span>
        <span>6 – {t('rating.grade_6_short')}</span>
      </div>
    </div>
  );
}
```

---

## Google Login Flow

```tsx
// src/pages/LoginPage.tsx
export function LoginPage() {
  const { t } = useTranslation();

  // Backend-driven redirect — no client-side OAuth needed
  const handleGoogleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/api/v1/auth/google`;
  };

  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 p-8">
        <h1 className="text-2xl font-bold text-center">{t('auth.login_title')}</h1>
        <Button onClick={handleGoogleLogin} className="w-full" size="lg">
          <GoogleIcon className="mr-2 h-5 w-5" aria-hidden />
          {t('auth.login_with_google')}
        </Button>
        <div className="text-center">
          <Link to="/rate-login" className="text-primary underline text-sm">
            {t('auth.rater_login_link')}
          </Link>
        </div>
      </div>
    </main>
  );
}

// src/pages/OAuthCallbackPage.tsx
export function OAuthCallbackPage() {
  const [searchParams] = useSearchParams();
  const { login } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');
    const refreshToken = searchParams.get('refresh');
    if (token) {
      login(token, refreshToken);
      navigate('/dashboard');
    } else {
      navigate('/login?error=oauth_failed');
    }
  }, []);

  return <LoadingSpinner />;
}
```

---

## Photo Upload with Camera Access

```tsx
// src/components/table/PhotoUpload.tsx
export function PhotoUpload({ tableId, category, onSuccess }: PhotoUploadProps) {
  const { t } = useTranslation();
  const inputRef = useRef<HTMLInputElement>(null);
  const { mutate: upload, isPending } = useUploadPhoto(tableId);

  const handleFiles = (files: FileList | null) => {
    if (!files) return;
    Array.from(files).forEach(file => upload({ file, category }));
  };

  return (
    <div>
      <p className="text-sm font-medium mb-2" id={`upload-label-${category}`}>
        {t(`photo.category_${category}`)}
      </p>
      <div
        role="button"
        tabIndex={0}
        aria-labelledby={`upload-label-${category}`}
        onClick={() => inputRef.current?.click()}
        onKeyDown={(e) => e.key === 'Enter' && inputRef.current?.click()}
        className="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer
                   hover:border-primary focus:outline-none focus:ring-2 focus:ring-primary"
      >
        {isPending
          ? <Spinner aria-label={t('photo.uploading')} />
          : <><CameraIcon className="mx-auto h-10 w-10 text-muted-foreground" aria-hidden />
              <p className="mt-2 text-sm">{t('photo.tap_to_upload')}</p></>
        }
      </div>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        capture="environment"   // Opens camera on mobile
        multiple
        className="hidden"
        aria-hidden
        onChange={(e) => handleFiles(e.target.files)}
      />
    </div>
  );
}
```

---

## Routing

```tsx
// src/App.tsx
const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/t/:slug', element: <TournamentPage /> },
  { path: '/rate/:slug/:tableNum', element: <RatePage /> },  // QR landing
  { path: '/login', element: <LoginPage /> },
  { path: '/auth/callback', element: <OAuthCallbackPage /> },
  { path: '/rate-login/:slug?', element: <RaterLoginPage /> },
  { path: '/imprint', element: <ImprintPage /> },
  { path: '/privacy', element: <PrivacyPage /> },

  {
    path: '/dashboard',
    element: <ProtectedRoute roles={['organizer', 'helper']}><AppLayout /></ProtectedRoute>,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'tournaments/new', element: <TournamentCreatePage /> },
      { path: 'tournaments/:slug', element: <TournamentEditPage /> },
      { path: 'tournaments/:slug/tables', element: <TableManagementPage /> },
      {
        path: 'tournaments/:slug/results',
        element: <ProtectedRoute roles={['organizer']}><ResultsPage /></ProtectedRoute>
      },
    ],
  },
]);
```

---

## Auth Store (Zustand)

```typescript
// src/stores/authStore.ts
interface AuthState {
  token: string | null;
  refreshToken: string | null;
  userId: string | null;
  role: 'organizer' | 'helper' | 'rater' | null;
  login: (token: string, refreshToken: string | null) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
  hasRole: (...roles: string[]) => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null, refreshToken: null, userId: null, role: null,
      login: (token, refreshToken) => {
        const claims = parseJWT(token); // { sub, role }
        set({ token, refreshToken, userId: claims.sub, role: claims.role });
      },
      logout: () => set({ token: null, refreshToken: null, userId: null, role: null }),
      isAuthenticated: () => !!get().token,
      hasRole: (...roles) => roles.includes(get().role ?? ''),
    }),
    { name: 'auth' }
  )
);
```

---

## Testing

### Unit + Accessibility Test

```tsx
// src/components/rating/CriteriaItem.test.tsx
import { render, screen } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
expect.extend(toHaveNoViolations);

describe('CriteriaItem', () => {
  it('renders label and current grade', () => {
    render(<CriteriaItem criteriaKey="balance" label="Balance" value={3} onChange={() => {}} />);
    expect(screen.getByLabelText('Balance')).toBeInTheDocument();
    expect(screen.getByText(/3/)).toBeInTheDocument();
  });

  it('has no accessibility violations', async () => {
    const { container } = render(
      <CriteriaItem criteriaKey="balance" label="Balance" value={1} onChange={() => {}} />
    );
    expect(await axe(container)).toHaveNoViolations();
  });

  it('calls onChange on keyboard navigation', async () => {
    const onChange = vi.fn();
    render(<CriteriaItem criteriaKey="balance" label="Balance" value={3} onChange={onChange} />);
    const slider = screen.getByRole('slider');
    slider.focus();
    await userEvent.keyboard('{ArrowRight}');
    expect(onChange).toHaveBeenCalledWith(4);
  });
});
```

### E2E Test (Playwright)

```typescript
// e2e/rating-flow.spec.ts
test('Rater can rate a table after QR code scan', async ({ page }) => {
  await page.goto('/rate/test-tournament-2025/3');

  // Rater login
  await page.fill('[name="nickname"]', 'TestPlayer');
  await page.fill('[name="code"]', '1234');
  await page.click('button[type="submit"]');

  // Rating form
  await expect(page.locator('h1')).toContainText('Table 3');

  // Set slider to grade 2 (keyboard)
  const balanceSlider = page.locator('[aria-label*="Balance"]');
  await balanceSlider.focus();
  await page.keyboard.press('ArrowLeft'); // 3 → 2

  // Select played zone
  await page.click('[data-value="zone_a"]');

  // Comment
  await page.fill('textarea', 'Good table, a bit sparse on Zone B terrain');

  // Submit
  await page.click('button:has-text("Submit Rating")');
  await expect(page.locator('[role="alert"]')).toContainText('Rating saved');

  // Duplicate rating prevention
  await page.goto('/rate/test-tournament-2025/3');
  await expect(page.locator('[role="alert"]')).toContainText('already rated');
});
```

### Coverage

```bash
npx vitest run --coverage
# Target: >75% overall, >90% for components/rating/

npx playwright test --reporter=html
```

---

## Conventions

- **No manual API types** — always use `src/api/generated/api.ts`
- **No text hardcoding** in JSX — always use `t('key')` from i18n
- **All interactive elements** have `aria-label` or a visible label
- **Forms:** React Hook Form + Zod schema, no uncontrolled inputs
- **API calls:** Always via TanStack Query hooks, no direct `fetch`/`axios` in components
- **Errors:** Error Boundaries at page level; toast for API errors with `role="alert"`
- **Theme:** `data-theme` attribute on `<html>` — never inline styles for theme colors
- **Tests:** Every new component gets at least one `axe` accessibility test
- **Mobile:** Minimum 44px touch target for all interactive elements
- **Licenses:** After adding any new npm package, run `make licenses`
