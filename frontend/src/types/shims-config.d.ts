// Type definitions for runtime config loader
// Auto-generated to satisfy TypeScript module resolution for '@/config'

export interface SiteConfig {
  signature: string;
  avatarPath: string;
  server: string;
  comment_enabled: boolean;
}

export function loadConfig(): Promise<SiteConfig>;
export function getCachedConfig(): SiteConfig | null;

declare const _default: {
  loadConfig: typeof loadConfig;
  getCachedConfig: typeof getCachedConfig;
};
export default _default;

declare module '@/config' {
  export interface SiteConfig {
    signature: string;
    avatarPath: string;
    server: string;
    comment_enabled: boolean;
  }
  export function loadConfig(): Promise<SiteConfig>;
  export function getCachedConfig(): SiteConfig | null;
  const _default: {
    loadConfig: typeof loadConfig;
    getCachedConfig: typeof getCachedConfig;
  };
  export default _default;
}
