### 5. Medium: Изоляция имен инструментов (Namespacing)

**Проблема:** Коллизии имен инструментов при подключении нескольких MCP-серверов.

**Решение:** Добавление префикса сервера к имени инструмента.

```go
func (m *Manager) GetTools(ctx context.Context) ([]core.Tool, error) {
    // ...
    qualifiedName := fmt.Sprintf("%s__%s", res.serverName, t.Name)
    allTools = append(allTools, core.Tool{
        Function: core.Function{
            Name:        qualifiedName, // Пример: "filesystem__read_file"
            Description: fmt.Sprintf("[%s] %s", res.serverName, t.Description),
        },
    })
    newToolToClient[qualifiedName] = clientsSnapshot[res.serverName]
}
```

---

### 6. Medium: Health Checking и авто-удаление

**Проблема:** Отсутствие детекции "мертвых" MCP-серверов после инициализации.

**Решение:** Фоновая горутина для проверки здоровья соединений.

```go
func (m *Manager) healthCheck(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            m.mu.Lock()
            for name, cli := range m.clients {
                if !cli.IsHealthy(ctx) {
                    cli.Close()
                    delete(m.clients, name)
                    m.cacheValid = false
                }
            }
            m.mu.Unlock()
        }
    }
}
```

---

### 7. Low: Оптимизация блокировок в ManageMCP

**Проблема:** Переусложненная логика блокировок, увеличивающая риск deadlock.

**Решение:** Вынос I/O операций за пределы критической секции.

```go
func (m *Manager) addServer(ctx context.Context, input struct{...}) (string, error) {
    newClient, err := m.connectToServer(ctx, newCfg)
    if err != nil {
        return "", err
    }

    m.mu.Lock()
    if oldCli, exists := m.clients[input.ServerName]; exists {
        go oldCli.Close() 
    }
    m.clients[input.ServerName] = newClient
    m.config.MCPServers[input.ServerName] = newCfg
    m.cacheValid = false
    m.mu.Unlock()

    if err := m.configStore.Save(ctx, m.config); err != nil {
        return "Server started but config save failed", err
    }
    return fmt.Sprintf("Server %s added", input.ServerName), nil
}
```

---

### 8. Low: Фабрика клиентов для тестирования

**Проблема:** Прямой вызов `client.NewStdioMCPClient` делает невозможным unit-тестирование.

**Решение:** Внедрение интерфейса фабрики.

```go
type MCPClientFactory interface {
    Create(ctx context.Context, cfg ServerConfig) (*client.Client, error)
}

type StdioClientFactory struct{}

func (f *StdioClientFactory) Create(ctx context.Context, cfg ServerConfig) (*client.Client, error) {
    return client.NewStdioMCPClient(cfg.Command, cfg.Env, cfg.Args...)
}
```

---

### Итоговая структура изменений

| Файл | Описание |
| :--- | :--- |
| `internal/providers/mcp/pool.go` | **New:** Реализация ConnectionPool |
| `internal/providers/mcp/config.go` | **New:** Интерфейс ConfigStore и файловая реализация |
| `internal/providers/mcp/client.go` | **New:** Обертка ManagedClient |
| `internal/providers/mcp/manager.go` | **Refactor:** Использование новых типов, исправление контекстов и блокировок |