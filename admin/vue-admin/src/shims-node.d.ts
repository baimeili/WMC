// 本仓库未安装 @types/node，而 tsconfig.json 的 include 包含 vite.config.ts，
// 该文件只用到 node:url 的少量类型。这里补充最小模块声明，避免为了构建配置引入额外开发依赖。
declare module 'node:url' {
  export function fileURLToPath(url: string | URL): string
  export function pathToFileURL(path: string): URL
}

declare module 'node:path' {
  export function resolve(...paths: string[]): string
  export function dirname(path: string): string
  export function join(...paths: string[]): string
}
