import { QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './index.css'
import { queryClient } from './lib/queryClient'
import CreateOrderPage from './pages/CreateOrderPage'
import OrderDetailPage from './pages/OrderDetailPage'
import ProductsPage from './pages/ProductsPage'

const router = createBrowserRouter([
  { path: '/', element: <ProductsPage /> },
  { path: '/orders/new', element: <CreateOrderPage /> },
  { path: '/orders/:id', element: <OrderDetailPage /> },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
