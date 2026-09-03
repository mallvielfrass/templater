<script setup lang="ts">
defineProps<{
  columns: string[]
}>()

const emit = defineEmits<{
  insert: [name: string]
}>()
</script>

<template>
  <div>
    <div class="text-subtitle-2 mb-2">
      Столбцы
    </div>
    <div class="d-flex flex-wrap ga-2">
      <v-chip
        v-for="name in columns"
        :key="name"
        color="primary"
        variant="outlined"
        size="small"
        clickable
        @click="emit('insert', name)"
      >
        {{ '{' + name + '}' }}
      </v-chip>
    </div>
    <div
      v-if="!columns.length"
      class="text-medium-emphasis text-caption"
    >
      Загрузите xlsx
    </div>
  </div>
</template>
