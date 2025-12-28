<script setup lang="ts">
import { ref, onMounted } from 'vue'

// const stats = ref({
//   total_requests: 2598,
//   error_count: 2347,
//   avg_latency_ms: .457
// })
const stats = ref()
const error = ref(null)

async function fetchStats() {
  try {
    const response = await fetch('/latios-api/stats')

    if (!response.ok) {
      if (response.status === 401) {
        window.location.href = '/latios-api/login'
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    stats.value = await response.json()
    console.log(stats.value)

  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(() => {
  fetchStats()
})

function convertNum(num: number): string {
  if (num >= 1000) {
    return (num / 1000).toFixed(2).toString() + "K"
  }
  return num.toString()
}
</script>

<template>
  <div class="bg-base-200 border-base-300 rounded-box border p-4 --box">
    <div class="header">
      <h2>Stats (30d)</h2>

      <div class="btn-group flex gap-2">
        <button class="btn btn-outline" @click="fetchStats">Refresh</button>
      </div>

    </div>

    <div v-if="error">Error: {{ error }}</div>

    <div v-if="stats != undefined" class="overflow-x-auto">
      <div class="stats shadow flex ">
        <div class="stat">
          <!-- <div class="stat-figure text-secondary">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              class="inline-block h-8 w-8 stroke-current"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              ></path>
            </svg>
          </div> -->
          <div class="stat-title">Total Requests</div>
          <div class="stat-value">{{ convertNum(stats.total_requests) }}</div>
          <div class="stat-desc">Last 30 days</div>
        </div>

        <div class="stat">
          <div class="stat-title">Errors</div>
          <div class="stat-value">{{ convertNum(stats.error_count) }}</div>
          <div class="stat-desc">400+ status</div>
        </div>

         <div class="stat">
          <div class="stat-title">Response time</div>
          <div class="stat-value">{{ stats.avg_latency_ms.toFixed(2) }}ms</div>
          <div class="stat-desc"></div>
        </div>
      </div>

    </div>

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