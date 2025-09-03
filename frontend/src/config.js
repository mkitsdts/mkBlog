// 运行时通过 /config.yaml 加载完整 YAML，并返回 site 段。
import yaml from 'js-yaml';

let cachedSite = null;

export async function loadConfig() {
  if (cachedSite) return cachedSite;
  try {
    const res = await fetch('/config.yaml', { cache: 'no-store' });
    if (!res.ok) throw new Error('HTTP ' + res.status);
    const text = await res.text();
    const parsed = yaml.load(text) || {};
    const site = parsed.site || {};
    cachedSite = {
      signature: site.signature || '签名未配置',
      avatarPath: site.avatarPath || 'avatar.jpg',
      server: site.server || ''
    };
  } catch (e) {
    console.error('加载 config.yaml 失败, 使用默认站点配置', e);
    cachedSite = { signature: '签名未配置', avatarPath: 'avatar.jpg', server: '' };
  }
  return cachedSite;
}

export function getCachedConfig() { return cachedSite; }

export default { loadConfig, getCachedConfig };
