# lease-renew

分布式租约的授予与续约。状态落在一个 JSON 文件里，子命令读写同一份状态。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/leasectl early
go run ./cmd/leasectl expired
go run ./cmd/leasectl replay
go run ./cmd/leasectl limit
go run ./cmd/leasectl work
```

## 口径

- **授予**：`expiry = at + ttl`，续约次数归零；同一个 id 重复授予保留原来那份。
- **续约窗口**：只有落在 `[expiry - 1000, expiry)` 内的续约才生效；更早的报 `too-early`。
- **过期**：`at >= expiry` 的续约报 `expired`，租约要重新授予。
- **次数上限**：一份租约最多续约 3 次，超出的报 `limit`。
- **幂等**：同一个 `req_id` 只生效一次，重复请求报 `replay` 且不再顺延。
- **顺延**：续约成功时 `expiry = at + ttl`。

## 不变量

- 只有窗口内的续约能改变 `expiry`。
- 续约次数不超过 3。
- `scanned` 不随租约数放大：两千份租约加两千次续约的 `scanned` 不超过 9000。

## 输出契约

每次操作打印一行 JSON：`id`、`status`（`granted` / `exists` / `ok` / `replay` / `too-early` / `expired` / `limit` / `unknown`）、`expiry_ms`、`renewals`、`scanned`。
`--sample=work` 只打印 `{"leases": N, "scanned": N}`。
