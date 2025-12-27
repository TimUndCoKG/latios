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
    height: 100dvh;
    overflow: hidden;
    padding: 1rem;
  }

  @media (max-width: 1200px) {
    .wrapper{
      width: fit-content;
      height: fit-content;
    }
  }

</style>
