<script setup lang="ts">
import Routes from '@/components/Routes.vue';
import Logs from '@/components/Logs.vue';
import Stats from '@/components/Stats.vue';

</script>

<template>
  <div class="main">
    <Routes></Routes>
    <div class="sub">
      <Stats></Stats>
      <Logs></Logs>
    </div>
  </div>
</template>

<style scoped>
  .main {
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-template-rows: auto;
    gap: 2rem;
    padding: 2rem;
  }

  .sub {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
</style>