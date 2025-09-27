// Explicit module typing for path alias '@/config'
// This references the real implementation types to avoid duplication.
declare module '@/config' {
  export type { SiteConfig } from '../config'
  export function loadConfig(): Promise<import('../config').SiteConfig>
  export function getCachedConfig(): import('../config').SiteConfig | null
}
