<template>
  <Teleport to="body" v-if="opened">
    <div class="dropdown-backdrop" @click.stop.prevent="close" @contextmenu.stop.prevent="close">
      <div class="dropdown is-active" :style="{ top: `${coords.y}px`, left: `${coords.x}px` }">
        <div class="dropdown-menu" role="menu">
          <div class="dropdown-content">
            <slot></slot>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const emit = defineEmits(['close']);
defineExpose({ open });

const coords = ref({ x: 0, y: 0 });
const opened = ref(false);

function open(anchor: MouseEvent | HTMLElement) {
  let x = 200;
  let y = 0;

  if (anchor instanceof MouseEvent) {
    anchor.stopPropagation();
    anchor.preventDefault();

    const el = (anchor.currentTarget as HTMLElement | null) ?? null;
    if (el) {
      const rect = el.getBoundingClientRect();
      x = rect.right + 6;
      y = rect.top;
    } else {
      x = Math.max(anchor.clientX + 6, 200);
      y = anchor.clientY;
    }
  } else {
    const rect = anchor.getBoundingClientRect();
    x = rect.right + 6;
    y = rect.top;
  }

  coords.value = { x, y };
  opened.value = true;
}

function close() {
  opened.value = false;
  emit('close');
}
</script>

<style lang="scss" scoped>
.dropdown-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  width: 100vw;
  z-index: var(--z-dropdown);

  .dropdown {
    z-index: calc(var(--z-dropdown) + 1);
  }

  .dropdown-content {
    user-select: none;
    transform: none;
    box-shadow: 0px 0px 7px 2px rgba(0, 0, 0, 0.5);
  }
}
</style>
