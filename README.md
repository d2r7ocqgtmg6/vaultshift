# vaultshift

> CLI tool to migrate secrets between HashiCorp Vault namespaces with dry-run and audit logging support.

---

## Installation

```bash
go install github.com/yourorg/vaultshift@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/vaultshift.git
cd vaultshift && go build -o vaultshift .
```

---

## Usage

```bash
# Migrate secrets from one namespace to another
vaultshift migrate \
  --src-namespace "team-a/prod" \
  --dst-namespace "team-b/prod" \
  --path "secret/data/app"

# Preview changes without writing (dry-run mode)
vaultshift migrate \
  --src-namespace "team-a/prod" \
  --dst-namespace "team-b/prod" \
  --path "secret/data/app" \
  --dry-run

# Enable audit logging to a file
vaultshift migrate \
  --src-namespace "team-a/prod" \
  --dst-namespace "team-b/prod" \
  --path "secret/data/app" \
  --audit-log ./migration-audit.log
```

### Flags

| Flag | Description |
|------|-------------|
| `--src-namespace` | Source Vault namespace |
| `--dst-namespace` | Destination Vault namespace |
| `--path` | Secret path to migrate |
| `--dry-run` | Preview changes without applying them |
| `--audit-log` | Path to write the audit log |

---

## Requirements

- Go 1.21+
- HashiCorp Vault 1.12+
- `VAULT_ADDR` and `VAULT_TOKEN` environment variables set

---

## License

[MIT](LICENSE)