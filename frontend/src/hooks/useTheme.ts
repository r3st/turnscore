export function useTheme() {
  const applyTheme = (type: 'fantasy' | 'scifi') => {
    document.documentElement.setAttribute('data-theme', type);
  };

  const clearTheme = () => {
    document.documentElement.removeAttribute('data-theme');
  };

  return { applyTheme, clearTheme };
}
