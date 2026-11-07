import { z } from 'zod'

export const errorCodes = {
  validation: 'VALIDATION_ERROR',
  skuAlreadyExists: 'SKU_ALREADY_EXISTS',
  productNotFound: 'PRODUCT_NOT_FOUND',
  productHasHistory: 'PRODUCT_HAS_HISTORY',
  insufficientStock: 'INSUFFICIENT_STOCK',
  orderNotFound: 'ORDER_NOT_FOUND',
  orderAlreadyCancelled: 'ORDER_ALREADY_CANCELLED',
  notFound: 'NOT_FOUND',
  methodNotAllowed: 'METHOD_NOT_ALLOWED',
  internal: 'INTERNAL_ERROR',
  network: 'NETWORK_ERROR',
  unknown: 'UNKNOWN_ERROR',
  invalidResponse: 'INVALID_RESPONSE',
} as const

export const movementTypeSchema = z.enum(['in', 'out'])
export const orderStatusSchema = z.enum(['CREATED', 'CANCELLED'])

export const productSchema = z.object({
  id: z.number().int(),
  sku: z.string(),
  name: z.string(),
  price: z.number().int().nonnegative(),
  stock: z.number().int(),
  is_active: z.boolean(),
  created_at: z.string(),
  last_movement_at: z.string().nullable(),
})

export const stockMovementSchema = z.object({
  id: z.number().int(),
  product_id: z.number().int(),
  type: movementTypeSchema,
  qty: z.number().int().positive(),
  note: z.string().nullable(),
  created_at: z.string(),
})

export const orderItemSchema = z.object({
  id: z.number().int(),
  sku: z.string(),
  qty: z.number().int().positive(),
  unit_price: z.number().int().nonnegative(),
})

export const orderSchema = z.object({
  id: z.number().int(),
  customer: z.string(),
  status: orderStatusSchema,
  total: z.number().int().nonnegative(),
  created_at: z.string(),
  cancelled_at: z.string().nullable(),
  items: z.array(orderItemSchema),
})

export const pagedSchema = <T extends z.ZodType>(item: T) =>
  z.object({
    data: z.array(item),
    page: z.number().int().positive(),
    limit: z.number().int().positive(),
    total: z.number().int().nonnegative(),
  })

export const productListSchema = pagedSchema(productSchema)

export const movementListSchema = z.object({
  data: z.array(stockMovementSchema),
})

export const apiErrorSchema = z.object({
  error: z.object({
    code: z.string(),
    message: z.string().optional(),
    sku: z.string().optional(),
    available: z.number().int().optional(),
    fields: z.record(z.string(), z.string()).optional(),
  }),
})

export const createProductSchema = z.object({
  sku: z.string().trim().min(1),
  name: z.string().trim().min(1),
  price: z.number().int().nonnegative(),
})

export const updateProductSchema = createProductSchema.extend({
  is_active: z.boolean(),
})

export const adjustStockSchema = z.object({
  type: movementTypeSchema,
  qty: z.number().int().positive(),
  note: z.string().optional(),
})

export const createOrderItemSchema = z.object({
  sku: z.string().trim().min(1),
  qty: z.number().int().positive(),
})

export const createOrderSchema = z.object({
  customer: z.string().trim().min(1),
  items: z.array(createOrderItemSchema).min(1),
})

export type MovementType = z.infer<typeof movementTypeSchema>
export type OrderStatus = z.infer<typeof orderStatusSchema>
export type Product = z.infer<typeof productSchema>
export type StockMovement = z.infer<typeof stockMovementSchema>
export type OrderItem = z.infer<typeof orderItemSchema>
export type Order = z.infer<typeof orderSchema>
export type ProductList = z.infer<typeof productListSchema>
export type MovementList = z.infer<typeof movementListSchema>
export type ApiError = z.infer<typeof apiErrorSchema>
export type CreateProductInput = z.infer<typeof createProductSchema>
export type UpdateProductInput = z.infer<typeof updateProductSchema>
export type AdjustStockInput = z.infer<typeof adjustStockSchema>
export type CreateOrderItemInput = z.infer<typeof createOrderItemSchema>
export type CreateOrderInput = z.infer<typeof createOrderSchema>
