<template>
  <div class="friend-links-page">
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
      <div class="nav-right">
        <input v-model="searchQuery" @keyup.enter="searchArticles" placeholder="搜索文章..." class="search-input">
        <button @click="searchArticles" class="search-button">搜索</button>
      </div>
    </div>
    <div class="options">
      <span :class="{ active: selectedOption === 'neighbors' }" @click="selectOption('neighbors')">邻居</span>
      <span :class="{ active: selectedOption === 'apply' }" @click="selectOption('apply')">申请</span>
    </div>
    <div v-if="selectedOption === 'neighbors'" class="neighbors">
      <div v-for="neighbor in neighbors" :key="neighbor.email" class="neighbor-card" @click="goToWebsite(neighbor.website)">
        <p>{{ neighbor.name }}</p>
        <p>{{ neighbor.email }}</p>
      </div>
    </div>
    <div v-if="selectedOption === 'apply'" class="apply-form">
      <h1>申请友链</h1>
      <form @submit.prevent="submitApplication">
        <div>
          <label for="name">名字:</label>
          <input type="text" id="name" v-model="application.name" required>
        </div>
        <div>
          <label for="email">邮箱:</label>
          <input type="email" id="email" v-model="application.email" required>
        </div>
        <div>
          <label for="website">网址:</label>
          <input type="url" id="website" v-model="application.website" required>
        </div>
        <button type="submit">提交</button>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      profileImage: new URL('@/assets/avatar.png', import.meta.url).href,
      username: 'mkitsdts',
      searchQuery: '',
      selectedOption: 'neighbors', // 默认选中的选项
      neighbors: [], // 存储邻居数据
      application: {
        name: '',
        email: '',
        website: ''
      }
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
    async searchArticles() {
      try {
        const response = await fetch(`http://localhost:8080/search?keyword=${this.searchQuery}&page=1`);
        if (!response.ok) {
          throw new Error('网络响应失败');
        }
        const data = await response.json();
        this.$router.push({ name: 'Home', query: { articles: data.articles, maxPage: data.maxpage, currentPage: 1 } });
      } catch (error) {
        console.error('搜索文章数据失败:', error);
      }
    },
    selectOption(option) {
      this.selectedOption = option;
      if (option === 'neighbors') {
        this.fetchNeighbors();
      }
    },
    async fetchNeighbors() {
      try {
        const response = await fetch('http://localhost:8080/friend');
        if (!response.ok) {
          throw new Error('网络响应失败');
        }
        const data = await response.json();
        this.neighbors = data.neighbors;
      } catch (error) {
        console.error('获取邻居数据失败:', error);
      }
    },
    goToWebsite(website) {
      window.open(website, '_blank');
    },
    async submitApplication() {
      const applicationData = {
        ...this.application,
        date: new Date().toISOString()
      };
      try {
        const response = await fetch('http://localhost:8080/friend/apply', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(applicationData)
        });
        console.log(JSON.stringify(applicationData))
        if (!response.ok) {
          throw new Error('提交申请失败');
        }
        alert('申请提交成功');
        this.application.name = '';
        this.application.email = '';
        this.application.website = '';
      } catch (error) {
        console.error('提交申请失败:', error);
      }
    }
  }
};
</script>

<style scoped>
.friend-links-page {
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

.search-input {
  border: none;
  border-bottom: 1px solid #ccc;
  padding: 5px;
  font-size: 16px;
  background: none;
}

.search-button {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
}

.options {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.options span {
  cursor: pointer;
  font-size: 16px;
}

.options .active {
  font-weight: bold;
  color: #ff69b4; /* 高亮颜色 */
}

.neighbors {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.neighbor-card {
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 5px;
  width: calc(100% / 6 - 10px);
  text-align: center;
  cursor: pointer;
}

.neighbor-card:hover {
  background-color: #f0f0f0;
}

.apply-form {
  width: 50vw;
  border: 1px solid #ccc;
  padding: 20px;
  border-radius: 10px;
  background-color: #f9f9f9;
}

.apply-form form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.apply-form label {
  font-weight: bold;
}

.apply-form input {
  padding: 5px;
  font-size: 16px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.apply-form button {
  padding: 10px;
  font-size: 16px;
  background-color: #ff69b4;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.apply-form button:hover {
  background-color: #ff1493;
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

.friend-links-page ul {
  list-style: none;
  padding: 0;
}

.friend-links-page li {
  margin-bottom: 10px;
}
</style>