import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { CreateOrderInput } from '../../types/api'
import { productKeys } from '../product/queries'
import { cancelOrder, createOrder, getOrder } from './api'

export const orderKeys = {
  detail: (id: number) => ['orders', id] as const,
}

export const useOrder = (id: number) =>
  useQuery({
    queryKey: orderKeys.detail(id),
    queryFn: () => getOrder(id),
  })

export const useCreateOrder = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateOrderInput) => createOrder(input),
    onSuccess: (order) => {
      queryClient.setQueryData(orderKeys.detail(order.id), order)
      return queryClient.invalidateQueries({ queryKey: productKeys.all })
    },
  })
}

export const useCancelOrder = (id: number) => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => cancelOrder(id),
    onSuccess: (order) => {
      queryClient.setQueryData(orderKeys.detail(id), order)
      return queryClient.invalidateQueries({ queryKey: productKeys.all })
    },
  })
}
