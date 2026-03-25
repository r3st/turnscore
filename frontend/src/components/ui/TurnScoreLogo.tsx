import logoFantasy from '@/assets/logo-fantasy.png';
import logoScifi from '@/assets/logo-scifi.png';

interface TurnScoreLogoProps {
  height?: number;
  className?: string;
}

/**
 * Renders both logo variants; CSS (globals.css .ts-logo-*) hides
 * the inactive one based on the current data-theme on <html>.
 * Falls back to the Fantasy logo on pages without a theme.
 */
export function TurnScoreLogo({ height = 36, className = '' }: TurnScoreLogoProps) {
  const style = { height: `${height}px`, width: 'auto' };

  return (
    <>
      <img
        src={logoFantasy}
        alt="TurnScore"
        style={style}
        className={`ts-logo-fantasy ${className}`}
      />
      <img
        src={logoScifi}
        alt="TurnScore"
        style={style}
        className={`ts-logo-scifi ${className}`}
      />
    </>
  );
}
