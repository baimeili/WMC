<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  // 应用启动：此处只做本地初始化，不发起网络请求
  console.log('WMC 应用启动')
})
</script>

<style lang="scss">
/* 全站公共样式：App.vue 的样式在 uni-app 中为全局样式（不会被 scoped 化）。
   注意：
   1. uni-app 只向页面/组件注入 uni.scss，App.vue 需要显式 @import 设计令牌；
   2. 不要把公共样式放进普通组件，否则会被加 data-v 作用域导致页面内容取不到样式。 */
@import './uni.scss';
@import './styles/global.scss';
</style>
