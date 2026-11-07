import { useState } from 'react'
import ConfirmDialog from '../../../components/ConfirmDialog'
import { ApiError } from '../../../lib/apiClient'
import { errorCodes, type Product } from '../../../types/api'
import { useDeleteProduct } from '../queries'

type DeleteProductDialogProps = {
  product: Product
  onClose: () => void
}

export default function DeleteProductDialog({ product, onClose }: DeleteProductDialogProps) {
  const [error, setError] = useState<string>()
  const remove = useDeleteProduct(product.id)

  const handleConfirm = () => {
    setError(undefined)
    remove.mutate(undefined, {
      onSuccess: onClose,
      onError: (failure) => {
        if (failure instanceof ApiError && failure.code === errorCodes.productHasHistory) {
          setError('This product has stock movement history and cannot be deleted. Mark it inactive instead.')
        } else if (failure instanceof ApiError && failure.code === errorCodes.productNotFound) {
          setError('This product no longer exists.')
        } else {
          setError(failure.message)
        }
      },
    })
  }

  return (
    <ConfirmDialog
      title="Delete product?"
      message={`${product.name} (${product.sku}) will be removed permanently.`}
      confirmLabel="Delete"
      busy={remove.isPending}
      error={error}
      onConfirm={handleConfirm}
      onCancel={onClose}
    />
  )
}
