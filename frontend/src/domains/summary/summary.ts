import type { Product } from '../../types/api'
import { isLowStock } from '../../utils/format'

export type Summary = {
  totalProducts: number
  activeProducts: number
  stockValue: number
  lowStock: number
}

export const summarize = (products: Product[], total: number): Summary => ({
  totalProducts: total,
  activeProducts: products.filter((product) => product.is_active).length,
  stockValue: products.reduce((sum, product) => sum + product.price * product.stock, 0),
  lowStock: products.filter((product) => isLowStock(product.stock)).length,
})
