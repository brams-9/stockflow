export type KeyHolder = {
  get: () => string
  reset: () => void
}

export const createKeyHolder = (generate: () => string = () => crypto.randomUUID()): KeyHolder => {
  let key: string | null = null

  return {
    get: () => {
      if (key === null) key = generate()
      return key
    },
    reset: () => {
      key = null
    },
  }
}
