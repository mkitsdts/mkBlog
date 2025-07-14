<template>
  <div class="home">
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
    <div class="main-content">
      <div class="articles">
        <h1>文章</h1>
        <div class="article-list">
          <div v-for="article in filteredArticles" :key="article.title" class="article-summary">
            <h2 @click="goToArticle(article.title)">{{ article.title }}</h2>
            <p>更新时间: {{ article.updateAt }}</p>
            <p>概要： {{ article.summary }}</p>
          </div>
        </div>
        <div class="pagination">
          <button @click="prevPage" :disabled="currentPage === 1">上一页</button>
          <span>第 {{ currentPage }} 页</span>
          <button @click="nextPage" :disabled="maxPage <= currentPage">下一页</button>
        </div>
      </div>
      <div class="sidebar">
        <div class="contact-links">
          <ul>
            <li>
              <a :href="links.github" target="_blank">
                <img :src="githubImage" alt="github" class="github"></a>
            </li>
          </ul>
        </div>
        <div class="archive">
          <h2>文章归档</h2>
          <ul>
            <li v-for="category in filteredCategories">
              <h3 @click="filterByCategory(category)">{{ category }}</h3>
              <ul>
                <li v-for="article in category.articles" :key="article.title">
                  <a @click="goToArticle(article.title)">{{ article.title }}</a>
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      profileImage: new URL('@/assets/avatar.png', import.meta.url).href,
      githubImage: new URL('@/assets/github-mark.svg', import.meta.url).href,
      username: 'mkitsdts',
      searchQuery: '',
      links: {
        github: 'https://github.com/mkitsdts',
      },
      articles: [],
      categories: new Set(), // 添加文章分类数据
      currentPage: 1,
      maxPage: 1,
      selectedCategory: null // 添加选中的分类
    };
  },
  computed: {
    filteredArticles() {
      if (this.selectedCategory != null) {
        return this.articles.filter(article => article.category === this.selectedCategory);
      }
      return this.articles;
    },
    filteredCategories(){
      for(let i = 0; i < this.articles.length; i++){
        this.categories.add(this.articles[i].category);
        console.log(this.articles[i].category);
      }
      return this.categories;
    }
  },
  created() {
    this.fetchArticles();
  },
  methods: {
    async fetchArticles(page = 1) {
      try {
        const response = await fetch(`http://localhost:8080/home?page=${page}`);
        if (!response.ok) {
          throw new Error('网络响应失败');
        }
        const data = await response.json();
        this.articles = data.articles;
        this.maxPage = data.maxpage;
        this.currentPage = page;
      } catch (error) {
        console.error('获取文章数据失败:', error);
      }
    },
    async searchArticles() {
      try {
        const response = await fetch(`http://localhost:8080/search?keyword=${this.searchQuery}&page=1`);
        if (!response.ok) {
          throw new Error('网络响应失败');
        }
        const data = await response.json();
        this.articles = data.articles;
        this.maxPage = data.maxpage;
        this.currentPage = 1;
      } catch (error) {
        console.error('搜索文章数据失败:', error);
      }
    },
    goToArticle(title) {
      this.$router.push({ name: 'Article', params: { title } });
    },
    goHome() {
      this.$router.push({ name: 'Home' });
    },
    goAbout() {
      this.$router.push({ name: 'About' });
    },
    goFriendLinks() {
      this.$router.push({ name: 'FriendLinks' });
    },
    prevPage() {
      if (this.currentPage > 1) {
        this.fetchArticles(this.currentPage - 1);
      }
    },
    nextPage() {
      if (this.maxPage > this.currentPage) {
        this.fetchArticles(this.currentPage + 1);
      }
    },
    filterByCategory(category) {
      this.selectedCategory = category;
    }
  }
};
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  min-height: 100vh;
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

.main-content {
  display: flex;
  width: 100%;
  max-width: 1200px;
}

.articles {
  flex: 3;
  margin-right: 20px;
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 10px;
  background-color: #f9f9f9;
}

.article-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.article-summary {
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 10px;
  background-color: #fff;
}

.article-summary h2 {
  cursor: pointer;
  color: blue;
  text-decoration: underline;
}

.sidebar {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.contact-links, .archive {
  border: 1px solid #ccc;
  padding: 10px;
  border-radius: 10px;
  background-color: #f9f9f9;
}

.contact-links ul, .archive ul {
  list-style: none;
  padding: 0;
}

.contact-links li, .archive li {
  margin-bottom: 10px;
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

.search-bar {
  margin-bottom: 20px;
}

.search-bar input {
  padding: 5px;
  font-size: 16px;
}

.search-bar button {
  padding: 5px 10px;
  font-size: 16px;
}

.content {
  width: 100%;
  max-width: 800px;
  margin-bottom: 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 20px;
}

.pagination button {
  padding: 5px 10px;
  margin: 0 5px;
}

.friend-links {
  position: fixed;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  width: 200px;
  text-align: center;
}

.friend-links ul {
  list-style: none;
  padding: 0;
}

.friend-links li {
  margin-bottom: 10px;
}

.github {
  width: 35px;
  height: 35px;
}
</style>