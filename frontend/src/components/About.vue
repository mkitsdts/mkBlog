<template>
  <div class="about-page">
    <div class="profile">
      <div class="profile-background">
        <img :src="profileImage" alt="头像" class="avatar">
        <p class="username">{{ username }}</p>
      </div>
    </div>
    <div class="nav-bar">
      <div class="nav-left">
        <button @click="goHome" class="nav-button">首页</button>
        <button @click="goFriendLinks" class="nav-button">友链</button>
        <button @click="goAbout" class="nav-button">关于</button>
      </div>
    </div>
    <h1>自我介绍</h1>
    <p>本人鼠鼠，长期生活在阴沟里。</p>
    <p>鼠鼠知道没人会联系 mkitsdts@outlook.com ， 因为鼠鼠穿越50年后依旧没有收到任何邮箱...</p>
  </div>
</template>

<script>
export default {
  data() {
    return {
      profileImage: new URL('@/assets/avatar.png', import.meta.url).href,
      username: 'mkitsdts',
      searchQuery: ''
    };
  },
  methods: {
    goHome() {
      this.$router.push({ name: 'Home' });
    },
    goAbout() {
      this.$router.push({ name: 'About' });
    },
    goFriendLinks() {
      this.$router.push({ name: 'FriendLinks' });
    },
  }
};
</script>

<style scoped>
.about-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.profile {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  margin-bottom: 20px;
}

.profile-background {
  text-align: center;
  border: 1px solid #ccc;
  padding: 20px;
  background-image: url('@/assets/background.jpg'); /* 替换为你的背景图片路径 */
  background-size: cover;
  background-position: center;
  border-radius: 10px;
  width: 50vw;
  height: 6.18vw;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.nav-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 50vw;
  margin-bottom: 20px;
  border: 1px solid #ccc;
  padding: 10px;
  background-color: #ffe4e1; /* 浅粉红色背景 */
  border-radius: 10px;
}

.nav-left {
  display: flex;
  gap: 10px;
}

.nav-right {
  display: flex;
  gap: 10px;
}

.nav-button {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
}

.nav-button:hover {
  text-decoration: underline;
}

.search-button {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
}

.username {
  margin-top: 10px;
  font-size: 18px;
  font-weight: bold;
  color: white;
}
</style>