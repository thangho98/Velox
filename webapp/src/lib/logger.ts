const DEBUG = import.meta.env.VITE_DEBUG === 'true' || import.meta.env.DEV

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const LEVEL_COLORS: Record<LogLevel, string> = {
  debug: '#8B8B8B',
  info: '#4FC3F7',
  warn: '#FFB74D',
  error: '#EF5350',
}

function createLogger(namespace: string) {
  const prefix = `[${namespace}]`

  function log(level: LogLevel, msg: string, ...args: unknown[]) {
    if (level === 'debug' && !DEBUG) return
    const color = LEVEL_COLORS[level]
    const method = level === 'error' ? 'error' : level === 'warn' ? 'warn' : 'log'
    console[method](`%c${prefix} ${msg}`, `color: ${color}; font-weight: bold`, ...args)
  }

  return {
    debug: (msg: string, ...args: unknown[]) => log('debug', msg, ...args),
    info: (msg: string, ...args: unknown[]) => log('info', msg, ...args),
    warn: (msg: string, ...args: unknown[]) => log('warn', msg, ...args),
    error: (msg: string, ...args: unknown[]) => log('error', msg, ...args),
    /** Log with timing — returns a function to call when the operation completes */
    time: (label: string) => {
      if (!DEBUG) return () => {}
      const start = performance.now()
      log('debug', `⏱ ${label} started`)
      return (extra?: string) => {
        const ms = (performance.now() - start).toFixed(1)
        log('debug', `⏱ ${label} completed in ${ms}ms${extra ? ` — ${extra}` : ''}`)
      }
    },
  }
}

export { createLogger, DEBUG }
