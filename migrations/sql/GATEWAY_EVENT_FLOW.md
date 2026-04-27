# Gateway 事件流完整指南

## 简答

流式输出的完整链路：

```
Agent 发出事件 (l.emit)
  ↓
事件总线 (bus.EventPublisher)
  ├─ emitRun(ChatEventChunk) → 投递到 bus
  └─ emitRun(AgentEventBlockReply) → 投递到 bus
  ↓
Gateway 订阅 (registerClient:609)
  ├─ s.eventPub.Subscribe(clientID, callback)
  └─ callback 每个事件调用一次
  ↓
事件过滤 (clientCanReceiveEvent)
  ├─ 权限检查（租户隔离）
  ├─ UserID 过滤（只发给相关用户）
  └─ 团队隔离检查
  ↓
协议转换 (protocol.NewEvent)
  └─ bus.Event → protocol.EventFrame
  ↓
发送到客户端 (c.SendEvent)
  ├─ JSON 序列化
  ├─ 投放到 c.send channel（256 个缓冲）
  └─ writePump 异步读取和发送
  ↓
WebSocket 文本帧
  └─ 客户端接收 (readMessage)
```

---

## 完整代码链路

### 1️⃣ Agent 发出事件

**位置：** `internal/agent/loop_pipeline_callbacks.go:23-38`

```go
func (l *Loop) pipelineCallbacks(req *RunRequest, bridgeRS *runState) pipelineCallbackSet {
    // 共享 emitRun 闭包，捕获 *Loop 的状态
    emitRun := func(event AgentEvent) {
        // 丰富路由上下文
        event.RunKind = req.RunKind
        event.DelegationID = req.DelegationID
        event.TeamID = req.TeamID
        event.ParentAgentID = req.ParentAgentID
        event.UserID = req.UserID
        event.Channel = req.Channel
        event.SessionKey = req.SessionKey
        event.TenantID = l.tenantID

        // ← 关键：发出到事件总线
        l.emit(event)
    }
    // ...
}
```

**什么时候调用 emitRun？**

```go
// 1. Token 级流式（makeCallLLM:266-282）
if req.Stream {
    resp, err = provider.ChatStream(ctx, chatReq, func(chunk providers.StreamChunk) {
        if chunk.Content != "" {
            emitRun(AgentEvent{
                Type: protocol.ChatEventChunk,
                Payload: map[string]string{"content": chunk.Content},
            })  // ← 每个 token 都发出
        }
    })
}

// 2. 块状回复（think_stage.go:150-151）
if resp.Content != "" && s.deps.EmitBlockReply != nil {
    s.deps.EmitBlockReply(resp.Content)
    // ← EmitBlockReply 是 emitRun 的包装
}
```

### 2️⃣ Loop.emit() 投递到事件总线

**位置：** `internal/agent/resolver.go` 或 `types.go`（需要查找 emit 实现）

推测实现（根据上下文）：

```go
func (l *Loop) emit(event AgentEvent) {
    busEvent := bus.Event{
        Name:     protocol.ChatEventChunk,  // 或其他事件名
        Payload:  event,
        TenantID: l.tenantID,  // 租户隔离
        UserID:   event.UserID,  // 用户过滤
    }

    // 投递到全局事件总线
    // 例如通过 eventbus.DomainEventBus.Publish()
    l.domainBus.Publish(busEvent)
}
```

### 3️⃣ Gateway 订阅事件总线

**位置：** `internal/gateway/server.go:603-627`

```go
func (s *Server) registerClient(c *Client) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.clients[c.id] = c

    // ← 客户端连接时注册订阅
    // 每个客户端一个独立的订阅回调
    s.eventPub.Subscribe(c.id, func(event bus.Event) {
        // 事件过滤（权限检查）
        if clientCanReceiveEvent(c, event) {
            // ← 只发送有权限的事件
            c.SendEvent(*protocol.NewEvent(event.Name, event.Payload))
        }
    })

    slog.Info("client connected", "id", c.id)
}

func (s *Server) unregisterClient(c *Client) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.clients, c.id)

    // ← 客户端断开连接时取消订阅
    s.eventPub.Unsubscribe(c.id)
    ...
}
```

**关键点：**
- 每个连接的客户端有独立的 subscription
- subscription ID = client.id
- callback 在 event publisher 线程执行
- 过滤决策在 callback 内部进行

### 4️⃣ 事件过滤（权限和隔离）

**位置：** `internal/gateway/event_filter.go:17-136`

```go
func clientCanReceiveEvent(c *Client, event bus.Event) bool {
    // 1. 内部事件过滤（不转发）
    if strings.HasPrefix(event.Name, "cache.") {
        return false
    }

    // 2. 系统事件广播给所有人
    if isSystemEvent(event.Name) {
        return true  // protocol.EventTick 等
    }

    // 3. 租户隔离：关键的安全检查
    // - 客户端连接时获得具体的 tenantID
    // - Owner 角色看到租户内的所有事件 + 系统事件
    // - 普通用户只看到自己租户的事件（fail-closed）
    if c.tenantID == uuid.Nil {
        return false  // fail-closed: 无租户分配
    }

    // 事件有租户范围 → 必须匹配
    if event.TenantID != uuid.Nil && event.TenantID != c.tenantID {
        return false
    }

    // 事件无租户范围 → 只有 owner 可以看到
    if event.TenantID == uuid.Nil && !c.IsOwner() {
        return false  // fail-closed: 普通用户不看无范围事件
    }

    // 4. 角色权限
    if permissions.HasMinRole(c.role, permissions.RoleAdmin) {
        return true  // Admin 看到所有内容（租户内）
    }

    // 5. Agent/Chat 事件：按 UserID 过滤
    // 只有消息对应的用户才能看到这个 Agent 运行的事件
    if event.Name == protocol.EventAgent || event.Name == protocol.EventChat {
        if uid := extractEventUserID(event); uid != "" {
            return uid == c.userID  // ← 只发给原始用户
        }
        return true  // 无路由上下文→广播
    }

    // 6. Team 事件：按 TeamID 过滤
    if strings.HasPrefix(event.Name, "team.") {
        if tid := extractTeamID(event); tid != "" {
            return c.hasTeamAccess(tid)  // ← 只发给有权限的团队成员
        }
        return true
    }

    // 7. 默认：拒绝未知事件（fail-closed）
    return false
}
```

**过滤规则总结：**

```
事件类型           | 过滤维度        | 实现
-------------------|-----------------|------------------
ChatEventChunk      | UserID          | event.Payload.UserID == c.userID
ChatEventThinking   | UserID          | 同上
AgentEventBlockReply| UserID          | 同上
EventSessionUpdated | UserID + TenantID| 同上
EventCron           | UserID + TenantID| 同上
team.*              | TeamID + TenantID| c.hasTeamAccess(teamID)
System events       | None            | 广播给所有人
Internal cache.*    | None            | 永远拒绝
Admin events        | Role check      | 仅 Admin
```

### 5️⃣ 协议转换

**位置：** `internal/gateway/server.go:611`

```go
// bus.Event → protocol.EventFrame
c.SendEvent(*protocol.NewEvent(event.Name, event.Payload))
```

**protocol.NewEvent 做什么？**（推测）

```go
func NewEvent(name string, payload any) *EventFrame {
    return &EventFrame{
        Type: FrameTypeEvent,
        Name: name,
        Payload: payload,  // JSON 序列化到 wire
    }
}
```

### 6️⃣ 发送到客户端

**位置：** `internal/gateway/client.go:176-192`

```go
func (c *Client) SendEvent(event protocol.EventFrame) {
    // 1. JSON 序列化
    data, err := json.Marshal(event)
    if err != nil {
        slog.Error("marshal event failed", "error", err)
        return
    }

    // 2. 错误恢复（客户端可能已断开）
    defer func() {
        if r := recover(); r != nil {
            slog.Debug("client gone, dropping event", "client", c.id)
        }
    }()

    // 3. 非阻塞发送到 channel（256 个缓冲）
    select {
    case c.send <- data:
        // ✅ 投放成功
    default:
        // ⚠️ 缓冲满，丢弃消息（背压）
        slog.Warn("client send buffer full, dropping event", "client", c.id)
    }
}
```

**关键点：**
- `c.send` 是大小为 256 的缓冲 channel
- 非阻塞 send（select + default）
- 缓冲满时丢弃（背压处理）
- 客户端主动读取（writePump）

### 7️⃣ WebSocket 写泵

**位置：** `internal/gateway/client.go:96-122`

```go
func (c *Client) writePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        // ← 从 c.send channel 读取消息
        case msg, ok := <-c.send:
            if !ok {
                // ← 管道关闭，连接终止
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            // 设置写超时
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

            // ← 写入 WebSocket 文本帧
            if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
                // ← 网络错误，写泵退出
                return
            }

        // 定期 ping 保活
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

**流程：**
1. Agent 发出 ChatEventChunk
2. 投放到 c.send channel
3. writePump 从 channel 读取
4. JSON 序列化数据写入 WebSocket
5. 客户端接收文本帧

### 8️⃣ 客户端接收

**位置：** `internal/gateway/client.go:68-93`

```go
func (c *Client) readPump(ctx context.Context) {
    defer c.conn.Close()

    c.conn.SetReadLimit(maxWSMessageSize)  // 512KB
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

    for {
        // ← 从 WebSocket 读取消息
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, ...) {
                slog.Warn("websocket read error", "client", c.id, "error", err)
            }
            return
        }

        // 处理帧
        c.handleFrame(ctx, data)
    }
}
```

---

## 完整时间线

```
时间 →

1. Client 连接
   ├─ handleWebSocket()
   ├─ newClient()
   └─ registerClient()
      └─ s.eventPub.Subscribe(clientID, callback)  ← 订阅注册

2. Agent 运行 (go func in chat.go:244-355)
   ├─ loop.Run(ctx, req)
   └─ runViaPipeline()
      └─ Pipeline.Run()
         └─ ThinkStage.Execute()
            ├─ CallLLM(Stream=true)
            │  └─ provider.ChatStream(callback)
            │     └─ [Token 1] → emitRun(ChatEventChunk)

3. emitRun 投递事件
   ├─ AgentEvent.UserID = req.UserID
   ├─ AgentEvent.TenantID = l.tenantID
   └─ l.emit(AgentEvent)
      └─ l.domainBus.Publish(event)  ← 投放到总线

4. Gateway 事件总线分发
   ├─ eventPub 调用所有 subscription callback
   └─ 对每个连接的客户端：
      └─ callback(event)
         ├─ clientCanReceiveEvent()
         │  ├─ Check: event.TenantID == c.tenantID?
         │  ├─ Check: event.UserID == c.userID?
         │  └─ return true/false
         │
         └─ 如果 true:
            └─ c.SendEvent(*protocol.NewEvent(...))

5. SendEvent 投放到 channel
   ├─ JSON 序列化 EventFrame
   └─ select c.send <- data  ← 非阻塞投放

6. writePump 从 channel 读取
   ├─ msg := <-c.send
   └─ c.conn.WriteMessage(TextMessage, msg)  ← 写 WS

7. 客户端接收 (JS 代码)
   ├─ ws.addEventListener('message', (e) => {
   │  └─ event = JSON.parse(e.data)  ← 解析
   │     └─ 显示内容
   └─ })

时间 →
```

---

## 数据结构

### bus.Event

```go
type Event struct {
    Name     string       // 事件名："chat.chunk", "agent.thinking", etc.
    Payload  any          // 事件内容：AgentEvent, map[string]any, 等
    TenantID uuid.UUID    // 租户 ID（uuid.Nil = 跨租户）
    // UserID 可能在 Payload 中（AgentEvent.UserID）
}
```

### protocol.EventFrame

```go
type EventFrame struct {
    Type    string `json:"type"`     // "event"
    Name    string `json:"name"`     // 事件名
    Payload any    `json:"payload"` // JSON 序列化的 payload
}
```

### AgentEvent

```go
type AgentEvent struct {
    Type:           string           // protocol.ChatEventChunk, etc.
    AgentID:        uuid.UUID        // Agent ID
    RunID:          string           // Run ID（用于去重）
    UserID:         string           // ← 用于 UserID 过滤
    TenantID:       uuid.UUID        // ← 用于租户隔离
    SessionKey:     string
    Channel:        string
    TeamID:         string           // ← 用于团队过滤
    Payload:        map[string]any   // 实际内容
    // ... 其他路由上下文
}
```

---

## 关键 API

### 1. EventPublisher.Subscribe

```go
// 在 gateway/server.go 中调用
s.eventPub.Subscribe(clientID, func(event bus.Event) {
    // 每个事件调用一次
})
```

### 2. EventPublisher.Publish

```go
// 在 agent loop 中调用（通过 l.emit()）
l.domainBus.Publish(busEvent)
```

### 3. Client.SendEvent

```go
// gateway 调用
c.SendEvent(eventFrame)
// → json.Marshal(eventFrame)
// → select c.send <- data (non-blocking)
```

---

## 权限模型（Fail-Closed）

### 租户隔离（T0: 强制）

```go
// 所有客户端必须有 tenantID（不能为 Nil）
if c.tenantID == uuid.Nil {
    return false  // ← 拒绝
}

// 事件和客户端的租户必须匹配
if event.TenantID != uuid.Nil && event.TenantID != c.tenantID {
    return false  // ← 拒绝
}
```

### 角色权限（T1: 基于角色）

```go
// Owner: 看到租户内所有 + 系统事件
if c.role == owner {
    return true
}

// Admin: 看到租户内所有
if c.role == admin {
    return true
}

// User: 有限制
if c.role == user {
    return checkUserScope()
}
```

### 用户范围（T2: 基于用户）

```go
// Agent/Chat 事件只发给原始用户
if event.Name == ChatEvent {
    if event.Payload.UserID != c.userID {
        return false  // ← 只发给原始用户
    }
}
```

### 团队范围（T3: 基于团队）

```go
// Team 事件检查团队成员资格
if event.Name.startsWith("team.") {
    if !c.hasTeamAccess(event.TeamID) {
        return false  // ← 拒绝
    }
}
```

---

## 去重机制

由于有多个事件来源，网关需要去重：

**场景：流式输出**
```
迭代 #0:
  ├─ ThinkStage: emitRun(ChatEventChunk: "我")  ← 实时 token
  ├─ ThinkStage: emitRun(ChatEventChunk: "现")  ← 实时 token
  ├─ ObserveStage: (无，有工具调用)
  └─ FinalizeStage: (无，未完成)

迭代 #1:
  ├─ ThinkStage: emitRun(ChatEventChunk: "好") ← 实时 token
  ├─ ObserveStage: (无，无工具调用 → BreakLoop)
  └─ FinalizeStage: emitRun(completed: "好的...") ← 最终答案
```

**网关去重规则（推测）：**
```
如果 ChatEventChunk 内容已发送 → 在 FinalizeStage completed 时跳过重复内容
否则 → 按顺序转发所有事件
```

---

## 性能特性

### 背压处理

```go
// client.send channel 大小为 256
// 如果满了（背压），新消息被丢弃
select {
case c.send <- data:
    // 成功
default:
    slog.Warn("client send buffer full, dropping event")
    // ← 丢弃（背压）
}
```

**意义：** 防止 Agent 发送过快导致内存爆炸

### 非阻塞设计

- emitRun() 不等待事件投放成功
- 事件投放失败时静默丢弃
- 网络背压不会阻塞 Agent 执行

### 订阅模型

- 每个 client 一个 subscription
- subscription ID = client ID
- 断开连接时自动清理（unregisterClient）

---

## 总结

✅ **完整流式输出链路：**

```
Agent.emit(event)
  → bus.EventPublisher.Publish()
  → gateway.registerClient() 订阅的 callback
  → clientCanReceiveEvent() 权限过滤
  → c.SendEvent() 投放到 channel
  → writePump() 从 channel 读取
  → WebSocket.WriteMessage() 发送
  → 客户端 JSON.parse() 接收
```

✅ **安全隔离：**
- 租户隔离（T0）：所有事件都检查 tenantID
- 角色权限（T1）：admin/owner/user 不同权限
- 用户过滤（T2）：Agent 事件只发给相关用户
- 团队隔离（T3）：Team 事件只发给成员

✅ **性能优化：**
- 背压处理：缓冲满时丢弃
- 非阻塞设计：发送不等待
- 定期清理：断开连接时自动注销

✅ **关键文件：**
- `gateway/server.go:609` - 事件订阅注册
- `gateway/event_filter.go:17` - 权限过滤规则
- `gateway/client.go:176` - 事件发送
- `gateway/client.go:96` - WebSocket 写泵
- `agent/loop_pipeline_callbacks.go:23` - emitRun 闭包
