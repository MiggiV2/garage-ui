<script setup lang="ts">
import { nextTick, ref, useTemplateRef } from "vue";

const props = defineProps<{ title: string; confirmText: string; description?: string; action?: string }>();
const emit = defineEmits<{ confirm: [] }>();

const dialog = useTemplateRef<HTMLDialogElement>("dialog");
const input = useTemplateRef<HTMLInputElement>("input");
const typed = ref("");

async function open() {
  typed.value = "";
  dialog.value?.showModal();
  await nextTick();
  input.value?.focus();
}

function submit() {
  if (typed.value !== props.confirmText) return;
  dialog.value?.close();
  emit("confirm");
}

defineExpose({ open });
</script>

<template>
  <dialog ref="dialog" class="modal" @click.self="dialog?.close()">
    <form @submit.prevent="submit">
      <h2>{{ title }}</h2>
      <p v-if="description" class="muted">{{ description }}</p>
      <label class="field">
        <span class="field-label">
          Type <code class="inline-code">{{ confirmText }}</code> to confirm
        </span>
        <input ref="input" v-model="typed" autocomplete="off" spellcheck="false" />
      </label>
      <div class="actions">
        <button type="button" class="btn" @click="dialog?.close()">Cancel</button>
        <button type="submit" class="btn danger-solid" :disabled="typed !== confirmText">
          {{ action ?? "Delete" }}
        </button>
      </div>
    </form>
  </dialog>
</template>
