<template>
  <div class="friends">
    <el-tabs v-model="activeName">
      <el-tab-pane label="My Friends" name="friendsList">
        <el-row :gutter="20">
          <el-col :span="6" v-for="friend in friends" :key="friend.id">
            <el-card class="friend-card">
              <a :href="friend.url" target="_blank">
                <el-avatar :src="friend.avatar"></el-avatar>
                <h4>{{ friend.name }}</h4>
              </a>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>
      <el-tab-pane label="Apply for Friendship" name="applyFriend">
        <el-form :model="form" label-width="120px">
          <el-form-item label="Your Name">
            <el-input v-model="form.name"></el-input>
          </el-form-item>
          <el-form-item label="Your Website">
            <el-input v-model="form.url"></el-input>
          </el-form-item>
          <el-form-item label="Your Avatar URL">
            <el-input v-model="form.avatar"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="onSubmit">Apply</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { ref, onMounted, reactive } from 'vue';
import api from '@/api';
import { ElMessage } from 'element-plus';

export default {
  name: 'Friends',
  setup() {
    const activeName = ref('friendsList');
    const friends = ref([]);
    const form = reactive({
      name: '',
      url: '',
      avatar: '',
    });

    const fetchFriends = async () => {
      try {
        const response = await api.getFriends();
        friends.value = response.data;
      } catch (error) {
        console.error('Failed to fetch friends:', error);
      }
    };

    const onSubmit = async () => {
      try {
        await api.applyFriend(form);
        ElMessage.success('Application submitted successfully!');
        form.name = '';
        form.url = '';
        form.avatar = '';
      } catch (error) {
        ElMessage.error('Failed to submit application.');
      }
    };

    onMounted(() => {
      fetchFriends();
    });

    return {
      activeName,
      friends,
      form,
      onSubmit,
    };
  },
};
</script>

<style scoped>
.friends {
  padding: 20px;
  background-color: rgba(255, 255, 255, 0.7);
  border-radius: 10px;
}
.friend-card {
  text-align: center;
  background-color: rgba(255, 255, 255, 0.8);
}
.friend-card a {
  text-decoration: none;
  color: inherit;
}
</style>
