<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  // 应用启动：此处只做本地初始化，不发起网络请求
  console.log('WMC 应用启动')
})
</script>

<style lang="scss">
page {
  background-color: $uni-bg-color;
  color: $uni-text-color;
  font-size: $uni-font-size-base;
  font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
</style>
