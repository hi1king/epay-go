export interface PaymentOption {
  label: string
  value: string
  payType: string
  payMethod?: string
  mode: 'qrcode' | 'redirect' | 'jsapi'
  provider: 'wechat' | 'alipay'
}

const ALL_OPTIONS: PaymentOption[] = [
  { label: 'WX_NATIVE', value: 'WX_NATIVE', payType: 'wxpay', payMethod: 'native', mode: 'qrcode', provider: 'wechat' },
  { label: 'WX_JSAPI', value: 'WX_JSAPI', payType: 'wxpay', payMethod: 'jsapi', mode: 'jsapi', provider: 'wechat' },
  { label: 'WX_H5', value: 'WX_H5', payType: 'wxpay', payMethod: 'h5', mode: 'redirect', provider: 'wechat' },
  { label: 'ALIPAY_SCAN', value: 'ALIPAY_SCAN', payType: 'alipay', payMethod: 'scan', mode: 'qrcode', provider: 'alipay' },
  { label: 'ALIPAY_H5', value: 'ALIPAY_H5', payType: 'alipay', payMethod: 'h5', mode: 'redirect', provider: 'alipay' },
  { label: 'ALIPAY_WEB', value: 'ALIPAY_WEB', payType: 'alipay', payMethod: 'web', mode: 'redirect', provider: 'alipay' },
]

export function getPaymentOptionsByPlugin(plugin: string): PaymentOption[] {
  const normalized = plugin.toLowerCase()
  if (normalized.includes('wechat') || normalized.includes('wxpay')) {
    return ALL_OPTIONS.filter(option => option.provider === 'wechat')
  }
  if (normalized.includes('alipay') || normalized.includes('ali')) {
    return ALL_OPTIONS.filter(option => option.provider === 'alipay')
  }
  return []
}

export function getPaymentOptionsByProvider(provider: 'wechat' | 'alipay') {
  return ALL_OPTIONS.filter(option => option.provider === provider)
}

export function findPaymentOption(value: string) {
  return ALL_OPTIONS.find(option => option.value === value)
}
