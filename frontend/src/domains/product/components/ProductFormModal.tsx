import { useState, type FormEvent } from 'react'
import Modal from '../../../components/Modal'
import { ApiError } from '../../../lib/apiClient'
import { errorCodes, type Product } from '../../../types/api'
import {
  hasProductErrors,
  mapProductFields,
  validateProductForm,
  type ProductFormErrors,
} from '../../../utils/validation'
import { useCreateProduct, useUpdateProduct } from '../queries'

type ProductFormModalProps = {
  product?: Product
  onClose: () => void
}

export default function ProductFormModal({ product, onClose }: ProductFormModalProps) {
  const editing = Boolean(product)
  const [sku, setSku] = useState(product?.sku ?? '')
  const [name, setName] = useState(product?.name ?? '')
  const [price, setPrice] = useState(product ? String(product.price) : '')
  const [active, setActive] = useState(product?.is_active ?? true)
  const [errors, setErrors] = useState<ProductFormErrors>({})
  const [formError, setFormError] = useState<string>()

  const create = useCreateProduct()
  const update = useUpdateProduct(product?.id ?? 0)
  const pending = create.isPending || update.isPending

  const handleError = (error: Error) => {
    if (!(error instanceof ApiError)) {
      setFormError(error.message)
    } else if (error.code === errorCodes.skuAlreadyExists) {
      setErrors({ sku: 'This SKU is already in use' })
    } else if (error.code === errorCodes.validation && error.fields) {
      setErrors(mapProductFields(error.fields))
      setFormError(error.message)
    } else if (error.code === errorCodes.productNotFound) {
      setFormError('This product no longer exists')
    } else {
      setFormError(error.message)
    }
  }

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    setFormError(undefined)

    const found = validateProductForm({ sku, name, price })
    setErrors(found)
    if (hasProductErrors(found)) return

    const input = { sku: sku.trim(), name: name.trim(), price: Number(price) }
    if (product) {
      update.mutate({ ...input, is_active: active }, { onSuccess: onClose, onError: handleError })
    } else {
      create.mutate(input, { onSuccess: onClose, onError: handleError })
    }
  }

  return (
    <Modal title={editing ? `Edit ${product?.name}` : 'New product'} onClose={onClose}>
      <form onSubmit={handleSubmit} noValidate>
        <label className="field">
          <span className="field-label">SKU</span>
          <input
            value={sku}
            onChange={(event) => setSku(event.target.value)}
            aria-invalid={Boolean(errors.sku)}
          />
          {errors.sku && <span className="field-error">{errors.sku}</span>}
        </label>
        <label className="field">
          <span className="field-label">Name</span>
          <input
            value={name}
            onChange={(event) => setName(event.target.value)}
            aria-invalid={Boolean(errors.name)}
          />
          {errors.name && <span className="field-error">{errors.name}</span>}
        </label>
        <label className="field">
          <span className="field-label">Price (IDR)</span>
          <input
            inputMode="numeric"
            value={price}
            onChange={(event) => setPrice(event.target.value)}
            aria-invalid={Boolean(errors.price)}
          />
          {errors.price && <span className="field-error">{errors.price}</span>}
        </label>
        {editing && (
          <label className="field checkbox">
            <input
              type="checkbox"
              checked={active}
              onChange={(event) => setActive(event.target.checked)}
            />
            <span>Active</span>
          </label>
        )}
        {formError && (
          <p className="form-error" role="alert">
            {formError}
          </p>
        )}
        <div className="modal-actions">
          <button type="button" className="btn" onClick={onClose} disabled={pending}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={pending}>
            {pending ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
