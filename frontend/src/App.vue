<template>
  <div id="app">
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
import { ref, onMounted } from 'vue'
import { loadConfig } from '@/config'

export default {
  name: 'App',
  setup() {
    const icp = ref('')
    onMounted(async () => {
      try {
        const conf = await loadConfig()
        icp.value = conf.icp || ''
      } catch {}
    })
    return { icp }
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  background-image: url('https://source.unsplash.com/random/1920x1080'); /* Add a random background image */
  background-size: cover;
  background-attachment: fixed;
  min-height: 100vh;
}
.el-menu-demo {
  background-color: rgba(255, 255, 255, 0.7) !important; /* Make menu transparent */
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

