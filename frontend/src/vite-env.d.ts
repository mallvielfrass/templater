/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ONLYOFFICE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
