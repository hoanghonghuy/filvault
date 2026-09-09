/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it } from 'vitest'
import {
  CHAT_BUBBLE_STYLES,
  getBubbleStyle,
  setBubbleStyle,
  activeBubbleStyleId,
} from './chatBubbles'

describe('chatBubbles utility', () => {
  beforeEach(() => {
    localStorage.clear()
    setBubbleStyle('default')
  })

  it('contains exactly 21 unique bubble styles', () => {
    expect(CHAT_BUBBLE_STYLES.length).toBe(21)
    const ids = new Set(CHAT_BUBBLE_STYLES.map((s) => s.id))
    expect(ids.size).toBe(21)
  })

  it('contains all referenced styles from design photos', () => {
    const expectedIds = [
      'default',
      'otter',
      'diva',
      'crocodile',
      'olivia',
      'frog_duck',
      'buff_cat',
      'capybara',
      'frog',
      'cat_dog',
      'facepalm',
      'pepe_heart',
      'shark_pig',
      'bear_tongue',
      'doge',
      'frog_chick',
      'hungry_pup',
      'dino',
      'dachshund',
      'hungry_frog',
      'screaming',
    ]
    for (const id of expectedIds) {
      const found = CHAT_BUBBLE_STYLES.find((s) => s.id === id)
      expect(found).toBeDefined()
      expect(found?.name).toBeTruthy()
      expect(found?.bg).toBeTruthy()
    }
  })

  it('gets default style when id is invalid or not found', () => {
    const style = getBubbleStyle('non_existent_style')
    expect(style.id).toBe('default')
  })

  it('sets and updates active bubble style and persists to localStorage', () => {
    setBubbleStyle('frog_chick')
    expect(activeBubbleStyleId.value).toBe('frog_chick')
    expect(localStorage.getItem('filvault.chatBubbleStyle')).toBe('frog_chick')

    const style = getBubbleStyle()
    expect(style.id).toBe('frog_chick')
    expect(style.name).toBe('Ếch gà con')
  })
})
