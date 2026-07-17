import { View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { getUser } from '@/stores/user'

export default function Index() {
  const user = getUser()
  if (!user.isLoggedIn) {
    Taro.redirectTo({ url: '/pages/login/index' })
  }

  const go = (url: string) => Taro.navigateTo({ url })

  return (
    <View className='index'>
      <View className='hero'>
        <View className='title'>知途 AI</View>
        <View className='slogan'>知己，知路，走好下一步</View>
      </View>
      <View className='menu'>
        <View className='item' onClick={() => go('/pages/chat/index')}>立即咨询</View>
        <View className='item' onClick={() => go('/pages/mbti/index')}>MBTI 测试</View>
        <View className='item' onClick={() => go('/pages/profile/index')}>我的画像</View>
        <View className='item' onClick={() => go('/pages/report/index')}>深度报告</View>
        <View className='item' onClick={() => go('/pages/orders/index')}>订单</View>
        <View className='item' onClick={() => go('/pages/settings/index')}>设置</View>
      </View>
    </View>
  )
}
