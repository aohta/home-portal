<script setup lang="ts">
import ElAlert from "element-plus/es/components/alert/index";

withDefaults(defineProps<{ severity?: "error" | "info" }>(), { severity: "info" });
</script>

<template>
  <ElAlert :type="severity" :closable="false" :role="severity === 'error' ? 'alert' : 'status'" :aria-live="severity === 'error' ? 'assertive' : 'polite'" aria-atomic="true"><slot /></ElAlert>
</template>
