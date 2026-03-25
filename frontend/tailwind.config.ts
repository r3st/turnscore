import type { Config } from 'tailwindcss';
import animate from 'tailwindcss-animate';

const config: Config = {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // Mapped to CSS custom properties — theme switching via data-theme attribute
        primary: 'var(--color-primary)',
        secondary: 'var(--color-secondary)',
        background: 'var(--color-background)',
        surface: 'var(--color-surface)',
        accent: 'var(--color-accent)',
        foreground: 'var(--color-text)',
        muted: {
          DEFAULT: 'var(--color-surface)',
          foreground: 'color-mix(in srgb, var(--color-text) 60%, transparent)',
        },
        border: 'color-mix(in srgb, var(--color-text) 15%, transparent)',
        input: 'color-mix(in srgb, var(--color-text) 15%, transparent)',
        ring: 'var(--color-primary)',
      },
      fontFamily: {
        heading: 'var(--font-heading)',
        body: 'var(--font-body)',
      },
      borderRadius: {
        DEFAULT: 'var(--border-radius)',
        sm: 'calc(var(--border-radius) * 0.5)',
        md: 'var(--border-radius)',
        lg: 'calc(var(--border-radius) * 2)',
      },
      boxShadow: {
        DEFAULT: 'var(--shadow)',
      },
    },
  },
  plugins: [animate],
};

export default config;
