import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '../../../lib/apiClient'
import type { Product } from '../../../types/api'
import AdjustStockModal from './AdjustStockModal'

vi.mock('../api', () => ({
  adjustStock: vi.fn(),
  listMovements: vi.fn(),
  listProducts: vi.fn(),
}))

import { adjustStock } from '../api'

const product: Product = {
  id: 1,
  sku: 'A-1',
  name: 'Widget',
  price: 1000,
  stock: 3,
  is_active: true,
  created_at: '2026-01-01T00:00:00Z',
  last_movement_at: null,
}

const renderModal = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <AdjustStockModal product={product} onClose={() => {}} />
    </QueryClientProvider>,
  )

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

describe('AdjustStockModal', () => {
  it('blocks submit with an invalid quantity', async () => {
    renderModal()
    await userEvent.type(screen.getByLabelText('Quantity'), '0')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText(/whole number above 0/)).toBeTruthy()
    expect(adjustStock).not.toHaveBeenCalled()
  })

  it('shows how many are available on insufficient stock', async () => {
    vi.mocked(adjustStock).mockRejectedValue(
      new ApiError({
        status: 409,
        code: 'INSUFFICIENT_STOCK',
        message: 'not enough stock',
        available: 3,
      }),
    )
    renderModal()
    await userEvent.click(screen.getByRole('button', { name: 'Stock out' }))
    await userEvent.type(screen.getByLabelText('Quantity'), '9')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('Only 3 available')).toBeTruthy()
  })
})
