---
name: fix-issue-v2
description: Systematically debug and fix bugs using a structured workflow
user-invocable: true
---

# Fix Issue Skill

Fix a reported bug using a structured debugging workflow.

## Input
Issue description:
$ARGUMENTS

---

## Workflow

### 1. Understand the Issue
- Restate the problem clearly
- Identify expected vs actual behavior
- List at least 2 possible causes

### 2. Gather Context
- Search the codebase for relevant files
- Identify relevant functions/components
- Ask for missing context if needed

### 3. Diagnose Root Cause
- Identify the most likely cause
- Explain WHY the issue occurs

### 4. Implement Fix
- Make minimal, targeted changes
- Show before -> after code

### 5. Validate Fix
- Provide test steps
- Include at least 1 edge case

### 6. Prevent Future Issues
- Suggest improvements (tests, patterns, safeguards)

---

## Constraints
- Do NOT assume missing details without stating assumptions
- Do NOT rewrite large parts of the code
- Prefer simple, maintainable fixes

---

## Output Format

1. Problem Summary
2. Root Cause
3. Fix (code changes)
4. Test Plan
5. Prevention Tips
