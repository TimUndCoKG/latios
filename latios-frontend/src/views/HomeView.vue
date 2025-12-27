<script setup lang="ts">
import Routes from '@/components/Routes.vue';
import Logs from '@/components/Logs.vue';
import Stats from '@/components/Stats.vue';

</script>

<template>
  <div class="main">
    <div class="sub">
      <Stats></Stats>
      <Routes></Routes>
    </div>
    <Logs></Logs>
  </div>
</template>

<style scoped>
  .main {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-template-rows: auto;
    gap: 2rem;
    padding: 2rem;
    overflow: hidden;
  }

  .sub {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  @media (max-width: 1200px) {
    .main {
      overflow-y: auto;
      grid-template-columns: 1fr;
      grid-template-rows: 1fr 1fr;
      padding: 1rem;
      gap: 1rem;
    }
  }
</style>