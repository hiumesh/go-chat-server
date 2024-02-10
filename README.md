# go-chat-server

A websocket based chat system in go

## Run the devlopment server

`CompileDaemon -command="./go-chat-server" `

## Direct Message JSON Format

```json
{
  "type": "direct_message",
  "payload": {
    "body": "hi",
    "to": "<user_id>"
  }
}
```
