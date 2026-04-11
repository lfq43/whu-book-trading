<template>
  <!-- 缩略图 -->
  <div class="image-thumbnail" @click="openViewer">
    <el-image
        :src="src"
        fit="cover"
        :style="{ width: width, height: height }"
        class="thumbnail-img"
    >
      <template #error>
        <div class="image-placeholder">
          <el-icon><Picture /></el-icon>
        </div>
      </template>
    </el-image>
    <div class="image-overlay">
      <el-icon><ZoomIn /></el-icon>
    </div>
  </div>

  <!-- 全屏预览弹窗 -->
  <el-dialog
      v-model="visible"
      :title="title"
      :width="dialogWidth"
      :align-center="true"
      class="image-viewer-dialog"
      @close="handleClose"
  >
    <div class="viewer-container" @click="toggleControls">
      <div class="image-wrapper" :style="{ height: viewerHeight }">
        <el-image
            ref="viewerImageRef"
            :src="src"
            fit="contain"
            class="viewer-image"
            :preview-src-list="previewList"
            :initial-index="currentIndex"
            :hide-on-click-modal="true"
        />
      </div>

      <!-- 控制栏 -->
      <div class="viewer-controls" :class="{ hidden: controlsHidden }">
        <div class="control-buttons">
          <el-button circle @click="zoomOut" :disabled="scale <= 0.5">
            <el-icon><ZoomOut /></el-icon>
          </el-button>
          <el-button circle @click="resetZoom">
            <span class="scale-text">{{ Math.round(scale * 100) }}%</span>
          </el-button>
          <el-button circle @click="zoomIn" :disabled="scale >= 3">
            <el-icon><ZoomIn /></el-icon>
          </el-button>
          <el-button circle @click="rotateLeft">
            <el-icon><RefreshLeft /></el-icon>
          </el-button>
          <el-button circle @click="rotateRight">
            <el-icon><RefreshRight /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <div class="viewer-footer" v-if="showInfo">
      <div class="image-info">
        <span>{{ title || '图片预览' }}</span>
        <span v-if="size">大小: {{ formatSize(size) }}</span>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Picture, ZoomIn, ZoomOut, RefreshLeft, RefreshRight } from '@element-plus/icons-vue'

const props = defineProps({
  src: {
    type: String,
    required: true
  },
  title: {
    type: String,
    default: ''
  },
  width: {
    type: String,
    default: '100%'
  },
  height: {
    type: String,
    default: '150px'
  },
  previewList: {
    type: Array,
    default: () => []
  },
  currentIndex: {
    type: Number,
    default: 0
  },
  showInfo: {
    type: Boolean,
    default: true
  },
  size: {
    type: Number,
    default: null
  }
})

const visible = ref(false)
const controlsHidden = ref(false)
let controlsTimer = null
const scale = ref(1)
const rotation = ref(0)
const viewerImageRef = ref(null)

const dialogWidth = computed(() => {
  return window.innerWidth < 768 ? '95%' : '80%'
})

const viewerHeight = computed(() => {
  return window.innerHeight < 600 ? '400px' : '70vh'
})

const openViewer = () => {
  visible.value = true
  resetZoom()
  startControlsTimer()
}

const handleClose = () => {
  if (controlsTimer) {
    clearTimeout(controlsTimer)
  }
  controlsHidden.value = false
}

const toggleControls = () => {
  controlsHidden.value = !controlsHidden.value
  if (!controlsHidden.value) {
    startControlsTimer()
  } else if (controlsTimer) {
    clearTimeout(controlsTimer)
  }
}

const startControlsTimer = () => {
  if (controlsTimer) {
    clearTimeout(controlsTimer)
  }
  controlsTimer = setTimeout(() => {
    if (visible.value && !controlsHidden.value) {
      controlsHidden.value = true
    }
  }, 3000)
}

const zoomIn = () => {
  scale.value = Math.min(scale.value + 0.25, 3)
  applyTransform()
}

const zoomOut = () => {
  scale.value = Math.max(scale.value - 0.25, 0.5)
  applyTransform()
}

const resetZoom = () => {
  scale.value = 1
  rotation.value = 0
  applyTransform()
}

const rotateLeft = () => {
  rotation.value = (rotation.value - 90) % 360
  applyTransform()
}

const rotateRight = () => {
  rotation.value = (rotation.value + 90) % 360
  applyTransform()
}

const applyTransform = () => {
  const img = document.querySelector('.viewer-image .el-image__inner')
  if (img) {
    img.style.transform = `scale(${scale.value}) rotate(${rotation.value}deg)`
    img.style.transition = 'transform 0.2s ease'
  }
}

const formatSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// 监听弹窗关闭，重置状态
watch(visible, (newVal) => {
  if (!newVal) {
    scale.value = 1
    rotation.value = 0
    if (controlsTimer) {
      clearTimeout(controlsTimer)
    }
  }
})
</script>

<style scoped>
.image-thumbnail {
  position: relative;
  cursor: pointer;
  overflow: hidden;
  border-radius: 8px;
  transition: all 0.2s;
}

.image-thumbnail:hover .image-overlay {
  opacity: 1;
}

.thumbnail-img {
  width: 100%;
  height: 100%;
  transition: transform 0.2s;
}

.image-thumbnail:hover .thumbnail-img {
  transform: scale(1.05);
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
  color: #fff;
  font-size: 32px;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 32px;
}

/* 弹窗样式 */
:deep(.image-viewer-dialog .el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

:deep(.image-viewer-dialog .el-dialog__body) {
  padding: 0;
}

.viewer-container {
  position: relative;
  background: #1a1a1a;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  width: 100%;
}

.viewer-image {
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.viewer-image .el-image__inner) {
  transition: transform 0.2s ease;
  cursor: grab;
}

:deep(.viewer-image .el-image__inner:active) {
  cursor: grabbing;
}

.viewer-controls {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  transition: opacity 0.3s;
}

.viewer-controls.hidden {
  opacity: 0;
}

.control-buttons {
  display: flex;
  gap: 8px;
  background: rgba(0, 0, 0, 0.6);
  padding: 8px 16px;
  border-radius: 40px;
  backdrop-filter: blur(8px);
}

.control-buttons .el-button {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: #fff;
}

.control-buttons .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
}

.scale-text {
  font-size: 12px;
  min-width: 45px;
}

.viewer-footer {
  padding: 12px 16px;
  background: #f5f5f5;
  border-top: 1px solid #eee;
}

.image-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #666;
}
</style>