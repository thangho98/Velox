const TOKEN_URL = 'https://anilist.co/api/v2/oauth/token'

function readJsonBody(req) {
  return new Promise((resolve, reject) => {
    let raw = ''

    req.on('data', (chunk) => {
      raw += chunk
    })

    req.on('end', () => {
      if (!raw) {
        resolve({})
        return
      }

      try {
        resolve(JSON.parse(raw))
      } catch (error) {
        reject(error)
      }
    })

    req.on('error', reject)
  })
}

async function extractRequestData(req) {
  if (req.method === 'GET') {
    return {
      code: req.query.code,
      redirectUri: req.query.redirectUri,
      clientId: req.query.clientId,
    }
  }

  const body = await readJsonBody(req)

  return {
    code: body.code,
    redirectUri: body.redirectUri,
    clientId: body.clientId,
  }
}

export default async function handler(req, res) {
  if (req.method !== 'POST' && req.method !== 'GET') {
    res.status(405).json({ error: 'method_not_allowed' })
    return
  }

  try {
    const { code, redirectUri, clientId } = await extractRequestData(req)
    const clientSecret = process.env.ANILIST_CLIENT_SECRET?.trim() ?? ''

    if (!clientId || !redirectUri || !code) {
      res.status(400).json({
        error: 'invalid_request',
        message: 'clientId, redirectUri, and code are required',
      })
      return
    }

    if (!clientSecret) {
      res.status(500).json({
        error: 'server_not_configured',
        message: 'ANILIST_CLIENT_SECRET is missing on the server',
      })
      return
    }

    const tokenResponse = await fetch(TOKEN_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify({
        grant_type: 'authorization_code',
        client_id: Number(clientId),
        client_secret: clientSecret,
        redirect_uri: redirectUri,
        code,
      }),
    })

    const body = await tokenResponse.json()

    if (!tokenResponse.ok) {
      res.status(tokenResponse.status).json({
        error: body.error ?? 'token_exchange_failed',
        message: body.error_description ?? body.message ?? 'AniList token exchange failed',
        hint: body.hint ?? null,
      })
      return
    }

    res.status(200).json({
      accessToken: body.access_token,
      tokenType: body.token_type,
    })
  } catch (error) {
    res.status(500).json({
      error: 'unexpected_error',
      message: error instanceof Error ? error.message : 'Unexpected exchange failure',
    })
  }
}
