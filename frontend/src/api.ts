export type SheetInfo = {
  name: string
  start_column: string
  end_column: string
  start_row: number
  end_row: number
}

export type FileInfo = {
  sheets: SheetInfo[]
  file_name: string
  size: number
}

export type CreateTaskResponse = {
  task_id: string
  exel_info: FileInfo
  doc_info: FileInfo
  doc_hash: string
}

export type XlsxInfoResponse = {
  sheets: SheetInfo[]
  file_name: string
}

export type ColumnsResponse = {
  columns: string[]
  sheet: string
}

export type OnlyOfficeConfigResponse = {
  config: Record<string, unknown>
  doc_hash: string
  key: string
}

export type RunTaskResponse = {
  task_id: string
  doc_hashes: string[]
  total_docs: number
}

async function readError(res: Response): Promise<string> {
  const text = await res.text()
  return text || res.statusText
}

export async function apiFetch(path: string, jwt: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers)
  if (jwt) {
    headers.set('Authorization', `Bearer ${jwt}`)
  }
  const res = await fetch(path, { ...init, headers })
  if (!res.ok) {
    throw new Error(await readError(res))
  }
  return res
}

export async function createSession(): Promise<{ user: string; jwt: string }> {
  const res = await fetch('/api/user', { method: 'POST' })
  if (!res.ok) {
    throw new Error(await readError(res))
  }
  return res.json()
}

export async function createTask(jwt: string, xlsx: File, docx: File | null): Promise<CreateTaskResponse> {
  const body = new FormData()
  body.append('exel_file', xlsx)
  if (docx) {
    body.append('doc_file', docx)
  }
  const res = await apiFetch('/api/create_task', jwt, { method: 'POST', body })
  return res.json()
}

export async function fetchXlsxInfo(jwt: string, xlsx: File): Promise<XlsxInfoResponse> {
  const body = new FormData()
  body.append('exel_file', xlsx)
  const res = await apiFetch('/api/xlsx_info', jwt, { method: 'POST', body })
  return res.json()
}

export async function fetchColumns(jwt: string, taskId: string, sheet: string): Promise<ColumnsResponse> {
  const q = new URLSearchParams({ task_id: taskId, sheet })
  const res = await apiFetch(`/api/columns?${q}`, jwt)
  return res.json()
}

export async function fetchEditorConfig(
  jwt: string,
  taskId: string,
): Promise<OnlyOfficeConfigResponse> {
  const q = new URLSearchParams({ task_id: taskId })
  const res = await apiFetch(`/api/onlyoffice/config?${q}`, jwt)
  return res.json()
}

export async function runTask(
  jwt: string,
  taskId: string,
  sheetName: string,
  useFirstRowAsColumns: boolean,
  minRow: number,
  maxRow: number,
): Promise<RunTaskResponse> {
  const q = new URLSearchParams({
    task_id: taskId,
    sheet_name: sheetName,
    use_first_row_as_columns: String(useFirstRowAsColumns),
    min_row: String(minRow),
    max_row: String(maxRow),
  })
  const res = await apiFetch(`/api/run_task?${q}`, jwt, { method: 'POST' })
  return res.json()
}

export async function downloadArchive(jwt: string, hashes: string[]): Promise<Blob> {
  const res = await apiFetch('/api/download_zip', jwt, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ hashes }),
  })
  return res.blob()
}
