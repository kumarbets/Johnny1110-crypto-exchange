<template>
  <div class="user-profile" v-if="user">
    <div class="profile-info">
      <span class="welcome-text">Welcome, {{ user.username }}!</span>

    </div>
<!--    <button @click="handleLogout" class="logout-btn">Logout</button>-->
  </div>
  <div class="user-details" v-if="user">

      <div class="detail-item">VIP: {{ user.vip_level }}</div>
      <div class="detail-item">Maker Fee: {{ (user.maker_fee * 100).toFixed(3) }}%</div>
      <div class="detail-item">Taker Fee: {{ (user.taker_fee * 100).toFixed(3) }}%</div>

<!--    <button @click="handleLogout" class="logout-btn">Logout</button>-->
  </div>
</template>

<script>
import {authUtils} from "@/services/auth";

export default {
  name: 'UserProfile',
  data() {
    return {
      user: {}
    }
  },

  mounted() {
    // eslint-disable-next-line vue/no-mutating-props
    this.user = authUtils.getUserProfile()
    console.log("u_prof", this.user)
  },
  methods: {
    handleLogout() {
      this.$emit('logout')
    }
  }
}
</script>

<style scoped>
.user-profile {
  display: flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid #ff99ff;
  padding: 8px 12px;
  margin-bottom: 10px;
  border-radius: 3px;
  justify-content: center;

}

.profile-info {
  display: flex;
  flex-direction: column;
  gap: 5px;

}
.user-details{
  justify-content: end;
}
.welcome-text {
  color: #ff66cc;
  font-size: 24px;
  text-shadow: 5px 5px #330033;
  box-shadow: 0 0 144px #f4de5f, 0 0 20px #eae585;
}

.user-details {
  display: flex;
  gap: 15px;
}

.detail-item {
  font-size: 10px;
  color: #ffccff;
}

.logout-btn {
  background: #ff3366;
  border: 2px solid #ff99ff;
  padding: 4px 8px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 8px;
  color: #ffffff;
  transition: all 0.2s;
}

.logout-btn:hover {
  background: #cc0033;
  box-shadow: 0 0 5px #ff3366;
}
</style>