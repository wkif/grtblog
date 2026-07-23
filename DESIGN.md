# GrtBlog Design System

## Design Direction

Warm editorial utility: a calm, tactile reading environment paired with a precise, capable workspace. The system should feel personal because of its material choices, typography, content rhythm, and considerate feedback—not because of ornament.

## Theme

Default toward a light, paper-like canvas for reading and publishing. The physical scene is a creator reviewing a draft beside a window at midday: warm ambient light, a quiet desk, readable contrast, and enough depth to separate paper, panels, and controls. Support dark mode as a first-class alternate with softened ink tones rather than pure black.

## Color

Use the existing ink/jade vocabulary as the foundation and keep semantic colors consistent across the public site and admin.

- **Canvas:** warm near-white or softly tinted paper for primary reading surfaces.
- **Ink:** deep charcoal/ink for headings and body text; avoid absolute black for long-form reading.
- **Muted ink:** desaturated gray-green for metadata, secondary labels, and supporting copy.
- **Jade accent:** the primary brand accent for links, active navigation, primary actions, and live presence.
- **Warm highlight:** restrained amber or apricot for personal notes, featured moments, and non-critical emphasis.
- **Semantic states:** calm blue for information, amber for warnings, red for destructive/error states, green/jade for success; do not rely on color alone.
- **Dark mode:** use layered ink, slate, and softened jade surfaces with reduced saturation and clear focus rings.

Accent color is for action, selection, and status—not broad decoration. Prefer surface tint, type hierarchy, and spacing to create emphasis.

## Typography

Keep the existing Google Sans Flex and Noto Sans SC foundation for UI and multilingual content, with a readable editorial serif where long-form article presentation benefits from it. Use JetBrains Mono for code and technical metadata.

- UI and controls: Google Sans Flex / Noto Sans SC Variable.
- Long-form prose: a high-legibility serif paired with the existing sans UI system; use sparingly and consistently.
- Code: JetBrains Mono.
- Use a compact product scale for admin controls and a more generous editorial scale for public headings.
- Keep prose measure around 65–75 characters where possible; allow data tables and operational layouts to be denser.
- Use weight and size before color to establish hierarchy.

## Spacing and Shape

Use a consistent rem-based spacing scale with generous vertical rhythm on reading surfaces and tighter grouping in forms and tables. Prefer subtle corner radii that feel tactile but not toy-like. Use full borders, tonal surfaces, and elevation sparingly; never use colored side-stripe accents as the primary hierarchy device.

## Layout

- Public pages: content-first compositions with a stable reading measure, clear article metadata, and generous breathing room.
- Navigation: keep the author's identity and core destinations easy to find without competing with the article.
- Admin: use familiar sidebar or top navigation, dense but breathable panels, clear section headings, and predictable action placement.
- Responsive behavior is structural: collapse navigation, reflow metadata, convert tables to useful mobile summaries, and preserve touch target sizes.
- Shared shell elements should align across public and admin where it helps recognition, while density and content rhythm can differ.

## Components

Build a shared vocabulary for buttons, links, inputs, cards, tabs, menus, badges, dialogs, notifications, pagination, and content metadata.

Every interactive component should define default, hover, focus-visible, active, disabled, loading, success, and error behavior where applicable. Use skeletons for content loading, instructional empty states, inline validation, and recoverable error messages. Realtime connection state should be visible without interrupting reading or editing.

## Motion

Use restrained 150–250ms transitions for state changes, navigation, overlays, and live updates. Motion should explain what changed: a notification arriving, content refreshing, a panel opening, or an action completing. Avoid page-load choreography, gratuitous parallax, and animation that delays reading or task completion.

## Imagery and Identity

Let author avatars, article artwork, screenshots, and open-graph imagery carry personal identity. Prefer authentic, context-rich imagery over generic stock photography. Keep image treatment consistent with the warm paper-and-ink foundation and provide meaningful alt text and resilient fallbacks.

## Accessibility

Maintain strong text contrast, visible keyboard focus, semantic landmarks, reduced-motion support, adequate touch targets, logical heading order, and non-color state communication. Long-form content, code blocks, comments, dialogs, and data tables must remain usable with keyboard navigation and assistive technology.

## Implementation Anchors

- Public app: SvelteKit, Svelte 5, Tailwind CSS v4, TanStack Query.
- Admin app: Vue 3.5, Naive UI, Tailwind CSS, Pinia, Vite.
- Existing type foundations: Google Sans Flex, Noto Sans SC Variable, and JetBrains Mono.
- Existing semantic foundations: ink/jade public tokens and Naive UI theme variables in the admin.
