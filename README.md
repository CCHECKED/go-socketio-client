# Socket.IO Go Client Library

![Go](https://img.shields.io/badge/Go-1.18%2B-blue) ![Socket.IO](https://img.shields.io/badge/Socket.IO-4.x-green) ![License](https://img.shields.io/badge/License-MIT-lightgrey)

Socket.IO клиентская библиотека для Go с поддержкой TLS и простым API для управления подключениями.

## 🚀 Возможности

- Поддержка **Socket.IO 4.x**
- **TLS**-шифрование подключений
- Поддержка нескольких **неймспейсов**
- Обработчики **сообщений** и **сервисных событий**
- Гибкая конфигурация
- Легкость в использовании

## 📦 Установка

```sh
go get github.com/CCHECKED/go-socketio-client
```

## 🔧 Использование

```go
manager := socketio.NewSocketManager(
	&url.URL{
        Scheme: "https",
        Host:   "localhost:3000",
        Path:   "/devices/",
    },
    &socketio.SocketManagerConfig{
        AllowUpgrade: false,
        Tls:          &tls.Config{},
        Headers:      &http.Header{},
        Debug:        false,
    },
)

// Подключение к namespace "/"
client := manager.DefaultSocket("/")
// Подключение к namespace "/chat"
clientChat := manager.DefaultSocket("/chat")

client.On("message", func(data interface{}) {
    fmt.Println("Recieved message:", data)
})

client.OnService(consts.PING, func() {
    fmt.Println("Recieved Ping")
})

clientChat.On("message", func(data interface{}) {
    fmt.Println("Recieved chat message:", data)
})
```

## ⚙ Конфигурация

При создании `SocketManager` можно передать объект конфигурации `SocketManagerConfig`:

| Поле          | Тип             | Описание                              |
|--------------|-----------------|---------------------------------------|
| `AllowUpgrade` | `bool`          | Разрешить обновление соединения       |
| `Tls` | `*tls.Config`   | Конфигурация TLS                      |
| `Headers` | `*http.Header`  | Заголовки запроса                     |
| `Debug` | `bool`          | Включить режим отладки                |
| `ReconnectWait` | `time.Duration` | Время между попытками переподключения |

## 🛠 Подключение обработчиков событий

```go
client.On("event_name", func(data interface{}) {
	fmt.Println("Получено событие:", data)
})
```

Для сервисных событий:

```go
client.OnService(consts.PING, func() {
	fmt.Println("Получен PING")
})
```