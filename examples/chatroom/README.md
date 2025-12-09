# Chatroom Example

This example demonstrates a simple chatroom application using WebSocket.

## Features

- Multiple clients can be connected at the same time
- Real-time message broadcasting
- User join/leave notifications
- Online count statistics
- Customizable usernames

## How to Run

### 1. Start the Server

```bash
cd examples/chatroom
go run server.go
```

The server will start at `ws://0.0.0.0:8080/ws/chat`.

### 2. Start the Client

Run in a new terminal window (multiple clients can be started):

```bash
go run client.go
```

Enter your username and start chatting.

## Usage

- Type a message and press Enter to send
- Type `/quit` to leave the chatroom
- All messages are broadcasted to other online users

## Architecture

### Server

- `ChatRoom`: Manages all connected clients
- `chatEventHandler`: Handles connection, disconnection, and message events
- Uses `sync.RWMutex` to ensure concurrency safety

### Client

- `chatClientHandler`: Handles received messages
- Uses `bufio.Reader` to read user input
- Passes username via HTTP Header

## Example Output

**Server**:
```
Chatroom server started at ws://0.0.0.0:8080/ws/chat
User [Alice] joined chatroom, online: 1
User [Bob] joined chatroom, online: 2
Received message: [Alice]: Hello!
Received message: [Bob]: Hello Alice
```

**Client**:
```
Enter your username: Alice
✓ Connected to chatroom, your username: Alice
Type a message and press Enter to send, type /quit to leave
> Hello!
System: [Bob] joined the chatroom
[Bob]: Hello Alice
```

