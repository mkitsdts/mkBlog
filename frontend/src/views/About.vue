<template>
	<div class="about">
		<el-card class="about-card" shadow="hover">
			<el-skeleton v-if="loading" :rows="5" animated />
			<template v-else>
				<h1 class="about-title">关于本站</h1>
				<p v-if="error" class="error-text">{{ error }}</p>
				<div v-else class="about-content">{{ aboutText }}</div>
			</template>
		</el-card>
	</div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { loadConfig } from '@/config'

export default {
	name: 'About',
	setup() {
		const aboutText = ref('')
		const loading = ref(true)
		const error = ref('')

		onMounted(async () => {
			try {
				const config = await loadConfig()
				aboutText.value = config.about || '鼠鼠很懒，什么都没有留下'
			} catch (e) {
				console.error('加载 config 失败', e)
				error.value = '暂时无法加载关于信息，请稍后再试。'
			} finally {
				loading.value = false
			}
		})

		return {
			aboutText,
			loading,
			error
		}
	}
}
</script>

<style scoped>
.about {
	padding: 24px;
	display: flex;
	justify-content: center;
}

.about-card {
	max-width: 780px;
	width: 100%;
	background: rgba(255, 255, 255, 0.78);
	border-radius: 16px;
	border: none;
}

.about-title {
	margin: 0 0 16px;
	font-size: 1.8rem;
	font-weight: 600;
	color: #1f2d3d;
	text-align: center;
}

.error-text {
	color: #f56c6c;
	text-align: center;
	margin: 24px 0;
}

.about-content {
	line-height: 1.8;
	font-size: 1.05rem;
	color: #34495e;
	white-space: pre-line;
}
</style>
