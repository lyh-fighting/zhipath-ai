import { View, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { wechatLogin } from '@/services/api'
import { setUser } from '@/stores/user'

export default function Login() {
  const handleLogin = async () => {
    try {
      const { code } = await Taro.login()
      const data = await wechatLogin(code)
      setUser({ userId: data.user_id, token: data.access_token, isLoggedIn: true })
      Taro.redirectTo({ url: '/pages/index/index' })
    } catch {
      Taro.showToast({ title: '登录失败', icon: 'error' })
    }
  }

  return (
    <View className='login'>
      <View className='title'>知途</View>
      <View className='subtitle'>知己，知路，走好下一步</View>
      <Button onClick={handleLogin}>微信登录</Button>
    </View>
  )
}
