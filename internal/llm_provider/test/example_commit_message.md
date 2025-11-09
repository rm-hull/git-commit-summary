# ✨ feat(ui): Add dark mode toggle

## Summary
Added a **dark mode toggle** to the main navigation bar.  
This allows users to switch between 🌞 *light* and 🌙 *dark* themes.

## Details
- Implemented theme state management via `ThemeContext`
- Added `usePrefersDarkMode()` hook
- Updated `Navbar.tsx` and `AppLayout.tsx`
- Persisted user preference in `localStorage`

```typescript
// Example usage
const { theme, toggleTheme } = useTheme();
toggleTheme(); // switches between light/dark
```

## Screenshots
| Mode | Preview |
|------|----------|
| Light ☀️ | ![light-mode](docs/img/light.png) |
| Dark 🌑 | ![dark-mode](docs/img/dark.png) |

## Checklist
- [x] Implement toggle
- [x] Persist user preference
- [ ] Add unit tests
- [ ] Update documentation

> “Darkness cannot drive out darkness; only light can do that.” — Martin Luther King Jr.

---

**Breaking Changes:**  
⚠️  `ThemeProvider` must now wrap the root of the app (`index.tsx`).

**Related Issues:**  
Closes #42, Relates to #56
