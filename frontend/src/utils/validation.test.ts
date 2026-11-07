import { describe, expect, it } from 'vitest'
import {
  hasOrderErrors,
  hasProductErrors,
  mapProductFields,
  mapServerFields,
  validateOrder,
  validatePrice,
  validateProductForm,
  validateQty,
} from './validation'

describe('validateQty', () => {
  it('accepts positive integers', () => {
    expect(validateQty('3')).toBeUndefined()
    expect(validateQty(12)).toBeUndefined()
  })

  it('rejects blank, zero, negative and non-integer values', () => {
    expect(validateQty('')).toBeDefined()
    expect(validateQty('0')).toBeDefined()
    expect(validateQty('-2')).toBeDefined()
    expect(validateQty('1.5')).toBeDefined()
    expect(validateQty('abc')).toBeDefined()
  })
})

describe('validateOrder', () => {
  it('passes a valid order', () => {
    const errors = validateOrder('Budi', [
      { sku: 'A-1', qty: '2' },
      { sku: 'B-2', qty: '1' },
    ])
    expect(hasOrderErrors(errors)).toBe(false)
  })

  it('requires a customer', () => {
    const errors = validateOrder('  ', [{ sku: 'A-1', qty: '1' }])
    expect(errors.customer).toBeDefined()
  })

  it('requires at least one line', () => {
    const errors = validateOrder('Budi', [])
    expect(errors.items).toBeDefined()
  })

  it('flags blank sku and bad quantities per row', () => {
    const errors = validateOrder('Budi', [
      { sku: '', qty: '1' },
      { sku: 'A-1', qty: '0' },
      { sku: 'B-2', qty: '1.5' },
    ])
    expect(errors.lines[0].sku).toBeDefined()
    expect(errors.lines[1].qty).toBeDefined()
    expect(errors.lines[2].qty).toBeDefined()
  })

  it('flags duplicate skus on the later row', () => {
    const errors = validateOrder('Budi', [
      { sku: 'A-1', qty: '1' },
      { sku: 'A-1', qty: '2' },
    ])
    expect(errors.lines[0]).toBeUndefined()
    expect(errors.lines[1].sku).toBeDefined()
  })
})

describe('mapServerFields', () => {
  it('maps customer, items and per-row keys', () => {
    const errors = mapServerFields({
      customer: 'required',
      items: 'at least one item',
      'items[0].qty': 'must be a positive integer',
      'items[1].sku': 'unknown sku',
    })
    expect(errors.customer).toBe('required')
    expect(errors.items).toBe('at least one item')
    expect(errors.lines[0].qty).toBe('must be a positive integer')
    expect(errors.lines[1].sku).toBe('unknown sku')
  })

  it('ignores unknown keys and handles no fields', () => {
    expect(mapServerFields({ other: 'x' }).lines).toEqual({})
    expect(hasOrderErrors(mapServerFields())).toBe(false)
  })
})

describe('validatePrice', () => {
  it('accepts zero and positive integers', () => {
    expect(validatePrice('0')).toBeUndefined()
    expect(validatePrice('15000')).toBeUndefined()
  })

  it('rejects blank, negative and fractional values', () => {
    expect(validatePrice('')).toBeDefined()
    expect(validatePrice('-1')).toBeDefined()
    expect(validatePrice('10.5')).toBeDefined()
    expect(validatePrice('abc')).toBeDefined()
  })
})

describe('validateProductForm', () => {
  it('passes valid values', () => {
    const errors = validateProductForm({ sku: 'A-1', name: 'Widget', price: '1000' })
    expect(hasProductErrors(errors)).toBe(false)
  })

  it('flags each invalid field', () => {
    const errors = validateProductForm({ sku: ' ', name: '', price: '-5' })
    expect(errors.sku).toBeDefined()
    expect(errors.name).toBeDefined()
    expect(errors.price).toBeDefined()
  })
})

describe('mapProductFields', () => {
  it('keeps known keys and ignores the rest', () => {
    const errors = mapProductFields({ sku: 'required', price: 'bad', other: 'x' })
    expect(errors).toEqual({ sku: 'required', price: 'bad' })
  })

  it('handles no fields', () => {
    expect(hasProductErrors(mapProductFields())).toBe(false)
  })
})
