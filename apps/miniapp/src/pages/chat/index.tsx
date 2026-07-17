import { View, Input, Button, Text } from '@tarojs/components'
import { useState } from 'react'
import { sendMessage } from '@/services/api'

export default function Chat() {
  const [message, setMessage] = useState('')
  const [reply, setReply] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSend = async () => {
    if (!message.trim()) return
    setLoading(true)
    const requestId = `req_${Date.now()}` // 网络重试复用相同 request_id
    try {
      const data = await sendMessage('current', message, requestId)
      setReply((data as { content_summary?: string })?.content_summary || '已收到')
    } catch {
      setReply('发送失败，请重试')
    }
    setLoading(false)
    setMessage('')
  }

  return (
    <View className='chat'>
      <View className='messages'>{reply && <Text>{reply}</Text>}</View>
      <View className='input-bar'>
        <Input value={message} onInput={(e) => setMessage(e.detail.value)} placeholder='说说你的困惑' />
        <Button onClick={handleSend} loading={loading}>发送</Button>
      </View>
    </View>
  )
}
