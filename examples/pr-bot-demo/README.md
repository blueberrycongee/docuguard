# PR Bot Demo

This example demonstrates DocuGuard's PR bot functionality with intentional documentation-code inconsistencies.

## Structure

```
pr-bot-demo/
├── code/           # Simulated code repository
├── docs/           # Corresponding documentation
└── README.md       # This file
```

## Test Scenarios

### Scenario 1: Function Signature Change
- **File**: `code/user.go` - `CreateUser` function
- **Doc**: `docs/user-api.md`
- **Issue**: Function parameters don't match documentation

### Scenario 2: Return Value Change
- **File**: `code/order.go` - `GetOrder` function
- **Doc**: `docs/order-guide.md`
- **Issue**: Documentation missing error return value

### Scenario 3: Business Logic Change
- **File**: `code/payment.go` - Free shipping threshold
- **Doc**: `docs/payment-flow.md`
- **Issue**: Threshold value differs between code and docs

## How to Test

1. Create a test branch and modify code:
   ```bash
   git checkout -b test-pr-bot
   # Modify one of the code files
   git commit -am "test: modify user function"
   git push origin test-pr-bot
   ```

2. Create a PR on GitHub

3. Run PR bot locally:
   ```bash
   docuguard pr --github --pr <PR_NUMBER> --comment --token <GITHUB_TOKEN>
   ```

4. Check the PR comment to see detected inconsistencies

## Expected Results

The bot should detect and report:
- Parameter mismatches in function signatures
- Missing error handling in documentation
- Inconsistent business logic values
