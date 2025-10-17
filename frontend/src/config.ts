// @ts-ignore
import yaml from 'js-yaml'

export interface SiteConfig {
  signature: string
  avatarPath: string
  server: string
  comment_enabled: boolean
  about?: string // 添加 about 字段
  icp?: string   // 备案号，可选
}

let cachedSite: SiteConfig | null = null

export async function loadConfig(): Promise<SiteConfig> {
  if (cachedSite) return cachedSite
  try {
    const res = await fetch('/config.yaml', { cache: 'no-store' })
    if (!res.ok) throw new Error('HTTP ' + res.status)
    const text = await res.text()
    const parsed: any = yaml.load(text) || {}
    const site: any = parsed.site || {}
    cachedSite = {
      signature: site.signature || '鼠鼠很懒，什么都没有留下',
      avatarPath: site.avatarPath || 'avatar.jpg',
      server: site.server || '',
      comment_enabled: site.comment_enabled !== false && site.comment_enabled !== 'false',
      about: site.about || '鼠鼠已经离开了星球', // 加载 about 字段
      icp: site.icp || ''
    }
  } catch (e) {
    console.error('加载 config.yaml 失败, 使用默认站点配置', e)
    cachedSite = { signature: '签名未配置', avatarPath: 'avatar.jpg', server: '', comment_enabled: true, about: '未配置关于内容' } // 默认 about 内容
  }
  return cachedSite
}

export function getCachedConfig(): SiteConfig | null { return cachedSite }

// 无默认导出，使用具名导入避免重复声明