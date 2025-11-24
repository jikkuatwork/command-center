# Interface Consistency

I wish for something like this:

```
  cc-server
    set-credentials
      --username
      --password

    deploy
      no-flag <PATH>
      --domain <SUB-DOMAIN>

    start
      --port
      --db
      --config

    stop


Not sure whats the function of this:
  --env <environment>   Load environment-specific config (development/production)
```

Note:
- lets have all config including keys/tokens in ~/.config/cc/config.json
- lets remove all other unwanted commands/flags
- also have sane defaults
