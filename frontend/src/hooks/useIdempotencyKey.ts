import { useRef } from 'react'
import { createKeyHolder, type KeyHolder } from '../utils/idempotency'

export function useIdempotencyKey(): KeyHolder {
  const holder = useRef<KeyHolder>(null)
  if (holder.current === null) holder.current = createKeyHolder()
  return holder.current
}
