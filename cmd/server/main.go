package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"aura-trade/api"
	"aura-trade/internal/db"
)

func main() {
	loadDotEnv()

	// Initialize DB
	database := db.GetDB()
	if database == nil {
		log.Fatal("failed to initialize database")
	}

	// Background alert checker
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			api.CheckAlerts()
		}
	}()

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/register", cors(api.Handler)) // mapped to register.go
	mux.HandleFunc("/api/login", cors(api.Handler))    // mapped to login.go
	// Wait, the local server needs to know WHICH handler to call.
	// Since I split them, I need to import the specific files or use a different approach for local dev.
	// However, for local dev, the user usually runs 'go run ./cmd/server/main.go'.
	// But now I have multiple files in package 'api'.
	
	// Actually, in Go, multiple files in the same package 'api' are all part of the same package.
	// But they all have a function named 'Handler' now! This will cause a name collision in the 'api' package.
	
	// CRITICAL ERROR IN MY PLAN: I named them all 'Handler' in package 'api'.
	// This works for Vercel because it compiles each file INDEPENDENTLY.
	// But for local Go development (cmd/server), it will FAIL to compile because of multiple 'Handler' definitions in package 'api'.

	// To fix this for both Vercel AND local development:
	// Each file in 'api/' should have its OWN package name if they all want a 'Handler' function.
	// OR, for Vercel, I can keep them in package 'api' but name the functions differently and use 'vercel.json' to map them.
	// BUT, Vercel Go runtime's simplest path is 'package main' in each file in 'api/'.
	
	// If I change every file in 'api/' to 'package main', then they are independent.
	
	// Let's check Vercel docs again.
	// "Go files in the api directory must export a function that matches the http.HandlerFunc signature. The exported function name must be Handler."
	// And "Each file is a separate Serverless Function."
	
	// If they are separate functions, they don't share a package space in Vercel's build process.
	// But in a local 'go build' of the whole project, they DO if they are in the same directory.
	
	// THE FIX: Change the package name in each 'api/*.go' file to 'package handler' or something unique, OR just 'package main'.
	// Vercel doesn't care about the package name as long as 'Handler' is exported.
	
	// Actually, the most robust way for Vercel + Local is:
	// 1. Give each file in 'api/' a unique package name OR keep them in 'package api' but give them unique handler names.
	// 2. If they have unique handler names, I need to tell Vercel WHICH function to use.
	// BUT Vercel Go runtime ONLY looks for 'Handler' by default.
	
	// Wait, if I use 'package main' in each file in 'api/', they won't collide.
	// But then 'cmd/server/main.go' can't import them all easily.
	
	// Let's try this:
	// Keep the unique handler names (AnalyzeHandler, LoginHandler, etc.) in 'package api'.
	// And for Vercel, create a small 'Handler' wrapper in each file OR use 'vercel.json' configuration if possible.
	// Actually, Vercel Go runtime DOES support other names if you use 'vercel.json' with 'functions' property, but it's simpler to just have 'Handler'.
	
	// What if I put each handler in its own subdirectory?
	// api/analyze/handler.go (package analyze)
	// api/login/handler.go (package login)
	// Then Vercel maps them to /api/analyze and /api/login.
	// And 'cmd/server/main.go' can import them as:
	// import "aura-trade/api/analyze"
	// import "aura-trade/api/login"
	
	// THIS IS THE CLEANEST WAY.
	
	// Let's redo the directory structure for 'api/'.
	
	// Current files in 'api/':
	// alerts.go, analyze.go, chat.go, login.go, models.go, profile.go, register.go, screener.go, top-picks.go
	
	// New structure:
	// api/alerts/handler.go
	// api/analyze/handler.go
	// api/chat/handler.go
	// api/login/handler.go
	// api/models/handler.go
	// api/profile/handler.go
	// api/register/handler.go
	// api/screener/handler.go
	// api/top-picks/handler.go

	// This way:
	// 1. Vercel maps them perfectly.
	// 2. No name collisions.
	// 3. Local server can import them individually.
	
	// Let's do this.
