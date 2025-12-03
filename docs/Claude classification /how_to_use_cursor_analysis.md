# How to Use This Analysis with Cursor

## Step 1: Prepare Cursor

1. Open your classification codebase in Cursor
2. Make sure Cursor has access to your entire project structure
3. Ensure you're using Cursor's Composer mode (Cmd+I or Ctrl+I)

## Step 2: Provide Context to Cursor

Copy and paste this message into Cursor Composer:

---

**IMPORTANT: Please read the attached `cursor_analysis_prompt.md` file completely before beginning your analysis.**

I need you to perform a comprehensive analysis of this codebase's industry classification implementation. The `cursor_analysis_prompt.md` file contains:
- Context about the proposed ideal-state architecture
- Detailed instructions for your analysis
- The exact output format required

Please follow these steps:

1. First, read and acknowledge that you understand the full scope of the analysis from the markdown file
2. Explore the codebase systematically as outlined in the analysis instructions
3. Produce the complete analysis report in the exact markdown format specified
4. Be thorough - this analysis will be used to re-scope the implementation plan

**Focus Areas:**
- Database schema analysis (Supabase tables: codes, keywords, trigram, crosswalks, and others)
- Current classification logic and approach
- Comparison against the proposed 3-layer hybrid system
- Identification of reusable components
- Gap analysis and migration path

**Output:**
Produce a complete markdown document following the template in section "Output Format" of the prompt file.

Begin your analysis now.

---

## Step 3: Let Cursor Analyze

Cursor will now:
- Explore your codebase
- Ask you clarifying questions if needed
- Generate the comprehensive analysis report

This may take 5-10 minutes depending on codebase size.

## Step 4: Save the Analysis

Once Cursor produces the analysis:

1. Copy the entire markdown output
2. Save it as `classification_analysis_report.md`
3. Review it for accuracy and completeness
4. Fill in any gaps Cursor couldn't determine (marked with [Unknown] or similar)

## Step 5: Return to Claude

Send the completed analysis report back in your chat with Claude with a message like:

"Here's the analysis report from Cursor. Please review this and re-scope the proposed implementation to leverage our existing infrastructure and identify the optimal migration path."

---

## Troubleshooting

**If Cursor doesn't find certain files:**
- Manually point Cursor to key directories
- Use `@folder` mentions in Cursor to guide it

**If Cursor's output is incomplete:**
- Ask it to continue: "Please complete section X of the analysis"
- Request more detail: "Can you provide more detail on the database schema?"

**If Cursor is unsure about something:**
- That's fine! The analysis should note unknowns
- You can manually investigate and fill in gaps

---

## Expected Timeline

- Cursor analysis: 5-10 minutes
- Your review and gap-filling: 15-30 minutes  
- Claude's re-scoping: 5-10 minutes
- Total: ~30-50 minutes

This investment will save significant time by ensuring the implementation plan leverages what you already have.
