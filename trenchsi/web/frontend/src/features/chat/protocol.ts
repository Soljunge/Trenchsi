import { normalizeUnixTimestamp } from "@/features/chat/state"
import { updateChatStore } from "@/store/chat"

export interface JameMessage {
  type: string
  id?: string
  session_id?: string
  timestamp?: number | string
  payload?: Record<string, unknown>
}

export function handleJameMessage(
  message: JameMessage,
  expectedSessionId: string,
) {
  if (message.session_id && message.session_id !== expectedSessionId) {
    return
  }

  const payload = message.payload || {}

  switch (message.type) {
    case "message.create": {
      const content = (payload.content as string) || ""
      const messageId = (payload.message_id as string) || `jame-${Date.now()}`
      const timestamp =
        message.timestamp !== undefined &&
        Number.isFinite(Number(message.timestamp))
          ? normalizeUnixTimestamp(Number(message.timestamp))
          : Date.now()

      updateChatStore((prev) => ({
        messages: [
          ...prev.messages,
          {
            id: messageId,
            role: "assistant",
            content,
            timestamp,
          },
        ],
        isTyping: false,
        errorDetail: "",
      }))
      break
    }

    case "message.update": {
      const content = (payload.content as string) || ""
      const messageId = payload.message_id as string
      if (!messageId) {
        break
      }

      updateChatStore((prev) => ({
        messages: prev.messages.map((msg) =>
          msg.id === messageId ? { ...msg, content } : msg,
        ),
      }))
      break
    }

    case "typing.start":
      updateChatStore({ isTyping: true, errorDetail: "" })
      break

    case "typing.stop":
      updateChatStore({ isTyping: false })
      break

    case "error":
      console.error("Jame error:", payload)
      updateChatStore({
        isTyping: false,
        errorDetail:
          (payload.message as string) ||
          (payload.error as string) ||
          "The model returned an error while processing your request.",
      })
      break

    case "pong":
      break

    default:
      console.log("Unknown jame message type:", message.type)
  }
}
