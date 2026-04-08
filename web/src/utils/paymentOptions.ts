export interface PaymentOption {
  label: string
  value: string
  payType: string
  payMethod?: string
  mode: 'qrcode' | 'redirect' | 'jsapi'
  provider: 'wechat' | 'alipay' | 'stripe'
}

const ALL_OPTIONS: PaymentOption[] = [
  { label: 'WX_NATIVE', value: 'WX_NATIVE', payType: 'wxpay', payMethod: 'native', mode: 'qrcode', provider: 'wechat' },
  { label: 'WX_JSAPI', value: 'WX_JSAPI', payType: 'wxpay', payMethod: 'jsapi', mode: 'jsapi', provider: 'wechat' },
  { label: 'WX_H5', value: 'WX_H5', payType: 'wxpay', payMethod: 'h5', mode: 'redirect', provider: 'wechat' },
  { label: 'ALIPAY_SCAN', value: 'ALIPAY_SCAN', payType: 'alipay', payMethod: 'scan', mode: 'qrcode', provider: 'alipay' },
  { label: 'ALIPAY_H5', value: 'ALIPAY_H5', payType: 'alipay', payMethod: 'h5', mode: 'redirect', provider: 'alipay' },
  { label: 'ALIPAY_WEB', value: 'ALIPAY_WEB', payType: 'alipay', payMethod: 'web', mode: 'redirect', provider: 'alipay' },
  { label: 'STRIPE_ALIPAY', value: 'STRIPE_ALIPAY', payType: 'stripe', payMethod: 'alipay', mode: 'redirect', provider: 'stripe' },
  { label: 'STRIPE_WXPAY', value: 'STRIPE_WXPAY', payType: 'stripe', payMethod: 'wechat_pay', mode: 'qrcode', provider: 'stripe' },
  { label: 'STRIPE_PAYPAL', value: 'STRIPE_PAYPAL', payType: 'stripe', payMethod: 'paypal', mode: 'redirect', provider: 'stripe' },
  { label: 'STRIPE_BANK', value: 'STRIPE_BANK', payType: 'stripe', payMethod: 'bank', mode: 'redirect', provider: 'stripe' },
  { label: 'STRIPE_CHECKOUT', value: 'STRIPE_CHECKOUT', payType: 'stripe', payMethod: 'checkout', mode: 'redirect', provider: 'stripe' },
]

export function getPaymentOptionsByPlugin(plugin: string): PaymentOption[] {
  const normalized = plugin.toLowerCase()
  if (normalized.includes('wechat') || normalized.includes('wxpay')) {
    return ALL_OPTIONS.filter(option => option.provider === 'wechat')
  }
  if (normalized.includes('alipay') || normalized.includes('ali')) {
    return ALL_OPTIONS.filter(option => option.provider === 'alipay')
  }
  if (normalized.includes('stripe')) {
    return ALL_OPTIONS.filter(option => option.provider === 'stripe')
  }
  return []
}

export function getPaymentOptionsByProvider(provider: 'wechat' | 'alipay' | 'stripe') {
  return ALL_OPTIONS.filter(option => option.provider === provider)
}

export function findPaymentOption(value: string) {
  return ALL_OPTIONS.find(option => option.value === value)
}
