import { useEffect, useMemo, useState } from 'react'

type AuthFragment = {
  error?: string
  errorDescription?: string
}

type Viewer = {
  id: number
  name: string
  avatar?: {
    large?: string | null
  } | null
}

const SESSION_STORAGE_KEY = 'velox.anilist.connect.token'
const CLIENT_ID = import.meta.env.VITE_ANILIST_CLIENT_ID?.trim() ?? ''
const CONFIGURED_REDIRECT_URI = import.meta.env.VITE_ANILIST_REDIRECT_URI?.trim() ?? ''

function buildRedirectUri() {
  if (CONFIGURED_REDIRECT_URI) {
    return CONFIGURED_REDIRECT_URI
  }

  if (typeof window === 'undefined') {
    return ''
  }

  return `${window.location.origin}${window.location.pathname}`
}

function parseAuthFragment(hash: string): AuthFragment {
  const fragment = hash.startsWith('#') ? hash.slice(1) : hash
  const params = new URLSearchParams(fragment)

  return {
    error: params.get('error') ?? undefined,
    errorDescription: params.get('error_description') ?? undefined,
  }
}

function createAuthorizeUrl(clientId: string, redirectUri: string) {
  const params = new URLSearchParams({
    client_id: clientId,
    redirect_uri: redirectUri,
    response_type: 'code',
  })

  return `https://anilist.co/api/v2/oauth/authorize?${params.toString()}`
}

async function exchangeCode(code: string, redirectUri: string, clientId: string) {
  const response = await fetch('/api/oauth/exchange', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify({
      code,
      redirectUri,
      clientId,
    }),
  })

  const body = await response.json()

  if (!response.ok) {
    throw new Error(body.message ?? body.error ?? 'Failed to exchange AniList authorization code')
  }

  return body as { accessToken: string; tokenType?: string }
}

async function fetchViewer(accessToken: string): Promise<Viewer> {
  const response = await fetch('https://graphql.anilist.co', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({
      query: `
        query Viewer {
          Viewer {
            id
            name
            avatar {
              large
            }
          }
        }
      `,
    }),
  })

  const body = await response.json()

  if (!response.ok || body.errors?.length) {
    const errorMessage =
      body.errors?.[0]?.message ??
      `AniList returned ${response.status}`

    throw new Error(errorMessage)
  }

  return body.data.Viewer as Viewer
}

export default function App() {
  const redirectUri = useMemo(() => buildRedirectUri(), [])
  const [accessToken, setAccessToken] = useState('')
  const [viewer, setViewer] = useState<Viewer | null>(null)
  const [isChecking, setIsChecking] = useState(false)
  const [statusMessage, setStatusMessage] = useState('')
  const [errorMessage, setErrorMessage] = useState('')
  const [copied, setCopied] = useState(false)
  const [isExchanging, setIsExchanging] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    const fragment = parseAuthFragment(window.location.hash)
    const query = new URLSearchParams(window.location.search)
    const authCode = query.get('code')
    const sessionToken = window.sessionStorage.getItem(SESSION_STORAGE_KEY) ?? ''

    if (fragment.error) {
      setErrorMessage(fragment.errorDescription ?? fragment.error)
    }

    if (authCode && CLIENT_ID) {
      setIsExchanging(true)
      setStatusMessage('Exchanging AniList authorization code...')
      setErrorMessage('')

      void exchangeCode(authCode, redirectUri, CLIENT_ID)
        .then(({ accessToken: nextToken }) => {
          setAccessToken(nextToken)
          window.sessionStorage.setItem(SESSION_STORAGE_KEY, nextToken)
          window.history.replaceState({}, document.title, redirectUri)
        })
        .catch((error: unknown) => {
          const message =
            error instanceof Error ? error.message : 'Failed to exchange AniList authorization code'
          setErrorMessage(message)
          setStatusMessage('')
        })
        .finally(() => {
          setIsExchanging(false)
        })

      return
    }

    if (sessionToken) {
      setAccessToken(sessionToken)
    }
  }, [redirectUri])

  useEffect(() => {
    if (!accessToken) {
      setViewer(null)
      setStatusMessage('')
      return
    }

    let cancelled = false

    setIsChecking(true)
    setErrorMessage('')
    setStatusMessage('Checking AniList token...')

    void fetchViewer(accessToken)
      .then((nextViewer) => {
        if (cancelled) {
          return
        }

        setViewer(nextViewer)
        setStatusMessage(`Connected as ${nextViewer.name}`)
      })
      .catch((error: unknown) => {
        if (cancelled) {
          return
        }

        const message = error instanceof Error ? error.message : 'Failed to verify AniList token'
        setViewer(null)
        setErrorMessage(message)
        setStatusMessage('')
      })
      .finally(() => {
        if (!cancelled) {
          setIsChecking(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [accessToken])

  const authorizeUrl = CLIENT_ID && redirectUri
    ? createAuthorizeUrl(CLIENT_ID, redirectUri)
    : ''

  const hasConfig = Boolean(CLIENT_ID && redirectUri)

  const handleConnect = () => {
    if (!authorizeUrl) {
      return
    }

    window.location.assign(authorizeUrl)
  }

  const handleCopy = async () => {
    if (!accessToken) {
      return
    }

    await navigator.clipboard.writeText(accessToken)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 2000)
  }

  const handleClear = () => {
    setAccessToken('')
    setViewer(null)
    setStatusMessage('')
    setErrorMessage('')
    setCopied(false)
    window.sessionStorage.removeItem(SESSION_STORAGE_KEY)
  }

  return (
    <main className="page-shell">
      <section className="hero-card">
        <div className="hero-copy">
          <p className="eyebrow">Velox Helper</p>
          <h1>Velox AniList Connect (Unofficial)</h1>
          <p className="lede">
            Sign in with AniList, verify the returned access token, then copy the raw token and
            paste it into your Velox user settings.
          </p>
          <div className="disclaimer-banner">
            <strong>Personal use only.</strong> This helper is intended for Velox homeserver media
            setups, not commercial services. If you copy, store, or share your AniList token
            insecurely, that risk stays with you.
          </div>
        </div>

        <div className="config-panel">
          <div>
            <span className="panel-label">Client ID</span>
            <code>{CLIENT_ID || 'Missing VITE_ANILIST_CLIENT_ID'}</code>
          </div>
          <div>
            <span className="panel-label">Redirect URI</span>
            <code>{redirectUri || 'Unavailable'}</code>
          </div>
          <div>
            <span className="panel-label">OAuth Flow</span>
            <code>Authorization code grant via Vercel function</code>
          </div>
        </div>
      </section>

      <section className="content-grid">
        <article className="step-card">
          <h2>How it works</h2>
          <ol>
            <li>Create an AniList OAuth client and set its redirect URI to this page URL.</li>
            <li>Press Connect AniList and approve access on AniList.</li>
            <li>This helper exchanges the returned authorization code on the server.</li>
            <li>The token is verified with the AniList Viewer query.</li>
            <li>Copy the token and paste it into Velox manually.</li>
          </ol>
          <p className="note">
            This helper keeps the token only in your current browser tab session so accidental
            reloads do not lose it immediately.
          </p>
          <p className="note">
            This project does not take responsibility for token leaks caused by screenshots,
            clipboard history, browser extensions, shared devices, or any other user-side handling.
          </p>
        </article>

        <article className="action-card">
          <h2>Connect</h2>
          <p className="section-copy">
            Use this only for account-specific AniList features. Basic anime metadata in Velox can
            still work without OAuth.
          </p>
          <div className="button-row">
            <button
              className="primary-button"
              onClick={handleConnect}
              disabled={!hasConfig || isExchanging}
              type="button"
            >
              {isExchanging ? 'Exchanging code...' : 'Connect AniList'}
            </button>
            <a
              className="secondary-button"
              href="https://anilist.gitbook.io/anilist-apiv2-docs"
              rel="noreferrer"
              target="_blank"
            >
              AniList Docs
            </a>
          </div>
          {!hasConfig ? (
            <p className="error-banner">
              Set <code>VITE_ANILIST_CLIENT_ID</code> before deploying.
            </p>
          ) : null}
          {statusMessage ? <p className="status-banner">{statusMessage}</p> : null}
          {errorMessage ? <p className="error-banner">{errorMessage}</p> : null}
          <p className="note legal-note">
            By using this helper, you acknowledge that AniList tokens are sensitive credentials and
            must be handled by you with care.
          </p>
        </article>

        <article className="token-card">
          <div className="token-header">
            <div>
              <h2>Access token</h2>
              <p className="section-copy">Copy the raw token value only. Do not share it publicly.</p>
            </div>
            {viewer?.avatar?.large ? (
              <img
                alt={`${viewer.name} AniList avatar`}
                className="viewer-avatar"
                src={viewer.avatar.large}
              />
            ) : null}
          </div>

          {viewer ? (
            <div className="viewer-badge">
              Connected account: <strong>{viewer.name}</strong> (Viewer ID {viewer.id})
            </div>
          ) : null}

          <div className="warning-box">
            Copy only into your own Velox homeserver. Do not paste this token into public chats,
            issue trackers, screenshots, shared notes, or commercial dashboards.
          </div>

          <textarea
            className="token-output"
            placeholder="AniList token will appear here after login"
            readOnly
            value={accessToken}
          />

          <div className="button-row">
            <button
              className="primary-button"
              disabled={!accessToken}
              onClick={() => {
                void handleCopy()
              }}
              type="button"
            >
              {copied ? 'Copied' : 'Copy token'}
            </button>
            <button
              className="ghost-button"
              disabled={!accessToken}
              onClick={handleClear}
              type="button"
            >
              Clear session token
            </button>
          </div>
        </article>
      </section>
    </main>
  )
}
