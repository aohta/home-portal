<script setup lang="ts">
import ElButton from "element-plus/es/components/button/index";

withDefaults(defineProps<{ loading?: boolean }>(), { loading: false });
</script>

<template>
  <ElButton native-type="button" type="primary" plain :loading="loading" :disabled="loading"><slot /></ElButton>
</template>
