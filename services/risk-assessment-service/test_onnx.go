package main

import (
	"fmt"
	"os"
	"path/filepath"

	ort "github.com/yalue/onnxruntime_go"
)

func main() {
	fmt.Println("🔍 Testing ONNX Runtime Installation...")

	// Check environment variables
	fmt.Printf("LD_LIBRARY_PATH: %s\n", os.Getenv("LD_LIBRARY_PATH"))
	fmt.Printf("CGO_ENABLED: %s\n", os.Getenv("CGO_ENABLED"))

	// Check if ONNX Runtime libraries exist
	libPath := "/app/onnxruntime/lib"
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		fmt.Printf("❌ ONNX Runtime lib directory not found: %s\n", libPath)
	} else {
		fmt.Printf("✅ ONNX Runtime lib directory exists: %s\n", libPath)

		// List files in the lib directory
		files, err := filepath.Glob(filepath.Join(libPath, "*"))
		if err != nil {
			fmt.Printf("❌ Error listing lib files: %v\n", err)
		} else {
			fmt.Printf("📁 Files in lib directory:\n")
			for _, file := range files {
				fmt.Printf("  - %s\n", file)
			}
		}
	}

	// Check if model file exists
	modelPath := "/app/models/risk_lstm_v1.onnx"
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Printf("❌ Model file not found: %s\n", modelPath)
	} else {
		fmt.Printf("✅ Model file exists: %s\n", modelPath)
	}

	// Try to initialize ONNX Runtime
	fmt.Println("\n🚀 Attempting to initialize ONNX Runtime...")
	err := ort.InitializeEnvironment()
	if err != nil {
		fmt.Printf("❌ Failed to initialize ONNX Runtime: %v\n", err)
		fmt.Println("\n🔧 Debugging information:")
		fmt.Println("- Check if ONNX Runtime C libraries are properly installed")
		fmt.Println("- Verify LD_LIBRARY_PATH includes the library directory")
		fmt.Println("- Ensure the application was built with CGO_ENABLED=1")
	} else {
		fmt.Println("✅ ONNX Runtime initialized successfully!")

		// Try to create a simple session
		fmt.Println("\n🧪 Testing session creation...")
		session, err := ort.NewDynamicSession[float32, float32](
			modelPath,
			[]string{"input"},
			[]string{"output"},
		)
		if err != nil {
			fmt.Printf("❌ Failed to create ONNX session: %v\n", err)
		} else {
			fmt.Println("✅ ONNX session created successfully!")
			session.Destroy()
		}
	}
}
