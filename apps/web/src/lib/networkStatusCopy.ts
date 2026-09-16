import type { Locale } from '@/lib/i18n'

export type NetworkStatusMessage = {
  title: string
  detail: string
}

export type NetworkStatusCopy = {
  offline: NetworkStatusMessage
  recovered: NetworkStatusMessage
}

const COPY: Record<Locale, NetworkStatusCopy> = {
  vi: {
    offline: {
      title: 'Bạn đang ngoại tuyến',
      detail: 'Các thao tác cần mạng sẽ khả dụng lại khi thiết bị kết nối.',
    },
    recovered: {
      title: 'Đã kết nối lại',
      detail: 'Kết nối mạng của thiết bị đã được khôi phục.',
    },
  },
  en: {
    offline: {
      title: 'You’re offline',
      detail: 'Network actions will be available again when your device reconnects.',
    },
    recovered: {
      title: 'Back online',
      detail: 'Your device connection has been restored.',
    },
  },
}

export function networkStatusCopy(locale: Locale): NetworkStatusCopy {
  return COPY[locale]
}
