import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import ErrorState from '../components/ErrorState'
import { useCreateOrder } from '../domains/order/queries'
import { useAllProducts } from '../domains/product/queries'
import { useIdempotencyKey } from '../hooks/useIdempotencyKey'
import { ApiError } from '../lib/apiClient'
import { errorCodes } from '../types/api'
import { formatRupiah } from '../utils/format'
import {
  hasOrderErrors,
  mapServerFields,
  validateOrder,
  type OrderErrors,
  type OrderLine,
} from '../utils/validation'

const emptyLine = (): OrderLine => ({ sku: '', qty: '1' })

export default function CreateOrderPage() {
  const navigate = useNavigate()
  const products = useAllProducts()
  const createOrder = useCreateOrder()
  const idempotency = useIdempotencyKey()

  const [customer, setCustomer] = useState('')
  const [lines, setLines] = useState<OrderLine[]>([emptyLine()])
  const [errors, setErrors] = useState<OrderErrors>({ lines: {} })
  const [formError, setFormError] = useState<string>()
  const [shortage, setShortage] = useState<{ sku: string; available: number }>()

  const options = (products.data?.data ?? []).filter((product) => product.is_active)

  const updateCustomer = (value: string) => {
    setCustomer(value)
    idempotency.reset()
  }

  const addLine = () => {
    setLines((current) => [...current, emptyLine()])
    idempotency.reset()
  }

  const updateLine = (index: number, patch: Partial<OrderLine>) => {
    idempotency.reset()
    setLines((current) =>
      current.map((line, position) => (position === index ? { ...line, ...patch } : line)),
    )
    setShortage(undefined)
  }

  const removeLine = (index: number) => {
    idempotency.reset()
    setLines((current) => current.filter((_, position) => position !== index))
    setErrors({ lines: {} })
    setShortage(undefined)
  }

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    setFormError(undefined)
    setShortage(undefined)

    const found = validateOrder(customer, lines)
    setErrors(found)
    if (hasOrderErrors(found)) return

    createOrder.mutate(
      {
        input: {
          customer: customer.trim(),
          items: lines.map((line) => ({ sku: line.sku, qty: Number(line.qty) })),
        },
        idempotencyKey: idempotency.get(),
      },
      {
        onSuccess: (order) => {
          idempotency.reset()
          navigate(`/orders/${order.id}`)
        },
        onError: (error) => {
          if (!(error instanceof ApiError)) {
            setFormError(error.message)
          } else if (error.code === errorCodes.insufficientStock && error.sku) {
            setShortage({ sku: error.sku, available: error.available ?? 0 })
            setFormError(error.message)
          } else if (error.code === errorCodes.idempotencyKeyReused) {
            idempotency.reset()
            setFormError('This order changed, please submit again')
          } else if (error.code === errorCodes.validation && error.fields) {
            setErrors(mapServerFields(error.fields))
            setFormError(error.message)
          } else {
            setFormError(error.message)
          }
        },
      },
    )
  }

  return (
    <div className="page page-narrow">
      <header className="page-header">
        <h1>New order</h1>
        <Link to="/" className="btn">
          Back to products
        </Link>
      </header>

      {products.isError ? (
        <ErrorState message={products.error.message} onRetry={() => products.refetch()} />
      ) : (
        <form className="card" onSubmit={handleSubmit} noValidate>
          <label className="field">
            <span className="field-label">Customer</span>
            <input
              value={customer}
              onChange={(event) => updateCustomer(event.target.value)}
              aria-invalid={Boolean(errors.customer)}
            />
            {errors.customer && <span className="field-error">{errors.customer}</span>}
          </label>

          <div className="lines">
            <span className="field-label">Items</span>
            {lines.map((line, index) => {
              const lineErrors = errors.lines[index]
              const short = shortage && shortage.sku === line.sku
              return (
                <div key={index} className={`line${short ? ' line-short' : ''}`}>
                  <div className="line-fields">
                    <label className="field">
                      <span className="sr-only">Product</span>
                      <select
                        value={line.sku}
                        onChange={(event) => updateLine(index, { sku: event.target.value })}
                        aria-invalid={Boolean(lineErrors?.sku)}
                        disabled={products.isPending}
                      >
                        <option value="">
                          {products.isPending ? 'Loading products...' : 'Select a product'}
                        </option>
                        {options.map((product) => (
                          <option key={product.id} value={product.sku}>
                            {product.sku} - {product.name} ({formatRupiah(product.price)})
                          </option>
                        ))}
                      </select>
                      {lineErrors?.sku && <span className="field-error">{lineErrors.sku}</span>}
                    </label>
                    <label className="field field-qty">
                      <span className="sr-only">Quantity</span>
                      <input
                        inputMode="numeric"
                        value={line.qty}
                        onChange={(event) => updateLine(index, { qty: event.target.value })}
                        aria-invalid={Boolean(lineErrors?.qty)}
                      />
                      {lineErrors?.qty && <span className="field-error">{lineErrors.qty}</span>}
                    </label>
                    <button
                      type="button"
                      className="btn btn-small"
                      onClick={() => removeLine(index)}
                      aria-label={`Remove item ${index + 1}`}
                    >
                      Remove
                    </button>
                  </div>
                  {short && (
                    <p className="field-error">Only {shortage.available} available for this item</p>
                  )}
                </div>
              )
            })}
            {errors.items && <span className="field-error">{errors.items}</span>}
            <button
              type="button"
              className="btn btn-small"
              onClick={addLine}
            >
              Add item
            </button>
          </div>

          {formError && (
            <p className="form-error" role="alert">
              {formError}
            </p>
          )}

          <div className="modal-actions">
            <button type="submit" className="btn btn-primary" disabled={createOrder.isPending}>
              {createOrder.isPending ? 'Placing order...' : 'Place order'}
            </button>
          </div>
        </form>
      )}
    </div>
  )
}
