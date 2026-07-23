---
target: public reading shell
total_score: 24
p0_count: 0
p1_count: 5
timestamp: 2026-07-21T02-57-33Z
slug: web-src-routes-page-svelte
---
# Public Reading Shell Critique

## Provenance

Two isolated assessments were completed: an independent design review and a detector/evidence review. Browser inspection was unavailable in this environment, so visual claims are based on source review; deterministic findings come from `detect.mjs`.

## AI-Slop Verdict

Not immediately AI-made, but showing AI-polish drift. The product has a recognizable editorial direction: warm paper tones, serif reading typography, jade/cinnabar semantics, author presence, and content-first layouts. The shell also layers familiar premium UI signals—noise texture, glass surfaces, cursor-following gradients, blur-in animations, glowing hover rails, floating status indicators, and morphing mobile chrome. Individually defensible, together they risk making the interface feel generically generated rather than authored.

The strongest anti-slop opportunity is subtraction: let author identity, typography, and writing carry more personality while reducing decorative interaction around ordinary links and list rows.

## Nielsen Scores

| Heuristic | Score | Key issue |
|---|---:|---|
| Visibility of system status | 3/4 | Active navigation, owner presence, route loading, live state, and metrics are visible; async loading/error states remain inconsistent. |
| Match between system and real world | 3/4 | Reading language is natural; icon-only navigation, “Archive,” “Admin,” and dense uppercase metadata introduce system vocabulary. |
| User control and freedom | 3/4 | Back, home, theme, search, pagination, TOC, and mobile exits exist; `history.back()` is unreliable for direct-entry readers. |
| Consistency and standards | 2/4 | Desktop/mobile navigation, post/moment list treatments, and hidden hover actions behave differently. |
| Error prevention | 2/4 | Reading has few destructive risks, but clipboard sharing, direct navigation, disabled zoom, and edge-state guardrails are weak. |
| Recognition rather than recall | 2/4 | Desktop users must infer icon meanings or hover to reveal labels. |
| Flexibility and efficiency | 3/4 | Keyboard search, pagination, TOC, responsive layouts, and direct links help; keyboard affordances are not advertised. |
| Aesthetic and minimalist design | 3/4 | Prose area is strong, but shell texture, activity, inspiration, and subscription compete for attention. |
| Error recognition and recovery | 2/4 | Missing content and route failures do not consistently explain what happened or what to do next. |
| Help and documentation | 1/4 | No contextual explanation for unfamiliar icons, federation/presence concepts, or interaction states. |

**Total: 24/40**

## Cognitive Load and Emotional Journey

- Intrinsic load is low to moderate: reading and archive scanning are simple.
- Extraneous load is moderate to high: icon-only navigation, hover-only labels, presence signals, changing mobile chrome, and multiple homepage modules require interpretation.
- The homepage presents at least five competing continuation paths: recent articles, recent moments, inspiration, activity pulse, and subscription (`web/src/routes/+page.svelte:16`).
- The article header exposes type, category, hot status, date, update date, reading time, views, likes, comments, and tags at once (`web/src/lib/features/post/components/post-detail/PostDetailHeader.svelte:60`).
- Arrival is welcoming; orientation is uncertain; discovery is rewarding but busy; article reading is the strongest moment; completion lacks a strong editorial handoff.

## Strengths

- Warm ink palette, jade accent, serif content typography, and subtle texture support the creator-first brief (`web/src/routes/layout.css:8`).
- Article detail has strong reading primitives: title/meta separation, reading time, TOC, related content, comments, and share actions (`web/src/lib/features/post/components/post-detail/PostDetailMain.svelte:31`).
- Responsive intent is real: desktop rail collapses into labeled mobile navigation and detail-page TOC/related content have mobile paths (`web/src/lib/ui/layout/sidebar/MobileNavBar.svelte:262`).

## Priority Issues

1. **P1 — Desktop navigation hides the information architecture.** Icon-only navigation with hover labels makes first-use discovery, keyboard use, zoomed browsing, and touch-adjacent pointer use dependent on inference (`web/src/lib/ui/layout/sidebar/Sidebar.svelte:43`, `:60`, `:73`). Preserve the compact rail if desired, but expose labels on focus, provide an expanded mode, and communicate the active section textually.
2. **P1 — Decorative motion is doing too much work in a reading product.** Blur, scale, stagger, cursor-following gradients, animated rails, and view transitions add cognitive and performance cost without improving comprehension (`web/src/routes/layout.css:181`, `:200`, `:241`; `web/src/lib/features/post/components/HomeArticleItem.svelte:14`, `:77`). Keep motion for state changes, remove it from ordinary list scanning, and add explicit reduced-motion behavior.
3. **P1 — The homepage has too many equal-weight invitations.** Recent articles and moments compete with inspiration, activity, and subscription, making the homepage simultaneously a publication front page, social room, and conversion surface (`web/src/routes/+page.svelte:16`, `:80`, `:83`, `:86`). Establish “read the latest work” as the dominant promise and demote the rest.
4. **P1 — Article metadata is overly dense.** Date, update date, reading time, views, likes, comments, hot status, category, and tags are treated as near-equal signals (`web/src/lib/features/post/components/post-detail/PostDetailHeader.svelte:83`, `:99`, `:124`). Keep date and reading time prominent; move engagement metrics into a quieter action row and demote tags/hot status.
5. **P1 — Failure and recovery states are too vague.** “The content could not be presented” does not distinguish not found, loading, offline, or server failure (`web/src/lib/features/post/components/PostDetail.svelte:31`, `web/src/routes/+layout.svelte:332`, `:437`). Provide explicit states, retry actions, and clear routes back to articles or home.

## Detector Findings

The deterministic scan returned 4 warnings and 0 errors:

- `bounce-easing`: `web/src/lib/features/home/Hero.svelte:197` (`animation: hero-scroll-bounce`).
- `layout-transition`: `web/src/lib/features/post/components/HomeArticleItem.svelte:108` (`transition: height 260ms ease`).
- `layout-transition`: `web/src/lib/features/moment/components/HomeMomentItem.svelte:108` (`transition: height 260ms ease`).
- `bounce-easing`: `web/src/lib/ui/layout/sidebar/Sidebar.svelte:134` (Tailwind `animate-bounce`).

Static evidence shows semantic `<main>` and `<nav>`, labelled sidebar controls, hidden decorative noise, avatar alt text, responsive shell padding, a mobile navigation path, and separate desktop/mobile hero structures. Browser verification of computed contrast, focus visibility, keyboard flow, touch-target sizes, and overflow was unavailable.

## Persona Red Flags

- First-time readers must infer desktop icon meanings until hover; “Archive,” “Admin,” and presence concepts are not fully explained (`web/src/lib/ui/layout/sidebar/Sidebar.svelte:60`).
- Accessibility-dependent readers encounter hover-dependent controls, a mobile TOC trigger that may lack an explicit label, and disabled user scaling (`web/src/lib/ui/layout/sidebar/MobileNavBar.svelte:247`, `web/src/routes/+layout.svelte:368`).
- Distracted mobile readers face a second full-screen menu layer and compact 36px controls (`web/src/lib/ui/layout/sidebar/MobileNavBar.svelte:183`).
- Direct-entry and stress readers may encounter weak recovery for missing content, clipboard failure, long titles, and zero-result archives (`web/src/lib/features/post/components/PostList.svelte:128`).

## Minor Observations

- `history.back()` can take a direct-entry reader outside the site (`web/src/lib/features/post/components/post-detail/PostDetailHeader.svelte:39`).
- “查看原文” is opacity-zero until hover, making it invisible to keyboard and non-pointer readers (`web/src/lib/features/post/components/ArticleItem.svelte:95`).
- Homepage and archive list items use different densities and interaction treatments (`web/src/lib/features/post/components/HomeArticleItem.svelte:28`, `web/src/lib/features/post/components/ArticleItem.svelte:30`).
- “没有知识的荒原” is memorable but displaces practical empty-state guidance (`web/src/lib/features/post/components/PostList.svelte:138`).

## Provocative Questions

- Is the homepage primarily a publication front page, a creator profile, or a realtime social room? What would be removed if it had to be only one?
- If hover effects, noise, blur transitions, and presence indicators disappeared, would the site still feel unmistakably like this creator’s place?
- Why should a reader care how many people are online while reading a long-form article?
- Does the desktop icon rail optimize for visual minimalism at the expense of first-use comprehension?
- What should the emotional end-state after an article be: comment, continue reading, follow the creator, or simply leave satisfied?
