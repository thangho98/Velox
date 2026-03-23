import { Providers } from '@/providers/Providers'
import { RouterProvider } from '@/providers/Router'
import { ErrorBoundary } from '@/components/ErrorBoundary'

export function App() {
  return (
    <ErrorBoundary>
      <Providers>
        <RouterProvider />
      </Providers>
    </ErrorBoundary>
  )
}
