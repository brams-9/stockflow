import { z } from 'zod'
import { request } from '../../lib/apiClient'
import {
  movementListSchema,
  productListSchema,
  productSchema,
  stockMovementSchema,
  type AdjustStockInput,
  type CreateProductInput,
  type MovementList,
  type Product,
  type ProductList,
  type StockMovement,
  type UpdateProductInput,
} from '../../types/api'

type ListProductsParams = {
  search?: string
  page?: number
  limit?: number
}

export const listProducts = (params: ListProductsParams = {}): Promise<ProductList> =>
  request({
    method: 'GET',
    path: '/api/products',
    query: params,
    schema: productListSchema,
  })

export const createProduct = (input: CreateProductInput): Promise<Product> =>
  request({
    method: 'POST',
    path: '/api/products',
    body: input,
    schema: productSchema,
  })

export const updateProduct = (id: number, input: UpdateProductInput): Promise<Product> =>
  request({
    method: 'PUT',
    path: `/api/products/${id}`,
    body: input,
    schema: productSchema,
  })

export const deleteProduct = (id: number): Promise<void> =>
  request({
    method: 'DELETE',
    path: `/api/products/${id}`,
    schema: z.undefined(),
  })

export const adjustStock = (id: number, input: AdjustStockInput): Promise<StockMovement> =>
  request({
    method: 'POST',
    path: `/api/products/${id}/stock`,
    body: input,
    schema: stockMovementSchema,
  })

export const listMovements = (id: number): Promise<MovementList> =>
  request({
    method: 'GET',
    path: `/api/products/${id}/movements`,
    schema: movementListSchema,
  })
