# Judge Service

Standalone HTTP service that evaluates attack outputs and returns severity/confidence.

## Endpoints

- `GET /health`
- `POST /evaluate-batch`

## Run

```bash
cd judge
make run
```

Environment:

- `JUDGE_PORT` (default `8091`)
