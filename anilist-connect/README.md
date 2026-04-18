# Velox AniList Connect (Unofficial)

Small static helper app that lets a user:

1. sign in with AniList
2. exchange the returned authorization code through a small Vercel serverless function
3. verify the AniList OAuth access token with the AniList `Viewer` query
4. copy the raw token and paste it into Velox manually

This app is designed for cheap static deployment on Vercel, Cloudflare Pages, or Netlify.

## Scope and disclaimer

- This helper is intended for personal Velox homeserver media setups.
- It is not designed or marketed as a commercial AniList integration service.
- AniList access tokens are sensitive credentials.
- Users are responsible for how they copy, store, share, or paste those tokens.
- This helper is provided as-is and does not assume liability for token leaks caused by user handling, local device compromise, clipboard history, screenshots, browser extensions, or shared environments.

## Why this exists

Velox is self-hosted, so each user may run it on a different NAS URL or local domain. That makes native OAuth callback handling awkward. This helper keeps a single fixed public callback URL and turns AniList login into a simple copy-paste flow.

## Environment variables

Create a `.env.local` from `.env.example`.

- `VITE_ANILIST_CLIENT_ID`
  Required. The AniList OAuth client ID for this helper app.
- `VITE_ANILIST_REDIRECT_URI`
  Optional. If omitted, the app uses the current page URL as the redirect URI.
- `ANILIST_CLIENT_SECRET`
  Required on Vercel. Used only by the serverless token exchange endpoint.

## AniList app setup

1. Open AniList developer settings and create an OAuth client.
2. Set the redirect URI to the final deployed URL of this helper app.
   Example:
   `https://velox-anilist-connect.vercel.app/`
3. Copy the AniList client ID into `VITE_ANILIST_CLIENT_ID`.

## Local development

```bash
pnpm install
pnpm --filter @velox/anilist-connect dev
```

## Build

```bash
pnpm --filter @velox/anilist-connect build
```

## Deploy on Vercel

1. Import this repo into Vercel.
2. Set the project root to `anilist-connect`.
3. `vercel.json` is already included for Vite SPA rewrites, so deep links keep working.
4. Add `VITE_ANILIST_CLIENT_ID` in project environment variables.
5. Add `ANILIST_CLIENT_SECRET` in project environment variables.
6. Optionally add `VITE_ANILIST_REDIRECT_URI` if you do not want to rely on the page URL automatically.
7. Deploy.

## Token handling

- The helper stores the token only in `sessionStorage`.
- The URL hash is cleared after AniList redirects back.
- The app does not send the token to Velox automatically.
- The user must copy and paste the raw token manually.

## Notes

- This app uses the AniList authorization code grant with a Vercel serverless exchange endpoint.
- Use this only when you need AniList account-specific features.
- Basic anime metadata fetching in Velox can still work without OAuth.
