/**
 * Chat Bubble Styles configuration and storage for Filvault.
 * Inspired by interactive customizable messenger bubble themes.
 */
import { ref } from 'vue'

export interface BubbleDecoration {
  position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'left' | 'right' | 'top' | 'bottom'
  svg: string
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
  previewSvg: string
}

export const CHAT_BUBBLE_STYLES: ChatBubbleStyle[] = [
  {
    id: 'default',
    name: 'Mặc định',
    bg: 'linear-gradient(135deg, #0084ff 0%, #0099ff 100%)',
    color: '#ffffff',
    decorations: [],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="5" y="6" width="80" height="28" rx="14" fill="#0084ff" />
      </svg>
    `,
  },
  {
    id: 'otter',
    name: 'Rái cá vui nhộn',
    bg: '#8ee0f4',
    color: '#153344',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 34 34" width="30" height="30"><circle cx="16" cy="18" r="12" fill="#9c6d48"/><circle cx="8" cy="10" r="4" fill="#7a4f2d"/><circle cx="24" cy="10" r="4" fill="#7a4f2d"/><circle cx="12" cy="16" r="2" fill="#222"/><circle cx="20" cy="16" r="2" fill="#222"/><ellipse cx="16" cy="20" rx="4" ry="3" fill="#ecd3b8"/><circle cx="16" cy="19" r="1.5" fill="#111"/><ellipse cx="16" cy="23" rx="1.5" ry="1" fill="#df6c72"/><circle cx="28" cy="8" r="3" fill="rgba(255,255,255,0.7)"/><circle cx="2" cy="20" r="2" fill="rgba(255,255,255,0.7)"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 20 20" width="18" height="18"><circle cx="10" cy="10" r="5" fill="none" stroke="rgba(255,255,255,0.8)" stroke-width="2"/><circle cx="14" cy="6" r="2" fill="rgba(255,255,255,0.8)"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="16" y="8" width="66" height="24" rx="12" fill="#8ee0f4" />
        <circle cx="14" cy="18" r="9" fill="#9c6d48" />
        <circle cx="9" cy="12" r="3" fill="#7a4f2d" />
        <ellipse cx="14" cy="20" rx="3.5" ry="2.5" fill="#ecd3b8" />
        <circle cx="14" cy="19" r="1" fill="#111" />
        <circle cx="11" cy="16" r="1" fill="#111" />
        <circle cx="17" cy="16" r="1" fill="#111" />
        <circle cx="78" cy="9" r="3" fill="rgba(255,255,255,0.8)" />
      </svg>
    `,
  },
  {
    id: 'diva',
    name: 'Diva lộng lẫy',
    bg: '#ffffff',
    color: '#1a1a1a',
    border: '1.5px solid #f3d4e6',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><path d="M4 18 C2 8 14 2 24 6 C28 10 30 18 26 22" fill="#222"/><ellipse cx="16" cy="14" rx="6" ry="7" fill="#ffe2d4"/><path d="M12 13 Q14 11 16 13" stroke="#9013fe" stroke-width="2" fill="none"/><path d="M14 18 Q16 21 18 18" stroke="#d0021b" stroke-width="2" fill="#ff4081"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 24 24" width="20" height="20"><path d="M4 14 C6 18 10 20 16 18" stroke="#d0021b" stroke-width="3" stroke-linecap="round" fill="none"/><circle cx="18" cy="16" r="2.5" fill="#e91e63"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="12" y="8" width="66" height="24" rx="12" fill="#ffffff" stroke="#f3d4e6" stroke-width="1.5" />
        <path d="M14 14 C12 6 22 4 28 8" fill="#222" />
        <ellipse cx="20" cy="12" rx="4" ry="4" fill="#ffe2d4" />
        <ellipse cx="20" cy="14" rx="2" ry="1" fill="#d0021b" />
        <path d="M74 24 C76 27 80 27 82 25" stroke="#d0021b" stroke-width="2" fill="none" />
      </svg>
    `,
  },
  {
    id: 'crocodile',
    name: 'Cá sấu thư thái',
    bg: '#6c7fe8',
    color: '#ffffff',
    decorations: [
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 36 30" width="32" height="26"><ellipse cx="18" cy="16" rx="14" ry="10" fill="#4d5fc7"/><path d="M6 10 Q14 4 22 10" fill="#e74c3c"/><circle cx="12" cy="14" r="3" fill="#fff"/><circle cx="13" cy="14" r="1.5" fill="#111"/><circle cx="20" cy="14" r="3" fill="#fff"/><circle cx="19" cy="14" r="1.5" fill="#111"/><ellipse cx="26" cy="18" rx="5" ry="3" fill="#4d5fc7"/></svg>`,
      },
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 24 16" width="20" height="14"><rect x="2" y="4" width="8" height="8" rx="4" fill="#4d5fc7"/><rect x="12" y="4" width="8" height="8" rx="4" fill="#4d5fc7"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="10" width="72" height="22" rx="11" fill="#6c7fe8" />
        <circle cx="68" cy="10" r="6" fill="#4d5fc7" />
        <path d="M64 7 Q68 4 72 7" fill="#e74c3c" />
        <circle cx="66" cy="10" r="1.5" fill="#fff" />
        <circle cx="70" cy="10" r="1.5" fill="#fff" />
        <rect x="16" y="28" width="6" height="4" rx="2" fill="#4d5fc7" />
        <rect x="26" y="28" width="6" height="4" rx="2" fill="#4d5fc7" />
      </svg>
    `,
  },
  {
    id: 'olivia',
    name: 'Olivia Rodrigo',
    bg: '#f8a8c7',
    color: '#461325',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 28 28" width="24" height="24"><circle cx="10" cy="10" r="4" fill="#e91e63"/><circle cx="18" cy="10" r="4" fill="#e91e63"/><circle cx="14" cy="6" r="4" fill="#e91e63"/><circle cx="14" cy="14" r="4" fill="#e91e63"/><circle cx="14" cy="10" r="3" fill="#ffeb3b"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 24 24" width="20" height="20"><path d="M4 12 Q8 4 12 12 Q16 4 20 12 Q16 20 12 12 Q8 20 4 12" fill="#9c27b0"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#f8a8c7" />
        <circle cx="12" cy="10" r="3" fill="#e91e63" />
        <circle cx="16" cy="10" r="3" fill="#e91e63" />
        <circle cx="14" cy="8" r="3" fill="#e91e63" />
        <circle cx="14" cy="12" r="3" fill="#e91e63" />
        <circle cx="14" cy="10" r="1.5" fill="#ffeb3b" />
        <circle cx="76" cy="27" r="2.5" fill="#9c27b0" />
      </svg>
    `,
  },
  {
    id: 'frog_duck',
    name: 'Ếch vịt',
    bg: '#fee679',
    color: '#423300',
    decorations: [
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 40 30" width="36" height="26"><ellipse cx="14" cy="16" rx="9" ry="7" fill="#4caf50"/><circle cx="10" cy="11" r="3" fill="#fff"/><circle cx="10" cy="11" r="1.5" fill="#111"/><circle cx="18" cy="11" r="3" fill="#fff"/><circle cx="18" cy="11" r="1.5" fill="#111"/><ellipse cx="28" cy="15" rx="8" ry="7" fill="#ffb300"/><circle cx="26" cy="13" r="1.5" fill="#111"/><path d="M34 15 L39 17 L34 19 Z" fill="#ff5722"/></svg>`,
      },
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 24 20" width="20" height="16"><circle cx="6" cy="12" r="4" fill="rgba(255,255,255,0.7)"/><circle cx="14" cy="8" r="3" fill="rgba(255,255,255,0.7)"/><circle cx="18" cy="14" r="2" fill="rgba(255,255,255,0.7)"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#fee679" />
        <ellipse cx="68" cy="9" rx="5" ry="4" fill="#4caf50" />
        <ellipse cx="78" cy="8" rx="5" ry="4" fill="#ffb300" />
        <circle cx="12" cy="27" r="3" fill="rgba(255,255,255,0.7)" />
      </svg>
    `,
  },
  {
    id: 'buff_cat',
    name: 'Mèo khoe cơ',
    bg: '#ffffff',
    color: '#1a1a1a',
    border: '1.5px solid #dcdcdc',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><path d="M4 18 L8 4 L16 12 L24 4 L28 18 Z" fill="#ffffff" stroke="#333" stroke-width="1.5"/><circle cx="11" cy="14" r="1.5" fill="#111"/><circle cx="21" cy="14" r="1.5" fill="#111"/><ellipse cx="16" cy="18" rx="3" ry="2" fill="#ffb6c1"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><path d="M4 18 C8 8 24 8 28 18" fill="#ffffff" stroke="#333" stroke-width="1.5"/><circle cx="16" cy="14" r="3" fill="#ffb6c1"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 32 20" width="28" height="18"><path d="M4 10 C10 4 20 4 26 10 C20 18 10 18 4 10" fill="#f5f5f5" stroke="#333" stroke-width="1.5"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#ffffff" stroke="#dcdcdc" stroke-width="1.5" />
        <polygon points="10,12 13,5 17,10" fill="#fff" stroke="#333" stroke-width="1" />
        <polygon points="73,10 77,5 80,12" fill="#fff" stroke="#333" stroke-width="1" />
        <circle cx="15" cy="27" r="2.5" fill="#f0f0f0" stroke="#333" stroke-width="1" />
        <circle cx="75" cy="27" r="2.5" fill="#f0f0f0" stroke="#333" stroke-width="1" />
      </svg>
    `,
  },
  {
    id: 'capybara',
    name: 'Capybara',
    bg: '#cc925e',
    color: '#ffffff',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 36 32" width="32" height="28"><ellipse cx="18" cy="18" rx="14" ry="11" fill="#9e663a"/><ellipse cx="18" cy="18" rx="11" ry="8" fill="#bb7e4b"/><circle cx="12" cy="14" r="1.5" fill="#111"/><ellipse cx="18" cy="20" rx="3.5" ry="2.5" fill="#6d4222"/><ellipse cx="18" cy="6" rx="4" ry="3.5" fill="#ff9800"/><circle cx="18" cy="3" r="1" fill="#4caf50"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 24 16" width="20" height="14"><ellipse cx="12" cy="8" rx="8" ry="5" fill="#9e663a"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#cc925e" />
        <ellipse cx="16" cy="12" rx="8" ry="6" fill="#9e663a" />
        <circle cx="16" cy="5" r="2.5" fill="#ff9800" />
        <circle cx="13" cy="11" r="1" fill="#111" />
      </svg>
    `,
  },
  {
    id: 'frog',
    name: 'Ếch',
    bg: '#27ae60',
    color: '#ffffff',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 28 26" width="24" height="22"><circle cx="14" cy="13" r="10" fill="#1e8449"/><circle cx="14" cy="13" r="6" fill="#fff"/><circle cx="15" cy="13" r="3" fill="#111"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 28 26" width="24" height="22"><circle cx="14" cy="13" r="10" fill="#1e8449"/><circle cx="14" cy="13" r="6" fill="#fff"/><circle cx="13" cy="13" r="3" fill="#111"/></svg>`,
      },
      {
        position: 'bottom',
        svg: `<svg viewBox="0 0 60 14" width="50" height="12"><rect x="2" y="3" width="56" height="7" rx="3.5" fill="#e67e22"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#27ae60" />
        <circle cx="22" cy="7" r="5" fill="#1e8449" />
        <circle cx="22" cy="7" r="2" fill="#fff" />
        <circle cx="68" cy="7" r="5" fill="#1e8449" />
        <circle cx="68" cy="7" r="2" fill="#fff" />
        <rect x="20" y="28" width="50" height="4" rx="2" fill="#e67e22" />
      </svg>
    `,
  },
  {
    id: 'cat_dog',
    name: 'Mèo chó',
    bg: '#fef0d2',
    color: '#422c0d',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 34 30" width="30" height="26"><polygon points="6,16 10,4 18,12" fill="#e67e22"/><polygon points="26,16 22,4 14,12" fill="#e67e22"/><ellipse cx="16" cy="18" rx="11" ry="9" fill="#f39c12"/><circle cx="12" cy="16" r="1.5" fill="#111"/><circle cx="20" cy="16" r="1.5" fill="#111"/><ellipse cx="16" cy="20" rx="2" ry="1.5" fill="#e74c3c"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 34 30" width="30" height="26"><ellipse cx="8" cy="12" rx="4" ry="8" fill="#5499c7"/><ellipse cx="24" cy="12" rx="4" ry="8" fill="#5499c7"/><ellipse cx="16" cy="18" rx="10" ry="9" fill="#7fb3d5"/><circle cx="12" cy="16" r="1.5" fill="#111"/><circle cx="20" cy="16" r="1.5" fill="#111"/><ellipse cx="16" cy="20" rx="3" ry="2" fill="#111"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#fef0d2" />
        <ellipse cx="16" cy="10" rx="6" ry="5" fill="#f39c12" />
        <polygon points="12,7 13,3 16,6" fill="#e67e22" />
        <ellipse cx="74" cy="10" rx="6" ry="5" fill="#7fb3d5" />
        <ellipse cx="78" cy="10" rx="2" ry="4" fill="#5499c7" />
      </svg>
    `,
  },
  {
    id: 'facepalm',
    name: 'Ôm mặt',
    bg: '#ffffff',
    color: '#1a1a1a',
    border: '1.5px solid #d5d8dc',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 34 30" width="30" height="26"><path d="M4 18 C6 6 22 4 28 14" fill="#2c3e50"/><ellipse cx="16" cy="16" rx="9" ry="8" fill="#fdebd0"/><circle cx="12" cy="15" r="1.5" fill="#111"/><circle cx="20" cy="15" r="1.5" fill="#111"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 30 24" width="26" height="20"><path d="M8 18 C14 8 24 10 26 18" fill="#fdebd0" stroke="#333" stroke-width="1.5"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#ffffff" stroke="#d5d8dc" stroke-width="1.5" />
        <path d="M12 12 C14 5 22 5 24 12" fill="#2c3e50" />
        <circle cx="18" cy="13" r="4" fill="#fdebd0" />
        <path d="M72 23 C75 18 80 18 82 23" fill="#fdebd0" stroke="#333" stroke-width="1" />
      </svg>
    `,
  },
  {
    id: 'pepe_heart',
    name: 'Pepe thả tim',
    bg: '#e8b3c6',
    color: '#381424',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 34 34" width="30" height="30"><ellipse cx="16" cy="18" rx="11" ry="10" fill="#7d9b4b"/><ellipse cx="10" cy="12" rx="4" ry="4.5" fill="#fff"/><circle cx="11" cy="12" r="2" fill="#111"/><ellipse cx="20" cy="12" rx="4" ry="4.5" fill="#fff"/><circle cx="19" cy="12" r="2" fill="#111"/><path d="M8 22 Q16 28 24 22" stroke="#b03a2e" stroke-width="2.5" fill="none"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 24 24" width="20" height="20"><path d="M12 20 L4 12 A5 5 0 0 1 12 6 A5 5 0 0 1 20 12 Z" fill="#e74c3c"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#e8b3c6" />
        <ellipse cx="16" cy="14" rx="7" ry="6" fill="#7d9b4b" />
        <circle cx="14" cy="12" r="2" fill="#fff" />
        <circle cx="18" cy="12" r="2" fill="#fff" />
        <path d="M74 10 L70 6 A2 2 0 0 1 74 3 A2 2 0 0 1 78 6 Z" fill="#e74c3c" />
      </svg>
    `,
  },
  {
    id: 'shark_pig',
    name: 'Cá mập heo',
    bg: '#dce6ff',
    color: '#1a284e',
    decorations: [
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 32 26" width="28" height="22"><ellipse cx="16" cy="14" rx="10" ry="8" fill="#f8bbd0"/><ellipse cx="16" cy="16" rx="4" ry="2.5" fill="#f48fb1"/><circle cx="14" cy="16" r="0.8" fill="#880e4f"/><circle cx="18" cy="16" r="0.8" fill="#880e4f"/><circle cx="12" cy="12" r="1.5" fill="#111"/><circle cx="20" cy="12" r="1.5" fill="#111"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 34 26" width="30" height="22"><path d="M4 16 Q18 4 28 16 Z" fill="#5c6bc0"/><circle cx="18" cy="14" r="1.5" fill="#fff"/><circle cx="18" cy="14" r="0.8" fill="#111"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#dce6ff" />
        <ellipse cx="16" cy="24" rx="6" ry="5" fill="#f8bbd0" />
        <path d="M72 26 Q80 16 84 26 Z" fill="#5c6bc0" />
      </svg>
    `,
  },
  {
    id: 'bear_tongue',
    name: 'Gấu thè lưỡi',
    bg: '#ffd4dc',
    color: '#481928',
    decorations: [
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 34 30" width="30" height="26"><circle cx="8" cy="8" r="4" fill="#8d6e63"/><circle cx="24" cy="8" r="4" fill="#8d6e63"/><ellipse cx="16" cy="16" rx="11" ry="10" fill="#a1887f"/><ellipse cx="16" cy="18" rx="5" ry="3.5" fill="#d7ccc8"/><circle cx="16" cy="17" r="1.5" fill="#3e2723"/><path d="M14 20 Q16 26 18 20" fill="#e91e63"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#ffd4dc" />
        <ellipse cx="16" cy="23" rx="6" ry="6" fill="#a1887f" />
        <circle cx="12" cy="18" r="2.5" fill="#8d6e63" />
        <path d="M15 25 Q16 29 17 25" fill="#e91e63" />
      </svg>
    `,
  },
  {
    id: 'doge',
    name: 'Doge',
    bg: '#fce18d',
    color: '#443103',
    decorations: [
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 34 30" width="30" height="26"><polygon points="6,14 10,4 16,12" fill="#d4ac0d"/><polygon points="26,14 22,4 16,12" fill="#d4ac0d"/><ellipse cx="16" cy="17" rx="10" ry="9" fill="#f4d03f"/><ellipse cx="16" cy="20" rx="4" ry="3" fill="#fcf3cf"/><circle cx="16" cy="19" r="1.5" fill="#111"/><circle cx="12" cy="15" r="1.5" fill="#111"/><circle cx="20" cy="15" r="1.5" fill="#111"/></svg>`,
      },
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 30 24" width="26" height="20"><path d="M4 14 C10 6 22 8 26 18" fill="#f4d03f" stroke="#b7950b" stroke-width="1.5"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#fce18d" />
        <ellipse cx="74" cy="10" rx="7" ry="6" fill="#f4d03f" />
        <polygon points="70,6 72,2 75,5" fill="#d4ac0d" />
        <path d="M12 24 C16 18 22 20 24 26" fill="#f4d03f" />
      </svg>
    `,
  },
  {
    id: 'frog_chick',
    name: 'Ếch gà con',
    bg: '#a3e2f5',
    color: '#123746',
    decorations: [
      {
        position: 'top-left',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><circle cx="10" cy="8" r="4" fill="#2ecc71"/><circle cx="10" cy="8" r="1.5" fill="#111"/><circle cx="22" cy="8" r="4" fill="#2ecc71"/><circle cx="22" cy="8" r="1.5" fill="#111"/><ellipse cx="16" cy="16" rx="12" ry="9" fill="#27ae60"/><ellipse cx="16" cy="19" rx="5" ry="3" fill="#a9dfbf"/></svg>`,
      },
      {
        position: 'top-right',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><ellipse cx="16" cy="16" rx="10" ry="9" fill="#f1c40f"/><circle cx="12" cy="14" r="1.5" fill="#111"/><circle cx="20" cy="14" r="1.5" fill="#111"/><polygon points="14,17 18,17 16,21" fill="#e67e22"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 24 20" width="20" height="16"><circle cx="8" cy="10" r="4" fill="rgba(255,255,255,0.7)"/><circle cx="16" cy="12" r="3" fill="rgba(255,255,255,0.7)"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#a3e2f5" />
        <ellipse cx="16" cy="10" rx="6" ry="5" fill="#27ae60" />
        <circle cx="13" cy="7" r="2" fill="#2ecc71" />
        <circle cx="19" cy="7" r="2" fill="#2ecc71" />
        <ellipse cx="74" cy="10" rx="5" ry="5" fill="#f1c40f" />
        <polygon points="72,11 76,11 74,13" fill="#e67e22" />
        <circle cx="78" cy="27" r="2.5" fill="rgba(255,255,255,0.7)" />
      </svg>
    `,
  },
  {
    id: 'hungry_pup',
    name: 'Cún thèm ăn',
    bg: '#d54d39',
    color: '#ffffff',
    decorations: [
      {
        position: 'left',
        svg: `<svg viewBox="0 0 30 34" width="26" height="30"><ellipse cx="14" cy="17" rx="10" ry="12" fill="#e59866"/><ellipse cx="8" cy="14" rx="3" ry="7" fill="#ba4a00"/><circle cx="16" cy="14" r="1.8" fill="#111"/><ellipse cx="20" cy="20" rx="3" ry="2" fill="#111"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="14" y="9" width="68" height="23" rx="11.5" fill="#d54d39" />
        <ellipse cx="14" cy="20" rx="7" ry="8" fill="#e59866" />
        <ellipse cx="10" cy="18" rx="2" ry="5" fill="#ba4a00" />
      </svg>
    `,
  },
  {
    id: 'dino',
    name: 'Khủng long',
    bg: '#8fe5eb',
    color: '#0f3d43',
    decorations: [
      {
        position: 'left',
        svg: `<svg viewBox="0 0 34 34" width="30" height="30"><path d="M8 24 L8 12 Q8 6 18 6 L24 6 Q28 6 28 14 L20 16 L28 20 L28 24 Z" fill="#48c9b0"/><circle cx="18" cy="10" r="2" fill="#fff"/><circle cx="19" cy="10" r="1" fill="#111"/><polygon points="20,16 23,19 26,16" fill="#fff"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="14" y="9" width="68" height="23" rx="11.5" fill="#8fe5eb" />
        <path d="M8 24 L8 12 Q8 8 16 8 L22 8 L18 16 L22 20 L22 24 Z" fill="#48c9b0" />
        <circle cx="14" cy="11" r="1.5" fill="#111" />
      </svg>
    `,
  },
  {
    id: 'dachshund',
    name: 'Chó lạp xưởng',
    bg: '#a5734e',
    color: '#ffffff',
    decorations: [
      {
        position: 'left',
        svg: `<svg viewBox="0 0 32 30" width="28" height="26"><ellipse cx="16" cy="16" rx="10" ry="8" fill="#6e2c00"/><ellipse cx="10" cy="14" rx="3" ry="8" fill="#4a235a"/><circle cx="18" cy="14" r="1.5" fill="#111"/><circle cx="24" cy="17" r="2" fill="#111"/></svg>`,
      },
      {
        position: 'bottom-right',
        svg: `<svg viewBox="0 0 24 16" width="20" height="14"><path d="M4 12 Q14 4 18 12" stroke="#6e2c00" stroke-width="3" fill="none"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="14" y="10" width="68" height="22" rx="11" fill="#a5734e" />
        <ellipse cx="14" cy="18" rx="6" ry="5" fill="#6e2c00" />
        <ellipse cx="10" cy="17" rx="2" ry="4" fill="#4a235a" />
        <circle cx="16" cy="17" r="1" fill="#111" />
      </svg>
    `,
  },
  {
    id: 'hungry_frog',
    name: 'Ếch đói bụng',
    bg: '#ddefda',
    color: '#1d3e1a',
    decorations: [
      {
        position: 'bottom-left',
        svg: `<svg viewBox="0 0 36 30" width="32" height="26"><ellipse cx="16" cy="16" rx="11" ry="9" fill="#27ae60"/><circle cx="12" cy="10" r="3" fill="#2ecc71"/><circle cx="12" cy="10" r="1.5" fill="#111"/><circle cx="20" cy="10" r="3" fill="#2ecc71"/><circle cx="20" cy="10" r="1.5" fill="#111"/><ellipse cx="16" cy="20" rx="3" ry="1.5" fill="#e74c3c"/></svg>`,
      },
      {
        position: 'bottom',
        svg: `<svg viewBox="0 0 80 10" width="70" height="8"><path d="M4 5 Q40 10 76 5" stroke="#e74c3c" stroke-width="2.5" fill="none"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="8" y="9" width="74" height="23" rx="11.5" fill="#ddefda" />
        <ellipse cx="16" cy="23" rx="6" ry="5" fill="#27ae60" />
        <circle cx="14" cy="19" r="1.5" fill="#2ecc71" />
        <path d="M22 25 Q50 28 70 25" stroke="#e74c3c" stroke-width="1.5" fill="none" />
      </svg>
    `,
  },
  {
    id: 'screaming',
    name: 'La hét',
    bg: '#f79c93',
    color: '#3d1612',
    decorations: [
      {
        position: 'left',
        svg: `<svg viewBox="0 0 30 30" width="26" height="26"><circle cx="14" cy="15" r="9" fill="#e74c3c"/><ellipse cx="14" cy="17" rx="4" ry="6" fill="#111"/><circle cx="11" cy="11" r="1.5" fill="#fff"/><circle cx="17" cy="11" r="1.5" fill="#fff"/></svg>`,
      },
    ],
    previewSvg: `
      <svg viewBox="0 0 90 40" width="100%" height="100%">
        <rect x="14" y="9" width="68" height="23" rx="11.5" fill="#f79c93" />
        <circle cx="14" cy="20" r="6" fill="#e74c3c" />
        <ellipse cx="14" cy="22" rx="2.5" ry="3.5" fill="#111" />
      </svg>
    `,
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
