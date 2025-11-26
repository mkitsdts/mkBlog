<template>
  <div class="home">
    <el-container :class="['home-container', { 'is-mobile': isMobile }]">
      <el-aside
        v-if="!isMobile"
        width="220px"
        class="sidebar"
      >
        <div class="user-profile">
          <el-avatar :size="100" :src="avatarUrl"></el-avatar>
          <h3>{{ signature }}</h3>
        </div>
        <div class="category-panel">
          <h4 class="cat-title">分类</h4>
          <el-tag
            :type="selectedCategories.length === 0 ? 'success' : 'info'"
            size="small"
            class="cat-item"
            effect="light"
            @click="clearCategories"
          >全部</el-tag>
          <el-tag
            v-for="c in categories"
            :key="c"
            :type="isActive(c) ? 'success' : 'info'"
            size="small"
            class="cat-item"
            effect="light"
            @click="toggleCategory(c)"
          >
            {{ c || '未分类' }}
          </el-tag>
        </div>
      </el-aside>
      <el-main :class="{ 'mobile-main': isMobile }">
        <div class="toolbar">
          <el-button
            v-if="isMobile"
            class="sidebar-toggle"
            type="primary"
            text
            @click="drawerVisible = true"
            aria-label="展开头像与分类"
          >
            <el-icon class="hamburger-icon"><Menu /></el-icon>
            <span class="sidebar-toggle-text">分类</span>
          </el-button>
          <el-input
            v-model="keyword"
            placeholder="搜索标题或摘要..."
            clearable
            class="search-input"
            @keyup.enter="doSearch"
            @clear="clearSearch"
          >
            <template #append>
              <el-button type="primary" @click="doSearch">搜索</el-button>
            </template>
          </el-input>
        </div>
        <div class="blog-list">
          <el-card
            v-for="article in articles" :key="article.title"
            class="blog-card"
            shadow="hover"
            @click="goDetail(article.title)"
          >
            <div class="card-inner">
              <h2 class="blog-title">
                <span class="title-text">{{ article.title }}</span>
                <el-tag size="small" type="success" v-if="article.category" class="cat-tag">{{ article.category }}</el-tag>
              </h2>
              <p class="blog-summary" v-if="article.summary">{{ article.summary }}</p>
              <div class="meta-row">
                <span class="meta-item" v-if="article.updateAt">{{ formatDate(article.updateAt) }}</span>
                <span class="meta-sep" v-if="article.tags && article.updateAt">•</span>
                <span class="meta-tags" v-if="article.tags">
                  <el-tag
                    v-for="t in splitTags(article.tags)"
                    :key="t"
                    size="small"
                    effect="plain"
                    class="tag-item"
                  >{{ t }}</el-tag>
                </span>
              </div>
            </div>
          </el-card>
        </div>
        <el-pagination
          @current-change="handlePageChange"
          :current-page="currentPage"
          :page-size="pageSize"
          layout="prev, pager, next"
          :total="total">
        </el-pagination>
      </el-main>
    </el-container>
    <el-drawer
      v-if="isMobile"
      v-model="drawerVisible"
      :with-header="false"
      direction="ltr"
      size="260px"
      class="sidebar-drawer"
    >
      <div class="sidebar">
        <div class="user-profile">
          <el-avatar :size="100" :src="avatarUrl"></el-avatar>
          <h3>{{ signature }}</h3>
        </div>
        <div class="category-panel">
          <h4 class="cat-title">分类</h4>
          <el-tag
            :type="selectedCategories.length === 0 ? 'success' : 'info'"
            size="small"
            class="cat-item"
            effect="light"
            @click="clearCategories"
          >全部</el-tag>
          <el-tag
            v-for="c in categories"
            :key="c"
            :type="isActive(c) ? 'success' : 'info'"
            size="small"
            class="cat-item"
            effect="light"
            @click="toggleCategory(c)"
          >
            {{ c || '未分类' }}
          </el-tag>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import api from '@/api';
import { loadConfig } from '@/config';
import { Menu } from '@element-plus/icons-vue';

export default {
  name: 'Home',
  setup() {
    const avatarUrl = ref('');
    const signature = ref('');
    const articles = ref([]);
    const keyword = ref('');
    const categories = ref([]);
    const selectedCategories = ref([]);
    const currentPage = ref(1);
    const pageSize = ref(18);
    const total = ref(0);
    const isMobile = ref(false);
    const drawerVisible = ref(false);

    const updateIsMobile = () => {
      isMobile.value = window.innerWidth <= 768;
      if (!isMobile.value) {
        // if desktop, close drawer
        drawerVisible.value = false;
      } else {
        pageSize.value = 10;
      }
    };

    const fetchArticles = async (page) => {
      try {
        const kw = keyword.value.trim();
        let response;
        if (kw) {
          // 全文搜索接口不支持分类过滤，清空分类以避免误导
          response = await api.searchArticles(kw, page, pageSize.value);
        } else {
          response = await api.getArticles(page, pageSize.value, selectedCategories.value);
        }
        articles.value = response.data.articles;
        total.value = response.data.total;
      } catch (error) {
        console.error('Failed to fetch articles:', error);
      }
    };

    const fetchCategories = async () => {
      try {
        const res = await api.getCategories();
        categories.value = res.data.categories || [];
      } catch(e) { console.error('Failed to fetch categories', e); }
    };

    const toggleCategory = (c) => {
      if (keyword.value.trim()) {
        // 有关键词时，分类筛选无效，清空关键字以回到分类模式
        keyword.value = ''
      }
      const idx = selectedCategories.value.indexOf(c);
      if (idx === -1) {
        selectedCategories.value.push(c);
      } else {
        selectedCategories.value.splice(idx,1);
      }
      currentPage.value = 1;
      fetchArticles(currentPage.value);
      if (isMobile.value) {
        drawerVisible.value = false;
      }
    };
    const clearCategories = () => {
      if (keyword.value.trim()) keyword.value = '';
      selectedCategories.value = [];
      fetchArticles(1);
      if (isMobile.value) {
        drawerVisible.value = false;
      }
    };
    const isActive = (c) => selectedCategories.value.includes(c);

    const router = useRouter();
    const route = useRoute();

    const parseCategoriesFromQuery = (raw) => {
      if (!raw) return [];
      const values = Array.isArray(raw) ? raw : [raw];
      return values
        .flatMap((item) => String(item).split(','))
        .map((item) => item.trim())
        .filter(Boolean);
    };

    const applyRouteState = () => {
      keyword.value = typeof route.query.keyword === 'string' ? route.query.keyword : '';
      selectedCategories.value = parseCategoriesFromQuery(route.query.categories);
      const parsedPage = Number(route.query.page);
      currentPage.value = Number.isFinite(parsedPage) && parsedPage > 0 ? Math.floor(parsedPage) : 1;
    };

    applyRouteState();

    const buildHomeQuery = () => {
      const query = {};
      const kw = keyword.value.trim();
      if (kw) query.keyword = kw;
      if (selectedCategories.value.length) query.categories = selectedCategories.value.join(',');
      if (currentPage.value > 1) query.page = String(currentPage.value);
      return query;
    };

    // 防抖：避免在每次输入时立即修改路由（会导致浏览器重复请求 favicon 等资源）
    let replaceTimer = null;
    const scheduleReplace = () => {
      if (replaceTimer) clearTimeout(replaceTimer);
      replaceTimer = setTimeout(() => {
        router.replace({ query: buildHomeQuery() });
        replaceTimer = null;
      }, 300);
    };

    // 只对关键字输入使用防抖，分页和分类变更立即更新路由
    watch(keyword, () => {
      scheduleReplace();
    });

    watch([currentPage, selectedCategories], () => {
      // 分页或分类变化时立即同步路由
      if (replaceTimer) {
        clearTimeout(replaceTimer);
        replaceTimer = null;
      }
      router.replace({ query: buildHomeQuery() });
    }, { deep: true });

    const goDetail = (title) => {
      router.push(`/article/${encodeURIComponent(title)}`)
    }
    const splitTags = (tags) => (tags || '').split(/[,;，\s]+/).filter(Boolean).slice(0,5);
    const formatDate = (dt) => {
      if (!dt) return ''
      try {
        const d = new Date(dt) // 支持带Z/时区的UTC字符串
        if (isNaN(d.getTime())) {
          // 兜底：退回旧逻辑（仅替换T并截断）
          return String(dt).replace('T',' ').substring(0,19)
        }
        const pad = (n) => String(n).padStart(2,'0')
        return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
      } catch {
        return String(dt).replace('T',' ').substring(0,19)
      }
    }

    const handlePageChange = (page) => {
      currentPage.value = page;
      fetchArticles(page);
    };

    const doSearch = () => {
      keyword.value = keyword.value.trim();
      currentPage.value = 1;
      fetchArticles(1);
    };
    const clearSearch = () => {
      keyword.value = '';
      currentPage.value = 1;
      fetchArticles(1);
    };

    onMounted(async () => {
      updateIsMobile();
      window.addEventListener('resize', updateIsMobile);
      try {
        const conf = await loadConfig();
        signature.value = conf.signature || '鼠鼠很懒，什么都没有留下';
        try {
          avatarUrl.value = new URL(`../assets/${conf.avatarPath}`, import.meta.url).href;
        } catch { avatarUrl.value = ''; }
      } catch (e) { console.error('load config failed', e); }
      fetchCategories();
      fetchArticles(currentPage.value);
    });

    onUnmounted(() => {
      window.removeEventListener('resize', updateIsMobile);
      if (replaceTimer) {
        clearTimeout(replaceTimer);
        replaceTimer = null;
      }
    });

    return {
      avatarUrl,
      signature,
      articles,
      keyword,
      currentPage,
      pageSize,
      total,
      handlePageChange,
      categories,
      selectedCategories,
      toggleCategory,
      clearCategories,
      isActive,
      goDetail,
      splitTags,
      formatDate,
      doSearch,
      clearSearch,
      isMobile,
      drawerVisible,
      Menu,
    };
  },
};
</script>

<style scoped>
.home {
  padding: 20px;
}
.home-container {
  min-height: calc(100vh - 80px);
}
.home-container.is-mobile {
  flex-direction: column;
}
.sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.user-profile {
  text-align: center;
  background-color: rgba(255, 255, 255, 0.7);
  padding: 20px;
  border-radius: 10px;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: flex-end;
  margin-bottom: 12px;
}
.sidebar-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px 6px;
  border-radius: 6px;
}
.sidebar-toggle .hamburger-icon {
  font-size: 20px;
  margin-right: 4px;
}
.sidebar-toggle-text {
  font-size: 14px;
  color: var(--el-color-primary);
  font-weight: 600;
}
.sidebar-toggle .el-icon {
  font-size: 20px;
}
.search-input {
  max-width: 360px;
}
.mobile-main {
  width: 100%;
}
.sidebar-drawer {
  padding: 0 12px 24px;
}
.sidebar-drawer .sidebar {
  gap: 16px;
}
.category-panel { margin-top:20px; background:rgba(255,255,255,.7); padding:12px 14px; border-radius:10px; }
.cat-title { margin:0 0 8px; font-size:14px; color:#333; }
.cat-item { margin: 4px 6px 4px 0; cursor:pointer; user-select:none; }
.blog-card {
  margin-bottom: 22px;
  background: linear-gradient(135deg, rgba(255,255,255,.92), rgba(255,255,255,0.85));
  border: 1px solid rgba(0,0,0,0.04);
  border-radius: 14px;
  transition: all .28s cubic-bezier(.4,0,.2,1);
  position: relative;
  overflow: hidden;
}
.blog-card::before { content:""; position:absolute; inset:0; background:linear-gradient(120deg, rgba(99,147,255,.15), rgba(140,223,255,.12), rgba(255,255,255,0)); opacity:0; transition:opacity .4s; }
.blog-card:hover { transform: translateY(-4px); box-shadow: 0 10px 28px -6px rgba(0,0,0,.12), 0 4px 10px -2px rgba(0,0,0,.06); }
.blog-card:hover::before { opacity:1; }
.blog-card .el-card__body { padding: 18px 22px 20px; }
.card-inner { display:flex; flex-direction:column; gap:8px; }
.blog-title { margin:0; font-size: 1.28rem; font-weight:600; line-height:1.35; display:flex; align-items:center; gap:10px; letter-spacing:.5px; }
.blog-title .title-text { background: linear-gradient(90deg,#2d3436,#0984e3); -webkit-background-clip:text; background-clip:text; color:transparent; }
.cat-tag { line-height:1; }
.blog-summary { margin:0; color:#4a4f55; font-size: .92rem; line-height:1.55; max-width:100%; display:-webkit-box; -webkit-line-clamp:3; line-clamp:3; -webkit-box-orient:vertical; overflow:hidden; position:relative; }
.blog-summary::after { content:""; position:absolute; bottom:0; left:0; right:0; height:1.2em; background:linear-gradient(to bottom, rgba(255,255,255,0), rgba(255,255,255,1)); }
.meta-row { display:flex; flex-wrap:wrap; align-items:center; gap:6px; font-size:11.5px; color:#5f6b76; }
.meta-item { font-variant-numeric: tabular-nums; }
.meta-sep { opacity:.6; }
.tag-item { margin-right:4px; }
.title-link { text-decoration:none; }
.blog-card:active { transform:translateY(-1px) scale(.995); }
.about-section {
  margin-top: 20px;
  padding: 20px;
  background-color: rgba(255, 255, 255, 0.7);
  border-radius: 10px;
}
.about-section h2 {
  margin-bottom: 10px;
  font-size: 1.5rem;
  color: #333;
}
.about-section p {
  font-size: 1rem;
  color: #555;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .home {
    padding: 12px;
  }
  .toolbar {
    justify-content: space-between;
  }
  .search-input {
    max-width: none;
    flex: 1;
  }
  .blog-card {
    margin-bottom: 16px;
  }
  .user-profile {
    padding: 16px;
  }
  .category-panel {
    margin-top: 12px;
  }
}
</style>
