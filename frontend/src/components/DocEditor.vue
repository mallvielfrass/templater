<script setup lang="ts">
import { DocumentEditor } from '@onlyoffice/document-editor-vue'
import type { IConfig } from '@onlyoffice/document-editor-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { fetchEditorConfig } from '../api'

type Connector = {
  executeMethod: (name: string, args: unknown[]) => void
}

const props = defineProps<{
  jwt: string
  taskId: string
}>()

const documentServerUrl = (import.meta.env.VITE_ONLYOFFICE_URL || 'http://localhost:8080').replace(/\/?$/, '/')
const editorId = 'docxEditor'
const config = ref<IConfig | null>(null)
const error = ref('')
const connector = ref<Connector | null>(null)
let pluginPort: MessagePort | null = null
const pluginReady = ref(false)
const editorVisible = ref(false)

function destroyEditor() {
  const instance = window.DocEditor?.instances?.[editorId]
  if (instance?.destroyEditor) {
    instance.destroyEditor()
    window.DocEditor.instances[editorId] = undefined
  }
}

function editorInstances(): Record<string, { createConnector?: () => Connector; executeMethod?: Connector['executeMethod'] }> {
  const w = window as Window & {
    DocEditor?: { instances?: Record<string, { createConnector?: () => Connector; executeMethod?: Connector['executeMethod'] }> }
    DocsAPI?: { DocEditor?: { instances?: Record<string, { createConnector?: () => Connector; executeMethod?: Connector['executeMethod'] }> } }
  }
  return { ...(w.DocEditor?.instances ?? {}), ...(w.DocsAPI?.DocEditor?.instances ?? {}) }
}

function grabConnector() {
  const instances = editorInstances()
  const instance = instances[editorId] ?? Object.values(instances).find(Boolean)
  if (instance?.createConnector) {
    connector.value = instance.createConnector()
  }
}

function findFrameApi(w: Window, depth = 0): Connector | null {
  if (depth > 10) {
    return null
  }
  try {
    const plugin = (w as Window & { Asc?: { plugin?: Connector } }).Asc?.plugin
    if (plugin && typeof plugin.executeMethod === 'function') {
      return plugin
    }
  } catch {}
  try {
    for (let i = 0; i < w.frames.length; i++) {
      const found = findFrameApi(w.frames[i], depth + 1)
      if (found) {
        return found
      }
    }
  } catch {}
  return null
}

function pasteText(text: string): boolean {
  if (pluginPort) {
    pluginPort.postMessage({ action: 'insertText', text })
    return true
  }
  if (!connector.value) {
    grabConnector()
  }
  if (connector.value) {
    connector.value.executeMethod('PasteText', [text])
    return true
  }
  for (const inst of Object.values(editorInstances())) {
    if (inst?.createConnector) {
      try {
        inst.createConnector().executeMethod('PasteText', [text])
        return true
      } catch {}
    }
    if (typeof inst?.executeMethod === 'function') {
      inst.executeMethod('PasteText', [text])
      return true
    }
  }
  const frameApi = findFrameApi(window)
  if (frameApi) {
    frameApi.executeMethod('PasteText', [text])
    return true
  }
  if (pluginReady.value) {
    broadcastInsert(text)
    return true
  }
  return false
}

function onDocumentReady() {
  grabConnector()
  window.setTimeout(grabConnector, 500)
  window.setTimeout(grabConnector, 2000)
}

function broadcastInsert(text: string) {
  const msg = { type: 'templater:insert', text }
  const walk = (w: Window) => {
    try {
      w.postMessage(msg, '*')
    } catch {}
    try {
      for (let i = 0; i < w.frames.length; i++) {
        walk(w.frames[i])
      }
    } catch {}
  }
  walk(window)
  document.querySelectorAll('iframe').forEach((el) => {
    if (el.contentWindow) {
      walk(el.contentWindow)
    }
  })
}

function onWindowMessage(e: MessageEvent) {
  if (e.data?.type !== 'templater:plugin-ready') {
    return
  }
  pluginReady.value = true
  if (e.ports?.[0]) {
    if (pluginPort) {
      try {
        pluginPort.close()
      } catch {}
    }
    pluginPort = e.ports[0]
    pluginPort.postMessage({ type: 'templater:ack' })
  }
  try {
    ;(e.source as Window | null)?.postMessage({ type: 'templater:ack' }, '*')
  } catch {}
}

onMounted(() => {
  window.addEventListener('message', onWindowMessage)
})

async function loadConfig() {
  error.value = ''
  connector.value = null
  pluginReady.value = false
  editorVisible.value = false
  if (pluginPort) {
    try {
      pluginPort.close()
    } catch {}
    pluginPort = null
  }
  destroyEditor()
  config.value = null
  if (!props.jwt || !props.taskId) {
    return
  }
  try {
    const res = await fetchEditorConfig(props.jwt, props.taskId)
    config.value = res.config as IConfig
    await nextTick()
    editorVisible.value = true
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

watch(
  () => [props.jwt, props.taskId],
  () => {
    void loadConfig()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  window.removeEventListener('message', onWindowMessage)
  if (pluginPort) {
    try {
      pluginPort.close()
    } catch {}
    pluginPort = null
  }
  destroyEditor()
  connector.value = null
})

const ready = computed(() => Boolean(config.value) && editorVisible.value)

defineExpose({
  async insertPlaceholder(name: string) {
    const text = '{' + name + '}'
    try {
      await navigator.clipboard.writeText(text)
    } catch {}
    if (!pasteText(text)) {
      throw new Error('Скопировано ' + text + '. Кликни в документ и нажми Ctrl+V')
    }
  },
})
</script>

<template>
  <div class="editor-wrap">
    <v-alert
      v-if="error"
      type="error"
      class="mb-2"
      density="compact"
    >
      {{ error }}
    </v-alert>
    <DocumentEditor
      v-if="ready && config"
      :id="editorId"
      :document-server-url="documentServerUrl"
      :config="config"
      height="100%"
      :events_onDocumentReady="onDocumentReady"
      :events_onAppReady="onDocumentReady"
    />
    <div
      v-else-if="!error"
      class="text-medium-emphasis"
    >
      Откройте docx
    </div>
  </div>
</template>

<style scoped>
.editor-wrap {
  height: 60vh;
  min-height: 360px;
}
@media (min-width: 960px) {
  .editor-wrap {
    height: calc(100vh - 120px);
    min-height: 480px;
  }
}
@media (min-width: 1920px) {
  .editor-wrap {
    height: calc(100vh - 160px);
    min-height: 800px;
  }
}
.editor-wrap :deep(.onlyoffice-editor),
.editor-wrap :deep(iframe) {
  height: 100%;
  width: 100%;
}
</style>
