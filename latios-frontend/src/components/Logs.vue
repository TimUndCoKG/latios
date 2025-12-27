<script setup lang="ts">
import { ref, onMounted } from 'vue'

const logs = ref([])
const error = ref(null)

function formatDate(dateString: string) {
  if (!dateString) return '';
  const date = new Date(dateString)

  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

async function fetchLogs() {
  try {
    const response = await fetch('/latios-api/logs')

    if (!response.ok) {
      if (response.status === 401) {
        window.location.href = '/latios-api/login'
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    logs.value = await response.json()
    console.log(logs.value)

  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(() => {
  fetchLogs()
})
</script>

<template>
  <div class="bg-base-200 border-base-300 rounded-box border p-4">
    <div class="header">
      <h2>Logs</h2>

      <div class="btn-group flex gap-2">
        <button class="btn btn-outline" @click="fetchLogs">Refresh</button>
      </div>

    </div>

    <div v-if="error">Error: {{ error }}</div>

    <div v-else-if="logs.length" class="overflow-x-auto">
      <table class="table">

        <thead>
          <tr>
            <th>Status</th>
            <th>Method</th>
            <th>Timestamp</th>
            <th>Target</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="log in (logs as any[])" :key="log.id"> 
            <td>{{ log.status_code }}</td>
            <td>{{ log.method }}</td>
            <td>{{ formatDate(log.timestamp) }}</td>
            <td>{{ log.host }}{{ log.path }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else>No logs found</div>

  </div>
</template>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  gap: 2rem;
}
</style>