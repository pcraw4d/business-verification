package classification

import (
	"context"
	"log"
	"testing"

	"kyb-platform/internal/classification/testutil"
)

// BenchmarkHybridCodeGeneration benchmarks the hybrid code generation performance
func BenchmarkHybridCodeGeneration(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology", "platform", "development", "cloud"}
	detectedIndustry := "Technology"
	confidence := 0.85

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkHybridCodeGeneration_MultiIndustry benchmarks multi-industry performance
func BenchmarkHybridCodeGeneration_MultiIndustry(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology", "finance", "banking"}
	detectedIndustry := "Technology"
	confidence := 0.85
	additionalIndustries := []IndustryResult{
		{IndustryName: "Software", Confidence: 0.75},
		{IndustryName: "Financial Services", Confidence: 0.70},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence, additionalIndustries...)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkHybridCodeGeneration_LargeKeywordSet benchmarks with large keyword sets
func BenchmarkHybridCodeGeneration_LargeKeywordSet(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	// Generate a large set of keywords
	keywords := make([]string, 100)
	for i := 0; i < 100; i++ {
		keywords[i] = "keyword" + string(rune('0'+i%10))
	}
	detectedIndustry := "Technology"
	confidence := 0.85

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkKeywordCodeLookup benchmarks keyword-based code lookup
func BenchmarkKeywordCodeLookup(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology", "platform"}
	codeType := "MCC"
	industryConfidence := 0.85

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := generator.generateCodesFromKeywords(ctx, keywords, codeType, industryConfidence)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// BenchmarkCodeMerging benchmarks the code merging operation
func BenchmarkCodeMerging(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	// Create sample data
	industryCodes := []*repository.ClassificationCode{
		{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test Code 1"},
		{ID: 2, Code: "5678", CodeType: "MCC", Description: "Test Code 2"},
		{ID: 3, Code: "9012", CodeType: "MCC", Description: "Test Code 3"},
	}

	keywordCodes := []CodeMatch{
		{
			Code:       &repository.ClassificationCode{ID: 1, Code: "1234", CodeType: "MCC", Description: "Test Code 1"},
			Source:     "keyword",
			Confidence: 0.8,
		},
		{
			Code:       &repository.ClassificationCode{ID: 4, Code: "3456", CodeType: "MCC", Description: "Test Code 4"},
			Source:     "keyword",
			Confidence: 0.75,
		},
	}

	industryConfidence := 0.85
	codeType := "MCC"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = generator.mergeCodeResults(industryCodes, keywordCodes, industryConfidence, codeType)
	}
}

// BenchmarkParallelCodeGeneration benchmarks parallel code generation for all types
func BenchmarkParallelCodeGeneration(b *testing.B) {
	mockRepo := testutil.NewMockKeywordRepository()
	generator := NewClassificationCodeGenerator(mockRepo, log.Default())

	ctx := context.Background()
	keywords := []string{"software", "technology"}
	industries := []IndustryResult{
		{IndustryName: "Technology", Confidence: 0.85},
		{IndustryName: "Software", Confidence: 0.75},
	}
	codes := &ClassificationCodesInfo{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		generator.generateCodesInParallel(ctx, codes, keywords, industries)
		// Reset codes for next iteration
		codes.MCC = nil
		codes.SIC = nil
		codes.NAICS = nil
	}
}

