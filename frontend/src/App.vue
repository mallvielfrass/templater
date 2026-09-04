<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import ColumnSidebar from './components/ColumnSidebar.vue'
import DocEditor from './components/DocEditor.vue'
import { createTask, downloadArchive, fetchColumns, fetchXlsxInfo, runTask } from './api'
import type { SheetInfo } from './api'
import { useSessionStore } from './stores/session'

const session = useSessionStore()
const xlsx = ref<File | File[] | null>(null)
const docx = ref<File | null>(null)
const newDoc = ref(false)
const docxInput = ref<HTMLInputElement | null>(null)

function firstFile(v: File | File[] | null): File | null {
  if (!v) {
    return null
  }
  return Array.isArray(v) ? (v[0] ?? null) : v
}

const docxLabel = computed(() => {
  if (newDoc.value) {
    return 'Новый документ'
  }
  return docx.value?.name ?? ''
})
const loading = ref(false)
const generating = ref(false)
const downloadingZip = ref(false)
const error = ref('')
const generated = ref<string[]>([])
const editorRef = ref<{ insertPlaceholder: (name: string) => void | Promise<void> } | null>(null)
const useFirstRow = ref(true)
const rangeMin = ref(0)
const rangeMax = ref(100)

const sheets = computed(() => session.sheets)
const currentSheet = computed(() => sheets.value.find((s: SheetInfo) => s.name === session.sheetName))
const sheetNames = computed(() => sheets.value.map((s: SheetInfo) => s.name))
const sheetStartRow = computed(() => currentSheet.value?.start_row || 1)
const sheetEndRow = computed(() => currentSheet.value?.end_row || 1)
const firstDataRow = computed(() => {
  if (!currentSheet.value) {
    return 1
  }
  return useFirstRow.value ? sheetStartRow.value + 1 : sheetStartRow.value
})
const dataRowCount = computed(() => {
  if (!currentSheet.value) {
    return 0
  }
  return Math.max(0, sheetEndRow.value - firstDataRow.value + 1)
})
const maxRangeIndex = computed(() => dataRowCount.value)

function resetRangeDefaults() {
  rangeMin.value = 0
  rangeMax.value = Math.min(100, dataRowCount.value)
}

watch([currentSheet, useFirstRow], resetRangeDefaults, { immediate: true })

const apiMinRow = computed(() => firstDataRow.value + rangeMin.value)
const apiMaxRow = computed(() => Math.max(apiMinRow.value, firstDataRow.value + rangeMax.value - 1))
const rangeValid = computed(() => (
  dataRowCount.value > 0
  && rangeMin.value >= 0
  && rangeMax.value > rangeMin.value
  && rangeMax.value <= dataRowCount.value
))
const selectedRowCount = computed(() => (
  rangeValid.value ? rangeMax.value - rangeMin.value : 0
))

function pickDocx() {
  docxInput.value?.click()
}

function onDocxPicked(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  docx.value = file
  newDoc.value = false
}

function createBlankDocx() {
  docx.value = null
  newDoc.value = true
  if (docxInput.value) {
    docxInput.value.value = ''
  }
}

onMounted(() => {
  session.ensureSession().catch((e: unknown) => {
    error.value = e instanceof Error ? e.message : String(e)
  })
})

async function onXlsxChange(value: File | File[] | null) {
  const file = firstFile(value)
  if (!file) {
    session.setSheets([], '')
    return
  }
  try {
    await session.ensureSession()
    const info = await fetchXlsxInfo(session.jwt, file)
    const sheets = info.sheets ?? []
    session.setSheets(sheets, sheets[0]?.name ?? '')
  } catch (e) {
    session.setSheets([], '')
    error.value = e instanceof Error ? e.message : String(e)
  }
}

function goBack() {
  error.value = ''
  generated.value = []
  xlsx.value = null
  docx.value = null
  newDoc.value = false
  if (docxInput.value) {
    docxInput.value.value = ''
  }
  session.clearTask()
}

async function onUpload() {
  error.value = ''
  generated.value = []
  const xlsxFile = firstFile(xlsx.value)
  const docxFile = newDoc.value ? null : docx.value
  if (!xlsxFile) {
    error.value = 'Нужен xlsx'
    return
  }
  if (!docxFile && !newDoc.value) {
    error.value = 'Откройте или создайте docx'
    return
  }
  loading.value = true
  try {
    await session.ensureSession()
    const created = await createTask(session.jwt, xlsxFile, docxFile)
    const sheets = created.exel_info.sheets ?? []
    const sheetName = session.sheetName || sheets[0]?.name || ''
    let columns: string[] = []
    if (sheetName) {
      const colRes = await fetchColumns(session.jwt, created.task_id, sheetName)
      columns = colRes.columns
    }
    session.setTask(created.task_id, created.doc_hash, sheets, columns, sheetName)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function onSheetChange(name: string) {
  session.sheetName = name
  if (!session.jwt || !session.taskId || !name) {
    return
  }
  try {
    const colRes = await fetchColumns(session.jwt, session.taskId, name)
    session.setColumns(colRes.columns, name)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function insertColumn(name: string) {
  try {
    await editorRef.value?.insertPlaceholder(name)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function onGenerate() {
  error.value = ''
  generated.value = []
  if (!session.jwt || !session.taskId || !session.sheetName) {
    error.value = 'Нет задачи'
    return
  }
  if (!rangeValid.value) {
    error.value = 'Некорректный диапазон строк'
    return
  }
  generating.value = true
  try {
    const res = await runTask(
      session.jwt,
      session.taskId,
      session.sheetName,
      useFirstRow.value,
      apiMinRow.value,
      apiMaxRow.value,
    )
    generated.value = res.doc_hashes
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    generating.value = false
  }
}

function downloadUrl(hash: string) {
  return `/api/files/${hash}`
}

async function downloadGenerated(hash: string) {
  const res = await fetch(downloadUrl(hash), {
    headers: { Authorization: `Bearer ${session.jwt}` },
  })
  if (!res.ok) {
    error.value = await res.text()
    return
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${hash.slice(0, 8)}.docx`
  a.click()
  URL.revokeObjectURL(url)
}

async function downloadAllZip() {
  if (!generated.value.length || !session.jwt) {
    return
  }
  downloadingZip.value = true
  try {
    const blob = await downloadArchive(session.jwt, generated.value)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'documents.zip'
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    downloadingZip.value = false
  }
}
</script>

<template>
  <v-app>
    <v-app-bar
      flat
      border
      height="56"
    >
      <div class="d-flex align-center w-100">
        <span class="text-h6">templater</span>
        <v-spacer />
        <span class="text-caption text-medium-emphasis">{{ session.user || 'нет сессии' }}</span>
      </div>
    </v-app-bar>
    <v-main>
      <div
        class="workspace d-flex flex-column flex-md-row pa-3 pa-md-6 pa-xl-8"
        :class="session.taskId ? 'align-start' : 'align-center justify-center'"
      >
        <v-card
          class="form-card pa-6"
          :class="session.taskId ? 'form-card--task' : 'form-card--start'"
          variant="outlined"
        >
          <v-alert
            v-if="error"
            type="error"
            class="mb-3"
            density="compact"
            closable
            @click:close="error = ''"
          >
            {{ error }}
          </v-alert>
          <template v-if="!session.taskId">
            <v-file-input
              v-model="xlsx"
              label="xlsx"
              accept=".xlsx"
              density="compact"
              prepend-icon="mdi-table"
              show-size
              @update:model-value="onXlsxChange"
            />
            <v-select
              :model-value="session.sheetName"
              :items="sheetNames"
              label="Лист"
              density="compact"
              class="mb-1"
              :disabled="!sheets.length"
              hide-details
              @update:model-value="onSheetChange"
            />
            <v-checkbox
              v-model="useFirstRow"
              label="Первая строка — имена столбцов"
              density="compact"
              hide-details
              class="mb-4"
              :disabled="!sheets.length"
            />
            <div class="text-subtitle-2 mb-2">
              Документ
            </div>
            <input
              ref="docxInput"
              type="file"
              accept=".docx"
              class="d-none"
              @change="onDocxPicked"
            >
            <div class="d-flex flex-wrap ga-2 mb-2">
              <v-btn
                variant="outlined"
                size="small"
                prepend-icon="mdi-file-word"
                @click="pickDocx"
              >
                Открыть
              </v-btn>
              <v-btn
                variant="outlined"
                size="small"
                prepend-icon="mdi-file-plus"
                @click="createBlankDocx"
              >
                Создать новый
              </v-btn>
            </div>
            <div
              v-if="docxLabel"
              class="text-caption text-medium-emphasis mb-4"
            >
              {{ docxLabel }}
            </div>
            <v-btn
              block
              color="primary"
              class="mb-4"
              :loading="loading"
              :disabled="!firstFile(xlsx) || (!docx && !newDoc)"
              @click="onUpload"
            >
              Открыть
            </v-btn>
          </template>
          <v-btn
            v-else
            block
            variant="outlined"
            class="mb-4"
            prepend-icon="mdi-arrow-left"
            @click="goBack"
          >
            Назад
          </v-btn>
          <ColumnSidebar
            :columns="session.columns"
            @insert="insertColumn"
          />
          <div
            v-if="session.taskId && currentSheet"
            class="mt-4"
          >
            <div class="text-subtitle-2 mb-1">
              Диапазон строк
            </div>
            <div class="text-caption text-medium-emphasis mb-2">
              Всего строк данных: {{ dataRowCount }}
              <span v-if="rangeValid">, будет сгенерировано: {{ selectedRowCount }}</span>
            </div>
            <div class="d-flex ga-2 mb-1">
              <v-text-field
                v-model.number="rangeMin"
                type="number"
                label="От"
                density="compact"
                hide-details
                :min="0"
                :max="maxRangeIndex"
              />
              <v-text-field
                v-model.number="rangeMax"
                type="number"
                label="До"
                density="compact"
                hide-details
                :min="0"
                :max="maxRangeIndex"
              />
            </div>
            <div
              v-if="!rangeValid && dataRowCount > 0"
              class="text-caption text-error"
            >
              Диапазон должен быть от 0 до {{ maxRangeIndex }}
            </div>
            <div
              v-else-if="dataRowCount === 0"
              class="text-caption text-medium-emphasis"
            >
              Нет строк для генерации
            </div>
          </div>
          <v-btn
            block
            class="mt-4"
            color="secondary"
            :loading="generating"
            :disabled="!session.taskId || !rangeValid"
            @click="onGenerate"
          >
            Сгенерировать
          </v-btn>
          <v-btn
            block
            class="mt-2"
            variant="outlined"
            prepend-icon="mdi-folder-zip"
            :loading="downloadingZip"
            :disabled="!generated.length"
            @click="downloadAllZip"
          >
            Скачать все архивом
          </v-btn>
          <v-list
            v-if="generated.length"
            density="compact"
            class="mt-3"
          >
            <v-list-item
              v-for="hash in generated"
              :key="hash"
              :title="hash.slice(0, 12)"
              @click="downloadGenerated(hash)"
            >
              <template #prepend>
                <v-icon>mdi-download</v-icon>
              </template>
            </v-list-item>
          </v-list>
        </v-card>
        <div
          v-if="session.taskId"
          class="editor-pane flex-grow-1 mt-4 mt-md-0 ml-md-6"
        >
          <DocEditor
            :key="session.taskId"
            ref="editorRef"
            :jwt="session.jwt"
            :task-id="session.taskId"
          />
        </div>
      </div>
    </v-main>
    <v-footer
      border
      class="text-medium-emphasis flex-grow-0"
    >
      <v-container
        class="footer-inner py-6 mx-auto"
      >
        <v-row>
          <v-col
            cols="12"
            md="5"
          >
            <div class="text-subtitle-2 text-high-emphasis mb-2">
              templater
            </div>
            <p class="text-body-2 mb-0">
              Собирает документы Word из таблицы Excel и шаблона docx: на каждую строку данных — отдельный файл.
            </p>
          </v-col>
          <v-col
            cols="12"
            md="7"
          >
            <div class="text-subtitle-2 text-high-emphasis mb-2">
              Как пользоваться шаблоном
            </div>
            <ol class="text-body-2 ps-4 mb-0">
              <li>Загрузите xlsx. Имена столбцов берутся из первой строки листа.</li>
              <li>Откройте готовый docx или создайте новый.</li>
              <li>
                В тексте шаблона вставьте плейсхолдеры вида
                <code>{ИмяСтолбца}</code>
                — кликом по столбцу в панели или вручную.
              </li>
              <li>Имя в фигурных скобках должно совпадать со столбцом, включая регистр.</li>
              <li>Нажмите «Сгенерировать»: по одной строке Excel получится один docx.</li>
            </ol>
          </v-col>
        </v-row>
      </v-container>
    </v-footer>
  </v-app>
</template>

<style>
.v-app-bar .v-toolbar__content {
  width: 100%;
  padding-inline: 24px;
}
.workspace {
  min-height: calc(100% - 24px);
}
.form-card {
  width: 100%;
}
.editor-pane {
  width: 100%;
  min-width: 0;
}
@media (min-width: 960px) {
  .form-card--task {
    width: 360px;
    flex-shrink: 0;
  }
  .form-card--start {
    width: 100%;
    max-width: 360px;
  }
  .editor-pane {
    width: auto;
  }
}
@media (min-width: 1920px) {
  .v-app-bar {
    height: 72px !important;
  }
  .v-app-bar .v-toolbar__content {
    height: 72px !important;
    padding-inline: 40px;
  }
  .form-card {
    padding: 32px !important;
  }
  .form-card--task {
    width: 40%;
    flex-shrink: 0;
  }
  .form-card--start {
    width: 40%;
    max-width: none;
  }
  .form-card .v-btn {
    min-height: 48px;
    font-size: 1rem;
  }
  .form-card .v-chip {
    height: 36px;
    font-size: 0.95rem;
  }
  .footer-inner {
    max-width: 1400px;
    padding-block: 40px;
  }
}
.v-application__wrap {
  min-height: 100%;
}
.footer-inner {
  max-width: 960px;
}
.v-footer {
  position: relative;
  flex: 0 0 auto;
}
</style>
