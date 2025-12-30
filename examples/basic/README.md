# Basic Example

This example demonstrates basic DocuGuard usage with a simple payment module.

## Files

- `docs/api.md` - Documentation with binding annotations
- `code/payment.go` - Go source code

## Run

```bash
docuguard check examples/basic/docs/api.md
```

The check will detect that the shipping threshold in the documentation (500) 
doesn't match the code implementation (1000).
