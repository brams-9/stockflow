import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Badge from '../components/Badge'
import ConfirmDialog from '../components/ConfirmDialog'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import { useCancelOrder, useOrder } from '../domains/order/queries'
import { ApiError } from '../lib/apiClient'
import { errorCodes } from '../types/api'
import { formatDateTime, formatRupiah } from '../utils/format'

export default function OrderDetailPage() {
  const params = useParams()
  const id = Number(params.id)
  const validId = Number.isInteger(id) && id > 0

  const order = useOrder(id)
  const cancel = useCancelOrder(id)
  const [confirming, setConfirming] = useState(false)
  const [cancelError, setCancelError] = useState<string>()

  const handleCancel = () => {
    setCancelError(undefined)
    cancel.mutate(undefined, {
      onSuccess: () => setConfirming(false),
      onError: (error) => {
        order.refetch()
        if (error instanceof ApiError && error.code === errorCodes.orderAlreadyCancelled) {
          setCancelError('This order has already been cancelled.')
        } else {
          setCancelError(error.message)
        }
      },
    })
  }

  const closeDialog = () => {
    setConfirming(false)
    setCancelError(undefined)
  }

  const renderContent = () => {
    if (!validId) {
      return <EmptyState title="Order not found" description="That order number is not valid." />
    }
    if (order.isPending) return <p className="muted">Loading order...</p>
    if (order.isError) {
      if (order.error instanceof ApiError && order.error.code === errorCodes.orderNotFound) {
        return <EmptyState title="Order not found" description={`There is no order #${id}.`} />
      }
      return <ErrorState message={order.error.message} onRetry={() => order.refetch()} />
    }

    const data = order.data
    const cancelled = data.status === 'CANCELLED'

    return (
      <div className="card">
        <div className="order-head">
          <div>
            <h2>Order #{data.id}</h2>
            <p className="muted">{data.customer}</p>
          </div>
          <Badge tone={cancelled ? 'red' : 'green'} label={cancelled ? 'Cancelled' : 'Created'} />
        </div>

        <dl className="meta">
          <div>
            <dt>Created</dt>
            <dd>{formatDateTime(data.created_at)}</dd>
          </div>
          {cancelled && (
            <div>
              <dt>Cancelled</dt>
              <dd>{formatDateTime(data.cancelled_at)}</dd>
            </div>
          )}
        </dl>

        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th className="num">Qty</th>
                <th className="num">Unit price</th>
                <th className="num">Line total</th>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <tr key={item.id}>
                  <td className="mono">{item.sku}</td>
                  <td className="num">{item.qty}</td>
                  <td className="num">{formatRupiah(item.unit_price)}</td>
                  <td className="num">{formatRupiah(item.qty * item.unit_price)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr>
                <td colSpan={3}>Total</td>
                <td className="num total">{formatRupiah(data.total)}</td>
              </tr>
            </tfoot>
          </table>
        </div>

        {!cancelled && (
          <div className="modal-actions">
            <button type="button" className="btn btn-danger" onClick={() => setConfirming(true)}>
              Cancel order
            </button>
          </div>
        )}

        {confirming && !cancelled && (
          <ConfirmDialog
            title="Cancel this order?"
            message="The reserved stock will be returned to the products. This cannot be undone."
            confirmLabel="Cancel order"
            cancelLabel="Keep order"
            busy={cancel.isPending}
            error={cancelError}
            onConfirm={handleCancel}
            onCancel={closeDialog}
          />
        )}
      </div>
    )
  }

  return (
    <div className="page page-narrow">
      <header className="page-header">
        <h1>Order</h1>
        <Link to="/" className="btn">
          Back to products
        </Link>
      </header>
      {renderContent()}
    </div>
  )
}
