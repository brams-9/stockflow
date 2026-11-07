import { useMemo, useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Badge from '../components/Badge'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'
import SkeletonRows from '../components/SkeletonRows'
import SummaryCards from '../components/SummaryCards'
import AdjustStockModal from '../domains/product/components/AdjustStockModal'
import DeleteProductDialog from '../domains/product/components/DeleteProductDialog'
import MovementsModal from '../domains/product/components/MovementsModal'
import ProductFormModal from '../domains/product/components/ProductFormModal'
import { useAllProducts, useProducts } from '../domains/product/queries'
import { summarize } from '../domains/summary/summary'
import { useDebounce } from '../hooks/useDebounce'
import type { Product } from '../types/api'
import { formatRelativeTime, formatRupiah, isLowStock } from '../utils/format'

type StatusFilter = 'all' | 'active' | 'inactive' | 'low'

const filters: { value: StatusFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Inactive' },
  { value: 'low', label: 'Low stock' },
]

const PAGE_SIZE = 10

type Dialog =
  | { kind: 'create' }
  | { kind: 'adjust' | 'edit' | 'history' | 'delete'; product: Product }

const matchesFilter = (product: Product, filter: StatusFilter) => {
  if (filter === 'active') return product.is_active
  if (filter === 'inactive') return !product.is_active
  if (filter === 'low') return isLowStock(product.stock)
  return true
}

export default function ProductsPage() {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<StatusFilter>('all')
  const [page, setPage] = useState(1)
  const [orderId, setOrderId] = useState('')
  const [dialog, setDialog] = useState<Dialog | null>(null)

  const debouncedSearch = useDebounce(search.trim(), 300)
  const products = useProducts({ search: debouncedSearch, page, limit: PAGE_SIZE })
  const everything = useAllProducts()

  const summary = useMemo(
    () => (everything.data ? summarize(everything.data.data, everything.data.total) : undefined),
    [everything.data],
  )

  const visible = useMemo(
    () => (products.data?.data ?? []).filter((product) => matchesFilter(product, filter)),
    [products.data, filter],
  )

  const total = products.data?.total ?? 0
  const lastPage = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const handleSearch = (value: string) => {
    setSearch(value)
    setPage(1)
  }

  const closeDialog = () => setDialog(null)

  const retry = () => {
    products.refetch()
    everything.refetch()
  }

  const openOrder = (event: FormEvent) => {
    event.preventDefault()
    const id = Number(orderId)
    if (Number.isInteger(id) && id > 0) navigate(`/orders/${id}`)
  }

  const renderBody = () => {
    if (products.isPending) return <SkeletonRows columns={7} />
    if (visible.length === 0) {
      return (
        <tr>
          <td colSpan={7}>
            <EmptyState
              title="No products found"
              description={
                debouncedSearch || filter !== 'all'
                  ? 'Try a different search or filter.'
                  : 'Products you add will show up here.'
              }
            />
          </td>
        </tr>
      )
    }
    return visible.map((product) => (
      <tr key={product.id}>
        <td className="mono">{product.sku}</td>
        <td>{product.name}</td>
        <td className="num">{formatRupiah(product.price)}</td>
        <td className="num">
          {product.stock}
          {isLowStock(product.stock) && <Badge tone="amber" label="Low" />}
        </td>
        <td>
          <Badge
            tone={product.is_active ? 'green' : 'gray'}
            label={product.is_active ? 'Active' : 'Inactive'}
          />
        </td>
        <td>{formatRelativeTime(product.last_movement_at)}</td>
        <td className="actions">
          <div className="row-actions">
            <button
              type="button"
              className="btn btn-small"
              onClick={() => setDialog({ kind: 'adjust', product })}
            >
              Adjust stock
            </button>
            <button
              type="button"
              className="btn btn-small"
              onClick={() => setDialog({ kind: 'history', product })}
            >
              History
            </button>
            <button
              type="button"
              className="btn btn-small"
              onClick={() => setDialog({ kind: 'edit', product })}
            >
              Edit
            </button>
            <button
              type="button"
              className="btn btn-small"
              onClick={() => setDialog({ kind: 'delete', product })}
            >
              Delete
            </button>
          </div>
        </td>
      </tr>
    ))
  }

  return (
    <div className="page">
      <header className="page-header">
        <h1>Products</h1>
        <div className="header-tools">
          <form className="inline-form" onSubmit={openOrder}>
            <input
              inputMode="numeric"
              placeholder="Open order #"
              aria-label="Open order number"
              value={orderId}
              onChange={(event) => setOrderId(event.target.value)}
            />
            <button type="submit" className="btn">
              Open
            </button>
          </form>
          <button type="button" className="btn" onClick={() => setDialog({ kind: 'create' })}>
            New product
          </button>
          <Link to="/orders/new" className="btn btn-primary">
            New order
          </Link>
        </div>
      </header>

      <SummaryCards summary={summary} />

      <div className="toolbar">
        <input
          type="search"
          placeholder="Search by SKU or name"
          aria-label="Search products"
          value={search}
          onChange={(event) => handleSearch(event.target.value)}
        />
        <div className="segmented">
          {filters.map((item) => (
            <button
              key={item.value}
              type="button"
              className={filter === item.value ? 'active' : ''}
              onClick={() => setFilter(item.value)}
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>

      {products.isError ? (
        <ErrorState message={products.error.message} onRetry={retry} />
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Name</th>
                <th className="num">Price</th>
                <th className="num">Stock</th>
                <th>Status</th>
                <th>Last movement</th>
                <th className="actions">Actions</th>
              </tr>
            </thead>
            <tbody>{renderBody()}</tbody>
          </table>
        </div>
      )}

      {!products.isError && (
        <div className="pager">
          <span className="muted">
            Page {page} of {lastPage} &middot; {total} products
          </span>
          <div>
            <button
              type="button"
              className="btn btn-small"
              disabled={page <= 1}
              onClick={() => setPage((current) => current - 1)}
            >
              Previous
            </button>
            <button
              type="button"
              className="btn btn-small"
              disabled={page >= lastPage}
              onClick={() => setPage((current) => current + 1)}
            >
              Next
            </button>
          </div>
        </div>
      )}

      {dialog?.kind === 'create' && <ProductFormModal onClose={closeDialog} />}
      {dialog?.kind === 'edit' && <ProductFormModal product={dialog.product} onClose={closeDialog} />}
      {dialog?.kind === 'adjust' && <AdjustStockModal product={dialog.product} onClose={closeDialog} />}
      {dialog?.kind === 'history' && <MovementsModal product={dialog.product} onClose={closeDialog} />}
      {dialog?.kind === 'delete' && (
        <DeleteProductDialog product={dialog.product} onClose={closeDialog} />
      )}
    </div>
  )
}
