package main

import (
	"camie-portal/backend/internal/config"
	"camie-portal/backend/internal/demo"
	"camie-portal/backend/internal/generator"
	"camie-portal/backend/internal/httpapi"
	"camie-portal/backend/internal/repository"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("用法: camie-static <preview|generate|serve|preview-server> --config config.example.yaml")
	}
	command := args[0]
	if command != "preview" && command != "generate" && command != "serve" && command != "preview-server" {
		return fmt.Errorf("未知命令 %q", command)
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	name := flags.String("config", "config.example.yaml", "配置文件")
	addr := flags.String("addr", "127.0.0.1:4173", "预览监听地址")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	cfg, err := config.Load(*name)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if command == "preview" || command == "preview-server" {
		cfg.Site.DistRoot = cfg.Site.PreviewRoot
		g := generator.New(cfg, demo.NewSource())
		result, err := g.GenerateSite(ctx)
		if err != nil {
			return err
		}
		_ = json.NewEncoder(os.Stdout).Encode(result)
		if command == "preview" {
			return nil
		}
		fmt.Printf("CAMIE preview: http://%s\n", *addr)
		// Serve only the generated public tree, never backend configuration or source files.
		return serve(ctx, *addr, http.FileServer(http.Dir(cfg.Site.PreviewRoot)))
	}
	dsn := os.Getenv(cfg.Database.DSNEnv)
	if dsn == "" {
		return fmt.Errorf("请设置环境变量 %s", cfg.Database.DSNEnv)
	}
	mysqlCfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("invalid MySQL DSN")
	}
	mysqlCfg.ParseTime = true
	mysqlCfg.Loc, _ = time.LoadLocation(cfg.Site.Timezone)
	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	ping, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = db.PingContext(ping)
	cancel()
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	id, err := repository.ResolvePageID(ctx, db, cfg.Site.PageName)
	if err != nil {
		return err
	}
	g := generator.New(cfg, repository.New(db, id))
	timeout, _ := time.ParseDuration(cfg.Server.RequestTimeout)
	if command == "generate" {
		runCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		result, err := g.GenerateSite(runCtx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	token := os.Getenv(cfg.Server.TokenEnv)
	if len(token) < 24 {
		return fmt.Errorf("%s 必须至少为 24 个字符", cfg.Server.TokenEnv)
	}
	return serve(ctx, cfg.Server.Addr, httpapi.New(ctx, g, token, timeout))
}
func serve(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	failures := make(chan error, 1)
	go func() { failures <- server.ListenAndServe() }()
	select {
	case err := <-failures:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
