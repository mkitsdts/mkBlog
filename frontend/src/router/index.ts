import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Friends from '../views/Friends.vue'
import Article from '../views/Article.vue'
import About from '../views/About.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/friends',
      name: 'friends',
      component: Friends
    },
    {
      path: '/about',
      name: 'about',
      component: About
    },
    {
      path: '/article/:title',
      name: 'article-detail',
      component: Article
    }
  ]
})

export default router
