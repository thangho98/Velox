import { Providers } from '@/providers/Providers'
import { RouterProvider } from '@/providers/Router'

export function App() {
  return (
    <Providers>
      <RouterProvider />
    </Providers>
  )
}
