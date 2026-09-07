export interface Sticker {
  id: string
  symbol: string
  name: string
  category: 'expressions' | 'gestures' | 'pets' | 'fun'
}

export interface StickerCategory {
  id: 'expressions' | 'gestures' | 'pets' | 'fun'
  label: string
  icon: string
}

export const STICKER_CATEGORIES: StickerCategory[] = [
  { id: 'expressions', label: 'Cảm xúc', icon: '🥰' },
  { id: 'gestures', label: 'Cử chỉ', icon: '👍' },
  { id: 'pets', label: 'Thú cưng', icon: '🐱' },
  { id: 'fun', label: 'Vui vẻ', icon: '🎉' },
]

export const STICKERS: Sticker[] = [
  // Expressions
  { id: 'exp_love', symbol: '🥰', name: 'Yêu thương', category: 'expressions' },
  { id: 'exp_heart_eyes', symbol: '😍', name: 'Thả tim', category: 'expressions' },
  { id: 'exp_lol', symbol: '😂', name: 'Cười lăn lộn', category: 'expressions' },
  { id: 'exp_cry', symbol: '😭', name: 'Khóc ròng', category: 'expressions' },
  { id: 'exp_cool', symbol: '😎', name: 'Cực ngầu', category: 'expressions' },
  { id: 'exp_party', symbol: '🥳', name: 'Tiệc tùng', category: 'expressions' },
  { id: 'exp_kiss', symbol: '😘', name: 'Hôn gió', category: 'expressions' },
  { id: 'exp_angry', symbol: '😡', name: 'Tức giận', category: 'expressions' },
  { id: 'exp_shock', symbol: '😱', name: 'Kinh ngạc', category: 'expressions' },
  { id: 'exp_think', symbol: '🤔', name: 'Suy nghĩ', category: 'expressions' },
  { id: 'exp_plead', symbol: '🥺', name: 'Cầu xin', category: 'expressions' },
  { id: 'exp_hug', symbol: '🤗', name: 'Ôm ấm áp', category: 'expressions' },

  // Gestures
  { id: 'ges_like', symbol: '👍', name: 'Like chuẩn', category: 'gestures' },
  { id: 'ges_wave', symbol: '👋', name: 'Vẫy tay', category: 'gestures' },
  { id: 'ges_clap', symbol: '👏', name: 'Vỗ tay', category: 'gestures' },
  { id: 'ges_celebrate', symbol: '🙌', name: 'Ăn mừng', category: 'gestures' },
  { id: 'ges_victory', symbol: '✌️', name: 'Chiến thắng', category: 'gestures' },
  { id: 'ges_handshake', symbol: '🤝', name: 'Bắt tay', category: 'gestures' },
  { id: 'ges_heart_sparkle', symbol: '💖', name: 'Trái tim', category: 'gestures' },
  { id: 'ges_fire', symbol: '🔥', name: 'Cháy quá', category: 'gestures' },
  { id: 'ges_hundred', symbol: '💯', name: '100 điểm', category: 'gestures' },
  { id: 'ges_muscle', symbol: '💪', name: 'Cố lên', category: 'gestures' },
  { id: 'ges_pray', symbol: '🙏', name: 'Cảm ơn', category: 'gestures' },
  { id: 'ges_star_struck', symbol: '🤩', name: 'Ngưỡng mộ', category: 'gestures' },

  // Pets
  { id: 'pet_cat', symbol: '🐱', name: 'Mèo ngoan', category: 'pets' },
  { id: 'pet_cat_grin', symbol: '😸', name: 'Mèo cười', category: 'pets' },
  { id: 'pet_cat_heart', symbol: '😻', name: 'Mèo mê đắm', category: 'pets' },
  { id: 'pet_cat_kiss', symbol: '😽', name: 'Mèo thơm', category: 'pets' },
  { id: 'pet_cat_cry', symbol: '😿', name: 'Mèo khóc', category: 'pets' },
  { id: 'pet_dog', symbol: '🐶', name: 'Cún cưng', category: 'pets' },
  { id: 'pet_shiba', symbol: '🐕', name: 'Shiba vui', category: 'pets' },
  { id: 'pet_panda', symbol: '🐼', name: 'Gấu trúc', category: 'pets' },
  { id: 'pet_bunny', symbol: '🐰', name: 'Thỏ ôm', category: 'pets' },
  { id: 'pet_fox', symbol: '🦊', name: 'Cáo nhỏ', category: 'pets' },
  { id: 'pet_bear', symbol: '🐻', name: 'Gấu con', category: 'pets' },
  { id: 'pet_unicorn', symbol: '🦄', name: 'Kỳ lân', category: 'pets' },

  // Fun
  { id: 'fun_tada', symbol: '🎉', name: 'Chúc mừng', category: 'fun' },
  { id: 'fun_cake', symbol: '🎂', name: 'Sinh nhật', category: 'fun' },
  { id: 'fun_pizza', symbol: '🍕', name: 'Pizza', category: 'fun' },
  { id: 'fun_coffee', symbol: '☕', name: 'Cà phê', category: 'fun' },
  { id: 'fun_cheers', symbol: '🍻', name: 'Cạn ly', category: 'fun' },
  { id: 'fun_rocket', symbol: '🚀', name: 'Bay lên', category: 'fun' },
  { id: 'fun_rainbow', symbol: '🌈', name: 'Cầu vồng', category: 'fun' },
  { id: 'fun_ghost', symbol: '👻', name: 'Ma tinh nghịch', category: 'fun' },
  { id: 'fun_sparkle', symbol: '✨', name: 'Lấp lánh', category: 'fun' },
  { id: 'fun_gift', symbol: '🎁', name: 'Hộp quà', category: 'fun' },
  { id: 'fun_idea', symbol: '💡', name: 'Sáng kiến', category: 'fun' },
  { id: 'fun_sleep', symbol: '💤', name: 'Ngủ thôi', category: 'fun' },
]

export function isStickerMessage(body: string): boolean {
  if (!body) return false
  return (
    body.startsWith('[sticker:') && body.endsWith(']')
  )
}

export function parseStickerSymbol(body: string): string | null {
  if (!isStickerMessage(body)) return null
  const content = body.slice(9, -1) // remove [sticker: and ]
  const parts = content.split(':')
  return parts[parts.length - 1] || null
}

export function formatStickerMessage(sticker: Sticker): string {
  return `[sticker:${sticker.id}:${sticker.symbol}]`
}
