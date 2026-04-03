import { useState, useEffect, useRef } from 'react'
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
} from 'react-native'
import { Eye, EyeOff, Wifi, WifiOff, RefreshCw } from 'lucide-react-native'
import { useLogin } from '@velox/shared/hooks/auth'
import { getServerUrl, setServerUrl } from '../platform/mobile-adapter'
import { useResponsiveLayout, scaledFont, scaledSpacing } from '../lib/responsive'

interface LoginScreenProps {
  onLoginSuccess?: () => void
}

type ConnectionStatus = 'unknown' | 'checking' | 'connected' | 'failed'

/** Ping GET /api/health and return { ok, latencyMs, error? } */
async function checkServerConnection(url: string): Promise<{
  ok: boolean
  latencyMs: number
  error?: string
}> {
  const start = Date.now()
  try {
    const apiUrl = `${url.replace(/\/+$/, '')}/api/health`
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), 8000)
    const res = await fetch(apiUrl, { method: 'GET', signal: controller.signal })
    clearTimeout(timeout)
    const latencyMs = Date.now() - start

    if (!res.ok) {
      return { ok: false, latencyMs, error: `Server returned HTTP ${res.status}` }
    }

    // Verify response is the expected health check
    try {
      const body = await res.json()
      if (body?.status === 'ok') {
        return { ok: true, latencyMs }
      }
    } catch {
      // Response wasn't JSON but status was 200 — still consider it alive
    }

    return { ok: true, latencyMs }
  } catch (err) {
    const latencyMs = Date.now() - start
    if (err instanceof Error && err.name === 'AbortError') {
      return { ok: false, latencyMs, error: 'Connection timed out (8s)' }
    }
    const message =
      err instanceof TypeError
        ? 'Cannot reach server. Check URL and network.'
        : err instanceof Error
          ? err.message
          : 'Connection failed'
    return { ok: false, latencyMs, error: message }
  }
}

export function LoginScreen({ onLoginSuccess }: LoginScreenProps) {
  const hasServer = !!getServerUrl()
  const [step, setStep] = useState<'server' | 'login'>(hasServer ? 'login' : 'server')
  const [serverUrl, setServerUrlInput] = useState(getServerUrl() ?? '')
  const [isConnecting, setIsConnecting] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  // Connection check state
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('unknown')
  const [connectionInfo, setConnectionInfo] = useState<{
    latencyMs?: number
    error?: string
  }>({})
  const hasCheckedRef = useRef(false)

  const login = useLogin()
  const layout = useResponsiveLayout()

  // On tablet/TV: constrain the form width
  const formMaxWidth =
    layout.device === 'tv' ? 480 : layout.device === 'tablet' ? 400 : undefined
  const inputPaddingVertical = layout.largeControls ? 20 : 16

  // Auto-check connection when entering login step with a saved server
  useEffect(() => {
    if (step === 'login' && hasServer && !hasCheckedRef.current) {
      hasCheckedRef.current = true
      handleTestConnection()
    }
  }, [step])

  const handleTestConnection = async () => {
    const url = getServerUrl()
    if (!url) return

    setConnectionStatus('checking')
    setConnectionInfo({})
    const result = await checkServerConnection(url)

    if (result.ok) {
      setConnectionStatus('connected')
      setConnectionInfo({ latencyMs: result.latencyMs })
      setErrorMessage(null)
    } else {
      setConnectionStatus('failed')
      setConnectionInfo({ error: result.error, latencyMs: result.latencyMs })
      setErrorMessage(result.error || 'Connection failed')
    }
  }

  const handleConnect = async () => {
    setErrorMessage(null)
    let url = serverUrl.trim()
    if (!url) {
      setErrorMessage('Please enter your server URL')
      return
    }
    // Auto-add http:// if no protocol
    if (!/^https?:\/\//i.test(url)) {
      url = `http://${url}`
      setServerUrlInput(url)
    }

    setIsConnecting(true)
    const result = await checkServerConnection(url)
    setIsConnecting(false)

    if (result.ok) {
      await setServerUrl(url)
      setConnectionStatus('connected')
      setConnectionInfo({ latencyMs: result.latencyMs })
      hasCheckedRef.current = true
      setStep('login')
    } else {
      setErrorMessage(result.error || 'Could not connect to server.')
    }
  }

  const handleChangeServer = () => {
    setStep('server')
    setErrorMessage(null)
    setConnectionStatus('unknown')
    setConnectionInfo({})
    hasCheckedRef.current = false
  }

  const handleLogin = async () => {
    setErrorMessage(null)
    if (!username.trim() || !password.trim()) {
      setErrorMessage('Please enter username and password')
      return
    }

    try {
      await login.mutateAsync({ username: username.trim(), password })
      onLoginSuccess?.()
    } catch (error) {
      // Detect connection errors vs auth errors
      const message = error instanceof Error ? error.message : 'Login failed'
      if (message.includes('Network') || message.includes('fetch')) {
        setConnectionStatus('failed')
        setConnectionInfo({ error: 'Lost connection to server' })
        setErrorMessage('Cannot reach server. Check your connection and try again.')
      } else {
        setErrorMessage(message)
      }
    }
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <View style={[styles.content, { padding: layout.screenPadding }]}>
        {/* Centered form wrapper for tablet/TV */}
        <View
          style={[
            styles.formWrapper,
            formMaxWidth ? { maxWidth: formMaxWidth, width: '100%', alignSelf: 'center' } : undefined,
          ]}
        >
          {/* Logo */}
          <Text style={[styles.logo, { fontSize: scaledFont(48, layout.fontScale) }]}>VELOX</Text>
          <Text
            style={[styles.subtitle, { fontSize: scaledFont(16, layout.fontScale), marginBottom: scaledSpacing(48, layout.fontScale) }]}
          >
            Home Media Server
          </Text>

          {step === 'server' ? (
            <View style={styles.form}>
              <Text
                style={[styles.label, { fontSize: scaledFont(14, layout.fontScale) }]}
              >
                Server URL
              </Text>
              <TextInput
                style={[
                  styles.input,
                  {
                    fontSize: scaledFont(16, layout.fontScale),
                    padding: inputPaddingVertical,
                  },
                ]}
                placeholder="http://192.168.1.100:8080"
                placeholderTextColor="#666"
                value={serverUrl}
                onChangeText={setServerUrlInput}
                autoCapitalize="none"
                autoCorrect={false}
                keyboardType="url"
                editable={!isConnecting}
              />
              <Text
                style={[styles.hint, { fontSize: scaledFont(13, layout.fontScale) }]}
              >
                Enter the URL of your Velox server
              </Text>

              {/* Inline Error */}
              {errorMessage && (
                <View style={styles.errorBanner}>
                  <Text
                    style={[styles.errorText, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {errorMessage}
                  </Text>
                </View>
              )}

              <TouchableOpacity
                style={[
                  styles.button,
                  { padding: inputPaddingVertical },
                  isConnecting && styles.buttonDisabled,
                ]}
                onPress={handleConnect}
                disabled={isConnecting}
              >
                {isConnecting ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text
                    style={[styles.buttonText, { fontSize: scaledFont(16, layout.fontScale) }]}
                  >
                    Connect
                  </Text>
                )}
              </TouchableOpacity>
            </View>
          ) : (
            <View style={styles.form}>
              {/* Server info + connection status */}
              <TouchableOpacity onPress={handleChangeServer} style={styles.serverInfo}>
                <Text
                  style={[styles.serverLabel, { fontSize: scaledFont(13, layout.fontScale) }]}
                >
                  Server
                </Text>
                <Text
                  style={[styles.serverUrl, { fontSize: scaledFont(14, layout.fontScale) }]}
                  numberOfLines={1}
                >
                  {getServerUrl()}
                </Text>
                <Text
                  style={[styles.changeText, { fontSize: scaledFont(13, layout.fontScale) }]}
                >
                  Change
                </Text>
              </TouchableOpacity>

              {/* Connection Status Banner */}
              <View style={[
                styles.connectionBanner,
                connectionStatus === 'connected' && styles.connectionOk,
                connectionStatus === 'failed' && styles.connectionFail,
                connectionStatus === 'checking' && styles.connectionChecking,
              ]}>
                <View style={styles.connectionRow}>
                  {connectionStatus === 'checking' ? (
                    <ActivityIndicator size="small" color="#888" />
                  ) : connectionStatus === 'connected' ? (
                    <Wifi size={scaledFont(16, layout.fontScale)} color="#4ade80" />
                  ) : connectionStatus === 'failed' ? (
                    <WifiOff size={scaledFont(16, layout.fontScale)} color="#e50914" />
                  ) : (
                    <Wifi size={scaledFont(16, layout.fontScale)} color="#666" />
                  )}
                  <Text style={[
                    styles.connectionText,
                    { fontSize: scaledFont(13, layout.fontScale) },
                    connectionStatus === 'connected' && { color: '#4ade80' },
                    connectionStatus === 'failed' && { color: '#e50914' },
                  ]}>
                    {connectionStatus === 'checking'
                      ? 'Checking connection...'
                      : connectionStatus === 'connected'
                        ? `Connected${connectionInfo.latencyMs ? ` · ${connectionInfo.latencyMs}ms` : ''}`
                        : connectionStatus === 'failed'
                          ? connectionInfo.error || 'Connection failed'
                          : 'Not checked'}
                  </Text>
                  <TouchableOpacity
                    onPress={handleTestConnection}
                    disabled={connectionStatus === 'checking'}
                    style={styles.retryButton}
                  >
                    <RefreshCw
                      size={scaledFont(14, layout.fontScale)}
                      color={connectionStatus === 'checking' ? '#444' : '#888'}
                    />
                  </TouchableOpacity>
                </View>
              </View>

              {/* Inline Error */}
              {errorMessage && connectionStatus !== 'failed' && (
                <View style={styles.errorBanner}>
                  <Text
                    style={[styles.errorText, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {errorMessage}
                  </Text>
                </View>
              )}

              {/* Connection failed — show retry guidance */}
              {connectionStatus === 'failed' && (
                <View style={styles.errorBanner}>
                  <Text
                    style={[styles.errorText, { fontSize: scaledFont(14, layout.fontScale), marginBottom: 8 }]}
                  >
                    {connectionInfo.error || 'Cannot connect to server'}
                  </Text>
                  <Text style={[styles.errorHint, { fontSize: scaledFont(12, layout.fontScale) }]}>
                    • Check that the server is running{'\n'}
                    • Verify the URL is correct{'\n'}
                    • Make sure you're on the same network
                  </Text>
                  <TouchableOpacity
                    style={styles.retryBannerButton}
                    onPress={handleTestConnection}
                  >
                    <RefreshCw size={scaledFont(14, layout.fontScale)} color="#fff" />
                    <Text style={[styles.retryBannerText, { fontSize: scaledFont(13, layout.fontScale) }]}>
                      Retry Connection
                    </Text>
                  </TouchableOpacity>
                </View>
              )}

              <TextInput
                style={[
                  styles.input,
                  {
                    fontSize: scaledFont(16, layout.fontScale),
                    padding: inputPaddingVertical,
                  },
                ]}
                placeholder="Username"
                placeholderTextColor="#666"
                value={username}
                onChangeText={setUsername}
                autoCapitalize="none"
                autoCorrect={false}
                editable={!login.isPending}
              />

              <View style={styles.passwordContainer}>
                <TextInput
                  style={[
                    styles.passwordInput,
                    {
                      fontSize: scaledFont(16, layout.fontScale),
                      padding: inputPaddingVertical,
                    },
                  ]}
                  placeholder="Password"
                  placeholderTextColor="#666"
                  value={password}
                  onChangeText={setPassword}
                  secureTextEntry={!showPassword}
                  editable={!login.isPending}
                />
                <TouchableOpacity
                  style={styles.eyeButton}
                  onPress={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <Eye size={scaledFont(20, layout.fontScale)} color="#888" />
                  ) : (
                    <EyeOff size={scaledFont(20, layout.fontScale)} color="#888" />
                  )}
                </TouchableOpacity>
              </View>

              <TouchableOpacity
                style={[
                  styles.button,
                  { padding: inputPaddingVertical },
                  login.isPending && styles.buttonDisabled,
                ]}
                onPress={handleLogin}
                disabled={login.isPending}
              >
                {login.isPending ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <Text
                    style={[styles.buttonText, { fontSize: scaledFont(16, layout.fontScale) }]}
                  >
                    Sign In
                  </Text>
                )}
              </TouchableOpacity>

              {/* New user text */}
              <Text
                style={[styles.newUserText, { fontSize: scaledFont(14, layout.fontScale) }]}
              >
                New user? Contact administrator
              </Text>
            </View>
          )}
        </View>

        {/* Footer Links */}
        <View style={styles.footer}>
          <TouchableOpacity>
            <Text style={[styles.footerLink, { fontSize: scaledFont(13, layout.fontScale) }]}>
              Privacy
            </Text>
          </TouchableOpacity>
          <Text style={[styles.footerDot, { fontSize: scaledFont(13, layout.fontScale) }]}>
            •
          </Text>
          <TouchableOpacity>
            <Text style={[styles.footerLink, { fontSize: scaledFont(13, layout.fontScale) }]}>
              Terms
            </Text>
          </TouchableOpacity>
          <Text style={[styles.footerDot, { fontSize: scaledFont(13, layout.fontScale) }]}>
            •
          </Text>
          <TouchableOpacity>
            <Text style={[styles.footerLink, { fontSize: scaledFont(13, layout.fontScale) }]}>
              Help
            </Text>
          </TouchableOpacity>
        </View>
      </View>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },
  formWrapper: {
    // Max-width applied dynamically for tablet/TV
  },
  logo: {
    fontSize: 48,
    fontWeight: 'bold',
    color: '#e50914',
    textAlign: 'center',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#888',
    textAlign: 'center',
    marginBottom: 48,
  },
  form: {
    gap: 16,
  },
  label: {
    fontSize: 14,
    color: '#aaa',
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
  },
  input: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    padding: 16,
    fontSize: 16,
    color: '#fff',
    borderWidth: 1,
    borderColor: '#333',
  },
  hint: {
    fontSize: 13,
    color: '#666',
  },
  errorBanner: {
    backgroundColor: 'rgba(229, 9, 20, 0.15)',
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
    borderColor: '#e50914',
  },
  errorText: {
    color: '#e50914',
    fontSize: 14,
    textAlign: 'center',
  },
  button: {
    backgroundColor: '#e50914',
    borderRadius: 8,
    padding: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonDisabled: {
    opacity: 0.7,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  serverInfo: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
    borderColor: '#333',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  serverLabel: {
    fontSize: 13,
    color: '#888',
  },
  serverUrl: {
    fontSize: 14,
    color: '#fff',
    flex: 1,
  },
  changeText: {
    fontSize: 13,
    color: '#e50914',
    fontWeight: '600',
  },
  connectionBanner: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    padding: 10,
    borderWidth: 1,
    borderColor: '#2a2a2a',
  },
  connectionOk: {
    borderColor: 'rgba(74, 222, 128, 0.3)',
    backgroundColor: 'rgba(74, 222, 128, 0.05)',
  },
  connectionFail: {
    borderColor: 'rgba(229, 9, 20, 0.3)',
    backgroundColor: 'rgba(229, 9, 20, 0.05)',
  },
  connectionChecking: {
    borderColor: '#333',
  },
  connectionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  connectionText: {
    flex: 1,
    fontSize: 13,
    color: '#888',
  },
  retryButton: {
    padding: 6,
  },
  errorHint: {
    color: '#999',
    fontSize: 12,
    lineHeight: 20,
  },
  retryBannerButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    backgroundColor: '#e50914',
    borderRadius: 6,
    paddingVertical: 10,
    marginTop: 12,
  },
  retryBannerText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  passwordContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#333',
  },
  passwordInput: {
    flex: 1,
    padding: 16,
    fontSize: 16,
    color: '#fff',
  },
  eyeButton: {
    padding: 12,
  },
  newUserText: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
    marginTop: 8,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 48,
    gap: 12,
  },
  footerLink: {
    fontSize: 13,
    color: '#666',
  },
  footerDot: {
    fontSize: 13,
    color: '#444',
  },
})
