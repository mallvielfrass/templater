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
const editorVisible = ref(false)

function destroyEditor() {
  const instance = window.DocEditor?.instances?.[editorId]
  if (instance?.destroyEditor) {
    instance.destroyEditor()
    window.DocEditor.instances[editorId] = undefined
  }
}

function grabConnector() {
  const instance = window.DocEditor?.instances?.[editorId]
  if (instance?.createConnector) {
    connector.value = instance.createConnector()
  }
}

function onDocumentReady() {
  grabConnector()
  window.setTimeout(grabConnector, 500)
  window.setTimeout(grabConnector, 2000)
}

function onWindowMessage(e: MessageEvent) {
  if (e.data?.type === 'templater:plugin-ready' && e.ports?.[0]) {
    if (pluginPort) {
      try {
        pluginPort.close()
      } catch {}
    }
    pluginPort = e.ports[0]
    pluginPort.postMessage({ type: 'templater:ack' })
  }
}

onMounted(() => {
  window.addEventListener('message', onWindowMessage)
})

async function loadConfig() {
  error.value = ''
  connector.value = null
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
  insertPlaceholder(name: string) {
    const text = '{' + name + '}'
    if (pluginPort) {
      pluginPort.postMessage({ action: 'insertText', text })
      return
    }
    if (!connector.value) {
      grabConnector()
    }
    if (connector.value) {
      connector.value.executeMethod('PasteText', [text])
      return
    }
    throw new Error('Редактор ещё не готов')
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
  height: calc(100vh - 120px);
  min-height: 480px;
}
.editor-wrap :deep(.onlyoffice-editor),
.editor-wrap :deep(iframe) {
  height: 100%;
  width: 100%;
}
</style>
