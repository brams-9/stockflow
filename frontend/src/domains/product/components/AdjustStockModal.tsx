import { useState, type FormEvent } from 'react'
import Modal from '../../../components/Modal'
import { ApiError } from '../../../lib/apiClient'
import { errorCodes, type MovementType, type Product } from '../../../types/api'
import { validateQty } from '../../../utils/validation'
import { useAdjustStock } from '../queries'

type AdjustStockModalProps = {
  product: Product
  onClose: () => void
}

export default function AdjustStockModal({ product, onClose }: AdjustStockModalProps) {
  const [type, setType] = useState<MovementType>('in')
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const [qtyError, setQtyError] = useState<string>()
  const [formError, setFormError] = useState<string>()
  const adjust = useAdjustStock(product.id)

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    const problem = validateQty(qty)
    setQtyError(problem)
    setFormError(undefined)
    if (problem) return

    adjust.mutate(
      { type, qty: Number(qty), note: note.trim() || undefined },
      {
        onSuccess: onClose,
        onError: (error) => {
          if (error instanceof ApiError && error.code === errorCodes.insufficientStock) {
            setFormError(`Only ${error.available ?? 0} available`)
          } else if (error instanceof ApiError && error.fields?.qty) {
            setQtyError(error.fields.qty)
          } else {
            setFormError(error.message)
          }
        },
      },
    )
  }

  return (
    <Modal title={`Adjust stock: ${product.name}`} onClose={onClose}>
      <p className="muted">
        {product.sku} &middot; current stock {product.stock}
      </p>
      <form onSubmit={handleSubmit} noValidate>
        <div className="field">
          <span className="field-label">Direction</span>
          <div className="segmented">
            <button
              type="button"
              className={type === 'in' ? 'active' : ''}
              onClick={() => setType('in')}
            >
              Stock in
            </button>
            <button
              type="button"
              className={type === 'out' ? 'active' : ''}
              onClick={() => setType('out')}
            >
              Stock out
            </button>
          </div>
        </div>
        <label className="field">
          <span className="field-label">Quantity</span>
          <input
            inputMode="numeric"
            value={qty}
            onChange={(event) => setQty(event.target.value)}
            aria-invalid={Boolean(qtyError)}
          />
          {qtyError && <span className="field-error">{qtyError}</span>}
        </label>
        <label className="field">
          <span className="field-label">Note</span>
          <input value={note} onChange={(event) => setNote(event.target.value)} />
        </label>
        {formError && (
          <p className="form-error" role="alert">
            {formError}
          </p>
        )}
        <div className="modal-actions">
          <button type="button" className="btn" onClick={onClose} disabled={adjust.isPending}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={adjust.isPending}>
            {adjust.isPending ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
