import Taro from '@tarojs/taro'

interface UserState {
  userId: string
  token: string
  isLoggedIn: boolean
}

const STORAGE_KEY = 'zhipath_user'

export function getUser(): UserState {
  return Taro.getStorageSync(STORAGE_KEY) || { userId: '', token: '', isLoggedIn: false }
}

export function setUser(user: UserState) {
  Taro.setStorageSync(STORAGE_KEY, user)
}

export function logout() {
  Taro.removeStorageSync(STORAGE_KEY)
}
