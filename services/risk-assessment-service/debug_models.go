package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🔍 Debugging Model Files in Container...")
	
	// Check if models directory exists
	modelsDir := "/app/models"
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		fmt.Printf("❌ Models directory not found: %s\n", modelsDir)
	} else {
		fmt.Printf("✅ Models directory exists: %s\n", modelsDir)
		
		// List all files in models directory
		files, err := filepath.Glob(filepath.Join(modelsDir, "*"))
		if err != nil {
			fmt.Printf("❌ Error listing files: %v\n", err)
		} else {
			fmt.Printf("📁 Files in models directory:\n")
			for _, file := range files {
				info, err := os.Stat(file)
				if err != nil {
					fmt.Printf("  - %s (error getting info: %v)\n", file, err)
				} else {
					fmt.Printf("  - %s (%d bytes, %s)\n", file, info.Size(), info.Mode())
				}
			}
		}
	}
	
	// Check specific model files
	modelFiles := []string{
		"/app/models/risk_lstm_v1.onnx",
		"/app/models/xgb_model.json",
	}
	
	for _, file := range modelFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("❌ Model file not found: %s\n", file)
		} else {
			info, err := os.Stat(file)
			if err != nil {
				fmt.Printf("❌ Error getting file info for %s: %v\n", file, err)
			} else {
				fmt.Printf("✅ Model file found: %s (%d bytes)\n", file, info.Size())
			}
		}
	}
	
	// Check environment variables
	fmt.Println("\n🔧 Environment Variables:")
	fmt.Printf("LSTM_MODEL_PATH: %s\n", os.Getenv("LSTM_MODEL_PATH"))
	fmt.Printf("XGBOOST_MODEL_PATH: %s\n", os.Getenv("XGBOOST_MODEL_PATH"))
}
