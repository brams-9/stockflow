import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AdjustStockInput, CreateProductInput, UpdateProductInput } from '../../types/api'
import {
  adjustStock,
  createProduct,
  deleteProduct,
  listMovements,
  listProducts,
  updateProduct,
} from './api'

type ProductListParams = {
  search?: string
  page?: number
  limit?: number
}

export const productKeys = {
  all: ['products'] as const,
  list: (params: ProductListParams) => ['products', 'list', params] as const,
  movements: (id: number) => ['products', 'movements', id] as const,
}

export const useProducts = (params: ProductListParams) =>
  useQuery({
    queryKey: productKeys.list(params),
    queryFn: () => listProducts(params),
    placeholderData: keepPreviousData,
  })

export const useAllProducts = () => useProducts({ limit: 100 })

export const useMovements = (id: number) =>
  useQuery({
    queryKey: productKeys.movements(id),
    queryFn: () => listMovements(id),
  })

export const useAdjustStock = (id: number) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: AdjustStockInput) => adjustStock(id, input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: productKeys.all }),
  })
}

export const useCreateProduct = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateProductInput) => createProduct(input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: productKeys.all }),
  })
}

export const useUpdateProduct = (id: number) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateProductInput) => updateProduct(id, input),
    onSettled: () => queryClient.invalidateQueries({ queryKey: productKeys.all }),
  })
}

export const useDeleteProduct = (id: number) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => deleteProduct(id),
    onSettled: () => queryClient.invalidateQueries({ queryKey: productKeys.all }),
  })
}
