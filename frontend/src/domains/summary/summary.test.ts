import { describe, expect, it } from 'vitest'
import type { Product } from '../../types/api'
import { summarize } from './summary'

const make = (overrides: Partial<Product>): Product => ({
  id: 1,
  sku: 'A-1',
  name: 'Item',
  price: 1000,
  stock: 20,
  is_active: true,
  created_at: '2026-01-01T00:00:00Z',
  last_movement_at: null,
  ...overrides,
})

describe('summarize', () => {
  it('derives the dashboard numbers from the product list', () => {
    const products = [
      make({ id: 1, price: 1000, stock: 20 }),
      make({ id: 2, price: 500, stock: 10, is_active: false }),
      make({ id: 3, price: 200, stock: 3 }),
    ]
    expect(summarize(products, 3)).toEqual({
      totalProducts: 3,
      activeProducts: 2,
      stockValue: 20000 + 5000 + 600,
      lowStock: 2,
    })
  })

  it('handles an empty list', () => {
    expect(summarize([], 0)).toEqual({
      totalProducts: 0,
      activeProducts: 0,
      stockValue: 0,
      lowStock: 0,
    })
  })
})
