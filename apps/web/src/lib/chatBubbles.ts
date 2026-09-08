/**
 * Chat Bubble Styles configuration and storage for Filvault.
 * Inspired by interactive customizable messenger bubble themes.
 */
import { ref } from 'vue'

export interface BubbleDecoration {
  position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'left' | 'right' | 'top' | 'bottom'
  imgUrl?: string
  svg?: string
  style?: Record<string, string>
}

export interface ChatBubbleStyle {
  id: string
  name: string
  bg: string
  color: string
  border?: string
  borderRadius?: string
  decorations: BubbleDecoration[]
  thumbUrl: string
}

export const CHAT_BUBBLE_STYLES: ChatBubbleStyle[] = [
  {
    id: 'default',
    name: 'Mặc định',
    bg: '#286CFF',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/default.png',
    decorations: [],
  },
  {
    id: 'otter',
    name: 'Rái cá vui nhộn',
    bg: '#8AE1F4',
    color: '#153344',
    thumbUrl: '/bubbles/thumbs/otter.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/otter_mascot.png',
      },
    ],
  },
  {
    id: 'diva',
    name: 'Diva lộng lẫy',
    bg: '#FFFFFF',
    color: '#1A1A1A',
    border: '1.5px solid #E5E7EB',
    thumbUrl: '/bubbles/thumbs/diva.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/diva_mascot.png',
      },
    ],
  },
  {
    id: 'crocodile',
    name: 'Cá sấu thư thái',
    bg: '#6879D5',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/crocodile.png',
    decorations: [
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/crocodile_mascot.png',
      },
    ],
  },
  {
    id: 'olivia',
    name: 'Olivia Rodrigo',
    bg: '#F5B8C7',
    color: '#461325',
    thumbUrl: '/bubbles/thumbs/olivia.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/olivia_mascot.png',
      },
    ],
  },
  {
    id: 'frog_duck',
    name: 'Ếch vịt',
    bg: '#F8E879',
    color: '#423300',
    thumbUrl: '/bubbles/thumbs/frog_duck.png',
    decorations: [
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/frog_duck_mascot.png',
      },
    ],
  },
  {
    id: 'buff_cat',
    name: 'Mèo khoe cơ',
    bg: '#FCFCFC',
    color: '#1A1A1A',
    border: '1.5px solid #E5E7EB',
    thumbUrl: '/bubbles/thumbs/buff_cat.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/buff_cat_mascot.png',
      },
    ],
  },
  {
    id: 'capybara',
    name: 'Capybara',
    bg: '#D9AC6B',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/capybara.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/capybara_mascot.png',
      },
    ],
  },
  {
    id: 'frog',
    name: 'Ếch',
    bg: '#009D58',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/frog.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/frog_eye_mascot.png',
      },
    ],
  },
  {
    id: 'cat_dog',
    name: 'Mèo chó',
    bg: '#FDF4D3',
    color: '#422C0D',
    thumbUrl: '/bubbles/thumbs/cat_dog.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/cat_mascot.png',
      },
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/dog_mascot.png',
      },
    ],
  },
  {
    id: 'facepalm',
    name: 'Ôm mặt',
    bg: '#FCFCFC',
    color: '#1A1A1A',
    border: '1.5px solid #E5E7EB',
    thumbUrl: '/bubbles/thumbs/facepalm.png',
    decorations: [
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/facepalm_mascot.png',
      },
    ],
  },
  {
    id: 'pepe_heart',
    name: 'Pepe thả tim',
    bg: '#DCBAC8',
    color: '#381424',
    thumbUrl: '/bubbles/thumbs/pepe_heart.png',
    decorations: [
      {
        position: 'bottom-left',
        imgUrl: '/bubbles/decorations/pepe_mascot.png',
      },
    ],
  },
  {
    id: 'shark_pig',
    name: 'Cá mập heo',
    bg: '#E1E9FE',
    color: '#1A284E',
    thumbUrl: '/bubbles/thumbs/shark_pig.png',
    decorations: [
      {
        position: 'bottom-left',
        imgUrl: '/bubbles/decorations/pig_mascot.png',
      },
      {
        position: 'bottom-right',
        imgUrl: '/bubbles/decorations/shark_mascot.png',
      },
    ],
  },
  {
    id: 'bear_tongue',
    name: 'Gấu thè lưỡi',
    bg: '#FFDFE0',
    color: '#481928',
    thumbUrl: '/bubbles/thumbs/bear_tongue.png',
    decorations: [
      {
        position: 'bottom-left',
        imgUrl: '/bubbles/decorations/bear_mascot.png',
      },
    ],
  },
  {
    id: 'doge',
    name: 'Doge',
    bg: '#F6DEA0',
    color: '#443103',
    thumbUrl: '/bubbles/thumbs/doge.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/doge_thumb.png',
      },
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/doge_mascot.png',
      },
    ],
  },
  {
    id: 'frog_chick',
    name: 'Ếch gà con',
    bg: '#B2E6FC',
    color: '#123746',
    thumbUrl: '/bubbles/thumbs/frog_chick.png',
    decorations: [
      {
        position: 'top-left',
        imgUrl: '/bubbles/decorations/frog_head.png',
      },
      {
        position: 'top-right',
        imgUrl: '/bubbles/decorations/chick_head.png',
      },
      {
        position: 'bottom-right',
        imgUrl: '/bubbles/decorations/water_bubbles.png',
      },
    ],
  },
  {
    id: 'hungry_pup',
    name: 'Cún thèm ăn',
    bg: '#CE503A',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/hungry_pup.png',
    decorations: [
      {
        position: 'left',
        imgUrl: '/bubbles/decorations/pup_mascot.png',
      },
    ],
  },
  {
    id: 'dino',
    name: 'Khủng long',
    bg: '#AEE3E9',
    color: '#0F3D43',
    thumbUrl: '/bubbles/thumbs/dino.png',
    decorations: [
      {
        position: 'left',
        imgUrl: '/bubbles/decorations/dino_mascot.png',
      },
    ],
  },
  {
    id: 'dachshund',
    name: 'Chó lạp xưởng',
    bg: '#90644B',
    color: '#FFFFFF',
    thumbUrl: '/bubbles/thumbs/dachshund.png',
    decorations: [
      {
        position: 'left',
        imgUrl: '/bubbles/decorations/dachshund_mascot.png',
      },
    ],
  },
  {
    id: 'hungry_frog',
    name: 'Ếch đói bụng',
    bg: '#DAE6D2',
    color: '#1D3E1A',
    thumbUrl: '/bubbles/thumbs/hungry_frog.png',
    decorations: [
      {
        position: 'bottom-left',
        imgUrl: '/bubbles/decorations/hungry_frog_mascot.png',
      },
    ],
  },
  {
    id: 'screaming',
    name: 'La hét',
    bg: '#FFA289',
    color: '#3D1612',
    thumbUrl: '/bubbles/thumbs/screaming.png',
    decorations: [
      {
        position: 'left',
        imgUrl: '/bubbles/decorations/screaming_mascot.png',
      },
    ],
  },
]

const STORAGE_KEY = 'filvault.chatBubbleStyle'

export const activeBubbleStyleId = ref<string>(loadInitialBubbleStyle())

function loadInitialBubbleStyle(): string {
  if (typeof window === 'undefined') return 'default'
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && CHAT_BUBBLE_STYLES.some((s) => s.id === saved)) {
      return saved
    }
  } catch {
    // Ignore storage access errors
  }
  return 'default'
}

export function getBubbleStyle(id?: string | null): ChatBubbleStyle {
  const targetId = id || activeBubbleStyleId.value
  const found = CHAT_BUBBLE_STYLES.find((s) => s.id === targetId)
  return found || CHAT_BUBBLE_STYLES[0]
}

export function setBubbleStyle(id: string): void {
  if (!CHAT_BUBBLE_STYLES.some((s) => s.id === id)) return
  activeBubbleStyleId.value = id
  try {
    localStorage.setItem(STORAGE_KEY, id)
  } catch {
    // Ignore storage quota or access errors
  }
}
