export type OrderLine = {
  sku: string
  qty: string
}

export type LineErrors = {
  sku?: string
  qty?: string
}

export type OrderErrors = {
  customer?: string
  items?: string
  lines: Record<number, LineErrors>
}

export const validateQty = (value: string | number): string | undefined => {
  const text = String(value).trim()
  if (text === '') return 'Quantity is required'
  const qty = Number(text)
  if (!Number.isInteger(qty) || qty <= 0) return 'Quantity must be a whole number above 0'
  return undefined
}

export const validateOrder = (customer: string, lines: OrderLine[]): OrderErrors => {
  const errors: OrderErrors = { lines: {} }

  if (!customer.trim()) errors.customer = 'Customer is required'
  if (lines.length === 0) errors.items = 'Add at least one item'

  const seen = new Set<string>()
  lines.forEach((line, index) => {
    const lineErrors: LineErrors = {}
    const sku = line.sku.trim()

    if (!sku) lineErrors.sku = 'Choose a product'
    else if (seen.has(sku)) lineErrors.sku = 'Product is already in this order'
    else seen.add(sku)

    const qtyError = validateQty(line.qty)
    if (qtyError) lineErrors.qty = qtyError

    if (lineErrors.sku || lineErrors.qty) errors.lines[index] = lineErrors
  })

  return errors
}

export const hasOrderErrors = (errors: OrderErrors): boolean =>
  Boolean(errors.customer || errors.items || Object.keys(errors.lines).length > 0)

const lineKey = /^items\[(\d+)\]\.(sku|qty)$/

export const mapServerFields = (fields: Record<string, string> = {}): OrderErrors => {
  const errors: OrderErrors = { lines: {} }

  for (const [key, message] of Object.entries(fields)) {
    if (key === 'customer') {
      errors.customer = message
    } else if (key === 'items') {
      errors.items = message
    } else {
      const match = lineKey.exec(key)
      if (!match) continue
      const index = Number(match[1])
      const field = match[2] as keyof LineErrors
      errors.lines[index] = { ...errors.lines[index], [field]: message }
    }
  }

  return errors
}

export type ProductFormValues = {
  sku: string
  name: string
  price: string
}

export type ProductFormErrors = {
  sku?: string
  name?: string
  price?: string
}

export const validatePrice = (value: string | number): string | undefined => {
  const text = String(value).trim()
  if (text === '') return 'Price is required'
  const price = Number(text)
  if (!Number.isInteger(price) || price < 0) return 'Price must be a whole number, 0 or more'
  return undefined
}

export const validateProductForm = (values: ProductFormValues): ProductFormErrors => {
  const errors: ProductFormErrors = {}
  if (!values.sku.trim()) errors.sku = 'SKU is required'
  if (!values.name.trim()) errors.name = 'Name is required'
  const priceError = validatePrice(values.price)
  if (priceError) errors.price = priceError
  return errors
}

export const hasProductErrors = (errors: ProductFormErrors): boolean =>
  Object.keys(errors).length > 0

export const mapProductFields = (fields: Record<string, string> = {}): ProductFormErrors => {
  const errors: ProductFormErrors = {}
  if (fields.sku) errors.sku = fields.sku
  if (fields.name) errors.name = fields.name
  if (fields.price) errors.price = fields.price
  return errors
}
