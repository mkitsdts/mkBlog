<template>
  <div id="app" :style="appStyle">
    <el-menu :default-active="$route.path" class="el-menu-demo" mode="horizontal" router>
      <el-menu-item index="/">Home</el-menu-item>
      <el-menu-item index="/friends">Friends</el-menu-item>
      <el-menu-item index="/about">About</el-menu-item>
    </el-menu>
    <router-view/>
    <footer v-if="icp" class="icp-footer">
      <a href="https://beian.miit.gov.cn/" target="_blank" rel="noopener noreferrer">{{ icp }}</a>
    </footer>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { imageExists, loadConfig, resolveSiteStaticAssetUrl } from '@/config'

export default {
  name: 'App',
  setup() {
    const icp = ref('')
    const bgUrl = ref('')

    const appStyle = computed(() => {
      if (!bgUrl.value) return {}
      return {
        backgroundImage: `url(${bgUrl.value})`
      }
    })

    onMounted(async () => {
      try {
        const conf = await loadConfig()
        icp.value = conf.icp || ''
        const preferredBg = resolveSiteStaticAssetUrl(conf.bgPicturePath)
        bgUrl.value = await imageExists(preferredBg) ? preferredBg : ''
      } catch (e) {
        console.error('Failed to load site config in App.vue', e)
      }
    })
    return { icp, appStyle }
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  background-size: cover;
  background-attachment: fixed;
  min-height: 100vh;
  transition: background-image 0.5s ease-in-out;
}
.el-menu-demo {
  background-color: rgba(255, 255, 255, 0.7) !important;
}
.el-menu-demo a {
  text-decoration: none;
}
.icp-footer {
  text-align: center;
  font-size: 12px;
  color: #888;
  padding: 12px 8px;
  background: rgba(255, 255, 255, 0.6);
}
.icp-footer a {
  color: #666;
  text-decoration: none;
}
.icp-footer a:hover { text-decoration: underline; }
</style>
