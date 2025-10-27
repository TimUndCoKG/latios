<script setup lang="ts">
import { ref } from 'vue'

const error = ref<string | null>(null)
const domain = ref('')
const targetPath = ref('')
const isStatic = ref(false)
const useHTTPS = ref(false)
const enforceAuth = ref(false)

async function addRoute() {
  try {
    const route = {
      domain: domain.value,
      target_path: targetPath.value,
      is_static: isStatic.value,
      use_https: useHTTPS.value,
      enforce_auth: enforceAuth.value
    }

    const response = await fetch('/latios-api/routes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(route)
    })

    if (!response.ok) {
      if (response.status === 401) {
        window.location.href = '/latios-api/login'
        return
      }
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    // Clear form on success
    domain.value = ''
    targetPath.value = ''
    isStatic.value = false
    useHTTPS.value = false
    enforceAuth.value = false
    error.value = null

    window.location.href = '/latios/'


  } catch (e: any) {
    error.value = e.message
  }
}

function cancel() {
    // Clear form on cancel
    
    domain.value = ''
    targetPath.value = ''
    isStatic.value = false
    useHTTPS.value = false
    enforceAuth.value = false
    error.value = null

    window.location.href = '/latios/'
}

</script>

<template>
  <div>
    <form class="header" @submit.prevent="addRoute" @reset.prevent="cancel">
      <fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4">
        <legend class="fieldset-legend">Add Route</legend>

        <label class="label">Domain</label>
        <input v-model="domain" class="input" placeholder="dummy.yourdomain.com" required />

        <label class="label">Target</label>
        <input v-model="targetPath" class="input" placeholder="http://docker-container:1234" required />

        <label class="label flex justify-between">
          <p>Static Path</p>
          <input v-model="isStatic" type="checkbox" class="checkbox" />
        </label>

        <label class="label flex justify-between">
          <p>Use HTTPS</p>
          <input v-model="useHTTPS" type="checkbox" class="checkbox" />
        </label>

        <label class="label flex justify-between">
          <p>Enforce Auth</p>
          <input v-model="enforceAuth" type="checkbox" class="checkbox" />
        </label>

        <div class="flex flex-col gap-2">
            <button type="submit" class="btn btn-primary mt-4">Submit</button>
            <button type="reset" class="btn btn-outline mt-4">Back</button>
        </div>
      </fieldset>
    </form>

    <div v-if="error" class="text-red-500 mt-4">Error: {{ error }}</div>
  </div>
</template>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}
</style>