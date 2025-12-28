<script setup lang="ts">
import { ref, onMounted } from 'vue'

const routes = ref([])
const error = ref(null)

async function fetchRoutes() {
  try {
    const response = await fetch('/latios-api/routes')

    if (!response.ok) {
      if (response.status === 401) {
        window.location.href = '/latios-api/login'
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    routes.value = await response.json()
    console.log(routes.value)

  } catch (e: any) {
    error.value = e.message
  }
}

async function deleteRoute(domain: string, id: number) {
  try {
    const response = await fetch(`/latios-api/routes`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ domain })
    })

    if (!response.ok) {
      if (response.status === 401) {
        window.location.href = '/latios-api/login'
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    routes.value = routes.value.filter((route: any) => route.id !== id)

  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(() => {
  fetchRoutes()
})
</script>

<template>
  <div class="bg-base-200 border-base-300 rounded-box border p-4 h-full --box">
    <div class="header">
      <h2>My Routes</h2>

      <div class="btn-group flex gap-2">
        <router-link to="/add-route" class="btn btn-primary">Add</router-link>
        <button class="btn btn-outline" @click="fetchRoutes">Refresh</button>
      </div>

    </div>

    <div v-if="error">Error: {{ error }}</div>

    <div v-else-if="routes.length" class="overflow-x-auto">
      <table class="table">

        <thead>
          <tr>
            <th>Domain</th>
            <th>Target</th>
            <th>Use Auth</th>
            <th>Static Path</th>
            <th>Actions</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="route in (routes as any[])" :key="route.id"> 
            
            <td>{{ route.domain }}</td>
            <td>{{ route.target_path }}</td>
            <td>
              <input type="checkbox" class="checkbox" :checked="route.enforce_auth" disabled />
            </td>
            <td>
              <input type="checkbox" class="checkbox" :checked="route.is_static" disabled />
            </td>
            <td>
              <button class="btn btn-outline btn-sm" @click="deleteRoute(route.domain, route.id)">Delete</button>
            </td>

          </tr>
        </tbody>
      </table>
    </div>

    <div v-else>No routes found</div>

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