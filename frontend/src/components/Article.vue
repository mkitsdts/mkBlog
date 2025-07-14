<template>
  <div class="article">
    <h1>{{ article.title }}</h1>
    <p>作者: {{ article.author }}</p>
    <p>创建时间: {{ article.createAt }}</p>
    <p>更新时间: {{ article.updateAt }}</p>
    <div v-html="parsedContent"></div>
  </div>
</template>

<script>
import { marked } from 'marked';

export default {
  data() {
    return {
      article: {},
      parsedContent: ''
    };
  },
  created() {
    this.fetchArticleDetail();
  },
  methods: {
    async fetchArticleDetail() {
      const title = this.$route.params.title;
      try {
        const response = await fetch(`http://localhost:8080/articles/${title}`);
        if (!response.ok) {
          throw new Error('网络响应失败');
        }
        const data = await response.json();
        this.article = data;
        this.parsedContent = marked(data.content);
        console.log(this.parsedContent);
        console.log('解析完成');
        this.fetchImages();
      } catch (error) {
        console.error('获取文章详细内容失败:', error);
      }
    },
    async fetchImages() {
      const parser = new DOMParser();
      const doc = parser.parseFromString(this.parsedContent, 'text/html');
      const images = doc.querySelectorAll('img');
      for (const img of images) {
        const src = img.getAttribute('src');
        try {
          const response = await fetch(`http://localhost:8080/images/`+ this.$route.params.title + `/${src}`);
          if (!response.ok) {
            throw new Error('网络响应失败');
          }
          const blob = await response.blob();
          const url = URL.createObjectURL(blob);
          img.setAttribute('src', url);
        } catch (error) {
          console.error('获取图片失败:', error);
        }
      }
      this.parsedContent = doc.body.innerHTML;
    }
  }
};
</script>

<style scoped>
.article {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.article h1 {
  margin-bottom: 20px;
}

.article p {
  margin-bottom: 10px;
}

.article h2, .article h3, .article h4, .article h5, .article h6 {
  margin-top: 20px;
  margin-bottom: 10px;
}

.article ul, .article ol {
  margin-left: 20px;
}

.article blockquote {
  margin: 20px 0;
  padding: 10px 20px;
  background-color: #f9f9f9;
  border-left: 5px solid #ccc;
}

.article pre {
  background-color: #f5f5f5;
  padding: 10px;
  overflow: auto;
}

.article code {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 4px;
}
</style>