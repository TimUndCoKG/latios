<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <header>
    <div class="wrapper">
      <RouterView />
    </div>
  </header>

</template>

<style scoped>
  .wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100vw;
    min-height: 100dvh;
    overflow: hidden;
    padding: 1rem;
  }
</style>
