import { defineStore } from 'pinia'
import { createSession } from '../api'
import type { SheetInfo } from '../api'

const STORAGE_KEY = 'templater-session'

type Persisted = {
  jwt: string
  user: string
}

type SessionState = Persisted & {
  taskId: string
  columns: string[]
  docHash: string
  sheets: SheetInfo[]
  sheetName: string
}

function loadPersisted(): Persisted {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return { jwt: '', user: '' }
    }
    const parsed = JSON.parse(raw) as Partial<Persisted>
    return {
      jwt: parsed.jwt ?? '',
      user: parsed.user ?? '',
    }
  } catch {
    return { jwt: '', user: '' }
  }
}

export const useSessionStore = defineStore('session', {
  state: (): SessionState => ({
    ...loadPersisted(),
    taskId: '',
    columns: [],
    docHash: '',
    sheets: [],
    sheetName: '',
  }),
  actions: {
    persist() {
      const payload: Persisted = {
        jwt: this.jwt,
        user: this.user,
      }
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
    },
    async ensureSession() {
      if (this.jwt && this.user) {
        return
      }
      const created = await createSession()
      this.jwt = created.jwt
      this.user = created.user
      this.persist()
    },
    setTask(taskId: string, docHash: string, sheets: SheetInfo[], columns: string[], sheetName: string) {
      this.taskId = taskId
      this.docHash = docHash
      this.sheets = sheets
      this.columns = columns
      this.sheetName = sheetName
      this.persist()
    },
    setSheets(sheets: SheetInfo[], sheetName: string) {
      this.sheets = sheets
      this.sheetName = sheetName
      this.columns = []
    },
    setColumns(columns: string[], sheetName: string) {
      this.columns = columns
      this.sheetName = sheetName
    },
    setDocHash(docHash: string) {
      this.docHash = docHash
    },
    clearTask() {
      this.taskId = ''
      this.docHash = ''
      this.sheets = []
      this.columns = []
      this.sheetName = ''
    },
  },
})
