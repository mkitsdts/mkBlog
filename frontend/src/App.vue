<template>
  <div id="app">
    <el-menu :default-active="$route.path" class="el-menu-demo" mode="horizontal" router>
      <el-menu-item index="/">Home</el-menu-item>
      <el-menu-item index="/friends">Friends</el-menu-item>
    </el-menu>
    <router-view/>
  </div>
</template>

<script>
export default {
  name: 'App',
};
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
</style>

