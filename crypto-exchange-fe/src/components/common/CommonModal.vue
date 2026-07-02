<template>
  <div class="modal-overlay" v-if="visible" @click="closeModal">
    <div class="modal-container" @click.stop>
      <div class="title-bar">
        <span>{{ commonData.title }}</span>
        <button class="close-button" @click="closeModal">X</button>
      </div>
      <div class="modal-content">
        <p>{{ commonData.context }}</p>
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { defineProps, defineEmits } from 'vue';

// Props
defineProps({
  visible: {
    type: Boolean,
    default: true,
    required: true,
  },
  commonData: {
    type: Object,
    default: () => ({
      title: '',
      context: '',
    }),
    required: true,
  },
});

// Emits
const emit = defineEmits(['close']);

// Methods
const closeModal = () => {
  emit('close');
};
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap');

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-container {
  background: rgba(51, 0, 51, 0.95);
  border: 3px solid #ff99ff;
  width: 450px;
  padding: 20px;
  box-shadow: 0 0 20px #ff66cc, 0 0 40px #9900cc;
  border-radius: 10px;
  font-family: 'Press Start 2P', cursive;
  color: #ffffff;
}

.title-bar {
  background: linear-gradient(90deg, #ff33cc, #cc00ff);
  padding: 10px;
  font-size: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #ff99ff;
  text-shadow: 1px 1px 2px #330033;
  margin-bottom: 15px;
}

.title-bar span {
  font-size: 16px;
}

.close-button {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 5px 15px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 12px;
  color: #ffffff;
  text-shadow: 1px 1px #330033;
  transition: all 0.2s;
}

.close-button:hover {
  background: #cc00ff;
  box-shadow: 0 0 5px #ff66cc;
}

.modal-content {
  font-size: 12px;
  line-height: 1.5;
  color: #ffccff;
}

.modal-content p {
  margin: 0 0 15px;
  text-shadow: 1px 1px #330033;
}

.modal-content ::slotted(*) {
  margin-top: 10px;
  font-size: 10px;
  color: #ff99ff;
}
</style>
