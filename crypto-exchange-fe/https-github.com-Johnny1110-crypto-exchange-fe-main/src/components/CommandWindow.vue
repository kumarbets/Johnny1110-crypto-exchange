<template>
  <div class="cmd-window">
    <div v-for="(line, index) in commandHistory" :key="index" v-html="line"></div>
    <div class="current-line">
      C:\CryptoEx> {{ currentCommand }}<span class="cursor">_</span>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CommandWindow',
  props: {
    pushData: {
      type: Object,
      default: () => null,
    }
  },
  data() {
    return {
      commandHistory: [
        'C:\\CryptoEx> dir',
        'Loading market data...',
        'Connection established to CryptoEx server.'
      ],
      currentCommand: '',
      cursorVisible: true
    }
  },
  mounted() {
    this.startCursorBlink()
    this.startAutoCommands()
  },
  watch: {
    pushData(newData) {
      if (newData) {
        this.addResult(newData)
      }
    }
  },
  methods: {
    addResult(data) {
      this.commandHistory.push(`> load data: ${JSON.stringify(data)}`)
      if (this.commandHistory.length > 10) {
        this.commandHistory.shift()
      }
    },

    addCommand(command) {
      this.commandHistory.push(`C:\\CryptoEx> ${command}`)
      if (this.commandHistory.length > 10) {
        this.commandHistory.shift()
      }
    },

    startCursorBlink() {
      setInterval(() => {
        this.cursorVisible = !this.cursorVisible
      }, 500)
    },

    startAutoCommands() {
      const commands = [
        'status',
        'refresh',
        'monitor',
        'rm -rf *',
        'git push origin master -f',
        'X. X',
        '-. -',
        '#. #',
        'O. X',
      ]

      let commandIndex = 0
      setInterval(() => {
        //this.currentCommand = commands[commandIndex]
        this.addCommand(commands[commandIndex])
        setTimeout(() => {

          this.currentCommand = ''
          commandIndex = (commandIndex + 1) % commands.length
        }, 20000)
      }, 50000)
    }
  }
}
</script>

<style scoped>
.cmd-window {
  background: #1a001a;
  color: #ff66cc;
  font-family: 'Courier New', monospace;
  padding: 12px;
  margin: 15px 0;
  border: 2px solid #ff99ff;
  height: 150px;
  overflow-y: auto;
  box-shadow: inset 0 0 10px #9900cc;
  font-size: 12px;
  text-align: left; /* 加這行讓文字靠左對齊 */
}

.current-line {
  color: #ff66cc;
}

.cursor {
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>