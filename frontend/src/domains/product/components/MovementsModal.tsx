import Badge from '../../../components/Badge'
import EmptyState from '../../../components/EmptyState'
import ErrorState from '../../../components/ErrorState'
import Modal from '../../../components/Modal'
import { ApiError } from '../../../lib/apiClient'
import { errorCodes, type Product } from '../../../types/api'
import { formatDateTime } from '../../../utils/format'
import { useMovements } from '../queries'

type MovementsModalProps = {
  product: Product
  onClose: () => void
}

export default function MovementsModal({ product, onClose }: MovementsModalProps) {
  const movements = useMovements(product.id)

  const renderContent = () => {
    if (movements.isPending) return <p className="muted">Loading history...</p>
    if (movements.isError) {
      if (movements.error instanceof ApiError && movements.error.code === errorCodes.productNotFound) {
        return <EmptyState title="Product not found" />
      }
      return <ErrorState message={movements.error.message} onRetry={() => movements.refetch()} />
    }
    if (movements.data.data.length === 0) {
      return <EmptyState title="No movements yet" description="Stock changes will be listed here." />
    }
    return (
      <div className="table-wrap modal-scroll">
        <table>
          <thead>
            <tr>
              <th>When</th>
              <th>Type</th>
              <th className="num">Qty</th>
              <th>Note</th>
            </tr>
          </thead>
          <tbody>
            {movements.data.data.map((movement) => (
              <tr key={movement.id}>
                <td>{formatDateTime(movement.created_at)}</td>
                <td>
                  <Badge
                    tone={movement.type === 'in' ? 'green' : 'amber'}
                    label={movement.type === 'in' ? 'In' : 'Out'}
                  />
                </td>
                <td className="num">{movement.qty}</td>
                <td>{movement.note ?? '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }

  return (
    <Modal title={`History: ${product.name}`} onClose={onClose} wide>
      <p className="muted">{product.sku}</p>
      {renderContent()}
      <div className="modal-actions">
        <button type="button" className="btn" onClick={onClose}>
          Close
        </button>
      </div>
    </Modal>
  )
}
