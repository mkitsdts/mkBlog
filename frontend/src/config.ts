// 站点配置接口
export interface SiteConfig {
  signature: string
  avatarPath: string
  bgPicturePath: string
  server: string
  comment_enabled: boolean
  about?: string // 添加 about 字段
  icp?: string   // 备案号，可选
}

let cachedSite: SiteConfig | null = null

export const DEFAULT_AVATAR_URL = new URL('./assets/avatar.jpg', import.meta.url).href

export function resolveSiteStaticAssetUrl(assetPath?: string): string {
  const value = String(assetPath || '').trim()
  if (!value) return ''
  if (/^https?:\/\//i.test(value) || value.startsWith('/')) return value
  return `/static/${value.replace(/^\/+/, '')}`
}

export function imageExists(url?: string): Promise<boolean> {
  if (!url) return Promise.resolve(false)
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve(true)
    img.onerror = () => resolve(false)
    img.src = url
  })
}

export async function loadConfig(): Promise<SiteConfig> {
  if (cachedSite) return cachedSite
  try {
    const res = await fetch('/api/site', { cache: 'no-store' })
    if (!res.ok) throw new Error('HTTP ' + res.status)
    const site = await res.json()
    cachedSite = {
      signature: site.signature || '鼠鼠很懒，什么都没有留下',
      avatarPath: site.avatarPath || '',
      bgPicturePath: site.bgPicturePath || '',
      server: site.server || '',
      comment_enabled: site.comment_enabled !== false && site.comment_enabled !== 'false',
      about: site.about || '鼠鼠已经离开了星球', // 加载 about 字段
      icp: site.icp || ''
    }
  } catch (e) {
    console.error('加载站点配置失败, 使用默认站点配置', e)
    cachedSite = {
      signature: '签名未配置',
      avatarPath: '',
      bgPicturePath: '',
      server: '',
      comment_enabled: true,
      about: '未配置关于内容'
    } // 默认 about 内容
  }
  return cachedSite
}

export function getCachedConfig(): SiteConfig | null { return cachedSite }

// 无默认导出，使用具名导入避免重复声明
