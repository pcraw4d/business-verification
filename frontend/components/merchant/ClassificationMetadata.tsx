import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Info, Database, Building2, Globe, Code2, Layers, TrendingUp } from 'lucide-react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ClassificationData } from '@/types/merchant';

interface ClassificationMetadataProps {
  metadata?: ClassificationData['metadata'];
  compact?: boolean;
}

export function ClassificationMetadata({ metadata, compact = false }: ClassificationMetadataProps) {
  if (!metadata) {
    return null;
  }

  // If metadata is empty object, return null
  if (Object.keys(metadata).length === 0) {
    return null;
  }

  if (compact) {
    return (
      <div className="flex flex-wrap gap-2">
        {metadata.pageAnalysis && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge variant="outline" className="cursor-help">
                  <Info className="h-3 w-3 mr-1" />
                  {metadata.pageAnalysis.pagesAnalyzed || 0} pages
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                <p>Analysis method: {metadata.pageAnalysis.analysisMethod || 'N/A'}</p>
                {metadata.pageAnalysis.structuredDataFound && (
                  <p>Structured data found</p>
                )}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
        {metadata.brandMatch?.isBrandMatch && (
          <Badge variant="secondary">
            <Building2 className="h-3 w-3 mr-1" />
            Brand: {metadata.brandMatch.brandName || 'Unknown'}
          </Badge>
        )}
        {metadata.pageAnalysis?.structuredDataFound && (
          <Badge variant="outline">
            <Database className="h-3 w-3 mr-1" />
            Structured Data
          </Badge>
        )}
        {metadata.codeGeneration && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Badge variant="outline" className="cursor-help">
                  <Code2 className="h-3 w-3 mr-1" />
                  {metadata.codeGeneration.method}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                <p>Method: {metadata.codeGeneration.method}</p>
                <p>Total codes: {metadata.codeGeneration.totalCodesGenerated}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-3 border-t pt-3 mt-3">
      <h4 className="text-sm font-semibold">Analysis Metadata</h4>
      
      {/* Page Analysis */}
      {metadata.pageAnalysis && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Info className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-medium">Pages Analyzed</span>
            </div>
            <Badge variant="outline">
              {metadata.pageAnalysis.pagesAnalyzed || 0} pages
            </Badge>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Analysis Method</span>
            <Badge variant="secondary">
              {metadata.pageAnalysis.analysisMethod === 'multi_page' && 'Multi-Page'}
              {metadata.pageAnalysis.analysisMethod === 'single_page' && 'Single-Page'}
              {metadata.pageAnalysis.analysisMethod === 'url_only' && 'URL Only'}
              {!metadata.pageAnalysis.analysisMethod && 'N/A'}
            </Badge>
          </div>
          {metadata.pageAnalysis.structuredDataFound && (
            <div className="flex items-center gap-2">
              <Database className="h-4 w-4 text-green-600" />
              <span className="text-sm text-green-600">Structured Data Found</span>
            </div>
          )}
        </div>
      )}

      {/* Brand Match */}
      {metadata.brandMatch?.isBrandMatch && (
        <div className="space-y-2 border-t pt-2">
          <div className="flex items-center gap-2">
            <Building2 className="h-4 w-4 text-blue-600" />
            <span className="text-sm font-medium">Brand Match</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Brand Name</span>
            <Badge variant="secondary">
              {metadata.brandMatch.brandName || 'Unknown'}
            </Badge>
          </div>
          {metadata.brandMatch.confidence && (
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Confidence</span>
              <span className="text-sm font-medium">
                {Math.round(metadata.brandMatch.confidence * 100)}%
              </span>
            </div>
          )}
        </div>
      )}

      {/* Data Source Priority */}
      {metadata.dataSourcePriority && (
        <div className="space-y-2 border-t pt-2">
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">Data Source Priority</span>
          </div>
          <div className="space-y-1">
            {metadata.dataSourcePriority.websiteContent && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Website Content</span>
                <Badge variant={metadata.dataSourcePriority.websiteContent === 'primary' ? 'default' : 'outline'}>
                  {metadata.dataSourcePriority.websiteContent}
                </Badge>
              </div>
            )}
            {metadata.dataSourcePriority.businessName && (
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Business Name</span>
                <Badge variant={metadata.dataSourcePriority.businessName === 'primary' ? 'default' : 'outline'}>
                  {metadata.dataSourcePriority.businessName}
                </Badge>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Code Generation Metadata */}
      {metadata.codeGeneration && (
        <div className="space-y-2 border-t pt-2">
          <div className="flex items-center gap-2">
            <Code2 className="h-4 w-4 text-purple-600" />
            <span className="text-sm font-medium">Code Generation</span>
          </div>
          <div className="space-y-1">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Method</span>
              <Badge variant={
                metadata.codeGeneration.method === 'hybrid' ? 'default' :
                metadata.codeGeneration.method === 'keyword_only' ? 'secondary' : 'outline'
              }>
                {metadata.codeGeneration.method === 'hybrid' && 'Hybrid'}
                {metadata.codeGeneration.method === 'keyword_only' && 'Keyword Only'}
                {metadata.codeGeneration.method === 'industry_only' && 'Industry Only'}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Total Codes</span>
              <span className="text-sm font-medium">
                {metadata.codeGeneration.totalCodesGenerated}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Industry Matches</span>
              <Badge variant="outline">
                <Layers className="h-3 w-3 mr-1" />
                {metadata.codeGeneration.industryMatches}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Keyword Matches</span>
              <Badge variant="outline">
                <TrendingUp className="h-3 w-3 mr-1" />
                {metadata.codeGeneration.keywordMatches}
              </Badge>
            </div>
            {metadata.codeGeneration.industriesAnalyzed && metadata.codeGeneration.industriesAnalyzed.length > 0 && (
              <div className="space-y-1">
                <span className="text-sm text-muted-foreground">Industries Analyzed</span>
                <div className="flex flex-wrap gap-1">
                  {metadata.codeGeneration.industriesAnalyzed.map((industry, idx) => (
                    <Badge key={idx} variant="outline" className="text-xs">
                      {industry}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

