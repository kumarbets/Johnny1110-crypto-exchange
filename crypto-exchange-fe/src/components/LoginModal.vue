<template>
  <div class="modal-overlay" v-if="isVisible" @click="closeModal">
    <div class="login-container" @click.stop>
      <div class="title-bar">
        <span>CryptoEx Pixel - {{ isLoginMode ? 'Login' : 'Register' }}</span>
        <div>
          <button @click="closeModal(false)">X</button>
        </div>
      </div>

      <div class="login-form">
        <label for="username">Username:</label>
        <input
            type="text"
            id="username"
            v-model="username"
            placeholder="Enter username"
            @keyup.enter="handleSubmit"
        >

        <label for="password">Password:</label>
        <input
            type="password"
            id="password"
            v-model="password"
            placeholder="Enter password"
            @keyup.enter="handleSubmit"
        >

        <button @click="handleSubmit" :disabled="loading">
          {{ loading ? 'Processing...' : (isLoginMode ? 'Login' : 'Register') }}
        </button>

        <div class="mode-switch">
          <a href="#" @click.prevent="toggleMode">
            {{ isLoginMode ? 'Need an account? Register' : 'Have an account? Login' }}
          </a>
        </div>
      </div>

      <div class="cmd-window">
        <div v-for="(line, index) in cmdOutput" :key="index" v-html="line"></div>
        <div class="cursor-line">C:\CryptoEx> <span class="cursor">_</span></div>
      </div>


    </div>
  </div>
</template>

<script>
import {userAPI} from '@/services/apiService'
import {authUtils} from "@/services/auth";

export default {
  name: 'LoginModal',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      username: '',
      password: '',
      loading: false,
      isLoginMode: true,
      cmdOutput: ['C:\\CryptoEx> auth', 'Enter credentials to authenticate']
    }
  },
  computed: {
    isVisible() {
      return this.visible
    }
  },
  watch: {
    visible(newVal) {
      if (newVal) {
        this.resetForm()
      }
    }
  },
  methods: {
    async handleSubmit() {
      if (!this.username || !this.password) {
        this.addCmdOutput('Error: Username and password are required')
        return
      }

      this.loading = true

      try {
        let response
        if (this.isLoginMode) {
          response = await userAPI.login(this.username, this.password)
        } else {
          response = await userAPI.register(this.username, this.password)
        }

        if (response.data.code === '0000000') {
          if (this.isLoginMode) {
            // 登入成功，儲存 token
            authUtils.setAuthData(response.data.data.token, this.username)

            this.addCmdOutput('Login successful!')
            this.addCmdOutput(`Welcome back, ${this.username}!`)

            setTimeout(() => {
              this.$emit('login-success', {
                username: this.username,
                token: response.data.data.token
              })
              this.closeModal(true)
            }, 500)
          } else {
            // 註冊成功
            this.addCmdOutput('Registration successful!')
            this.addCmdOutput(`User ID: ${response.data.data.user_id}`)
            this.addCmdOutput('Please login with your credentials')

            setTimeout(() => {
              this.isLoginMode = true
            }, 500)
          }
        } else {
          throw new Error(response.data.message || 'Authentication failed')
        }
      } catch (error) {
        const errorMsg = error.response?.data?.message || error.message || 'Network error'
        this.addCmdOutput(`Error: ${errorMsg}`)
      } finally {
        this.loading = false
      }
    },

    toggleMode() {
      this.isLoginMode = !this.isLoginMode
      this.resetCmdOutput()
    },

    closeModal(flag) {
      this.$emit('close', flag);
    },

    resetForm() {
      this.username = ''
      this.password = ''
      this.loading = false
      this.resetCmdOutput()
    },

    resetCmdOutput() {
      this.cmdOutput = [
        'C:\\CryptoEx> auth',
        this.isLoginMode ? 'Enter credentials to authenticate' : 'Create new account'
      ]
    },

    addCmdOutput(message) {
      this.cmdOutput.push(`> ${message}`)
      if (this.cmdOutput.length > 8) {
        this.cmdOutput.shift()
      }
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap');

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.login-container {
  background: rgba(51, 0, 51, 0.95);
  border: 3px solid #ff99ff;
  width: 450px;
  padding: 15px;
  box-shadow: 0 0 20px #ff66cc, 0 0 40px #9900cc;
  border-radius: 5px;
  font-family: 'Press Start 2P', cursive;
  color: #ffffff;
}

.title-bar {
  background: linear-gradient(90deg, #ff33cc, #cc00ff);
  color: #ffffff;
  padding: 8px;
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 2px solid #ff99ff;
  text-shadow: 1px 1px 2px #330033;
  margin-bottom: 15px;
}

.title-bar button {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 3px 10px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  text-shadow: 1px 1px #330033;
  transition: all 0.2s;
}

.title-bar button:hover {
  background: #cc00ff;
  box-shadow: 0 0 5px #ff66cc;
}

.login-form {
  padding: 15px;
}

.login-form label {
  display: block;
  font-size: 12px;
  margin: 10px 0 5px;
  color: #ffccff;
  text-shadow: 1px 1px #330033;
}

.login-form input {
  width: calc(100% - 14px);
  padding: 8px;
  border: 2px solid #ff99ff;
  background: rgba(255, 255, 255, 0.1);
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  box-shadow: inset 0 0 5px #9900cc;
  margin-bottom: 10px;
}

.login-form input:focus {
  outline: none;
  box-shadow: inset 0 0 5px #9900cc, 0 0 5px #ff66cc;
}

.login-form button {
  background: #ff66cc;
  border: 2px solid #ff99ff;
  padding: 10px 12px;
  cursor: pointer;
  font-family: 'Press Start 2P', cursive;
  font-size: 10px;
  color: #ffffff;
  width: 100%;
  margin-top: 10px;
  box-shadow: 0 0 5px #ff66cc;
  text-shadow: 1px 1px #330033;
  transition: all 0.2s;
}

.login-form button:hover:not(:disabled) {
  background: #cc00ff;
  box-shadow: 0 0 8px #ff66cc;
}

.login-form button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.mode-switch {
  text-align: center;
  margin-top: 15px;
}

.mode-switch a {
  color: #ffccff;
  text-decoration: none;
  font-size: 8px;
  cursor: pointer;
}

.mode-switch a:hover {
  color: #ff66cc;
  text-shadow: 0 0 3px #ff66cc;
}

.cmd-window {
  background: #1a001a;
  color: #ff66cc;
  font-family: 'Courier New', monospace;
  padding: 12px;
  margin: 15px 0;
  border: 2px solid #ff99ff;
  height: 120px;
  overflow-y: auto;
  box-shadow: inset 0 0 10px #9900cc;
  font-size: 11px;
}

.cursor-line {
  color: #ff66cc;
}

.cursor {
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% {
    opacity: 1;
  }
  51%, 100% {
    opacity: 0;
  }
}

.footer {
  text-align: center;
  font-size: 8px;
  color: #ffccff;
  margin-top: 10px;
  text-shadow: 1px 1px #330033;
}
</style>