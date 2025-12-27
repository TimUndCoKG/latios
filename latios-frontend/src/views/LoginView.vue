<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const redirectPath = (route.query.redirect as string) || '/'

const error = ref<string | null>(null)
const username = ref('')
const password = ref('')
const loading = ref(false)

async function login() {
  try {
    loading.value = true;

    const data = {
        "username": username.value,
        "password": password.value,
        "redirect": redirectPath,
    }

    const response = await fetch('/latios-api/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data) 
    })

    if (response.type === "opaqueredirect" || response.status === 302 || response.ok){
        window.location.href = redirectPath;
    } else {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

  } catch (e: any) {
    password.value = ""
    error.value = e.message
  } finally {
    loading.value = false;
  }
}

function cancel() {
    // Clear form on cancel
    username.value = ''
    password.value = ''

}

</script>

<template>
  <div class="bg-base-200 rounded-box p-4">
    <div v-if="loading">
        <span class="loading loading-spinner loading-lg"></span>
    </div>
    <form v-else-if="!loading" class="header" @submit.prevent="login" @reset.prevent="cancel">
      <fieldset class="fieldset w-xs">
        <legend class="fieldset-legend text-2xl">Latios Login</legend>

        <label class="label pt-2">Username</label>
        <input v-model="username" class="input" placeholder="user" required />

        <label class="label">Password</label>
        <input v-model="password" class="input" type="password" placeholder="********" required />

        <div class="flex flex-col gap-2 pt-3">
            <button type="submit" class="btn btn-primary mt-4">Submit</button>
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
}
</style>