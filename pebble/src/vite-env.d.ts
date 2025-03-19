/// <reference types="vite/client" />

interface ImportMetaEnv {
  VITE_BLOG_URL: string
  VITE_MEMO_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
